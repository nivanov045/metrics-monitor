package metricsagent

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/nivanov045/metrics-monitor/internal/agent/config"
	"github.com/nivanov045/metrics-monitor/internal/agent/metricsperformer"
	"github.com/nivanov045/metrics-monitor/internal/agent/requester"
	"github.com/nivanov045/metrics-monitor/internal/metrics"
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
			log.Debug().Msg("end select for update")
			return
		case <-ticker.C:
			log.Debug().Msg("start update")

			mOriginal := <-a.metricsChannel
			mCopy := mOriginal.Clone()
			a.metricsChannel <- mOriginal

			metricsperformer.UpdateRuntimeMetrics(mCopy)
			<-a.metricsChannel
			a.metricsChannel <- mCopy

			log.Debug().Msg("finish update")
		}
	}
}

func (a *metricsagent) updateAdditionalMetrics() {
	ticker := time.NewTicker(a.config.PollInterval)
	ctx := context.Background()
	for {
		select {
		case <-ctx.Done():
			log.Debug().Msg("end select for update addititonal metrics")
			return
		case <-ticker.C:
			log.Debug().Msg("start update of additional metrics")

			mOriginal := <-a.metricsChannel
			mCopy := mOriginal.Clone()
			a.metricsChannel <- mOriginal

			err := metricsperformer.UpdateAdditionalMetrics(mCopy)
			if err != nil {
				log.Error().Err(err).Stack()
			}
			<-a.metricsChannel
			a.metricsChannel <- mCopy

			log.Debug().Msg("finish update of additional metrics")
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
	log.Debug().Msg("metricsagent started")

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
			log.Debug().Msg("end select for sending metrics")
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
				log.Error().Err(err).Stack()
				return
			}

			err = a.requester.SendSeveral(marshalled)
			if err != nil {
				log.Error().Err(err).Stack()
				return
			}

			log.Debug().Msg("metrics were sent")
		}
	}
}
