package main

import (
	"flag"
	"net/http"
	"os"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"

	"github.com/szymon-planeta/predicted_cpu_exporter/exporter"
	"github.com/szymon-planeta/predicted_cpu_exporter/algorithm"
)

const (
	metricsEndpoint = "/metrics"
)

var (
	addr		= flag.String("listen-address", ":8080", "The address to listen on for HTTP requests.")
	promURL		= flag.String("prometheus-url", "http://localhost:9090", "URL to prometheus")
	mode		= flag.String("mode", "ARMA", "Prediction algorithm, possible values: ARMA, DES")
)



func main() {
	flag.Parse()

	var alg algorithm.Algorithm

	switch *mode {
		case "ARMA":
			alg = algorithm.NewArma()
		case "DES":
			alg = algorithm.NewDES()
		default:
			log.Fatalf("Undefined mode: %s", *mode)
			os.Exit(1)
	}

	exp := exporter.NewExporter(*promURL, alg)
	prometheus.MustRegister(exp)

	log.Infoln("Starting predicted_cpu_exporter")
	log.Infof("Starting server: %s", *addr)

	http.Handle(metricsEndpoint, prometheus.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
			<head><title>Predicted CPU Exporter</title><head>
			<body>
			<h1>Predicted CPU Exporter</h1>
			<p><a href='` + metricsEndpoint + `'>Metrics</a></p>
			</body>
			</html>`))
	})

	log.Fatal(http.ListenAndServe(*addr, nil))
}
