package inmemorystorage

import (
	"encoding/json"
	"log"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/nivanov045/silver-octo-train/internal/metrics"
)

type InMemoryStorage struct {
	Metrics       metrics.Metrics
	storeInterval time.Duration
	storeFile     string
	restore       bool
	hasUpdates    bool
	syncSave      bool
	mu            sync.Mutex
}

func New(storeInterval time.Duration, storeFile string, restore bool) *InMemoryStorage {
	var res = &InMemoryStorage{
		Metrics: metrics.Metrics{
			GaugeMetrics:   map[string]metrics.Gauge{},
			CounterMetrics: map[string]metrics.Counter{},
		},
		storeInterval: storeInterval,
		storeFile:     storeFile,
		restore:       restore,
		hasUpdates:    false,
		syncSave:      false,
	}

	if restore {
		res.doRestore()
	}
	runtime.SetFinalizer(res, func(s *InMemoryStorage) {
		log.Println("InMemoryStorage::StorageFinalizer::info: started")
		s.doSave()
	})
	if res.storeInterval > 0*time.Second {
		go res.saveByTimer()
	} else {
		res.syncSave = true
	}
	return res
}

func (s *InMemoryStorage) doRestore() {
	err := s.restoreFromFile()
	if err != nil {
		log.Println("storage::doRestore error:", err)
	}
}

func (s *InMemoryStorage) restoreFromFile() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	file, err := os.OpenFile(s.storeFile, os.O_RDONLY|os.O_CREATE, 0777)
	if err != nil {
		log.Println("InMemoryStorage::restoreFromFile::warning: can't open file:", err)
		return err
	}
	defer file.Close()
	encoder := json.NewDecoder(file)
	err = encoder.Decode(&s.Metrics)
	if err != nil {
		log.Println("InMemoryStorage::restoreFromFile::warning: can't read from file:", err)
		return err
	}
	return nil
}

func (s *InMemoryStorage) SetCounterMetrics(name string, val metrics.Counter) error {
	log.Println("InMemoryStorage::SetCounterMetrics::info: started")
	s.Metrics.CounterMetrics[name] = val
	if s.syncSave {
		s.doSave()
	} else {
		s.hasUpdates = true
	}
	return nil
}

func (s *InMemoryStorage) GetCounterMetrics(name string) (metrics.Counter, bool) {
	if val, ok := s.Metrics.CounterMetrics[name]; ok {
		return val, true
	}
	return 0, false
}

func (s *InMemoryStorage) SetGaugeMetrics(name string, val metrics.Gauge) error {
	log.Println("InMemoryStorage::SetGaugeMetrics::info: started")
	s.Metrics.GaugeMetrics[name] = val
	if s.syncSave {
		s.doSave()
	} else {
		s.hasUpdates = true
	}
	return nil
}

func (s *InMemoryStorage) GetGaugeMetrics(name string) (metrics.Gauge, bool) {
	if val, ok := s.Metrics.GaugeMetrics[name]; ok {
		return val, true
	}
	return 0, false
}

func (s *InMemoryStorage) GetKnownMetrics() []string {
	var res []string
	for key := range s.Metrics.CounterMetrics {
		res = append(res, key)
	}
	for key := range s.Metrics.GaugeMetrics {
		res = append(res, key)
	}
	return res
}

func (s *InMemoryStorage) saveByTimer() {
	ticker := time.NewTicker(s.storeInterval)
	for {
		<-ticker.C
		log.Println("InMemoryStorage::saveByTimer::info: ticker")
		s.doSave()
	}
}

func (s *InMemoryStorage) doSave() {
	err := s.saveToFile()
	if err != nil {
		log.Println("InMemoryStorage::doSave::warning:", err)
	}
}

func (s *InMemoryStorage) saveToFile() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	log.Println("InMemoryStorage::saveToFile::info: started")
	if !s.syncSave && !s.hasUpdates {
		log.Println("InMemoryStorage::saveToFile::info: nothing to update")
		return nil
	}
	file, err := os.OpenFile(s.storeFile, os.O_WRONLY|os.O_CREATE, 0777)
	if err != nil {
		log.Println("InMemoryStorage::saveToFile::error: can't open file:", err)
		return err
	}
	defer file.Close()
	encoder := json.NewEncoder(file)
	err = encoder.Encode(&s.Metrics)
	if err != nil {
		log.Println("InMemoryStorage::saveToFile::error: can't write to file:", err)
		return err
	}
	s.hasUpdates = false
	return nil
}

func (s *InMemoryStorage) IsDBConnected() bool {
	return false
}
