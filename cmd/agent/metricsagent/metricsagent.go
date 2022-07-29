package metricsagent

import (
	"encoding/json"
	"log"
	"time"

	"github.com/nivanov045/silver-octo-train/cmd/agent/metricsperformer"
	"github.com/nivanov045/silver-octo-train/cmd/agent/requester"
	"github.com/nivanov045/silver-octo-train/internal/metrics"
)

type metricsagent struct {
	Metrics        metrics.Metrics
	pollInterval   time.Duration
	reportInterval time.Duration
}

func (a *metricsagent) updateMetrics() {
	ticker := time.NewTicker(a.pollInterval)
	for {
		<-ticker.C
		metricsperformer.New().UpdateMetrics(a.Metrics)
		log.Println("metricsagent::updateMetrics: metrics were updated")
	}
}

func (a *metricsagent) sendMetrics() {
	ticker := time.NewTicker(a.reportInterval)
	for {
		<-ticker.C
		for key, val := range a.Metrics.GaugeMetrics {
			asFloat := float64(val)
			metricForSend := metrics.MetricsInterface{
				ID:    key,
				MType: "gauge",
				Delta: nil,
				Value: &asFloat,
			}
			marshalled, err := json.Marshal(metricForSend)
			if err != nil {
				log.Panicln("metricsagent::sendMetrics: can't marshal gauge metric for sand with", err)
			}
			err = requester.New().Send(marshalled)
			if err != nil {
				log.Println("metricsagent::sendMetrics: can't send gauge with", err)
			}
		}
		pc := a.Metrics.CounterMetrics["PollCount"]
		asint := int64(pc)
		metricForSend := metrics.MetricsInterface{
			ID:    "PollCount",
			MType: "counter",
			Delta: &asint,
			Value: nil,
		}
		marshalled, err := json.Marshal(metricForSend)
		if err != nil {
			log.Panicln("metricsagent::sendMetrics: can't marshal PollCount metric for sand with", err)
		}
		err = requester.New().Send(marshalled)
		if err != nil {
			log.Println("metricsagent::sendMetrics: can't send PollCount with", err)
		}
		log.Println("metricsagent::sendMetrics: metrics were sent")
	}
}

func (a *metricsagent) Start() {
	log.Println("metricsagent::Start: metricsagent started")
	go a.updateMetrics()
	go a.sendMetrics()
}

func New(pollInterval time.Duration, reportInterval time.Duration) *metricsagent {
	return &metricsagent{
		Metrics: metrics.Metrics{
			GaugeMetrics:   map[string]metrics.Gauge{},
			CounterMetrics: map[string]metrics.Counter{},
		},
		pollInterval:   pollInterval,
		reportInterval: reportInterval,
	}
}
