package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
	"strings"
	"strconv"
	"crypto/tls"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
//	"github.com/prometheus/common/version"
)

const (
	metricsEndpoint = "/metrics"
	namespace = "kubernetes"
)

var (
	addr		= flag.String("listen-address", ":8080", "The address to listen on for HTTP requests.")
	scrapeURI	= flag.String("scrape-uri", "http://localhost:9090/metrics", "URI to prometheus metrics")
)


type Exporter struct {
	URI	string
	mutex	sync.Mutex
	client	*http.Client

	up		*prometheus.Desc
	predictedCpu	prometheus.Gauge
}

func NewExporter(uri string) *Exporter {
	return &Exporter {
		URI: uri,
		up: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "up"),
			"Could the prometheus be reached",
			nil,
			nil),
		predictedCpu: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Name: "predicted_cpu_usage",
			Help: "Predicted CPU usage in milicores",
		}),
		client: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
		},
	}
}

func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	ch <- e.up
	e.predictedCpu.Describe(ch)
}

func (e *Exporter) collect(ch chan<- prometheus.Metric) error {
	resp, err := e.client.Get(e.URI)
	if err != nil {
		ch <- prometheus.MustNewConstMetric(e.up, prometheus.GaugeValue, 0)
		return fmt.Errorf("Error scraping prometheus: %v", err)
	}
	ch <- prometheus.MustNewConstMetric(e.up, prometheus.GaugeValue, 1)

	data, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if resp.StatusCode != 200 {
		if err != nil {
			data = []byte(err.Error())
		}
		return fmt.Errorf("Status %s (%d): %s", resp.Status, resp.StatusCode, data)
	}

	lines := strings.Split(string(data), "\n")

	for _, l := range lines {
		if strings.TrimSpace(l) == "" { continue }
		fields := strings.Fields(l)
		log.Infoln(fields)
		name := fields[0]
		switch {
		case name == "go_goroutines":
			log.Infoln(name)
			value := fields[1]
			val, err := strconv.ParseFloat(value, 64)
			if err != nil {
				return err
			}
			e.predictedCpu.Set(val)
		}
	}

	e.predictedCpu.Collect(ch)

	return nil
}

func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	e.mutex.Lock() // To protect metrics from concurrent collects.
	defer e.mutex.Unlock()
	if err := e.collect(ch); err != nil {
		log.Errorf("Error scraping apache: %s", err)
	}
	return
}

func main() {
	flag.Parse()

	exporter := NewExporter(*scrapeURI)
	prometheus.MustRegister(exporter)

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
