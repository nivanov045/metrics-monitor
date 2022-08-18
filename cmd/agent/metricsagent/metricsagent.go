package metricsagent

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/nivanov045/silver-octo-train/cmd/agent/config"
	"github.com/nivanov045/silver-octo-train/cmd/agent/metricsperformer"
	"github.com/nivanov045/silver-octo-train/cmd/agent/requester"
	"github.com/nivanov045/silver-octo-train/internal/metrics"
)

type metricsagent struct {
	metricsChannel chan metrics.Metrics
	config         config.Config
	requester      requester.Requester
}

func New(c config.Config) *metricsagent {
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
			log.Println("metricsagent::updateMetrics::info: ctx.Done")
			return
		case <-ticker.C:
			log.Println("metricsagent::updateMetrics::info: start update")
			mOriginal := <-a.metricsChannel
			mCopy := mOriginal.Clone()
			a.metricsChannel <- mOriginal
			metricsperformer.New().UpdateMetrics(mCopy)
			<-a.metricsChannel
			a.metricsChannel <- mCopy
			log.Println("metricsagent::updateMetrics::info: finish update")
		}
	}
}

func (a *metricsagent) sendMetrics() {
	ticker := time.NewTicker(a.config.ReportInterval)
	ctx := context.Background()
	for {
		select {
		case <-ctx.Done():
			log.Println("metricsagent::sendMetrics::info: ctx.Done")
			return
		case <-ticker.C:
			mOriginal := <-a.metricsChannel
			m := mOriginal.Clone()
			a.metricsChannel <- mOriginal
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
					//log.Println("metricsagent::sendMetrics::info: hash", hash)
					//log.Println("metricsagent::sendMetrics::info: string(hash)", string(hash))
				}
				marshalled, err := json.Marshal(metricForSend)
				if err != nil {
					log.Panicln("metricsagent::sendMetrics::error: can't marshal gauge metric for sand with:", err)
				}
				err = a.requester.Send(marshalled)
				if err != nil {
					log.Println("metricsagent::sendMetrics:error: can't send gauge with:", err)
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
				//log.Println("metricsagent::sendMetrics::info: hash", hash)
				//log.Println("metricsagent::sendMetrics::info: string(hash)", string(hash))
			}
			marshalled, err := json.Marshal(metricForSend)
			if err != nil {
				log.Panicln("metricsagent::sendMetrics::error: can't marshal PollCount metric for sand with:", err)
			}
			err = a.requester.Send(marshalled)
			if err != nil {
				log.Println("metricsagent::sendMetrics::error: can't send PollCount with:", err)
				return
			}
			log.Println("metricsagent::sendMetrics::info: metrics were sent")
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
	log.Println("metricsagent::Start::info: metricsagent started")
	a.metricsChannel <- metrics.Metrics{
		GaugeMetrics:   map[string]metrics.Gauge{},
		CounterMetrics: map[string]metrics.Counter{},
	}
	go a.updateMetrics()
	go a.sendSeveralMetrics()
}

func (a *metricsagent) sendSeveralMetrics() {
	ticker := time.NewTicker(a.config.ReportInterval)
	ctx := context.Background()
	for {
		select {
		case <-ctx.Done():
			log.Println("metricsagent::sendSeveralMetrics::info: ctx.Done")
			return
		case <-ticker.C:
			mOriginal := <-a.metricsChannel
			m := mOriginal.Clone()
			a.metricsChannel <- mOriginal
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
					//log.Println("metricsagent::sendSeveralMetrics::info: hash", hash)
					//log.Println("metricsagent::sendSeveralMetrics::info: string(hash)", string(hash))
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
				//log.Println("metricsagent::sendSeveralMetrics::info: hash", hash)
				//log.Println("metricsagent::sendSeveralMetrics::info: string(hash)", string(hash))
			}
			toSend = append(toSend, metricForSend)
			marshalled, err := json.Marshal(toSend)
			if err != nil {
				log.Panicln("metricsagent::sendSeveralMetrics::error: can't marshal PollCount metric for sand with", err)
			}
			err = a.requester.SendSeveral(marshalled)
			if err != nil {
				log.Println("metricsagent::sendSeveralMetrics::error: can't send PollCount with", err)
				return
			}
			log.Println("metricsagent::sendSeveralMetrics::info: metrics were sent")
		}
	}
}
