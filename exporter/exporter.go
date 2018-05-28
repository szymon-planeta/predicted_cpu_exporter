package exporter

import (
	"fmt"
	"sync"
	"time"

	"golang.org/x/net/context"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/api"
	"github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
)

type Exporter struct {
	mutex		sync.Mutex
	clientAPI	v1.API

	up		*prometheus.Desc
	predictedCpu	prometheus.Gauge
}

func NewExporter(url string) *Exporter {
	client, err := api.NewClient(api.Config{Address:url})
	cliAPI := v1.NewAPI(client)

	if err != nil {
		fmt.Errorf("%v", err)
	}

	return &Exporter {
		clientAPI: cliAPI,
		up: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "up"),
			"Could the prometheus be reached",
			nil,
			nil),
		predictedCpu: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Name: "cpu_usage",
			Help: "Predicted CPU usage in milicores",
			ConstLabels: prometheus.Labels{"namespace":"default", "service":"podinfo"},
		}),
	}
}

func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	ch <- e.up
	e.predictedCpu.Describe(ch)
}

func (e *Exporter) collect(ch chan<- prometheus.Metric) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	val, err := e.clientAPI.Query(ctx, "sum(rate(container_cpu_usage_seconds_total{namespace=\"default\",pod_name=~\"podinfo.*\"}[1m]))*1000", time.Now())

	if err != nil {
		ch <- prometheus.MustNewConstMetric(e.up, prometheus.GaugeValue, 0)
		return fmt.Errorf("Error scraping prometheus: %v", err)
	}
	ch <- prometheus.MustNewConstMetric(e.up, prometheus.GaugeValue, 1)

	switch {
		case val.Type() == model.ValVector:
			vectorVal := val.(model.Vector)
			if len(vectorVal) != 1 { return fmt.Errorf("Received vector with size different than 1") }
			e.predictedCpu.Set(float64(vectorVal[0].Value) + 100)
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
