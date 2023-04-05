package service

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"log"

	"github.com/nivanov045/silver-octo-train/cmd/server/crypto"
	"github.com/nivanov045/silver-octo-train/internal/metrics"
)

type service struct {
	storage   Storage
	crypto    Crypto
	useCrypto bool
}

func New(key string, storage Storage) *service {
	return &service{storage: storage, crypto: crypto.New(key), useCrypto: len(key) > 0}
}

const (
	gauge   string = "gauge"
	counter string = "counter"
)

// TODO: DRY ParseAndSave and ParseAndSaveSeveral
func (ser *service) ParseAndSave(s []byte) error {
	log.Println("service::ParseAndSave::info: started:", string(s))

	var m metrics.Metric
	err := json.Unmarshal(s, &m)
	if err != nil {
		log.Println("service::ParseAndSave::warning: can't unmarshal with error:", err)
		return errors.New("wrong query")
	}

	metricType := m.MType
	metricName := m.ID

	log.Println("service::ParseAndSave:info: type:", metricType, "|| name:", metricName)

	switch metricType {
	case gauge:
		value := m.Value
		if value == nil {
			log.Println("service::ParseAndSave::info: gauge value is empty")
			return errors.New("wrong query")
		}

		if !ser.crypto.CheckHash(m) {
			log.Println("service::ParseAndSave::info: wrong hash")
			return errors.New("wrong hash")
		}

		err = ser.storage.SetGaugeMetrics(metricName, metrics.Gauge(*value))
		if err != nil {
			log.Println("service::ParseAndSave::error: problem in metrics saving:", err)
			return errors.New("problem in metrics saving")
		}
	case counter:
		if m.Delta == nil {
			log.Println("service::ParseAndSave::info: counter delta is empty")
			return errors.New("wrong query")
		}

		value := *m.Delta

		if !ser.crypto.CheckHash(m) {
			log.Println("service::ParseAndSave::info: wrong hash")
			return errors.New("wrong hash")
		}

		exVal, isKnown := ser.storage.GetCounterMetrics(metricName)
		if !isKnown {
			log.Println("service::ParseAndSave::info: new counter metric")

			err = ser.storage.SetCounterMetrics(metricName, metrics.Counter(value))
			if err != nil {
				log.Println("service::ParseAndSave::error: problem in metrics saving:", err)
				return errors.New("problem in metrics saving")
			}
			return nil
		}

		log.Println("service::ParseAndSave::info: update counter metric")

		err = ser.storage.SetCounterMetrics(metricName, metrics.Counter(int64(exVal)+value))
		if err != nil {
			log.Println("service::ParseAndSave::error: problem in metrics saving:", err)
			return errors.New("problem in metrics saving")
		}
	default:
		log.Println("service::ParseAndSave::info: unknown metrics type")
		return errors.New("wrong metrics type")
	}

	return nil
}

func (ser *service) ParseAndGet(s []byte) ([]byte, error) {
	log.Println("service::ParseAndGet::info: started:", string(s))

	var m metrics.Metric
	err := json.Unmarshal(s, &m)
	if err != nil {
		log.Println("service::ParseAndGet::warning: can't unmarshal with error:", err)
		return nil, errors.New("wrong query")
	}

	metricType := m.MType
	metricName := m.ID

	log.Println("service::ParseAndGet::info: type:", metricType, "|| name:", metricName)

	switch metricType {
	case gauge:
		val, ok := ser.storage.GetGaugeMetrics(metricName)
		if !ok {
			log.Println("service::ParseAndGet::info: no such gauge metrics")
			return nil, errors.New("no such metric")
		}

		asFloat := float64(val)
		m.Value = &asFloat

		if ser.useCrypto {
			m.Hash = hex.EncodeToString(ser.crypto.CreateHash(m))
		}

		marshal, err := json.Marshal(m)
		if err != nil {
			log.Panic("service::ParseAndGet::error: can't marshal gauge metric with:", err)
			return nil, err
		}

		return marshal, nil
	case counter:
		val, ok := ser.storage.GetCounterMetrics(metricName)
		if !ok {
			log.Println("service::ParseAndGet::info: no such counter metrics")
			return nil, errors.New("no such metric")
		}

		asint := int64(val)
		m.Delta = &asint

		if ser.useCrypto {
			m.Hash = hex.EncodeToString(ser.crypto.CreateHash(m))
		}

		marshal, err := json.Marshal(m)
		if err != nil {
			log.Panic("service::ParseAndGet::error: can't marshal counter metric with:", err)
			return nil, err
		}

		return marshal, nil
	default:
		log.Println("service::ParseAndGet::info: unknown metrics type")

		return nil, errors.New("wrong metrics type")
	}
}

func (ser *service) GetKnownMetrics() []string {
	return ser.storage.GetKnownMetrics()
}

func (ser *service) IsDBConnected() bool {
	return ser.storage.IsDBConnected()
}

func (ser *service) ParseAndSaveSeveral(s []byte) error {
	log.Println("service::ParseAndSaveSeveral::info: started:", string(s))

	var mall []metrics.Metric
	err := json.Unmarshal(s, &mall)
	if err != nil {
		log.Println("service::ParseAndSaveSeveral::warning: can't unmarshal with error:", err)
		return errors.New("wrong query")
	}

	for _, m := range mall {
		metricType := m.MType
		metricName := m.ID

		log.Println("service::ParseAndSaveSeveral::info: type:", metricType, "|| name:", metricName)

		switch metricType {
		case gauge:
			value := m.Value
			if value == nil {
				log.Println("service::ParseAndSaveSeveral::info: gauge value is empty")
				continue
			}

			if !ser.crypto.CheckHash(m) {
				log.Println("service::ParseAndSaveSeveral::info: wrong hash")
				continue
			}

			err = ser.storage.SetGaugeMetrics(metricName, metrics.Gauge(*value))
			if err != nil {
				log.Println("service::ParseAndSave::error: problem in metrics saving:", err)
				continue
			}
		case counter:
			if m.Delta == nil {
				log.Println("service::ParseAndSaveSeveral::info: counter delta is empty")
				continue
			}

			value := *m.Delta

			if !ser.crypto.CheckHash(m) {
				log.Println("service::ParseAndSaveSeveral::info: wrong hash")
				continue
			}

			exVal, ok := ser.storage.GetCounterMetrics(metricName)
			if !ok {
				log.Println("service::ParseAndSaveSeveral::info: new counter metric")
				err = ser.storage.SetCounterMetrics(metricName, metrics.Counter(value))
				if err != nil {
					log.Println("service::ParseAndSave::error: problem in metrics saving:", err)
				}
				continue
			}
			log.Println("service::ParseAndSaveSeveral::info: update counter metric")
			err = ser.storage.SetCounterMetrics(metricName, metrics.Counter(int64(exVal)+value))
			if err != nil {
				log.Println("service::ParseAndSave::error: problem in metrics saving:", err)
			}
		default:
			log.Println("service::ParseAndSaveSeveral::info: unknown metrics type")
			continue
		}
	}
	return nil
}
