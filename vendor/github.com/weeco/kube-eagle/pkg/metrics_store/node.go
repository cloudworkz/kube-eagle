package metricsstore

import "k8s.io/metrics/pkg/apis/metrics/v1beta1"

// NodeMetrics defines the labels and values we expose with prometheus
type NodeMetrics struct {
	Node                   string
	AllocatableCPUCores    float64
	AllocatableMemoryBytes float64
	RequestedCPUCores      float64
	RequestedMemoryBytes   float64
	LimitCPUCores          float64
	LimitMemoryBytes       float64
	UsageCPUCores          float64
	UsageMemoryBytes       float64
}

type nodeResourceDetails struct {
	cpuCoresRequested    float64
	memoryBytesRequested float64
	cpuCoresLimit        float64
	memoryBytesLimit     float64
}

func getResourceDetailsByNode() map[string]nodeResourceDetails {
	nodeResourcesByNodeName := make(map[string]nodeResourceDetails)

	// iterate through all pod definitions to aggregate pods' resource requests and limits by node
	for _, podInfo := range podList.Items {
		nodeName := podInfo.Spec.NodeName
		for _, containerInfo := range podInfo.Spec.Containers {
			// TODO: Add container status as label
			requestedCPUCores := float64(containerInfo.Resources.Requests.Cpu().MilliValue()) / 1000
			requestedMemoryBytes := float64(containerInfo.Resources.Requests.Memory().MilliValue()) / 1000
			limitCPUCores := float64(containerInfo.Resources.Limits.Cpu().MilliValue()) / 1000
			limitMemoryBytes := float64(containerInfo.Resources.Limits.Memory().MilliValue()) / 1000

			newCPURequest := nodeResourcesByNodeName[nodeName].cpuCoresRequested + requestedCPUCores
			newMemoryBytesRequest := nodeResourcesByNodeName[nodeName].memoryBytesRequested + requestedMemoryBytes
			newCPULimit := nodeResourcesByNodeName[nodeName].cpuCoresLimit + limitCPUCores
			newMemoryBytesLimit := nodeResourcesByNodeName[nodeName].memoryBytesLimit + limitMemoryBytes
			nodeResourcesByNodeName[nodeName] = nodeResourceDetails{
				cpuCoresRequested:    newCPURequest,
				cpuCoresLimit:        newCPULimit,
				memoryBytesRequested: newMemoryBytesRequest,
				memoryBytesLimit:     newMemoryBytesLimit,
			}
		}
	}

	return nodeResourcesByNodeName
}

func getNodeMetricsByNode() map[string]v1beta1.NodeMetrics {
	resourceUsageMap := make(map[string]v1beta1.NodeMetrics)
	for _, localNodeMetrics := range nodeUsageList.Items {
		resourceUsageMap[localNodeMetrics.Name] = localNodeMetrics
	}

	return resourceUsageMap
}
