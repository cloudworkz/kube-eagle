package collector

import (
	"github.com/google-cloud-tools/kube-eagle/kubernetes"
	"github.com/google-cloud-tools/kube-eagle/options"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	v1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
	"sync"
)

type nodeResourcesCollector struct {
	kubernetesClient *kubernetes.Client

	// Allocatable
	allocatableCPUCoresDesc    *prometheus.Desc
	allocatableMemoryBytesDesc *prometheus.Desc

	// Resource limits
	limitCPUCoresDesc    *prometheus.Desc
	limitMemoryBytesDesc *prometheus.Desc

	// Resource requests
	requestCPUCoresDesc    *prometheus.Desc
	requestMemoryBytesDesc *prometheus.Desc

	// Resource usage
	usageCPUCoresDesc    *prometheus.Desc
	usageMemoryBytesDesc *prometheus.Desc
	usagePodCount        *prometheus.Desc
}

func init() {
	registerCollector("node_resource", newNodeResourcesCollector)
}

func newNodeResourcesCollector(opts *options.Options) (Collector, error) {
	subsystem := "node_resource"
	labels := []string{"node"}

	return &nodeResourcesCollector{
		// Prometheus metrics
		// Allocatable
		allocatableCPUCoresDesc: prometheus.NewDesc(
			prometheus.BuildFQName(opts.Namespace, subsystem, "allocatable_cpu_cores"),
			"Allocatable CPU cores on a specific node in Kubernetes",
			labels,
			prometheus.Labels{},
		),
		allocatableMemoryBytesDesc: prometheus.NewDesc(
			prometheus.BuildFQName(opts.Namespace, subsystem, "allocatable_memory_bytes"),
			"Allocatable memory bytes on a specific node in Kubernetes",
			labels,
			prometheus.Labels{},
		),
		// Resource limits
		limitCPUCoresDesc: prometheus.NewDesc(
			prometheus.BuildFQName(opts.Namespace, subsystem, "limits_cpu_cores"),
			"Total limit CPU cores of all specified pod resources on a node",
			labels,
			prometheus.Labels{},
		),
		limitMemoryBytesDesc: prometheus.NewDesc(
			prometheus.BuildFQName(opts.Namespace, subsystem, "limits_memory_bytes"),
			"Total limit of RAM bytes of all specified pod resources on a node",
			labels,
			prometheus.Labels{},
		),
		// Resource requests
		requestCPUCoresDesc: prometheus.NewDesc(
			prometheus.BuildFQName(opts.Namespace, subsystem, "requests_cpu_cores"),
			"Total request of CPU cores of all specified pod resources on a node",
			labels,
			prometheus.Labels{},
		),
		requestMemoryBytesDesc: prometheus.NewDesc(
			prometheus.BuildFQName(opts.Namespace, subsystem, "requests_memory_bytes"),
			"Total request of RAM bytes of all specified pod resources on a node",
			labels,
			prometheus.Labels{},
		),
		// Resource usage
		usageCPUCoresDesc: prometheus.NewDesc(
			prometheus.BuildFQName(opts.Namespace, subsystem, "usage_cpu_cores"),
			"Total number of used CPU cores on a node",
			labels,
			prometheus.Labels{},
		),
		usageMemoryBytesDesc: prometheus.NewDesc(
			prometheus.BuildFQName(opts.Namespace, subsystem, "usage_memory_bytes"),
			"Total number of RAM bytes used on a node",
			labels,
			prometheus.Labels{},
		),
		usagePodCount: prometheus.NewDesc(
			prometheus.BuildFQName(opts.Namespace, subsystem, "usage_pod_count"),
			"Total number of running pods for each kubernetes node",
			labels,
			prometheus.Labels{},
		),
	}, nil
}

func (c *nodeResourcesCollector) updateMetrics(ch chan<- prometheus.Metric) error {
	log.Debug("Collecting node metrics")

	var wg sync.WaitGroup
	var nodeList *corev1.NodeList
	var nodeListError error
	var podList *corev1.PodList
	var podListError error
	var nodeMetricsList *v1beta1.NodeMetricsList
	var nodeMetricsListError error

	// Get pod list
	wg.Add(1)
	go func() {
		defer wg.Done()
		podList, podListError = kubernetesClient.PodList()
	}()

	// Get node list
	wg.Add(1)
	go func() {
		defer wg.Done()
		nodeList, nodeListError = kubernetesClient.NodeList()
	}()

	// Get node resource usage metrics
	wg.Add(1)
	go func() {
		defer wg.Done()
		nodeMetricsList, nodeMetricsListError = kubernetesClient.NodeMetricses()
	}()

	wg.Wait()
	if podListError != nil {
		log.Warn("Failed to get podList from Kubernetes", podListError)
		return podListError
	}
	if nodeListError != nil {
		log.Warn("Failed to get nodeList from Kubernetes", nodeListError)
		return nodeListError
	}
	if nodeMetricsListError != nil {
		log.Warn("Failed to get podList from Kubernetes", nodeMetricsListError)
		return nodeMetricsListError
	}
	nodeMetricsByNodeName := getNodeMetricsByNodeName(nodeMetricsList)
	podMetricsByNodeName := getAggregatedPodMetricsByNodeName(podList)

	for _, n := range nodeList.Items {
		// allocatable
		allocatableCPU := n.Status.Allocatable.Cpu().Value()
		allocatableMemoryBytes := float64(n.Status.Allocatable.Memory().MilliValue()) / 1000
		ch <- prometheus.MustNewConstMetric(c.allocatableCPUCoresDesc, prometheus.GaugeValue, float64(allocatableCPU), n.Name)
		ch <- prometheus.MustNewConstMetric(c.allocatableMemoryBytesDesc, prometheus.GaugeValue, float64(allocatableMemoryBytes), n.Name)

		// resource usage
		usageMetrics := nodeMetricsByNodeName[n.Name]
		usageCPU := float64(usageMetrics.Usage.Cpu().MilliValue()) / 1000
		usageMemoryBytes := float64(usageMetrics.Usage.Memory().MilliValue()) / 1000
		ch <- prometheus.MustNewConstMetric(c.usageCPUCoresDesc, prometheus.GaugeValue, float64(usageCPU), n.Name)
		ch <- prometheus.MustNewConstMetric(c.usageMemoryBytesDesc, prometheus.GaugeValue, float64(usageMemoryBytes), n.Name)

		// aggregated pod metrics (e. g. resource requests by node)
		podMetrics := podMetricsByNodeName[n.Name]
		ch <- prometheus.MustNewConstMetric(c.requestCPUCoresDesc, prometheus.GaugeValue, podMetrics.requestedCPUCores, n.Name)
		ch <- prometheus.MustNewConstMetric(c.requestMemoryBytesDesc, prometheus.GaugeValue, float64(podMetrics.requestedMemoryBytes), n.Name)
		ch <- prometheus.MustNewConstMetric(c.limitCPUCoresDesc, prometheus.GaugeValue, podMetrics.limitCPUCores, n.Name)
		ch <- prometheus.MustNewConstMetric(c.limitMemoryBytesDesc, prometheus.GaugeValue, float64(podMetrics.limitMemoryBytes), n.Name)
		ch <- prometheus.MustNewConstMetric(c.usagePodCount, prometheus.GaugeValue, float64(podMetrics.podCount), n.Name)
	}

	return nil
}

// getNodeMetricsByNodeName returns a map of node metrics where the keys are the particular node names
func getNodeMetricsByNodeName(nodeMetricsList *v1beta1.NodeMetricsList) map[string]v1beta1.NodeMetrics {
	nodeMetricsByName := make(map[string]v1beta1.NodeMetrics)
	for _, metrics := range nodeMetricsList.Items {
		nodeMetricsByName[metrics.Name] = metrics
	}

	return nodeMetricsByName
}

type aggregatedPodMetrics struct {
	podCount             uint16
	containerCount       uint16
	requestedMemoryBytes int64
	requestedCPUCores    float64
	limitMemoryBytes     int64
	limitCPUCores        float64
}

// getAggregatedPodMetricsByNodeName returns a map of aggregated pod metrics grouped by node name.
func getAggregatedPodMetricsByNodeName(pods *corev1.PodList) map[string]aggregatedPodMetrics {
	podMetrics := make(map[string]aggregatedPodMetrics)

	// Iterate through all pod definitions to sum and group pods' resource requests and limits by node name
	for _, podInfo := range pods.Items {
		nodeName := podInfo.Spec.NodeName

		// skip not running pods (e. g. failed/succeeded jobs, evicted pods etc.)
		podPhase := podInfo.Status.Phase
		if podPhase == corev1.PodFailed || podPhase == corev1.PodSucceeded {
			continue
		}

		// Don't increment this counter for failed / non running pods
		podCount := podMetrics[nodeName].podCount + 1

		for _, c := range podInfo.Spec.Containers {
			requestedCPUCores := float64(c.Resources.Requests.Cpu().MilliValue()) / 1000
			requestedMemoryBytes := c.Resources.Requests.Memory().MilliValue() / 1000
			limitCPUCores := float64(c.Resources.Limits.Cpu().MilliValue()) / 1000
			limitMemoryBytes := c.Resources.Limits.Memory().MilliValue() / 1000

			podMetrics[nodeName] = aggregatedPodMetrics{
				podCount:             podCount,
				containerCount:       podMetrics[nodeName].containerCount + 1,
				requestedCPUCores:    podMetrics[nodeName].requestedCPUCores + requestedCPUCores,
				requestedMemoryBytes: podMetrics[nodeName].requestedMemoryBytes + requestedMemoryBytes,
				limitCPUCores:        podMetrics[nodeName].limitCPUCores + limitCPUCores,
				limitMemoryBytes:     podMetrics[nodeName].limitMemoryBytes + limitMemoryBytes,
			}
		}
	}

	return podMetrics
}
