package sink

import (
	"github.com/google-cloud-tools/kube-eagle/pkg/log"
	"os"
	"path/filepath"
	"time"

	"k8s.io/client-go/rest"

	"github.com/google-cloud-tools/kube-eagle/pkg/options"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/metrics/pkg/apis/metrics/v1beta1"
	metrics "k8s.io/metrics/pkg/client/clientset/versioned"

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

// Collect gathers all needed metrics and actually fires requests against the kubernetes master
func Collect() {
	var err error
	var errorCount int8

	start := time.Now()
	// get kubernetes' node metrics
	nodeList, err = clientset.CoreV1().Nodes().List(metav1.ListOptions{})
	if err != nil {
		errorCount++
		log.Warn("Couldn't get nodeList from Kubernetes master", err.Error())
	}

	// get kubernetes' pod metrics
	podList, err = clientset.CoreV1().Pods(metav1.NamespaceAll).List(metav1.ListOptions{})
	if err != nil {
		errorCount++
		log.Warn("Couldn't get podList from Kubernetes master", err.Error())
	}

	// get pods' resource usage metrics
	podUsageList, err = metricsClientset.MetricsV1beta1().PodMetricses(metav1.NamespaceAll).List(metav1.ListOptions{})
	if err != nil {
		errorCount++
		log.Warn("Couldn't get podUsageList from Kubernetes master", err.Error())
	}

	// get nodes' resource usage metrics
	nodeUsageList, err = metricsClientset.MetricsV1beta1().NodeMetricses().List(metav1.ListOptions{})
	if err != nil {
		errorCount++
		log.Warn("Couldn't get nodeUsageList from Kubernetes master", err.Error())
	}
	elapsed := time.Since(start)
	elapsedMs := int64(elapsed / time.Millisecond)
	log.Infof("Collected metrics with %v errors from Kubernetes cluster within %vms", errorCount, elapsedMs)
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
	var err error
	if opts.IsInCluster {
		log.Info("Creating InCluster config to communicate with Kubernetes master")
		config, err = rest.InClusterConfig()
		if err != nil {
			panic(err.Error())
		}
	} else {
		log.Info("Looking for Kubernetes config to communicate with Kubernetes master")
		home := homeDir()
		kubeconfigPath := filepath.Join(home, ".kube", "config")

		// use the current context in kubeconfig
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfigPath)
		if err != nil {
			panic(err.Error())
		}
	}

	// create the clientset
	clientset, err = kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatal("Error while creating kubernetes clientSet", err.Error())
	}

	metricsClientset, err = metrics.NewForConfig(config)
	if err != nil {
		log.Fatal("Error while creating metrics clientSet", err.Error())
	}
}

// IsClientHealthy tries to get PodList. If this is successful the client is considered healthy
func IsClientHealthy() bool {
	_, err := clientset.CoreV1().Pods(metav1.NamespaceDefault).List(metav1.ListOptions{})
	if err != nil {
		log.Warn("Kubernetes client is not healthy. Couldn't list pods in the default namespace.", err.Error())
		return false
	}

	return true
}

// returns os homedir
func homeDir() string {
	home := os.Getenv("HOME")
	if home != "" {
		return home
	}
	home = os.Getenv("USERPROFILE") // windows
	if home != "" {
		return home
	}
	log.Fatal("Couldn't find home directory to look for the kube config.")
	return ""
}
