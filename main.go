package main

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/kelseyhightower/envconfig"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"github.com/weeco/kube-eagle/pkg/collectors"
	"github.com/weeco/kube-eagle/pkg/metrics_store"
	"github.com/weeco/kube-eagle/pkg/options"
)

var (
	opts *options.Options
)

func init() {
	log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(os.Stdout)
	log.SetLevel(log.InfoLevel)

	opts = options.NewOptions()
	err := envconfig.Process("", opts)
	if err != nil {
		log.Fatal(err, "error parsing env vars into opts")
	}
}
func hea(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h.ServeHTTP(w, r)
	})
}

func healthcheck() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Debug("Healthcheck has been called")
		isKubernetesClientHealthy := metricsstore.IsClientHealthy()
		if isKubernetesClientHealthy == true {
			w.Write([]byte("Ok"))
		} else {
			http.Error(w, "Healthcheck failed", http.StatusServiceUnavailable)
		}
	})
}

func main() {
	go func() {
		metricsstore.InitKuberneterClient(opts)
		// Collect stats every 10s
		for {
			metricsstore.Collect()
			collectors.UpdateContainerMetrics()
			collectors.UpdateNodeMetrics()
			time.Sleep(10 * time.Second)
		}
	}()

	http.Handle("/metrics", promhttp.Handler())
	http.Handle("/health", healthcheck())
	address := fmt.Sprintf("0.0.0.0:%s", strconv.Itoa(opts.Port))
	log.Info("Listening on ", address)
	log.Fatal(http.ListenAndServe(address, nil))
}
