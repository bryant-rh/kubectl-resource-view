package kube

import (
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/kubectl/pkg/metricsutil"
	metricsapi "k8s.io/metrics/pkg/apis/metrics"
)

// CPUResources describes node allocated resources.
type CPUResources struct {
	// CPUUsages is number of allocated milicores.
	CPUUsages *CpuResource

	// CPURequests is number of allocated milicores.
	CPURequests *CpuResource

	// CPURequestsFraction is a fraction of CPU, that is allocated.
	CPURequestsFraction float64 `json:"cpuRequestsFraction"`

	// CPULimits is defined CPU limit.
	CPULimits *CpuResource

	// CPULimitsFraction is a fraction of defined CPU limit, can be over 100%, i.e.
	// overcommitted.
	CPULimitsFraction float64 `json:"cpuLimitsFraction"`

	// CPUCapacity is specified node CPU capacity in milicores.
	CPUCapacity *CpuResource
}

// MemoryResources describes node allocated resources.
type MemoryResources struct {
	// MemoryUsages is a fraction of memory, that is allocated.
	MemoryUsages *MemoryResource

	// MemoryRequests is a fraction of memory, that is allocated.
	MemoryRequests *MemoryResource

	// MemoryRequestsFraction is a fraction of memory, that is allocated.
	MemoryRequestsFraction float64 `json:"memoryRequestsFraction"`

	// MemoryLimits is defined memory limit.
	MemoryLimits *MemoryResource

	// MemoryLimitsFraction is a fraction of defined memory limit, can be over 100%, i.e.
	// overcommitted.
	MemoryLimitsFraction float64 `json:"memoryLimitsFraction"`

	// MemoryCapacity is specified node memory capacity in bytes.
	MemoryCapacity *MemoryResource
}

// PodResources describes node allocated resources.
type PodResources struct {
	// AllocatedPods in number of currently allocated pods on the node.
	AllocatedPods int `json:"allocatedPods"`

	// PodCapacity is maximum number of pods, that can be allocated on the node.
	PodCapacity int64 `json:"podCapacity"`

	// PodFraction is a fraction of pods, that can be allocated on given node.
	PodFraction float64 `json:"podFraction"`
}

// GPUResources describes node allocated resources.
type GPUResources struct {
	// NvidiaGpuCountsRequests is a fraction of NvidiaGpuCountsRequests, that is allocated.
	NvidiaGpuCountsRequests int64

	// NvidiaGpuCountsRequestsFraction is a fraction of NvidiaGpuCountsRequests, that is allocated.
	NvidiaGpuCountsRequestsFraction float64 `json:"NvidiaGpuCountsRequestsFraction"`

	// NvidiaGpuCountsLimits is defined NvidiaGpuCounts limit.
	NvidiaGpuCountsLimits int64

	// NvidiaGpuCountsLimitsFraction is a fraction of defined NvidiaGpuCounts limit, can be over 100%, i.e.
	// overcommitted.
	NvidiaGpuCountsLimitsFraction float64 `json:"NvidiaGpuCountsLimitsFraction"`

	// NvidiaGpuCountsCapacity is maximum number of pods, that can be allocated on the node.
	NvidiaGpuCountsCapacity int64 `json:"nvidiaGpuCountsCapacity"`

	// AliyunGpuMemRequests is a fraction of AliyunGpuMemRequests, that is allocated.
	AliyunGpuMemRequests int64

	// AliyunGpuMemRequestsFraction is a fraction of AliyunGpuMemRequests, that is allocated.
	AliyunGpuMemRequestsFraction float64 `json:"aliyunGpuMemsRequestsFraction"`

	// AliyunGpuMemLimits is defined AliyunGpuMem limit.
	AliyunGpuMemLimits int64

	// NvidiaGpuCountsLimitsFraction is a fraction of defined NvidiaGpuCounts limit, can be over 100%, i.e.
	// overcommitted.
	AliyunGpuMemLimitsFraction float64 `json:"aliyunGpuMemLimitsFraction"`

	// AliyunGpuMemCapacity is maximum number of pods, that can be allocated on the node.
	AliyunGpuMemCapacity int64 `json:"aliyunGpuMemCapacity"`
}

// NodeAllocatedResources describes node allocated resources.
type NodeAllocatedResources struct {
	CPUResources
	MemoryResources
	GPUResources
	PodResources
}

// PodAllocatedResources describes node allocated resources.
type PodAllocatedResources struct {
	// CPUUsages is number of allocated milicores.
	CPUUsages *CpuResource

	// CPURequestsFraction is a fraction of CPU, that is allocated.
	CPUUsagesFraction float64 `json:"cpuUsagesFraction"`

	// CPURequests is number of allocated milicores.
	CPURequests *CpuResource

	// CPULimits is defined CPU limit.
	CPULimits *CpuResource

	// MemoryUsages is a fraction of memory, that is allocated.
	MemoryUsages *MemoryResource

	// MemoryRequestsFraction is a fraction of memory, that is allocated.
	MemoryUsagesFraction float64 `json:"memoryUsagesFraction"`

	// MemoryRequests is a fraction of memory, that is allocated.
	MemoryRequests *MemoryResource

	// MemoryLimits is defined memory limit.
	MemoryLimits *MemoryResource

	// NvidiaGpuCountsRequests is a fraction of NvidiaGpuCounts, that is allocated.
	NvidiaGpuCountsRequests int64

	// NvidiaGpuCountsLimits is defined NvidiaGpuCounts limit.
	NvidiaGpuCountsLimits int64

	// AliyunGpuMemRequests is a fraction of AliyunGpuMem, that is allocated.
	AliyunGpuMemRequests int64

	// AliyunGpuMemLimits is defined AliyunGpuMem limit.
	AliyunGpuMemLimits int64
}

//NodeCapacity
func NodeCapacity(node *v1.Node) v1.ResourceList {
	allocatable := node.Status.Capacity
	if len(node.Status.Allocatable) > 0 {
		allocatable = node.Status.Allocatable
	}
	return allocatable
}

// getNodeMetricsByNodeName returns a map of node metrics where the keys are the particular node names
func getNodeMetricsByNodeName(nodeMetricsList *metricsapi.NodeMetricsList) map[string]metricsapi.NodeMetrics {
	nodeMetricsByName := make(map[string]metricsapi.NodeMetrics)
	for _, metrics := range nodeMetricsList.Items {
		nodeMetricsByName[metrics.Name] = metrics
	}

	return nodeMetricsByName
}

// // getNodeMetricsByPodName returns a map of node metrics where the keys are the particular pod names
// func getPodMetricsByPodName(podMetricsList []metricsapi.PodMetrics) map[string]metricsapi.PodMetrics {
// 	podMetricsByName := make(map[string]metricsapi.PodMetrics)
// 	for _, metrics := range podMetricsList {
// 		podMetricsByName[metrics.Name] = metrics
// 	}

// 	return podMetricsByName
// }

func getPodMetrics(m *metricsapi.PodMetrics) v1.ResourceList {
	podMetrics := make(v1.ResourceList)
	for _, res := range metricsutil.MeasuredResources {
		podMetrics[res], _ = resource.ParseQuantity("0")
	}

	for _, c := range m.Containers {
		for _, res := range metricsutil.MeasuredResources {
			quantity := podMetrics[res]
			quantity.Add(c.Usage[res])
			podMetrics[res] = quantity
		}
	}
	return podMetrics
}

//getNodeAllocatedResources https://github.com/kubernetes/dashboard/blob/d386ff60597b6eab0222f2c3c4aecf8e49b3014e/src/app/backend/resource/node/detail.go\#L171
func getNodeAllocatedResources(node v1.Node, podList *v1.PodList, nodeMetricsList *metricsapi.NodeMetricsList, resourceType string) (NodeAllocatedResources, error) {
	reqs, limits := map[v1.ResourceName]resource.Quantity{}, map[v1.ResourceName]resource.Quantity{}

	for _, pod := range podList.Items {
		podReqs, podLimits, err := PodRequestsAndLimits(&pod)
		if err != nil {
			return NodeAllocatedResources{}, err
		}
		for podReqName, podReqValue := range podReqs {
			if value, ok := reqs[podReqName]; !ok {
				reqs[podReqName] = podReqValue.DeepCopy()
			} else {
				value.Add(podReqValue)
				reqs[podReqName] = value
			}
		}
		for podLimitName, podLimitValue := range podLimits {
			if value, ok := limits[podLimitName]; !ok {
				limits[podLimitName] = podLimitValue.DeepCopy()
			} else {
				value.Add(podLimitValue)
				limits[podLimitName] = value
			}
		}
	}

	nodeMetricsByNodeName := getNodeMetricsByNodeName(nodeMetricsList)
	usageMetrics := nodeMetricsByNodeName[node.Name]

	capacity := NodeCapacity(&node)
	var nodeAllocatedResources = NodeAllocatedResources{}
	switch {
	case resourceType == "cpu":
		_cpuRequests, _cpuLimits := reqs[v1.ResourceCPU], limits[v1.ResourceCPU]
		_cpuUsages := usageMetrics.Usage.Cpu().MilliValue()

		cpuUsages := NewCpuResource(_cpuUsages)
		cpuRequests := NewCpuResource(_cpuRequests.MilliValue())
		cpuLimits := NewCpuResource(_cpuLimits.MilliValue())

		nodeAllocatedResources = NodeAllocatedResources{
			CPUResources{
				CPUUsages:           cpuUsages,
				CPURequests:         cpuRequests,
				CPURequestsFraction: cpuRequests.calcPercentage(capacity.Cpu()),
				CPULimits:           cpuLimits,
				CPULimitsFraction:   cpuLimits.calcPercentage(capacity.Cpu()),
				CPUCapacity:         NewCpuResource(capacity.Cpu().MilliValue()),
			},
			MemoryResources{},
			GPUResources{},
			PodResources{},
		}

	case resourceType == "memory":
		_memoryRequests, _memoryLimits := reqs[v1.ResourceMemory], limits[v1.ResourceMemory]
		_memoryUsages := usageMetrics.Usage.Memory().Value()

		memoryUsages := NewMemoryResource(_memoryUsages)
		memoryRequests := NewMemoryResource(_memoryRequests.Value())
		memoryLimits := NewMemoryResource(_memoryLimits.Value())
		nodeAllocatedResources = NodeAllocatedResources{
			CPUResources{},
			MemoryResources{
				MemoryUsages:           memoryUsages,
				MemoryRequests:         memoryRequests,
				MemoryRequestsFraction: memoryRequests.calcPercentage(capacity.Memory()),
				MemoryLimits:           memoryLimits,
				MemoryLimitsFraction:   memoryLimits.calcPercentage(capacity.Memory()),
				MemoryCapacity:         NewMemoryResource(capacity.Memory().Value()),
			},
			GPUResources{},
			PodResources{},
		}
	case resourceType == "pod":
		podCapacity := capacity.Pods().Value()
		podFraction := calcPercentage(int64(len(podList.Items)), podCapacity)
		nodeAllocatedResources = NodeAllocatedResources{
			CPUResources{},
			MemoryResources{},
			GPUResources{},
			PodResources{
				AllocatedPods: len(podList.Items),
				PodCapacity:   podCapacity,
				PodFraction:   podFraction,
			},
		}
	case resourceType == "gpu":
		_nvidiaGpuCountsRequests, _nvidiaGpuCountsLimits := reqs[ResourceNvidiaGpuCounts], limits[ResourceNvidiaGpuCounts]
		nvidiaGpuCountsRequests := _nvidiaGpuCountsRequests.Value()
		nvidiaGpuCountsLimits := _nvidiaGpuCountsLimits.Value()
		nvidiaGpuCountsCapacity := NewGpuResource(ResourceNvidiaGpuCounts, &capacity).Value()

		// _aliyunGpuMemRequests, _aliyunGpuMemLimits := reqs[ResourceAliyunGpuMem], limits[ResourceAliyunGpuMem]
		// aliyunGpuMemRequests := _aliyunGpuMemRequests.Value()
		// aliyunGpuMemLimits := _aliyunGpuMemLimits.Value()
		// aliyunGpuMemCapacity := NewGpuResource(ResourceAliyunGpuMem, &capacity).Value()

		nodeAllocatedResources = NodeAllocatedResources{
			CPUResources{},
			MemoryResources{},
			GPUResources{
				NvidiaGpuCountsRequests:         nvidiaGpuCountsRequests,
				NvidiaGpuCountsRequestsFraction: calcPercentage(nvidiaGpuCountsRequests, nvidiaGpuCountsCapacity),
				NvidiaGpuCountsLimits:           nvidiaGpuCountsLimits,
				NvidiaGpuCountsLimitsFraction:   calcPercentage(nvidiaGpuCountsLimits, nvidiaGpuCountsCapacity),
				NvidiaGpuCountsCapacity:         nvidiaGpuCountsCapacity,
				// AliyunGpuMemRequests:            aliyunGpuMemRequests,
				// AliyunGpuMemRequestsFraction:    calcPercentage(aliyunGpuMemRequests, aliyunGpuMemCapacity),
				// AliyunGpuMemLimits:              aliyunGpuMemLimits,
				// AliyunGpuMemLimitsFraction:      calcPercentage(aliyunGpuMemLimits, aliyunGpuMemCapacity),
				// AliyunGpuMemCapacity:            aliyunGpuMemCapacity,
			},
			PodResources{},
		}
	default:
		_cpuRequests, _cpuLimits, _memoryRequests, _memoryLimits := reqs[v1.ResourceCPU], limits[v1.ResourceCPU],
			reqs[v1.ResourceMemory], limits[v1.ResourceMemory]
		_cpuUsages, _memoryUsages := usageMetrics.Usage.Cpu().MilliValue(), usageMetrics.Usage.Memory().Value()

		cpuUsages := NewCpuResource(_cpuUsages)
		cpuRequests := NewCpuResource(_cpuRequests.MilliValue())
		cpuLimits := NewCpuResource(_cpuLimits.MilliValue())

		memoryUsages := NewMemoryResource(_memoryUsages)
		memoryRequests := NewMemoryResource(_memoryRequests.Value())
		memoryLimits := NewMemoryResource(_memoryLimits.Value())
		podCapacity := capacity.Pods().Value()
		podFraction := calcPercentage(int64(len(podList.Items)), podCapacity)

		_nvidiaGpuCountsRequests, _nvidiaGpuCountsLimits := reqs[ResourceNvidiaGpuCounts], limits[ResourceNvidiaGpuCounts]
		nvidiaGpuCountsRequests := _nvidiaGpuCountsRequests.Value()
		nvidiaGpuCountsLimits := _nvidiaGpuCountsLimits.Value()
		nvidiaGpuCountsCapacity := NewGpuResource(ResourceNvidiaGpuCounts, &capacity).Value()

		// _aliyunGpuMemRequests, _aliyunGpuMemLimits := reqs[ResourceAliyunGpuMem], limits[ResourceAliyunGpuMem]
		// aliyunGpuMemRequests := _aliyunGpuMemRequests.Value()
		// aliyunGpuMemLimits := _aliyunGpuMemLimits.Value()
		// aliyunGpuMemCapacity := NewGpuResource(ResourceAliyunGpuMem, &capacity).Value()

		nodeAllocatedResources = NodeAllocatedResources{
			CPUResources{
				CPUUsages:           cpuUsages,
				CPURequests:         cpuRequests,
				CPURequestsFraction: cpuRequests.calcPercentage(capacity.Cpu()),
				CPULimits:           cpuLimits,
				CPULimitsFraction:   cpuLimits.calcPercentage(capacity.Cpu()),
				CPUCapacity:         NewCpuResource(capacity.Cpu().MilliValue()),
			},
			MemoryResources{
				MemoryUsages:           memoryUsages,
				MemoryRequests:         memoryRequests,
				MemoryRequestsFraction: memoryRequests.calcPercentage(capacity.Memory()),
				MemoryLimits:           memoryLimits,
				MemoryLimitsFraction:   memoryLimits.calcPercentage(capacity.Memory()),
				MemoryCapacity:         NewMemoryResource(capacity.Memory().Value()),
			},
			GPUResources{
				NvidiaGpuCountsRequests:         nvidiaGpuCountsRequests,
				NvidiaGpuCountsRequestsFraction: calcPercentage(nvidiaGpuCountsRequests, nvidiaGpuCountsCapacity),
				NvidiaGpuCountsLimits:           nvidiaGpuCountsLimits,
				NvidiaGpuCountsLimitsFraction:   calcPercentage(nvidiaGpuCountsLimits, nvidiaGpuCountsCapacity),
				NvidiaGpuCountsCapacity:         nvidiaGpuCountsCapacity,
				// AliyunGpuMemRequests:            aliyunGpuMemRequests,
				// AliyunGpuMemRequestsFraction:    calcPercentage(aliyunGpuMemRequests, aliyunGpuMemCapacity),
				// AliyunGpuMemLimits:              aliyunGpuMemLimits,
				// AliyunGpuMemLimitsFraction:      calcPercentage(aliyunGpuMemLimits, aliyunGpuMemCapacity),
				// AliyunGpuMemCapacity:            aliyunGpuMemCapacity,
			},
			PodResources{
				AllocatedPods: len(podList.Items),
				PodCapacity:   podCapacity,
				PodFraction:   podFraction,
			},
		}
	}
	return nodeAllocatedResources, nil
}

//getPodAllocatedResources
func getPodAllocatedResources(pod *v1.Pod, podmetric *metricsapi.PodMetrics, resourceType string) (PodAllocatedResources, error) {

	reqs, limits := map[v1.ResourceName]resource.Quantity{}, map[v1.ResourceName]resource.Quantity{}

	podReqs, podLimits, err := PodRequestsAndLimits(pod)
	if err != nil {
		return PodAllocatedResources{}, err
	}
	for podReqName, podReqValue := range podReqs {
		if value, ok := reqs[podReqName]; !ok {
			reqs[podReqName] = podReqValue.DeepCopy()
		} else {
			value.Add(podReqValue)
			reqs[podReqName] = value
		}
	}
	for podLimitName, podLimitValue := range podLimits {
		if value, ok := limits[podLimitName]; !ok {
			limits[podLimitName] = podLimitValue.DeepCopy()
		} else {
			value.Add(podLimitValue)
			limits[podLimitName] = value
		}
	}
	//podMetricsByPodName := getPodMetricsByPodName(podMetricsList)
	//usageMetrics := podMetricsByPodName[pod.Name]
	usageMetrics := getPodMetrics(podmetric)

	var podAllocatedResources = PodAllocatedResources{}
	switch {
	case resourceType == "cpu":
		_cpuRequests, _cpuLimits := reqs[v1.ResourceCPU], limits[v1.ResourceCPU]
		_cpuUsages := usageMetrics[v1.ResourceCPU]

		cpuUsages := NewCpuResource(_cpuUsages.MilliValue())
		cpuRequests := NewCpuResource(_cpuRequests.MilliValue())
		cpuLimits := NewCpuResource(_cpuLimits.MilliValue())

		podAllocatedResources = PodAllocatedResources{
			CPUUsages:         cpuUsages,
			CPUUsagesFraction: cpuUsages.calcPercentage(&_cpuLimits),
			CPURequests:       cpuRequests,
			CPULimits:         cpuLimits,
		}
	case resourceType == "memory":
		_memoryRequests, _memoryLimits := reqs[v1.ResourceMemory], limits[v1.ResourceMemory]
		_memoryUsages := usageMetrics[v1.ResourceMemory]

		memoryUsages := NewMemoryResource(_memoryUsages.Value())
		memoryRequests := NewMemoryResource(_memoryRequests.Value())
		memoryLimits := NewMemoryResource(_memoryLimits.Value())
		podAllocatedResources = PodAllocatedResources{
			MemoryUsages:         memoryUsages,
			MemoryUsagesFraction: memoryUsages.calcPercentage(&_memoryLimits),
			MemoryRequests:       memoryRequests,
			MemoryLimits:         memoryLimits,
		}
	case resourceType == "gpu":
		_nvidiaGpuCountsRequests, _nvidiaGpuCountsLimits := reqs[ResourceNvidiaGpuCounts], limits[ResourceNvidiaGpuCounts]
		nvidiaGpuCountsRequests := _nvidiaGpuCountsRequests.Value()
		nvidiaGpuCountsLimits := _nvidiaGpuCountsLimits.Value()

		// _aliyunGpuMemRequests, _aliyunGpuMemLimits := reqs[ResourceAliyunGpuMem], limits[ResourceAliyunGpuMem]
		// aliyunGpuMemRequests := _aliyunGpuMemRequests.Value()
		// aliyunGpuMemLimits := _aliyunGpuMemLimits.Value()

		podAllocatedResources = PodAllocatedResources{
			NvidiaGpuCountsRequests: nvidiaGpuCountsRequests,
			NvidiaGpuCountsLimits:   nvidiaGpuCountsLimits,
			// AliyunGpuMemRequests:    aliyunGpuMemRequests,
			// AliyunGpuMemLimits:      aliyunGpuMemLimits,
		}

	default:
		_cpuRequests, _cpuLimits, _memoryRequests, _memoryLimits := reqs[v1.ResourceCPU], limits[v1.ResourceCPU],
			reqs[v1.ResourceMemory], limits[v1.ResourceMemory]
		_cpuUsages, _memoryUsages := usageMetrics[v1.ResourceCPU], usageMetrics[v1.ResourceMemory]

		cpuUsages := NewCpuResource(_cpuUsages.MilliValue())
		cpuRequests := NewCpuResource(_cpuRequests.MilliValue())
		cpuLimits := NewCpuResource(_cpuLimits.MilliValue())

		memoryUsages := NewMemoryResource(_memoryUsages.Value())
		memoryRequests := NewMemoryResource(_memoryRequests.Value())
		memoryLimits := NewMemoryResource(_memoryLimits.Value())

		_nvidiaGpuCountsRequests, _nvidiaGpuCountsLimits := reqs[ResourceNvidiaGpuCounts], limits[ResourceNvidiaGpuCounts]
		nvidiaGpuCountsRequests := _nvidiaGpuCountsRequests.Value()
		nvidiaGpuCountsLimits := _nvidiaGpuCountsLimits.Value()

		// _aliyunGpuMemRequests, _aliyunGpuMemLimits := reqs[ResourceAliyunGpuMem], limits[ResourceAliyunGpuMem]
		// aliyunGpuMemRequests := _aliyunGpuMemRequests.Value()
		// aliyunGpuMemLimits := _aliyunGpuMemLimits.Value()

		podAllocatedResources = PodAllocatedResources{
			CPUUsages:               cpuUsages,
			CPUUsagesFraction:       cpuUsages.calcPercentage(&_cpuLimits),
			CPURequests:             cpuRequests,
			CPULimits:               cpuLimits,
			MemoryUsages:            memoryUsages,
			MemoryUsagesFraction:    memoryUsages.calcPercentage(&_memoryLimits),
			MemoryRequests:          memoryRequests,
			MemoryLimits:            memoryLimits,
			NvidiaGpuCountsRequests: nvidiaGpuCountsRequests,
			NvidiaGpuCountsLimits:   nvidiaGpuCountsLimits,
			// AliyunGpuMemRequests:    aliyunGpuMemRequests,
			// AliyunGpuMemLimits:      aliyunGpuMemLimits,
		}
	}
	return podAllocatedResources, nil
}

// PodRequestsAndLimits returns a dictionary of all defined resources summed up for all
// containers of the pod. If pod overhead is non-nil, the pod overhead is added to the
// total container resource requests and to the total container limits which have a
// non-zero quantity.
func PodRequestsAndLimits(pod *v1.Pod) (reqs, limits v1.ResourceList, err error) {
	reqs, limits = v1.ResourceList{}, v1.ResourceList{}
	for _, container := range pod.Spec.Containers {
		addResourceList(reqs, container.Resources.Requests)
		addResourceList(limits, container.Resources.Limits)
	}
	// init containers define the minimum of any resource
	for _, container := range pod.Spec.InitContainers {
		maxResourceList(reqs, container.Resources.Requests)
		maxResourceList(limits, container.Resources.Limits)
	}

	// Add overhead for running a pod to the sum of requests and to non-zero limits:
	if pod.Spec.Overhead != nil {
		addResourceList(reqs, pod.Spec.Overhead)

		for name, quantity := range pod.Spec.Overhead {
			if value, ok := limits[name]; ok && !value.IsZero() {
				value.Add(quantity)
				limits[name] = value
			}
		}
	}
	return
}

// addResourceList adds the resources in newList to list
func addResourceList(list, new v1.ResourceList) {
	for name, quantity := range new {
		if value, ok := list[name]; !ok {
			list[name] = quantity.DeepCopy()
		} else {
			value.Add(quantity)
			list[name] = value
		}
	}
}

// maxResourceList sets list to the greater of list/newList for every resource
// either list
func maxResourceList(list, new v1.ResourceList) {
	for name, quantity := range new {
		if value, ok := list[name]; !ok {
			list[name] = quantity.DeepCopy()
			continue
		} else {
			if quantity.Cmp(value) > 0 {
				list[name] = quantity.DeepCopy()
			}
		}
	}
}
