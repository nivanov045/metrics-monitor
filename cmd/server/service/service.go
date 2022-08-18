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
	GetCounterMetrics(name string) (metrics.Counter, bool)
	GetGaugeMetrics(name string) (metrics.Gauge, bool)
	GetKnownMetrics() []string
	IsDBConnected() bool
	SetCounterMetrics(name string, val metrics.Counter) error
	SetGaugeMetrics(name string, val metrics.Gauge) error
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
	log.Println("service::ParseAndSave::info: started:", string(s))
	var m metrics.Interface
	err := json.Unmarshal(s, &m)
	if err != nil {
		log.Println("service::ParseAndSave::warning: can't unmarshal with error:", err)
		return errors.New("wrong query")
	}
	metricType := m.MType
	metricName := m.ID
	log.Println("service::ParseAndSave:info: type:", metricType, "|| name:", metricName)
	if metricType == gauge {
		value := m.Value
		if value == nil {
			log.Println("service::ParseAndSave::info: gauge value is empty")
			return errors.New("wrong query")
		}
		if !ser.checkHash(m) {
			log.Println("service::ParseAndSave::info: wrong hash")
			return errors.New("wrong hash")
		}
		err = ser.storage.SetGaugeMetrics(metricName, metrics.Gauge(*value))
		if err != nil {
			log.Println("service::ParseAndSave::error: problem in metrics saving:", err)
			return errors.New("problem in metrics saving")
		}
	} else if metricType == counter {
		if m.Delta == nil {
			log.Println("service::ParseAndSave::info: counter delta is empty")
			return errors.New("wrong query")
		}
		value := *m.Delta
		if !ser.checkHash(m) {
			log.Println("service::ParseAndSave::info: wrong hash")
			return errors.New("wrong hash")
		}
		exVal, ok := ser.storage.GetCounterMetrics(metricName)
		if !ok {
			log.Println("service::ParseAndSave::info: new counter metric")
			err = ser.storage.SetCounterMetrics(metricName, metrics.Counter(value))
			if err != nil {
				log.Println("service::ParseAndSave::error: problem in metrics saving:", err)
				return errors.New("problem in metrics saving")
			}
		} else {
			log.Println("service::ParseAndSave::info: update counter metric")
			err = ser.storage.SetCounterMetrics(metricName, metrics.Counter(int64(exVal)+value))
			if err != nil {
				log.Println("service::ParseAndSave::error: problem in metrics saving:", err)
				return errors.New("problem in metrics saving")
			}
		}
	} else {
		log.Println("service::ParseAndSave::info: unknown metrics type")
		return errors.New("wrong metrics type")
	}
	return nil
}

func (ser *service) ParseAndGet(s []byte) ([]byte, error) {
	log.Println("service::ParseAndGet::info: started:", string(s))
	var m metrics.Interface
	err := json.Unmarshal(s, &m)
	if err != nil {
		log.Println("service::ParseAndGet::warning: can't unmarshal with error:", err)
		return nil, errors.New("wrong query")
	}
	metricType := m.MType
	metricName := m.ID
	log.Println("service::ParseAndGet::info: type:", metricType, "|| name:", metricName)
	if metricType == gauge {
		val, ok := ser.storage.GetGaugeMetrics(metricName)
		if !ok {
			log.Println("service::ParseAndGet::info: no such gauge metrics")
			return nil, errors.New("no such metric")
		}
		asFloat := float64(val)
		m.Value = &asFloat
		if len(ser.key) > 0 {
			m.Hash = hex.EncodeToString(createHash([]byte(ser.key), m))
		}
		marshal, err := json.Marshal(m)
		if err != nil {
			log.Panic("service::ParseAndGet::error: can't marshal gauge metric with:", err)
			return nil, err
		}
		return marshal, nil
	} else if metricType == counter {
		val, ok := ser.storage.GetCounterMetrics(metricName)
		if !ok {
			log.Println("service::ParseAndGet::info: no such counter metrics")
			return nil, errors.New("no such metric")
		}
		asint := int64(val)
		m.Delta = &asint
		if len(ser.key) > 0 {
			m.Hash = hex.EncodeToString(createHash([]byte(ser.key), m))
		}
		marshal, err := json.Marshal(m)
		if err != nil {
			log.Panic("service::ParseAndGet::error: can't marshal counter metric with:", err)
			return nil, err
		}
		return marshal, nil
	}
	log.Println("service::ParseAndGet::info: unknown metrics type")
	return nil, errors.New("wrong metrics type")
}

func (ser *service) checkHash(m metrics.Interface) bool {
	hash := createHash([]byte(ser.key), m)
	received, _ := hex.DecodeString(m.Hash)
	if len(ser.key) > 0 && !hmac.Equal(received, hash) {
		log.Println("service::checkHash::info: wrong hash")
		return false
	}
	return true
}

func createHash(key []byte, m metrics.Interface) []byte {
	h := hmac.New(sha256.New, key)
	if m.MType == gauge {
		h.Write([]byte(fmt.Sprintf("%s:gauge:%f", m.ID, *m.Value)))
	} else {
		h.Write([]byte(fmt.Sprintf("%s:counter:%d", m.ID, *m.Delta)))
	}
	return h.Sum(nil)
}

func (ser *service) GetKnownMetrics() []string {
	return ser.storage.GetKnownMetrics()
}

func (ser *service) IsDBConnected() bool {
	return ser.storage.IsDBConnected()
}

func (ser *service) ParseAndSaveSeveral(s []byte) error {
	log.Println("service::ParseAndSaveSeveral::info: started:", string(s))
	var mall []metrics.Interface
	err := json.Unmarshal(s, &mall)
	if err != nil {
		log.Println("service::ParseAndSaveSeveral::warning: can't unmarshal with error:", err)
		return errors.New("wrong query")
	}
	for _, m := range mall {
		metricType := m.MType
		metricName := m.ID
		log.Println("service::ParseAndSaveSeveral::info: type:", metricType, "|| name:", metricName)
		if metricType == gauge {
			value := m.Value
			if value == nil {
				log.Println("service::ParseAndSaveSeveral::info: gauge value is empty")
				continue
			}
			if !ser.checkHash(m) {
				log.Println("service::ParseAndSaveSeveral::info: wrong hash")
				continue
			}
			err = ser.storage.SetGaugeMetrics(metricName, metrics.Gauge(*value))
			if err != nil {
				log.Println("service::ParseAndSave::error: problem in metrics saving:", err)
				continue
			}
		} else if metricType == counter {
			if m.Delta == nil {
				log.Println("service::ParseAndSaveSeveral::info: counter delta is empty")
				continue
			}
			value := *m.Delta
			if !ser.checkHash(m) {
				log.Println("service::ParseAndSaveSeveral::info: wrong hash")
				continue
			}
			exVal, ok := ser.storage.GetCounterMetrics(metricName)
			if !ok {
				log.Println("service::ParseAndSaveSeveral::info: new counter metric")
				err = ser.storage.SetCounterMetrics(metricName, metrics.Counter(value))
				if err != nil {
					log.Println("service::ParseAndSave::error: problem in metrics saving:", err)
					continue
				}
			} else {
				log.Println("service::ParseAndSaveSeveral::info: update counter metric")
				err = ser.storage.SetCounterMetrics(metricName, metrics.Counter(int64(exVal)+value))
				if err != nil {
					log.Println("service::ParseAndSave::error: problem in metrics saving:", err)
					continue
				}
			}
		} else {
			log.Println("service::ParseAndSaveSeveral::info: unknown metrics type")
			continue
		}
	}
	return nil
}
