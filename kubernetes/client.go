package kubernetes

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/google-cloud-tools/kube-eagle/options"
	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth" // Auth required for out of cluster connections
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	v1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
	metrics "k8s.io/metrics/pkg/client/clientset/versioned"
)

// Client provides methods to get all required metrics from Kubernetes
type Client struct {
	apiClient     *kubernetes.Clientset
	metricsClient *metrics.Clientset
}

// NewClient creates a new client to get data from kubernetes masters
func NewClient(opts *options.Options) (*Client, error) {
	// Get right config to connect to kubernetes
	var config *rest.Config
	if opts.IsInCluster {
		log.Info("Creating InCluster config to communicate with Kubernetes master")
		var err error
		config, err = rest.InClusterConfig()
		if err != nil {
			return nil, err
		}
	} else {
		// Try to read currently set kubernetes config from your local kube config
		log.Info("Looking for Kubernetes config to communicate with Kubernetes master")
		kubeConfigPath, err := getKubeConfigPath()
		if err != nil {
			return nil, err
		}
		// use the current context in kubeconfig
		config, err = clientcmd.BuildConfigFromFlags("", kubeConfigPath)
		if err != nil {
			return nil, fmt.Errorf("read kubeconfig: %v", err)
		}
	}

	// We got two clients, one for the common API and one explicitly for metrics
	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("error creating kubernetes main client: '%v'", err)
	}

	metricsClient, err := metrics.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("error creating kubernetes metrics client: '%v'", err)
	}

	return &Client{
		apiClient:     client,
		metricsClient: metricsClient,
	}, nil
}

// NodeList returns a list of all known nodes in a kubernetes cluster
func (c *Client) NodeList() (*corev1.NodeList, error) {
	nodeList, err := c.apiClient.CoreV1().Nodes().List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	return nodeList, nil
}

// PodList returns a list of all known pods in a kubernetes cluster
func (c *Client) PodList() (*corev1.PodList, error) {
	podList, err := c.apiClient.CoreV1().Pods(metav1.NamespaceAll).List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	return podList, nil
}

// PodMetricses returns all pods' usage metrics
func (c *Client) PodMetricses() (*v1beta1.PodMetricsList, error) {
	podMetricses, err := c.metricsClient.MetricsV1beta1().PodMetricses(metav1.NamespaceAll).List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	return podMetricses, nil
}

// NodeMetricses returns all nodes' usage metrics
func (c *Client) NodeMetricses() (*v1beta1.NodeMetricsList, error) {
	nodeMetricses, err := c.metricsClient.MetricsV1beta1().NodeMetricses().List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	return nodeMetricses, nil
}

// IsHealthy returns whether the kubernetes client is able to get a list of all pods
func (c *Client) IsHealthy() bool {
	_, err := c.apiClient.CoreV1().Pods(metav1.NamespaceSystem).List(metav1.ListOptions{})
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Error("kubernetes client is not healthy")
		return false
	}

	return true
}

// getKubeConfigPath returns the filepath to the local kubeConfig file or fails if it couldn't find it
func getKubeConfigPath() (string, error) {
	home := os.Getenv("HOME")

	// Mac OS
	if home != "" {
		configPath := filepath.Join(home, ".kube", "config")
		_, err := os.Stat(configPath)
		if err == nil {
			return configPath, nil
		}
	}

	// Windows
	home = os.Getenv("USERPROFILE")
	if home != "" {
		configPath := filepath.Join(home, ".kube", "config")
		_, err := os.Stat(configPath)
		if err == nil {
			return configPath, nil
		}
	}

	return "", fmt.Errorf("couldn't find home directory to look for the kube config")
}
