package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"os"
	"runtime"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"github.com/nivanov045/silver-octo-train/internal/metrics"
)

type storage struct {
	Metrics       metrics.Metrics
	storeInterval time.Duration
	storeFile     string
	restore       bool
	hasUpdates    bool
	syncSave      bool
	mu            sync.Mutex
	databasePath  string
	db            *sql.DB
}

func New(storeInterval time.Duration, storeFile string, restore bool, databasePath string) *storage {
	var res = &storage{
		Metrics: metrics.Metrics{
			GaugeMetrics:   map[string]metrics.Gauge{},
			CounterMetrics: map[string]metrics.Counter{},
		},
		storeInterval: storeInterval,
		storeFile:     storeFile,
		restore:       restore,
		hasUpdates:    false,
		syncSave:      false,
		databasePath:  databasePath,
	}

	if len(databasePath) > 0 {
		var err error
		res.db, err = sql.Open("sqlite3",
			databasePath)
		if err != nil {
			log.Println("storage::New: error in db open:", err)
		}
	}

	if restore {
		res.doRestore()
	}
	runtime.SetFinalizer(res, StorageFinalizer)
	if res.storeInterval > 0*time.Second {
		go res.saveByTimer()
	} else {
		res.syncSave = true
	}
	return res
}

func (s *storage) doRestore() {
	if s.db != nil {
		s.restoreFromDB()
	} else {
		err := s.restoreFromFile()
		if err != nil {
			log.Println("storage::doRestore error:", err)
		}
	}
}

func (s *storage) restoreFromDB() {}

func (s *storage) restoreFromFile() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	file, err := os.OpenFile(s.storeFile, os.O_RDONLY|os.O_CREATE, 0777)
	if err != nil {
		log.Println("storage::restoreFromFile: can't open file:", err)
		return err
	}
	defer file.Close()
	encoder := json.NewDecoder(file)
	err = encoder.Decode(&s.Metrics)
	if err != nil {
		log.Println("storage::restoreFromFile: can't read from file:", err)
		return err
	}
	return nil
}

func (s *storage) SetCounterMetrics(name string, val metrics.Counter) {
	log.Println("storage::SetCounterMetrics: started")
	s.Metrics.CounterMetrics[name] = val
	if s.syncSave {
		s.doSave()
	} else {
		s.hasUpdates = true
	}
}

func (s *storage) GetCounterMetrics(name string) (metrics.Counter, bool) {
	if val, ok := s.Metrics.CounterMetrics[name]; ok {
		return val, true
	}
	return 0, false
}

func (s *storage) SetGaugeMetrics(name string, val metrics.Gauge) {
	log.Println("storage::SetGaugeMetrics: started")
	s.Metrics.GaugeMetrics[name] = val
	if s.syncSave {
		s.doSave()
	} else {
		s.hasUpdates = true
	}
}

func (s *storage) GetGaugeMetrics(name string) (metrics.Gauge, bool) {
	if val, ok := s.Metrics.GaugeMetrics[name]; ok {
		return val, true
	}
	return 0, false
}

func (s *storage) GetKnownMetrics() []string {
	var res []string
	for key := range s.Metrics.CounterMetrics {
		res = append(res, key)
	}
	for key := range s.Metrics.GaugeMetrics {
		res = append(res, key)
	}
	return res
}

func StorageFinalizer(s *storage) {
	if s.db != nil {
		defer s.db.Close()
	}
	log.Println("storage::StorageFinalizer: started")
	s.doSave()
}

func (s *storage) saveByTimer() {
	ticker := time.NewTicker(s.storeInterval)
	for {
		<-ticker.C
		log.Println("storage::saveByTimer: ticker")
		s.doSave()
	}
}

func (s *storage) doSave() {
	if s.db != nil {
		s.saveToDB()
	} else {
		err := s.saveToFile()
		if err != nil {
			log.Println("storage::doSave: error:", err)
		}
	}
}

func (s *storage) saveToDB() {}

func (s *storage) saveToFile() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	log.Println("storage::saveToFile: started")
	if !s.syncSave && !s.hasUpdates {
		log.Println("storage::saveToFile: nothing to update")
		return nil
	}
	file, err := os.OpenFile(s.storeFile, os.O_WRONLY|os.O_CREATE, 0777)
	if err != nil {
		log.Println("storage::saveToFile: can't open file:", err)
		return err
	}
	defer file.Close()
	encoder := json.NewEncoder(file)
	err = encoder.Encode(&s.Metrics)
	if err != nil {
		log.Println("storage::saveToFile: can't write to file:", err)
		return err
	}
	s.hasUpdates = false
	return nil
}

func (s *storage) IsDbConnected() bool {
	if s.db == nil {
		return false
	}
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	if err := s.db.PingContext(ctx); err != nil {
		return false
	}
	return true
}
