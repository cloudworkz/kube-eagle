package metricsstore

import (
	"os"
	"path/filepath"

	"k8s.io/client-go/rest"

	log "github.com/sirupsen/logrus"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/metrics/pkg/apis/metrics/v1beta1"
	metrics "k8s.io/metrics/pkg/client/clientset/versioned"

	"github.com/weeco/kube-eagle/pkg/options"

	// Needed for GCP auth - only relevant for out of cluster communications (developers)
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

var (
	config           *rest.Config
	clientset        *kubernetes.Clientset
	metricsClientset *metrics.Clientset

	nodeList      *v1.NodeList
	podList       *v1.PodList
	podUsageList  *v1beta1.PodMetricsList
	nodeUsageList *v1beta1.NodeMetricsList
)

func init() {
	log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(os.Stdout)
	log.SetLevel(log.InfoLevel)
}

// Collect gathers all needed metrics and actually fires requests against the kubernetes master
func Collect() {
	var err error

	// get kubernetes' node metrics
	nodeList, err = clientset.CoreV1().Nodes().List(metav1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}

	// get kubernetes' pod metrics
	podList, err = clientset.CoreV1().Pods(metav1.NamespaceAll).List(metav1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}

	// get pods' resource usage metrics
	podUsageList, err = metricsClientset.MetricsV1beta1().PodMetricses(metav1.NamespaceAll).List(metav1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}

	// get nodes' resource usage metrics
	nodeUsageList, err = metricsClientset.MetricsV1beta1().NodeMetricses().List(metav1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}
}

// BuildNodeMetrics returns all node relevant exposed prometheus metrics
func BuildNodeMetrics() []NodeMetrics {
	var nodeMetricsPrepared []NodeMetrics
	nodeResourceConfig := getResourceDetailsByNode()
	nodeMetricsByName := getNodeMetricsByNode()

	for _, nodeInfo := range nodeList.Items {
		// resources allocatable
		allocatableCPUCores := float64(nodeInfo.Status.Allocatable.Cpu().MilliValue()) / 1000
		allocatableMemoryBytes := float64(nodeInfo.Status.Allocatable.Memory().MilliValue()) / 1000

		// resource usage
		nodeUsageMetrics := nodeMetricsByName[nodeInfo.Name]
		nodeCPUUsage := float64(nodeUsageMetrics.Usage.Cpu().MilliValue()) / 1000
		nodeMemoryBytesUsage := float64(nodeUsageMetrics.Usage.Memory().MilliValue()) / 1000

		// resources requested
		resourceConfig := nodeResourceConfig[nodeInfo.Name]

		nodeMetric := NodeMetrics{
			Node:                   nodeInfo.Name,
			AllocatableCPUCores:    allocatableCPUCores,
			AllocatableMemoryBytes: allocatableMemoryBytes,
			RequestedCPUCores:      resourceConfig.cpuCoresRequested,
			RequestedMemoryBytes:   resourceConfig.memoryBytesRequested,
			LimitCPUCores:          resourceConfig.cpuCoresLimit,
			LimitMemoryBytes:       resourceConfig.memoryBytesLimit,
			UsageCPUCores:          nodeCPUUsage,
			UsageMemoryBytes:       nodeMemoryBytesUsage,
		}
		nodeMetricsPrepared = append(nodeMetricsPrepared, nodeMetric)
	}

	return nodeMetricsPrepared
}

// InitKuberneterClient parses kubeconfig and creates kubernetes clientset
func InitKuberneterClient(opts *options.Options) {
	if opts.IsInCluster {
		log.Info("Creating InCluster config to communicate with Kubernetes master")
		var err error
		config, err = rest.InClusterConfig()
		if err != nil {
			panic(err.Error())
		}
	} else {
		log.Info("Looking for Kubernetes config to communicate with Kubernetes master")
		home := homeDir()
		kubeconfigPath := filepath.Join(home, ".kube", "config")

		// use the current context in kubeconfig
		var err error
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfigPath)
		if err != nil {
			panic(err.Error())
		}
	}

	// create the clientset
	var err error
	clientset, err = kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	metricsClientset, err = metrics.NewForConfig(config)
	if err != nil {
		panic(err)
	}
}

// IsClientHealthy tries to get PodList. If this is successful the client is considered healthy
func IsClientHealthy() bool {
	_, err := clientset.CoreV1().Pods(metav1.NamespaceDefault).List(metav1.ListOptions{})
	if err != nil {
		return false
	}

	return true
}

// returns os homedir
func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}
