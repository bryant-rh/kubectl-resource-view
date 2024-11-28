package kube

import (
	"context"
	"fmt"
	"log"
	"sort"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/fields"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"k8s.io/kubectl/pkg/metricsutil"
	metricsapi "k8s.io/metrics/pkg/apis/metrics"
	metricsV1beta1api "k8s.io/metrics/pkg/apis/metrics/v1beta1"

	"k8s.io/apimachinery/pkg/labels"

	metrics "k8s.io/metrics/pkg/client/clientset/versioned"
)

// KubeClient provides methods to get all required metrics from Kubernetes
type KubeClient struct {
	apiClient     *kubernetes.Clientset
	metricsClient *metrics.Clientset
}

// NewClient creates a new client to get data from kubernetes masters
func NewClient(config *rest.Config) (*KubeClient, error) {
	// Add rate limiting configuration to avoid client-side throttling
	config.QPS = 50    // Increase QPS (queries per second)
	config.Burst = 100 // Increase burst rate

	// We got two clients, one for the common API and one explicitly for metrics
	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("error creating kubernetes main client: '%v'", err)
	}

	metricsClient, err := metrics.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("error creating kubernetes metrics client: '%v'", err)
	}

	return &KubeClient{
		apiClient:     client,
		metricsClient: metricsClient,
	}, nil
}

//GetNodes
func (k *KubeClient) GetNodes(ctx context.Context, resourceName string, selector labels.Selector) (map[string]corev1.Node, error) {
	nodes := make(map[string]corev1.Node)
	if len(resourceName) > 0 {
		node, err := k.apiClient.CoreV1().Nodes().Get(ctx, resourceName, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}
		nodes[node.Name] = *node

	} else {
		nodeList, err := k.apiClient.CoreV1().Nodes().List(ctx, metav1.ListOptions{
			LabelSelector: selector.String(),
		})
		if err != nil {
			return nil, err
		}
		for _, i := range nodeList.Items {
			nodes[i.Name] = i

		}
		//	nodes = append(nodes, noderes)
		//nodes = append(nodes, nodeList.Items...)
	}
	return nodes, nil
}

//GetActivePodByNodename
func (k *KubeClient) GetActivePodByNodename(ctx context.Context, node corev1.Node) (*corev1.PodList, error) {
	fieldSelector, err := fields.ParseSelector("spec.nodeName=" + node.Name +
		",status.phase!=" + string(corev1.PodSucceeded) +
		",status.phase!=" + string(corev1.PodFailed))

	if err != nil {
		return nil, err
	}
	activePods, err := k.apiClient.CoreV1().Pods(corev1.NamespaceAll).List(ctx, metav1.ListOptions{FieldSelector: fieldSelector.String()})
	if err != nil {
		return nil, err
	}
	return activePods, err
}

//GetActivePodByPodname
func (k *KubeClient) GetPodByPodname(ctx context.Context, podName string, namespace string) (*corev1.Pod, error) {
	pod, err := k.apiClient.CoreV1().Pods(namespace).Get(ctx, podName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return pod, err
}

//NodeResources
func (k *KubeClient) GetNodeResources(ctx context.Context, resourceName string, resourceType []string, sortBy string, selector labels.Selector) ([][]string, error) {
	metrics, err := k.GetNodeMetricsFromMetricsAPI(ctx, resourceName, selector)
	if err != nil {
		return nil, err
	}

	if len(sortBy) > 0 {
		sort.Sort(metricsutil.NewNodeMetricsSorter(metrics.Items, sortBy))
	}

	var nodenames []string
	for _, i := range metrics.Items {
		nodenames = append(nodenames, i.Name)
	}

	nodes, err := k.GetNodes(ctx, resourceName, selector)
	if err != nil {
		return nil, err
	}

	// Create channels for results and errors
	type nodeResult struct {
		resource []string
		err      error
	}
	resultChan := make(chan nodeResult, len(nodenames))

	// Process nodes concurrently
	for _, nodename := range nodenames {
		go func(nodename string) {
			// Check context before starting work
			select {
			case <-ctx.Done():
				resultChan <- nodeResult{nil, ctx.Err()}
				return
			default:
			}

			var resource []string
			
			// Get active pods with context
			activePodsList, err := k.GetActivePodByNodename(ctx, nodes[nodename])
			if err != nil {
				resultChan <- nodeResult{nil, err}
				return
			}

			// Get node metrics with context
			NodeMetricsList, err := k.GetNodeMetricsFromMetricsAPI(ctx, resourceName, selector)
			if err != nil {
				resultChan <- nodeResult{nil, err}
				return
			}

			resource = append(resource, nodename)
			for _, t := range resourceType {
				// Check context periodically
				select {
				case <-ctx.Done():
					resultChan <- nodeResult{nil, ctx.Err()}
					return
				default:
				}

				noderesource, err := getNodeAllocatedResources(nodes[nodename], activePodsList, NodeMetricsList, t)
				if err != nil {
					log.Printf("Couldn't get allocated resources of %s node: %s\n", nodename, err)
					continue
				}

				switch {
				case t == "cpu":
					resource = append(resource,
						noderesource.CPUUsages.String(),
						newFormat(noderesource.CPURequests.String(), noderesource.CPUCapacity.String()),
						ExceedsCompare(float64ToString(noderesource.CPURequestsFraction)),
						newFormat(noderesource.CPULimits.String(), noderesource.CPUCapacity.String()),
						float64ToString(noderesource.CPULimitsFraction),
					)
				case t == "memory":
					resource = append(resource,
						noderesource.MemoryUsages.String(),
						newFormat(noderesource.MemoryRequests.String(), noderesource.MemoryCapacity.String()), ExceedsCompare(float64ToString(noderesource.MemoryRequestsFraction)),
						newFormat(noderesource.MemoryLimits.String(), noderesource.MemoryCapacity.String()), float64ToString(noderesource.MemoryLimitsFraction),
					)
				case t == "gpu":
					// resource = append(resource,
					// 	newFormat(int64ToString(noderesource.NvidiaGpuCountsRequests), int64ToString(noderesource.NvidiaGpuCountsCapacity)), ExceedsCompare(float64ToString(noderesource.NvidiaGpuCountsRequestsFraction)),
					// 	newFormat(int64ToString(noderesource.NvidiaGpuCountsLimits), int64ToString(noderesource.NvidiaGpuCountsCapacity)), float64ToString(noderesource.NvidiaGpuCountsLimitsFraction),
					// 	newFormat(int64ToString(noderesource.AliyunGpuMemRequests), int64ToString(noderesource.AliyunGpuMemCapacity)), ExceedsCompare(float64ToString(noderesource.AliyunGpuMemRequestsFraction)),
					// 	newFormat(int64ToString(noderesource.AliyunGpuMemLimits), int64ToString(noderesource.AliyunGpuMemCapacity)), float64ToString(noderesource.AliyunGpuMemLimitsFraction),
					// )
					resource = append(resource,
						newFormat(int64ToString(noderesource.NvidiaGpuCountsRequests), int64ToString(noderesource.NvidiaGpuCountsCapacity)), ExceedsCompare(float64ToString(noderesource.NvidiaGpuCountsRequestsFraction)),
						newFormat(int64ToString(noderesource.NvidiaGpuCountsLimits), int64ToString(noderesource.NvidiaGpuCountsCapacity)), float64ToString(noderesource.NvidiaGpuCountsLimitsFraction),
					)
				case t == "pod":
					resource = append(resource,
						newFormat(intToString(noderesource.AllocatedPods), int64ToString(noderesource.PodCapacity)), ExceedsCompare(float64ToString(noderesource.PodFraction)),
					)
				default:
					// resource = append(resource,
					// 	noderesource.CPUUsages.String(),
					// 	newFormat(noderesource.CPURequests.String(), noderesource.CPUCapacity.String()), ExceedsCompare(float64ToString(noderesource.CPURequestsFraction)),
					// 	newFormat(noderesource.CPULimits.String(), noderesource.CPUCapacity.String()), float64ToString(noderesource.CPULimitsFraction),
					// 	noderesource.MemoryUsages.String(),
					// 	newFormat(noderesource.MemoryRequests.String(), noderesource.MemoryCapacity.String()), ExceedsCompare(float64ToString(noderesource.MemoryRequestsFraction)),
					// 	newFormat(noderesource.MemoryLimits.String(), noderesource.MemoryCapacity.String()), float64ToString(noderesource.MemoryLimitsFraction),
					// 	newFormat(int64ToString(noderesource.NvidiaGpuCountsRequests), int64ToString(noderesource.NvidiaGpuCountsCapacity)), ExceedsCompare(float64ToString(noderesource.NvidiaGpuCountsRequestsFraction)),
					// 	newFormat(int64ToString(noderesource.NvidiaGpuCountsLimits), int64ToString(noderesource.NvidiaGpuCountsCapacity)), float64ToString(noderesource.NvidiaGpuCountsLimitsFraction),
					// 	newFormat(int64ToString(noderesource.AliyunGpuMemRequests), int64ToString(noderesource.AliyunGpuMemCapacity)), ExceedsCompare(float64ToString(noderesource.AliyunGpuMemRequestsFraction)),
					// 	newFormat(int64ToString(noderesource.AliyunGpuMemLimits), int64ToString(noderesource.AliyunGpuMemCapacity)), float64ToString(noderesource.AliyunGpuMemLimitsFraction),
					// 	newFormat(intToString(noderesource.AllocatedPods), int64ToString(noderesource.PodCapacity)), ExceedsCompare(float64ToString(noderesource.PodFraction)),
					// )
					resource = append(resource,
						noderesource.CPUUsages.String(),
						newFormat(noderesource.CPURequests.String(), noderesource.CPUCapacity.String()), ExceedsCompare(float64ToString(noderesource.CPURequestsFraction)),
						newFormat(noderesource.CPULimits.String(), noderesource.CPUCapacity.String()), float64ToString(noderesource.CPULimitsFraction),
						noderesource.MemoryUsages.String(),
						newFormat(noderesource.MemoryRequests.String(), noderesource.MemoryCapacity.String()), ExceedsCompare(float64ToString(noderesource.MemoryRequestsFraction)),
						newFormat(noderesource.MemoryLimits.String(), noderesource.MemoryCapacity.String()), float64ToString(noderesource.MemoryLimitsFraction),
						newFormat(int64ToString(noderesource.NvidiaGpuCountsRequests), int64ToString(noderesource.NvidiaGpuCountsCapacity)), ExceedsCompare(float64ToString(noderesource.NvidiaGpuCountsRequestsFraction)),
						newFormat(int64ToString(noderesource.NvidiaGpuCountsLimits), int64ToString(noderesource.NvidiaGpuCountsCapacity)), float64ToString(noderesource.NvidiaGpuCountsLimitsFraction),
						newFormat(intToString(noderesource.AllocatedPods), int64ToString(noderesource.PodCapacity)), ExceedsCompare(float64ToString(noderesource.PodFraction)),
					)
				}
			}
			resultChan <- nodeResult{resource, nil}
		}(nodename)
	}

	// Collect results with context awareness
	var resources [][]string
	var firstError error
	for i := 0; i < len(nodenames); i++ {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case result := <-resultChan:
			if result.err != nil {
				if firstError == nil {
					firstError = result.err
				}
				continue
			}
			if result.resource != nil {
				resources = append(resources, result.resource)
			}
		}
	}

	if firstError != nil {
		return nil, firstError
	}
	return resources, nil
}

func (k *KubeClient) GetPodResources(ctx context.Context, podmetrics []metricsapi.PodMetrics, namespace string, resourceName string, allNamespaces bool, resourceType []string, sortBy string, labelSelector labels.Selector, fieldSelector fields.Selector) ([][]string, error) {
	// 判断是否排序
	if len(sortBy) > 0 {
		sort.Sort(metricsutil.NewPodMetricsSorter(podmetrics, allNamespaces, sortBy))
	}

	// Create channels for results and errors
	type podResult struct {
		resource []string
		err      error
	}
	resultChan := make(chan podResult, len(podmetrics))

	// Process pods concurrently
	for _, podmetric := range podmetrics {
		go func(podmetric metricsapi.PodMetrics) {
			// Check context before starting work
			select {
			case <-ctx.Done():
				resultChan <- podResult{nil, ctx.Err()}
				return
			default:
			}

			var resource []string
			pod, err := k.GetPodByPodname(ctx, podmetric.Name, podmetric.Namespace)
			if err != nil {
				resultChan <- podResult{nil, err}
				return
			}

			resource = append(resource, podmetric.Namespace, podmetric.Name)
			for _, t := range resourceType {
				// Check context periodically during processing
				select {
				case <-ctx.Done():
					resultChan <- podResult{nil, ctx.Err()}
					return
				default:
				}

				podresource, err := getPodAllocatedResources(pod, &podmetric, t)
				if err != nil {
					resultChan <- podResult{nil, err}
					return
				}
				switch {
				case t == "cpu":
					resource = append(resource,
						podresource.CPUUsages.String(), ExceedsCompare(float64ToString(podresource.CPUUsagesFraction)),
						podresource.CPURequests.String(), podresource.CPULimits.String(),
					)
				case t == "memory":
					resource = append(resource,
						podresource.MemoryUsages.String(), ExceedsCompare(float64ToString(podresource.MemoryUsagesFraction)),
						podresource.MemoryRequests.String(), podresource.MemoryLimits.String(),
					)
				case t == "gpu":
					resource = append(resource,
						int64ToString(podresource.NvidiaGpuCountsRequests), int64ToString(podresource.NvidiaGpuCountsLimits),
					)
				default:
					resource = append(resource,
						podresource.CPUUsages.String(), ExceedsCompare(float64ToString(podresource.CPUUsagesFraction)),
						podresource.CPURequests.String(), podresource.CPULimits.String(),
						podresource.MemoryUsages.String(), ExceedsCompare(float64ToString(podresource.MemoryUsagesFraction)),
						podresource.MemoryRequests.String(), podresource.MemoryLimits.String(),
						int64ToString(podresource.NvidiaGpuCountsRequests), int64ToString(podresource.NvidiaGpuCountsLimits),
					)
				}
			}
			resultChan <- podResult{resource, nil}
		}(podmetric)
	}

	// Collect results with context awareness
	var resources [][]string
	var firstError error
	for i := 0; i < len(podmetrics); i++ {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case result := <-resultChan:
			if result.err != nil {
				if firstError == nil {
					firstError = result.err
				}
				continue
			}
			if result.resource != nil {
				resources = append(resources, result.resource)
			}
		}
	}

	if firstError != nil {
		return nil, firstError
	}
	return resources, nil
}

// PodMetricses returns all pods' usage metrics
func (k *KubeClient) PodMetricses(ctx context.Context) (*metricsV1beta1api.PodMetricsList, error) {
	podMetricses, err := k.metricsClient.MetricsV1beta1().PodMetricses(metav1.NamespaceAll).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	return podMetricses, nil
}

// GetNodeMetricsFromMetricsAPI with context
func (k *KubeClient) GetNodeMetricsFromMetricsAPI(ctx context.Context, resourceName string, selector labels.Selector) (*metricsapi.NodeMetricsList, error) {
	var err error
	versionedMetrics := &metricsV1beta1api.NodeMetricsList{}
	mc := k.metricsClient.MetricsV1beta1()
	nm := mc.NodeMetricses()
	
	if resourceName != "" {
		m, err := nm.Get(ctx, resourceName, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}
		versionedMetrics.Items = []metricsV1beta1api.NodeMetrics{*m}
	} else {
		versionedMetrics, err = nm.List(ctx, metav1.ListOptions{
			LabelSelector: selector.String(),
		})
		if err != nil {
			return nil, err
		}
	}
	
	metrics := &metricsapi.NodeMetricsList{}
	err = metricsV1beta1api.Convert_v1beta1_NodeMetricsList_To_metrics_NodeMetricsList(versionedMetrics, metrics, nil)
	if err != nil {
		return nil, err
	}
	return metrics, nil
}

// GetPodMetricsFromMetricsAPI
func (k *KubeClient) GetPodMetricsFromMetricsAPI(ctx context.Context, namespace, resourceName string, allNamespaces bool, labelSelector labels.Selector, fieldSelector fields.Selector) (*metricsapi.PodMetricsList, error) {
	var err error
	ns := metav1.NamespaceAll
	if !allNamespaces {
		ns = namespace
	}
	versionedMetrics := &metricsV1beta1api.PodMetricsList{}
	if resourceName != "" {
		m, err := k.metricsClient.MetricsV1beta1().PodMetricses(ns).Get(ctx, resourceName, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}
		versionedMetrics.Items = []metricsV1beta1api.PodMetrics{*m}
	} else {
		versionedMetrics, err = k.metricsClient.MetricsV1beta1().PodMetricses(ns).List(ctx, metav1.ListOptions{
			LabelSelector: labelSelector.String(), 
			FieldSelector: fieldSelector.String(),
		})
		if err != nil {
			return nil, err
		}
	}
	metrics := &metricsapi.PodMetricsList{}
	err = metricsV1beta1api.Convert_v1beta1_PodMetricsList_To_metrics_PodMetricsList(versionedMetrics, metrics, nil)
	if err != nil {
		return nil, err
	}
	return metrics, nil
}
