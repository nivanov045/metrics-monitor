package service

import (
	"encoding/json"
	"errors"
	"log"

	"github.com/nivanov045/silver-octo-train/internal/metrics"
)

type Storage interface {
	SetGaugeMetrics(name string, val metrics.Gauge)
	GetGaugeMetrics(name string) (metrics.Gauge, bool)
	SetCounterMetrics(name string, val metrics.Counter)
	GetCounterMetrics(name string) (metrics.Counter, bool)
	GetKnownMetrics() []string
}

type service struct {
	storage Storage
}

func New(storage Storage) *service {
	return &service{storage: storage}
}

const (
	gauge   string = "gauge"
	counter string = "counter"
)

func (ser *service) ParseAndSave(s []byte) error {
	log.Println("service::ParseAndSave: started", string(s))
	var m metrics.MetricsInterface
	err := json.Unmarshal(s, &m)
	if err != nil {
		log.Println("service::ParseAndSave: can't unmarshal with error", err)
		return errors.New("wrong query")
	}
	metricType := m.MType
	metricName := m.ID
	log.Println("service::ParseAndSave: type:", metricType, "; name:", metricName)
	if metricType == gauge {
		value := m.Value
		if value == nil {
			log.Println("service::ParseAndSave: gauge value is empty")
			return errors.New("wrong query")
		}
		ser.storage.SetGaugeMetrics(metricName, metrics.Gauge(*value))
	} else if metricType == counter {
		if m.Delta == nil {
			log.Println("service::ParseAndSave: counter delta is empty")
			return errors.New("wrong query")
		}
		value := *m.Delta
		exVal, ok := ser.storage.GetCounterMetrics(metricName)
		if !ok {
			log.Println("service::ParseAndSave: new counter metric")
			ser.storage.SetCounterMetrics(metricName, metrics.Counter(value))
		} else {
			log.Println("service::ParseAndSave: update counter metric")
			ser.storage.SetCounterMetrics(metricName, metrics.Counter(int64(exVal)+value))
		}
	} else {
		log.Println("service::ParseAndSave: unknown metrics type")
		return errors.New("wrong metrics type")
	}
	return nil
}

func (ser *service) ParseAndGet(s []byte) ([]byte, error) {
	log.Println("service::ParseAndGet: started", string(s))
	var m metrics.MetricsInterface
	err := json.Unmarshal(s, &m)
	if err != nil {
		log.Println("service::ParseAndGet: can't unmarshal with error", err)
		return nil, errors.New("wrong query")
	}
	metricType := m.MType
	metricName := m.ID
	log.Println("service::ParseAndGet: type:", metricType, "; name:", metricName)
	if metricType == gauge {
		val, ok := ser.storage.GetGaugeMetrics(metricName)
		if !ok {
			log.Println("service::ParseAndGet: no such gauge metrics")
			return nil, errors.New("no such metric")
		}
		asFloat := float64(val)
		m.Value = &asFloat
		marshal, err := json.Marshal(m)
		if err != nil {
			log.Panic("service::ParseAndGet: can't marshal gauge metric with", err)
			return nil, err
		}
		return marshal, nil
	} else if metricType == counter {
		val, ok := ser.storage.GetCounterMetrics(metricName)
		if !ok {
			log.Println("service::ParseAndGet: no such counter metrics")
			return nil, errors.New("no such metric")
		}
		asint := int64(val)
		m.Delta = &asint
		marshal, err := json.Marshal(m)
		if err != nil {
			log.Panic("service::ParseAndGet: can't marshal caunter metric with", err)
			return nil, err
		}
		return marshal, nil
	}
	log.Println("service::ParseAndGet: unknown metrics type")
	return nil, errors.New("wrong metrics type")
}

func (ser *service) GetKnownMetrics() []string {
	return ser.storage.GetKnownMetrics()
}
