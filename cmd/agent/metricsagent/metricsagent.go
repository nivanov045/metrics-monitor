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

func (a *metricsagent) updateRuntimeMetrics() {
	ticker := time.NewTicker(a.config.PollInterval)
	ctx := context.Background()
	for {
		select {
		case <-ctx.Done():
			log.Println("metricsagent::updateRuntimeMetrics::info: ctx.Done")
			return
		case <-ticker.C:
			log.Println("metricsagent::updateRuntimeMetrics::info: start update")
			mOriginal := <-a.metricsChannel
			mCopy := mOriginal.Clone()
			a.metricsChannel <- mOriginal
			metricsperformer.New().UpdateRuntimeMetrics(mCopy)
			<-a.metricsChannel
			a.metricsChannel <- mCopy
			log.Println("metricsagent::updateRuntimeMetrics::info: finish update")
		}
	}
}

func (a *metricsagent) updateAdditionalMetrics() {
	ticker := time.NewTicker(a.config.PollInterval)
	ctx := context.Background()
	for {
		select {
		case <-ctx.Done():
			log.Println("metricsagent::updateAdditionalMetrics::info: ctx.Done")
			return
		case <-ticker.C:
			log.Println("metricsagent::updateAdditionalMetrics::info: start update")
			mOriginal := <-a.metricsChannel
			mCopy := mOriginal.Clone()
			a.metricsChannel <- mOriginal
			err := metricsperformer.New().UpdateAdditionalMetrics(mCopy)
			if err != nil {
				log.Println("metricsagent::updateAdditionalMetrics::error:", err)
			}
			<-a.metricsChannel
			a.metricsChannel <- mCopy
			log.Println("metricsagent::updateAdditionalMetrics::info: finish update")
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
				metricForSend := metrics.Metric{
					ID:    key,
					MType: "gauge",
					Delta: nil,
					Value: &asFloat,
				}
				if len(a.config.Key) > 0 {
					hash := createHash([]byte(a.config.Key), metricForSend)
					metricForSend.Hash = hex.EncodeToString(hash)
				}
				marshalled, err := json.Marshal(metricForSend)
				if err != nil {
					log.Println("metricsagent::sendMetrics::error: can't marshal gauge metric for sand with:", err)
					return
				}
				err = a.requester.Send(marshalled)
				if err != nil {
					log.Println("metricsagent::sendMetrics:error: can't send gauge with:", err)
				}
			}
			pc := m.CounterMetrics["PollCount"]
			asInt := int64(pc)
			metricForSend := metrics.Metric{
				ID:    "PollCount",
				MType: "counter",
				Delta: &asInt,
				Value: nil,
			}
			if len(a.config.Key) > 0 {
				hash := createHash([]byte(a.config.Key), metricForSend)
				metricForSend.Hash = hex.EncodeToString(hash)
			}
			marshalled, err := json.Marshal(metricForSend)
			if err != nil {
				log.Println("metricsagent::sendMetrics::error: can't marshal PollCount metric for sand with:", err)
				return
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

func createHash(key []byte, m metrics.Metric) []byte {
	h := hmac.New(sha256.New, key)
	if m.MType == "gauge" {
		h.Write([]byte(fmt.Sprintf("%s:gauge:%f", m.ID, *m.Value)))
	} else {
		h.Write([]byte(fmt.Sprintf("%s:counter:%d", m.ID, *m.Delta)))
	}
	return h.Sum(nil)
}

func (a *metricsagent) Start() {
	log.Println("metricsagent::Start::info: metricsagent started")
	a.metricsChannel <- metrics.Metrics{
		GaugeMetrics:   map[string]metrics.Gauge{},
		CounterMetrics: map[string]metrics.Counter{},
	}
	go a.updateRuntimeMetrics()
	go a.updateAdditionalMetrics()
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
			var toSend []metrics.Metric
			for key, val := range m.GaugeMetrics {
				asFloat := float64(val)
				metricForSend := metrics.Metric{
					ID:    key,
					MType: "gauge",
					Delta: nil,
					Value: &asFloat,
				}
				if len(a.config.Key) > 0 {
					hash := createHash([]byte(a.config.Key), metricForSend)
					metricForSend.Hash = hex.EncodeToString(hash)
				}
				toSend = append(toSend, metricForSend)
			}
			pc := m.CounterMetrics["PollCount"]
			asInt := int64(pc)
			metricForSend := metrics.Metric{
				ID:    "PollCount",
				MType: "counter",
				Delta: &asInt,
				Value: nil,
			}
			if len(a.config.Key) > 0 {
				hash := createHash([]byte(a.config.Key), metricForSend)
				metricForSend.Hash = hex.EncodeToString(hash)
			}
			toSend = append(toSend, metricForSend)
			marshalled, err := json.Marshal(toSend)
			if err != nil {
				log.Println("metricsagent::sendSeveralMetrics::error: can't marshal PollCount metric for sand with", err)
				return
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
