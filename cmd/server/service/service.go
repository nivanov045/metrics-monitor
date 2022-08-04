package service

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"github.com/nivanov045/silver-octo-train/internal/metrics"
)

type Storage interface {
	SetGaugeMetrics(name string, val metrics.Gauge)
	GetGaugeMetrics(name string) (metrics.Gauge, bool)
	SetCounterMetrics(name string, val metrics.Counter)
	GetCounterMetrics(name string) (metrics.Counter, bool)
	GetKnownMetrics() []string
	IsDbConnected() bool
}

type service struct {
	storage Storage
	key     string
}

func New(key string, storage Storage) *service {
	return &service{storage: storage, key: key}
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
	log.Println("service::ParseAndSave: metrics", m)
	metricType := m.MType
	metricName := m.ID
	log.Println("service::ParseAndSave: type:", metricType, "; name:", metricName)
	if metricType == gauge {
		value := m.Value
		if value == nil {
			log.Println("service::ParseAndSave: gauge value is empty")
			return errors.New("wrong query")
		}
		if !ser.checkHash(m) {
			log.Println("service::ParseAndSave: wrong hash")
			return errors.New("wrong hash")
		}
		ser.storage.SetGaugeMetrics(metricName, metrics.Gauge(*value))
	} else if metricType == counter {
		if m.Delta == nil {
			log.Println("service::ParseAndSave: counter delta is empty")
			return errors.New("wrong query")
		}
		value := *m.Delta
		if !ser.checkHash(m) {
			log.Println("service::ParseAndSave: wrong hash")
			return errors.New("wrong hash")
		}
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
	//if ser.checkHash(m) == false {
	//	log.Println("service::ParseAndGet: wrong hash")
	//	return nil, errors.New("wrong hash")
	//}
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
		if len(ser.key) > 0 {
			m.Hash = hex.EncodeToString(createHash([]byte(ser.key), m))
		}
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
		if len(ser.key) > 0 {
			m.Hash = hex.EncodeToString(createHash([]byte(ser.key), m))
		}
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

func (ser *service) checkHash(m metrics.MetricsInterface) bool {
	hash := createHash([]byte(ser.key), m)
	received, _ := hex.DecodeString(m.Hash)
	if len(ser.key) > 0 && !hmac.Equal(received, hash) {
		log.Println("service::checkHash: wrong hash: made", hash, "but received as []byte", received, "as string", m.Hash)
		return false
	}
	return true
}

func createHash(key []byte, m metrics.MetricsInterface) []byte {
	h := hmac.New(sha256.New, key)
	if m.MType == gauge {
		h.Write([]byte(fmt.Sprintf("%s:gauge:%f", m.ID, *m.Value)))
		log.Println("service::createHash: hash by", fmt.Sprintf("%s:gauge:%f", m.ID, *m.Value), "with key", key)
	} else {
		h.Write([]byte(fmt.Sprintf("%s:counter:%d", m.ID, *m.Delta)))
		log.Println("service::createHash: hash by", fmt.Sprintf("%s:counter:%d", m.ID, *m.Delta), "with key", key)
	}
	return h.Sum(nil)
}

func (ser *service) GetKnownMetrics() []string {
	return ser.storage.GetKnownMetrics()
}

func (ser *service) IsDbConnected() bool {
	return ser.storage.IsDbConnected()
}
