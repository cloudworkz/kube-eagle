package sink

import "k8s.io/metrics/pkg/apis/metrics/v1beta1"

// ContainerMetrics defines the labels and values we expose with prometheus
type ContainerMetrics struct {
	Node                 string
	Container            string
	Pod                  string
	Qos                  string
	Namespace            string
	RequestedCPUCores    float64
	RequestedMemoryBytes float64
	LimitCPUCores        float64
	LimitMemoryBytes     float64
	UsageCPUCores        float64
	UsageMemoryBytes     float64
}

// containerUsageMap returns a map of of maps
// Where the outer map key is the podname which contains the map of
// ContainerMetrics (where container name is the key)
// It basically as a map of ContainerMetrics grouped by (podName, containerName)
func containerUsageMap() map[string]map[string]v1beta1.ContainerMetrics {
	containerUsageByName := make(map[string]map[string]v1beta1.ContainerMetrics)

	for _, podMetrics := range podUsageList.Items {
		podMetricsMap := make(map[string]v1beta1.ContainerMetrics)
		for _, containerMetrics := range podMetrics.Containers {
			podMetricsMap[containerMetrics.Name] = containerMetrics
		}
		containerUsageByName[podMetrics.Name] = podMetricsMap
	}

	return containerUsageByName
}

// BuildContainerMetrics returns all container relevant exported prometheus metrics
func BuildContainerMetrics() []ContainerMetrics {
	var containerMetrics []ContainerMetrics
	containerUsageByName := containerUsageMap()

	for _, podInfo := range podList.Items {
		for _, containerInfo := range podInfo.Spec.Containers {
			qos := string(podInfo.Status.QOSClass)

			// Resources requested
			requestedCPUCores := float64(containerInfo.Resources.Requests.Cpu().MilliValue()) / 1000
			requestedMemoryBytes := float64(containerInfo.Resources.Requests.Memory().MilliValue()) / 1000

			// Resources limit
			limitCPUCores := float64(containerInfo.Resources.Limits.Cpu().MilliValue()) / 1000
			limitMemoryBytes := float64(containerInfo.Resources.Limits.Memory().MilliValue()) / 1000

			// Resources usage
			containerUsageMetrics := containerUsageByName[podInfo.Name][containerInfo.Name]
			usageCPUCores := float64(containerUsageMetrics.Usage.Cpu().MilliValue()) / 1000
			usageMemoryBytes := float64(containerUsageMetrics.Usage.Memory().MilliValue()) / 1000

			nodeName := podInfo.Spec.NodeName
			metric := ContainerMetrics{
				Node:                 nodeName,
				Container:            containerInfo.Name,
				Pod:                  podInfo.Name,
				Qos:                  qos,
				Namespace:            podInfo.Namespace,
				RequestedCPUCores:    requestedCPUCores,
				RequestedMemoryBytes: requestedMemoryBytes,
				LimitCPUCores:        limitCPUCores,
				LimitMemoryBytes:     limitMemoryBytes,
				UsageCPUCores:        usageCPUCores,
				UsageMemoryBytes:     usageMemoryBytes,
			}
			containerMetrics = append(containerMetrics, metric)
		}
	}

	return containerMetrics
}
