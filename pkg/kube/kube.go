package kube

import (
	"context"
	"errors"
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
func (k *KubeClient) GetNodes(resourceName string, selector labels.Selector) (map[string]corev1.Node, error) {
	nodes := make(map[string]corev1.Node)
	if len(resourceName) > 0 {
		node, err := k.apiClient.CoreV1().Nodes().Get(context.TODO(), resourceName, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}
		nodes[node.Name] = *node

	} else {
		nodeList, err := k.apiClient.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{
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
func (k *KubeClient) GetActivePodByNodename(node corev1.Node) (*corev1.PodList, error) {
	fieldSelector, err := fields.ParseSelector("spec.nodeName=" + node.Name +
		",status.phase!=" + string(corev1.PodSucceeded) +
		",status.phase!=" + string(corev1.PodFailed))

	if err != nil {
		return nil, err
	}
	activePods, err := k.apiClient.CoreV1().Pods(corev1.NamespaceAll).List(context.TODO(), metav1.ListOptions{FieldSelector: fieldSelector.String()})
	if err != nil {
		return nil, err
	}
	return activePods, err
}

//GetActivePodByPodname
func (k *KubeClient) GetPodByPodname(podName string, namespace string) (*corev1.Pod, error) {
	pod, err := k.apiClient.CoreV1().Pods(namespace).Get(context.TODO(), podName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return pod, err
}

//NodeResources
func (k *KubeClient) GetNodeResources(resourceName string, resourceType string, sortBy string, selector labels.Selector) ([][]string, error) {
	//resources := make(map[string]map[string]interface{})
	var resources [][]string
	var nodenames []string

	metrics, err := k.GetNodeMetricsFromMetricsAPI(resourceName, selector)
	if err != nil {
		return nil, err
	}
	//判断是否排序
	if len(sortBy) > 0 {
		sort.Sort(metricsutil.NewNodeMetricsSorter(metrics.Items, sortBy))
	}
	for _, i := range metrics.Items {
		nodenames = append(nodenames, i.Name)
	}

	nodes, err := k.GetNodes(resourceName, selector)
	if err != nil {
		return nil, err
	}

	for _, nodename := range nodenames {
		//resource := make(map[string]interface{})
		var resource []string
		activePodsList, err := k.GetActivePodByNodename(nodes[nodename])
		if err != nil {
			return nil, err
		}
		NodeMetricsList, err := k.GetNodeMetricsFromMetricsAPI(nodename, selector)
		if err != nil {
			return nil, err
		}

		noderesource, err := getNodeAllocatedResources(nodes[nodename], activePodsList, NodeMetricsList, resourceType)
		if err != nil {
			log.Printf("Couldn't get allocated resources of %s node: %s\n", nodename, err)
		}

		//if cpu,ok := noderesource.(CpuResource); ok{
		switch v := noderesource.(type) {
		case NodeAllocatedResources:
			resource = append(resource, nodename,
				v.CPUUsages.String(), v.CPURequests.String(), ExceedsCompare(float64ToString(v.CPURequestsFraction)),
				v.CPULimits.String(), float64ToString(v.CPULimitsFraction), v.CPUCapacity.String(),
				v.MemoryUsages.String(), v.MemoryRequests.String(), ExceedsCompare(float64ToString(v.MemoryRequestsFraction)),
				v.MemoryLimits.String(), float64ToString(v.MemoryLimitsFraction), v.MemoryCapacity.String(),
				podFormat(intToString(v.AllocatedPods), int64ToString(v.PodCapacity)), ExceedsCompare(float64ToString(v.PodFraction)))
		case CPUResources:
			resource = append(resource, nodename,
				v.CPUUsages.String(), v.CPURequests.String(), ExceedsCompare(float64ToString(v.CPURequestsFraction)),
				v.CPULimits.String(), float64ToString(v.CPULimitsFraction), v.CPUCapacity.String())
		case MemoryResources:
			resource = append(resource, nodename,
				v.MemoryUsages.String(), v.MemoryRequests.String(), ExceedsCompare(float64ToString(v.MemoryRequestsFraction)),
				v.MemoryLimits.String(), float64ToString(v.MemoryLimitsFraction), v.MemoryCapacity.String())
		case PodResources:
			resource = append(resource, nodename,
				podFormat(intToString(v.AllocatedPods), int64ToString(v.PodCapacity)), ExceedsCompare(float64ToString(v.PodFraction)))
		default:
			return nil, errors.New("unexpected type")
		}
		resources = append(resources, resource)
	}
	return resources, err

}

func (k *KubeClient) GetPodResources(podmetrics []metricsapi.PodMetrics, namespace string, resourceName string, allNamespaces bool, resourceType string, sortBy string, labelSelector labels.Selector, fieldSelector fields.Selector) ([][]string, error) {
	var resources [][]string

	//判断是否排序
	if len(sortBy) > 0 {
		sort.Sort(metricsutil.NewPodMetricsSorter(podmetrics, allNamespaces, sortBy))
	}
	for _, podmetric := range podmetrics {
		var resource []string
		pod, err := k.GetPodByPodname(podmetric.Name, podmetric.Namespace)
		if err != nil {
			return nil, err
		}
		podresource, err := getPodAllocatedResources(pod, &podmetric, resourceType)
		if err != nil {
			return nil, err
		}

		switch {
		case resourceType == "cpu":
			resource = append(resource, podmetric.Namespace, podmetric.Name,
				podresource.CPUUsages.String(), ExceedsCompare(float64ToString(podresource.CPUUsagesFraction)),
				podresource.CPURequests.String(), podresource.CPULimits.String())
		case resourceType == "memory":
			resource = append(resource, podmetric.Namespace, podmetric.Name,
				podresource.MemoryUsages.String(), ExceedsCompare(float64ToString(podresource.MemoryUsagesFraction)),
				podresource.MemoryRequests.String(), podresource.MemoryLimits.String())
		default:
			resource = append(resource, podmetric.Namespace, podmetric.Name,
				podresource.CPUUsages.String(), ExceedsCompare(float64ToString(podresource.CPUUsagesFraction)),
				podresource.CPURequests.String(), podresource.CPULimits.String(),
				podresource.MemoryUsages.String(), ExceedsCompare(float64ToString(podresource.MemoryUsagesFraction)),
				podresource.MemoryRequests.String(), podresource.MemoryLimits.String())
		}
		resources = append(resources, resource)

	}

	return resources, nil
}

// PodMetricses returns all pods' usage metrics
func (k *KubeClient) PodMetricses() (*metricsV1beta1api.PodMetricsList, error) {
	podMetricses, err := k.metricsClient.MetricsV1beta1().PodMetricses(metav1.NamespaceAll).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	return podMetricses, nil
}

// GetNodeMetricsFromMetricsAPI
func (k *KubeClient) GetNodeMetricsFromMetricsAPI(resourceName string, selector labels.Selector) (*metricsapi.NodeMetricsList, error) {
	var err error
	versionedMetrics := &metricsV1beta1api.NodeMetricsList{}
	mc := k.metricsClient.MetricsV1beta1()
	nm := mc.NodeMetricses()
	if resourceName != "" {
		m, err := nm.Get(context.TODO(), resourceName, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}
		versionedMetrics.Items = []metricsV1beta1api.NodeMetrics{*m}
	} else {
		versionedMetrics, err = nm.List(context.TODO(), metav1.ListOptions{LabelSelector: selector.String()})
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
func (k *KubeClient) GetPodMetricsFromMetricsAPI(namespace, resourceName string, allNamespaces bool, labelSelector labels.Selector, fieldSelector fields.Selector) (*metricsapi.PodMetricsList, error) {
	var err error
	ns := metav1.NamespaceAll
	if !allNamespaces {
		ns = namespace
	}
	versionedMetrics := &metricsV1beta1api.PodMetricsList{}
	if resourceName != "" {
		m, err := k.metricsClient.MetricsV1beta1().PodMetricses(ns).Get(context.TODO(), resourceName, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}
		versionedMetrics.Items = []metricsV1beta1api.PodMetrics{*m}
	} else {
		versionedMetrics, err = k.metricsClient.MetricsV1beta1().PodMetricses(ns).List(context.TODO(), metav1.ListOptions{LabelSelector: labelSelector.String(), FieldSelector: fieldSelector.String()})
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

// func (k *KubeClient) ClusterCapacity() (capacity corev1.ResourceList, err error) {
// 	nodes, err := k.apiClient.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
// 	if err != nil {
// 		return nil, err
// 	}

// 	capacity = corev1.ResourceList{}

// 	for _, node := range nodes.Items {

// 		allocatable := NodeCapacity(&node)

// 		for name, quantity := range allocatable {
// 			if value, ok := capacity[name]; ok {
// 				value.Add(quantity)
// 				capacity[name] = value
// 			} else {
// 				capacity[name] = quantity
// 			}
// 		}

// 	}

// 	return capacity, nil
// }
