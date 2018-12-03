package collectors

import (
	"github.com/google-cloud-tools/kube-eagle/pkg/sink"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	labelsContainers = []string{"node", "container", "pod", "qos", "namespace", "phase"}
	namespace        = "eagle"

	// resources requested
	requestedContainerCPUCoresGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "pod_container_resource_requests",
			Name:      "cpu_cores",
			Help:      "Requested CPU cores in Kubernetes configuration",
		},
		labelsContainers,
	)
	requestedContainerRAMBytesGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "pod_container_resource_requests",
			Name:      "memory_bytes",
			Help:      "Requested memory bytes in Kubernetes configuration",
		},
		labelsContainers,
	)

	// resources limit
	limitContainerRAMBytesGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "pod_container_resource_limits",
			Name:      "memory_bytes",
			Help:      "Memory bytes limit in Kubernetes configuration",
		},
		labelsContainers,
	)
	limitContainerCPUCoresGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "pod_container_resource_limits",
			Name:      "cpu_cores",
			Help:      "CPU cores limit in Kubernetes configuration",
		},
		labelsContainers,
	)

	// resources usage
	usageContainerRAMBytesGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "pod_container_resource_usage",
			Name:      "memory_bytes",
			Help:      "Memory bytes usage",
		},
		labelsContainers,
	)
	usageContainerCPUCoresGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "pod_container_resource_usage",
			Name:      "cpu_cores",
			Help:      "CPU cores usage",
		},
		labelsContainers,
	)
)

func init() {
	prometheus.MustRegister(requestedContainerCPUCoresGauge)
	prometheus.MustRegister(requestedContainerRAMBytesGauge)

	prometheus.MustRegister(limitContainerCPUCoresGauge)
	prometheus.MustRegister(limitContainerRAMBytesGauge)

	prometheus.MustRegister(usageContainerCPUCoresGauge)
	prometheus.MustRegister(usageContainerRAMBytesGauge)
}

// UpdateContainerMetrics updates exposed container metrics in prometheus client
func UpdateContainerMetrics() {
	containerMetrics := sink.BuildContainerMetrics()

	requestedContainerCPUCoresGauge.Reset()
	requestedContainerRAMBytesGauge.Reset()
	limitContainerCPUCoresGauge.Reset()
	limitContainerRAMBytesGauge.Reset()
	usageContainerCPUCoresGauge.Reset()
	usageContainerRAMBytesGauge.Reset()

	for _, containerMetric := range containerMetrics {
		containerLabels := prometheus.Labels{
			"node":      containerMetric.Node,
			"container": containerMetric.Container,
			"qos":       containerMetric.Qos,
			"pod":       containerMetric.Pod,
			"namespace": containerMetric.Namespace,
			"phase":     string(containerMetric.Phase),
		}

		requestedContainerCPUCoresGauge.With(containerLabels).Set(containerMetric.RequestedCPUCores)
		requestedContainerRAMBytesGauge.With(containerLabels).Set(containerMetric.RequestedMemoryBytes)

		limitContainerCPUCoresGauge.With(containerLabels).Set(containerMetric.LimitCPUCores)
		limitContainerRAMBytesGauge.With(containerLabels).Set(containerMetric.LimitMemoryBytes)

		usageContainerCPUCoresGauge.With(containerLabels).Set(containerMetric.UsageCPUCores)
		usageContainerRAMBytesGauge.With(containerLabels).Set(containerMetric.UsageMemoryBytes)
	}
}
