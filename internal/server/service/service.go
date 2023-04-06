package service

import (
	"encoding/hex"
	"encoding/json"
	"errors"

	"github.com/rs/zerolog/log"

	"github.com/nivanov045/metrics-monitor/internal/metrics"
	"github.com/nivanov045/metrics-monitor/internal/server/crypto"
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
	log.Debug().Interface("data", string(s)).Msg("started parse and save:")

	var m metrics.Metric
	err := json.Unmarshal(s, &m)
	if err != nil {
		log.Error().Err(err).Stack()
		return errors.New("wrong query")
	}

	metricType := m.MType
	metricName := m.ID

	log.Debug().Interface("metricType", metricType).Interface("metricName", metricName)

	switch metricType {
	case gauge:
		value := m.Value
		if value == nil {
			log.Error().Msg("gauge value is empty")
			return errors.New("wrong query")
		}

		if !ser.crypto.CheckHash(m) {
			log.Error().Msg("wrong hash")
			return errors.New("wrong hash")
		}

		err = ser.storage.SetGaugeMetrics(metricName, metrics.Gauge(*value))
		if err != nil {
			log.Error().Err(err).Stack()
			return errors.New("problem in metrics saving")
		}
	case counter:
		if m.Delta == nil {
			log.Error().Msg("counter delta is empty")
			return errors.New("wrong query")
		}

		value := *m.Delta

		if !ser.crypto.CheckHash(m) {
			log.Error().Msg("wrong hash")
			return errors.New("wrong hash")
		}

		exVal, isKnown := ser.storage.GetCounterMetrics(metricName)
		if !isKnown {
			log.Debug().Msg("new counter metric")

			err = ser.storage.SetCounterMetrics(metricName, metrics.Counter(value))
			if err != nil {
				log.Error().Err(err).Stack()
				return errors.New("problem in metrics saving")
			}
			return nil
		}

		log.Debug().Msg("update counter metric")

		err = ser.storage.SetCounterMetrics(metricName, metrics.Counter(int64(exVal)+value))
		if err != nil {
			log.Error().Err(err).Stack()
			return errors.New("problem in metrics saving")
		}
	default:
		log.Error().Msg("unknown metrics type")
		return errors.New("wrong metrics type")
	}

	return nil
}

func (ser *service) ParseAndGet(s []byte) ([]byte, error) {
	log.Debug().Interface("data", string(s)).Msg("ParseAndGet started")

	var m metrics.Metric
	err := json.Unmarshal(s, &m)
	if err != nil {
		log.Error().Err(err).Stack()
		return nil, errors.New("wrong query")
	}

	metricType := m.MType
	metricName := m.ID

	log.Debug().Interface("metricType", metricType).Interface("metricName", metricName)

	switch metricType {
	case gauge:
		val, ok := ser.storage.GetGaugeMetrics(metricName)
		if !ok {
			log.Error().Msg("service::ParseAndGet::info: no such gauge metrics")
			return nil, errors.New("no such metric")
		}

		asFloat := float64(val)
		m.Value = &asFloat

		if ser.useCrypto {
			m.Hash = hex.EncodeToString(ser.crypto.CreateHash(m))
		}

		marshal, err := json.Marshal(m)
		if err != nil {
			log.Error().Err(err).Stack()
			return nil, err
		}

		return marshal, nil
	case counter:
		val, ok := ser.storage.GetCounterMetrics(metricName)
		if !ok {
			log.Error().Msg("no such counter metrics")
			return nil, errors.New("no such metric")
		}

		asint := int64(val)
		m.Delta = &asint

		if ser.useCrypto {
			m.Hash = hex.EncodeToString(ser.crypto.CreateHash(m))
		}

		marshal, err := json.Marshal(m)
		if err != nil {
			log.Error().Err(err).Stack()
			return nil, err
		}

		return marshal, nil
	default:
		log.Error().Msg("unknown metrics type")

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
	log.Debug().Interface("data", string(s)).Msg("ParseAndSaveSeveral started")

	var mall []metrics.Metric
	err := json.Unmarshal(s, &mall)
	if err != nil {
		log.Error().Err(err).Stack()
		return errors.New("wrong query")
	}

	for _, m := range mall {
		metricType := m.MType
		metricName := m.ID

		log.Debug().Interface("metricType", metricType).Interface("metricName", metricName)

		switch metricType {
		case gauge:
			value := m.Value
			if value == nil {
				log.Error().Msg("gauge value is empty")
				continue
			}

			if !ser.crypto.CheckHash(m) {
				log.Error().Msg("wrong hash")
				continue
			}

			err = ser.storage.SetGaugeMetrics(metricName, metrics.Gauge(*value))
			if err != nil {
				log.Error().Err(err).Stack()
				continue
			}
		case counter:
			if m.Delta == nil {
				log.Error().Msg("counter delta is empty")
				continue
			}

			value := *m.Delta

			if !ser.crypto.CheckHash(m) {
				log.Error().Msg("wrong hash")
				continue
			}

			exVal, ok := ser.storage.GetCounterMetrics(metricName)
			if !ok {
				log.Debug().Msg("new counter metric")
				err = ser.storage.SetCounterMetrics(metricName, metrics.Counter(value))
				if err != nil {
					log.Error().Err(err).Stack()
				}
				continue
			}
			log.Debug().Msg("update counter metric")
			err = ser.storage.SetCounterMetrics(metricName, metrics.Counter(int64(exVal)+value))
			if err != nil {
				log.Error().Err(err).Stack()
			}
		default:
			log.Error().Msg("unknown metrics type")
			continue
		}
	}
	return nil
}
