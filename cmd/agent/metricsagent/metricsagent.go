package metricsagent

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
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
			mCopy := m
			metricsperformer.New().UpdateMetrics(mCopy)
			<-a.metricsChannel
			a.metricsChannel <- mCopy
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
			m_original := <-a.metricsChannel
			a.metricsChannel <- m_original
			m := m_original
			for key, val := range m.GaugeMetrics {
				asFloat := float64(val)
				metricForSend := metrics.MetricsInterface{
					ID:    key,
					MType: "gauge",
					Delta: nil,
					Value: &asFloat,
				}
				if len(a.config.Key) > 0 {
					hash := createHash([]byte(a.config.Key), metricForSend)
					metricForSend.Hash = hex.EncodeToString(hash)
					log.Println("metricsagent::sendMetrics: hash", hash)
					log.Println("metricsagent::sendMetrics: string(hash)", string(hash))
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
			asInt := int64(pc)
			metricForSend := metrics.MetricsInterface{
				ID:    "PollCount",
				MType: "counter",
				Delta: &asInt,
				Value: nil,
			}
			if len(a.config.Key) > 0 {
				hash := createHash([]byte(a.config.Key), metricForSend)
				metricForSend.Hash = hex.EncodeToString(hash)
				log.Println("metricsagent::sendMetrics: hash", hash)
				log.Println("metricsagent::sendMetrics: string(hash)", string(hash))
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

func createHash(key []byte, m metrics.MetricsInterface) []byte {
	h := hmac.New(sha256.New, key)
	if m.MType == "gauge" {
		h.Write([]byte(fmt.Sprintf("%s:gauge:%f", m.ID, *m.Value)))
		log.Println("metricsagent::createHash: hash by", fmt.Sprintf("%s:gauge:%f", m.ID, *m.Value), "with key", key)
	} else {
		h.Write([]byte(fmt.Sprintf("%s:counter:%d", m.ID, *m.Delta)))
		log.Println("metricsagent::createHash: hash by", fmt.Sprintf("%s:counter:%d", m.ID, *m.Delta), "with key", key)
	}
	return h.Sum(nil)
}

func (a *metricsagent) Start() {
	log.Println("metricsagent::Start: metricsagent started")
	a.metricsChannel <- metrics.Metrics{
		GaugeMetrics:   map[string]metrics.Gauge{},
		CounterMetrics: map[string]metrics.Counter{},
	}
	go a.updateMetrics()
	//go a.sendMetrics()
	go a.sendSeveralMetrics()
}

func (a *metricsagent) sendSeveralMetrics() {
	ticker := time.NewTicker(a.config.ReportInterval)
	ctx := context.Background()
	for {
		select {
		case <-ctx.Done():
			log.Println("metricsagent::sendSeveralMetrics: ctx.Done")
			return
		case <-ticker.C:
			m_original := <-a.metricsChannel
			a.metricsChannel <- m_original
			m := m_original
			var toSend []metrics.MetricsInterface
			for key, val := range m.GaugeMetrics {
				asFloat := float64(val)
				metricForSend := metrics.MetricsInterface{
					ID:    key,
					MType: "gauge",
					Delta: nil,
					Value: &asFloat,
				}
				if len(a.config.Key) > 0 {
					hash := createHash([]byte(a.config.Key), metricForSend)
					metricForSend.Hash = hex.EncodeToString(hash)
					//log.Println("metricsagent::sendSeveralMetrics: hash", hash)
					//log.Println("metricsagent::sendSeveralMetrics: string(hash)", string(hash))
				}
				toSend = append(toSend, metricForSend)
			}
			pc := m.CounterMetrics["PollCount"]
			asInt := int64(pc)
			metricForSend := metrics.MetricsInterface{
				ID:    "PollCount",
				MType: "counter",
				Delta: &asInt,
				Value: nil,
			}
			if len(a.config.Key) > 0 {
				hash := createHash([]byte(a.config.Key), metricForSend)
				metricForSend.Hash = hex.EncodeToString(hash)
				//log.Println("metricsagent::sendSeveralMetrics: hash", hash)
				//log.Println("metricsagent::sendSeveralMetrics: string(hash)", string(hash))
			}
			toSend = append(toSend, metricForSend)
			marshalled, err := json.Marshal(toSend)
			if err != nil {
				log.Panicln("metricsagent::sendSeveralMetrics: can't marshal PollCount metric for sand with", err)
			}
			err = a.requester.SendSeveral(marshalled)
			if err != nil {
				log.Println("metricsagent::sendSeveralMetrics: can't send PollCount with", err)
			}
			log.Println("metricsagent::sendSeveralMetrics: metrics were sent")
		}
	}
}
