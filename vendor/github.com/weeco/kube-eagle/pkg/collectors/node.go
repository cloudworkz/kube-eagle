package collectors

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/weeco/kube-eagle/pkg/metrics_store"
)

var (
	labelsNodes = []string{"node"}

	// resources allocatable
	allocatableNodeCPUCoresGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "node_resource_allocatable",
			Name:      "cpu_cores",
			Help:      "Allocatable CPU cores",
		},
		labelsNodes,
	)
	allocatableNodeRAMBytesGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "node_resource_allocatable",
			Name:      "memory_bytes",
			Help:      "Allocatable memory bytes",
		},
		labelsNodes,
	)

	// resources requested
	requestedNodeCPUCoresGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "node_resource_requests",
			Name:      "cpu_cores",
			Help:      "Requested CPU cores in Kubernetes configuration",
		},
		labelsNodes,
	)
	requestedNodeRAMBytesGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "node_resource_requests",
			Name:      "memory_bytes",
			Help:      "Requested memory bytes in Kubernetes configuration",
		},
		labelsNodes,
	)

	// resources limit
	limitNodeRAMBytesGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "node_resource_limits",
			Name:      "memory_bytes",
			Help:      "Memory bytes limit in Kubernetes configuration",
		},
		labelsNodes,
	)
	limitNodeCPUCoresGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "node_resource_limits",
			Name:      "cpu_cores",
			Help:      "CPU cores limit in Kubernetes configuration",
		},
		labelsNodes,
	)

	// resources usage
	usageNodeRAMBytesGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "node_resource_usage",
			Name:      "memory_bytes",
			Help:      "Memory bytes usage",
		},
		labelsNodes,
	)
	usageNodeCPUCoresGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "node_resource_usage",
			Name:      "cpu_cores",
			Help:      "CPU cores usage",
		},
		labelsNodes,
	)
)

func init() {
	prometheus.MustRegister(allocatableNodeCPUCoresGauge)
	prometheus.MustRegister(allocatableNodeRAMBytesGauge)

	prometheus.MustRegister(requestedNodeCPUCoresGauge)
	prometheus.MustRegister(requestedNodeRAMBytesGauge)

	prometheus.MustRegister(limitNodeRAMBytesGauge)
	prometheus.MustRegister(limitNodeCPUCoresGauge)

	prometheus.MustRegister(usageNodeRAMBytesGauge)
	prometheus.MustRegister(usageNodeCPUCoresGauge)
}

// updates exposed node metrics in prometheus client
func UpdateNodeMetrics() {
	nodeMetrics := metricsstore.BuildNodeMetrics()

	for _, nodeMetric := range nodeMetrics {
		nodeLabels := prometheus.Labels{
			"node": nodeMetric.Node,
		}
		allocatableNodeCPUCoresGauge.With(nodeLabels).Set(nodeMetric.AllocatableCPUCores)
		allocatableNodeRAMBytesGauge.With(nodeLabels).Set(nodeMetric.AllocatableMemoryBytes)

		requestedNodeCPUCoresGauge.With(nodeLabels).Set(nodeMetric.RequestedCPUCores)
		requestedNodeRAMBytesGauge.With(nodeLabels).Set(nodeMetric.RequestedMemoryBytes)

		limitNodeCPUCoresGauge.With(nodeLabels).Set(nodeMetric.LimitCPUCores)
		limitNodeRAMBytesGauge.With(nodeLabels).Set(nodeMetric.LimitMemoryBytes)

		usageNodeCPUCoresGauge.With(nodeLabels).Set(nodeMetric.UsageCPUCores)
		usageNodeRAMBytesGauge.With(nodeLabels).Set(nodeMetric.UsageMemoryBytes)
	}
}
