package metricsagent

import (
	"context"
	"encoding/json"
	"github.com/nivanov045/silver-octo-train/cmd/agent/agentconfig"
	"log"
	"time"

	"github.com/nivanov045/silver-octo-train/cmd/agent/metricsperformer"
	"github.com/nivanov045/silver-octo-train/cmd/agent/requester"
	"github.com/nivanov045/silver-octo-train/internal/metrics"
)

type metricsagent struct {
	metricsChannel chan metrics.Metrics
	config         agentconfig.Config
	requester      requester.Requester
}

func New(c agentconfig.Config) *metricsagent {
	return &metricsagent{
		metricsChannel: make(chan metrics.Metrics, 1),
		config:         c,
		requester:      *requester.New(c.Address),
	}
}

func (a *metricsagent) updateMetrics() {
	ticker := time.NewTicker(a.config.PollInterval)
	ctx := context.Background()
	for {
		select {
		case <-ctx.Done():
			log.Println("metricsagent::updateMetrics: ctx.Done")
			return
		case <-ticker.C:
			log.Println("metricsagent::updateMetrics: time to update")
			m := <-a.metricsChannel
			a.metricsChannel <- m
			metricsperformer.New().UpdateMetrics(m)
			<-a.metricsChannel
			a.metricsChannel <- m
			log.Println("metricsagent::updateMetrics: metrics were updated")
		}
	}
}

func (a *metricsagent) sendMetrics() {
	ticker := time.NewTicker(a.config.ReportInterval)
	ctx := context.Background()
	for {
		select {
		case <-ctx.Done():
			log.Println("metricsagent::sendMetrics: ctx.Done")
			return
		case <-ticker.C:
			m := <-a.metricsChannel
			a.metricsChannel <- m
			for key, val := range m.GaugeMetrics {
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
				err = a.requester.Send(marshalled)
				if err != nil {
					log.Println("metricsagent::sendMetrics: can't send gauge with", err)
				}
			}
			pc := m.CounterMetrics["PollCount"]
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
			err = a.requester.Send(marshalled)
			if err != nil {
				log.Println("metricsagent::sendMetrics: can't send PollCount with", err)
			}
			log.Println("metricsagent::sendMetrics: metrics were sent")
		}
	}
}

func (a *metricsagent) Start() {
	log.Println("metricsagent::Start: metricsagent started")
	a.metricsChannel <- metrics.Metrics{
		GaugeMetrics:   map[string]metrics.Gauge{},
		CounterMetrics: map[string]metrics.Counter{},
	}
	go a.updateMetrics()
	go a.sendMetrics()
}
