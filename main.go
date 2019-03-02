package main

import (
	"fmt"
	"github.com/google-cloud-tools/kube-eagle/collector"
	"github.com/prometheus/client_golang/prometheus"
	"net/http"
	"os"
	"strconv"

	"github.com/google-cloud-tools/kube-eagle/options"
	"github.com/kelseyhightower/envconfig"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
)

func healthcheck(collector *collector.KubeEagleCollector) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Debug("Healthcheck has been called")
		if collector.IsHealthy() == true {
			w.Write([]byte("Ok"))
		} else {
			http.Error(w, "Healthcheck failed", http.StatusServiceUnavailable)
		}
	})
}

func main() {
	// Initialize logrus settings
	log.SetOutput(os.Stdout)
	log.SetFormatter(&log.JSONFormatter{})

	// Parse & validate environment variable
	opts := options.NewOptions()
	err := envconfig.Process("", opts)
	if err != nil {
		log.Fatal("Error parsing env vars into opts", err)
	}

	// Set log level from environment variable
	level, err := log.ParseLevel(opts.LogLevel)
	if err != nil {
		log.Panicf("Loglevel could not be parsed as one of the known loglevels. See logrus documentation for valid log level inputs. Given input was: '%s'", opts.LogLevel)
	}
	log.SetLevel(level)

	// Start kube eagle exporter
	log.Infof("Starting kube eagle v%v", opts.Version)
	collector, err := collector.NewKubeEagleCollector(opts)
	if err != nil {
		log.Fatalf("could not start kube eagle collector: '%v'", err)
	}
	prometheus.MustRegister(collector)

	http.Handle("/metrics", promhttp.Handler())
	http.Handle("/health", healthcheck(collector))
	address := fmt.Sprintf("%v:%s", opts.Host, strconv.Itoa(opts.Port))
	log.Info("Listening on ", address)
	log.Fatal(http.ListenAndServe(address, nil))
}
