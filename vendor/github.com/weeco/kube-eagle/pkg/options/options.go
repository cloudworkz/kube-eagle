package options

// Options are configuration options that can be set by Environment Variables
// Port - Port to listen on for the prometheus exporter
// IsInCluster - Whether to use in cluster communication (if deployed inside of Kubernetes) or to look for a kubeconfig in home directory
// Namespace - Prefix of exposed prometheus metrics
type Options struct {
	Port        int    `envconfig:"PORT" default:"8080"`
	IsInCluster bool   `envconfig:"IS_IN_CLUSTER" default:"true"`
	Namespace   string `envconfig:"METRICS_NAME" default:"eagle"`
}

// NewOptions provides Application Options
func NewOptions() *Options {
	return &Options{}
}
