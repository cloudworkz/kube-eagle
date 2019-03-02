package collector

import (
	"github.com/google-cloud-tools/kube-eagle/options"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	v1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
	"sync"
)

type containerResourcesCollector struct {
	// Resource limits
	limitCPUCoresDesc    *prometheus.Desc
	limitMemoryBytesDesc *prometheus.Desc

	// Resource requests
	requestCPUCoresDesc    *prometheus.Desc
	requestMemoryBytesDesc *prometheus.Desc

	// Resource usage
	usageCPUCoresDesc    *prometheus.Desc
	usageMemoryBytesDesc *prometheus.Desc
}

func init() {
	registerCollector("container_resources", newContainerResourcesCollector)
}

func newContainerResourcesCollector(opts *options.Options) (Collector, error) {
	subsystem := "pod_container_resource"
	labels := []string{"pod", "container", "qos", "phase", "namespace", "node"}

	return &containerResourcesCollector{
		// Prometheus metrics
		// Resource limits
		limitCPUCoresDesc: prometheus.NewDesc(
			prometheus.BuildFQName(opts.Namespace, subsystem, "limits_cpu_cores"),
			"The container's CPU limit in Kubernetes",
			labels,
			prometheus.Labels{},
		),
		limitMemoryBytesDesc: prometheus.NewDesc(
			prometheus.BuildFQName(opts.Namespace, subsystem, "limits_memory_bytes"),
			"The container's RAM limit in Kubernetes",
			labels,
			prometheus.Labels{},
		),
		// Resource requests
		requestCPUCoresDesc: prometheus.NewDesc(
			prometheus.BuildFQName(opts.Namespace, subsystem, "requests_cpu_cores"),
			"The container's requested CPU resources in Kubernetes",
			labels,
			prometheus.Labels{},
		),
		requestMemoryBytesDesc: prometheus.NewDesc(
			prometheus.BuildFQName(opts.Namespace, subsystem, "requests_memory_bytes"),
			"The container's requested RAM resources in Kubernetes",
			labels,
			prometheus.Labels{},
		),
		// Resource usage
		usageCPUCoresDesc: prometheus.NewDesc(
			prometheus.BuildFQName(opts.Namespace, subsystem, "usage_cpu_cores"),
			"CPU usage in number of cores",
			labels,
			prometheus.Labels{},
		),
		usageMemoryBytesDesc: prometheus.NewDesc(
			prometheus.BuildFQName(opts.Namespace, subsystem, "usage_memory_bytes"),
			"RAM usage in bytes",
			labels,
			prometheus.Labels{},
		),
	}, nil
}

func (c *containerResourcesCollector) updateMetrics(ch chan<- prometheus.Metric) error {
	log.Debug("Collecting container metrics")

	var wg sync.WaitGroup
	var podList *corev1.PodList
	var podListError error
	var podMetricses *v1beta1.PodMetricsList
	var podMetricsesError error

	// Get pod list
	wg.Add(1)
	go func() {
		defer wg.Done()
		podList, podListError = kubernetesClient.PodList()
	}()

	// Get node resource usage metrics
	wg.Add(1)
	go func() {
		defer wg.Done()
		podMetricses, podMetricsesError = kubernetesClient.PodMetricses()
	}()

	wg.Wait()
	if podListError != nil {
		log.Warn("Failed to get podList from Kubernetes", podListError)
		return podListError
	}
	if podMetricsesError != nil {
		log.Warn("Failed to get podMetricses from Kubernetes", podMetricsesError)
		return podMetricsesError
	}

	containerMetricses := buildEnrichedContainerMetricses(podList, podMetricses)

	for _, containerMetrics := range containerMetricses {
		cm := *containerMetrics
		log.Debugf("Test")
		labelValues := []string{cm.Pod, cm.Container, cm.Qos, cm.Phase, cm.Namespace, cm.Node}
		ch <- prometheus.MustNewConstMetric(c.requestCPUCoresDesc, prometheus.GaugeValue, cm.RequestCPUCores, labelValues...)
		ch <- prometheus.MustNewConstMetric(c.requestMemoryBytesDesc, prometheus.GaugeValue, cm.RequestMemoryBytes, labelValues...)
		ch <- prometheus.MustNewConstMetric(c.limitCPUCoresDesc, prometheus.GaugeValue, cm.LimitCPUCores, labelValues...)
		ch <- prometheus.MustNewConstMetric(c.limitMemoryBytesDesc, prometheus.GaugeValue, cm.LimitMemoryBytes, labelValues...)
		ch <- prometheus.MustNewConstMetric(c.usageCPUCoresDesc, prometheus.GaugeValue, cm.UsageCPUCores, labelValues...)
		ch <- prometheus.MustNewConstMetric(c.usageMemoryBytesDesc, prometheus.GaugeValue, cm.UsageMemoryBytes, labelValues...)
	}

	return nil
}

type enrichedContainerMetricses struct {
	Node               string
	Pod                string
	Container          string
	Qos                string
	Phase              string
	Namespace          string
	RequestCPUCores    float64
	RequestMemoryBytes float64
	LimitCPUCores      float64
	LimitMemoryBytes   float64
	UsageCPUCores      float64
	UsageMemoryBytes   float64
}

// buildEnrichedContainerMetricses merges the container metrics from two requests (podList request and podMetrics request) into
// one, so that we can expose valuable metadata (such as a nodename) as prometheus labels which is just present
// in one of the both responses.
func buildEnrichedContainerMetricses(podList *corev1.PodList, podMetricses *v1beta1.PodMetricsList) []*enrichedContainerMetricses {
	// Group container metricses by pod name
	containerMetricsesByPod := make(map[string]map[string]v1beta1.ContainerMetrics)
	for _, pm := range podMetricses.Items {
		containerMetricses := make(map[string]v1beta1.ContainerMetrics)
		for _, c := range pm.Containers {
			containerMetricses[c.Name] = c
		}
		containerMetricsesByPod[pm.Name] = containerMetricses
	}

	var containerMetricses []*enrichedContainerMetricses
	for _, podInfo := range podList.Items {
		containers := append(podInfo.Spec.Containers, podInfo.Spec.InitContainers...)

		for _, containerInfo := range containers {
			qos := string(podInfo.Status.QOSClass)

			// Resources requested
			requestCPUCores := float64(containerInfo.Resources.Requests.Cpu().MilliValue()) / 1000
			requestMemoryBytes := float64(containerInfo.Resources.Requests.Memory().MilliValue()) / 1000

			// Resources limit
			limitCPUCores := float64(containerInfo.Resources.Limits.Cpu().MilliValue()) / 1000
			limitMemoryBytes := float64(containerInfo.Resources.Limits.Memory().MilliValue()) / 1000

			// Resources usage
			containerUsageMetrics := containerMetricsesByPod[podInfo.Name][containerInfo.Name]
			usageCPUCores := float64(containerUsageMetrics.Usage.Cpu().MilliValue()) / 1000
			usageMemoryBytes := float64(containerUsageMetrics.Usage.Memory().MilliValue()) / 1000

			nodeName := podInfo.Spec.NodeName
			metric := &enrichedContainerMetricses{
				Node:               nodeName,
				Container:          containerInfo.Name,
				Pod:                podInfo.Name,
				Qos:                qos,
				Phase:              string(podInfo.Status.Phase),
				Namespace:          podInfo.Namespace,
				RequestCPUCores:    requestCPUCores,
				RequestMemoryBytes: requestMemoryBytes,
				LimitCPUCores:      limitCPUCores,
				LimitMemoryBytes:   limitMemoryBytes,
				UsageCPUCores:      usageCPUCores,
				UsageMemoryBytes:   usageMemoryBytes,
			}
			containerMetricses = append(containerMetricses, metric)
		}
	}

	return containerMetricses
}
