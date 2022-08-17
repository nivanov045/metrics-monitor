package storage

import (
	"time"

	_ "github.com/lib/pq"

	"github.com/nivanov045/silver-octo-train/cmd/server/storage/dbstorage"
	"github.com/nivanov045/silver-octo-train/cmd/server/storage/inmemorystorage"
	"github.com/nivanov045/silver-octo-train/internal/metrics"
)

type InnerStorage interface {
	SetGaugeMetrics(name string, val metrics.Gauge)
	GetGaugeMetrics(name string) (metrics.Gauge, bool)
	SetCounterMetrics(name string, val metrics.Counter)
	GetCounterMetrics(name string) (metrics.Counter, bool)
	GetKnownMetrics() []string
	IsDBConnected() bool
}

type storage struct {
	is InnerStorage
}

func New(storeInterval time.Duration, storeFile string, restore bool, databasePath string) *storage {
	var res = &storage{}
	if len(databasePath) > 0 {
		res.is = dbstorage.New(databasePath, restore)
	} else {
		res.is = inmemorystorage.New(storeInterval, storeFile, restore)
	}
	return res
}

func (s *storage) SetCounterMetrics(name string, val metrics.Counter) {
	s.is.SetCounterMetrics(name, val)
}

func (s *storage) GetCounterMetrics(name string) (metrics.Counter, bool) {
	return s.is.GetCounterMetrics(name)
}

func (s *storage) SetGaugeMetrics(name string, val metrics.Gauge) {
	s.is.SetGaugeMetrics(name, val)
}

func (s *storage) GetGaugeMetrics(name string) (metrics.Gauge, bool) {
	return s.is.GetGaugeMetrics(name)
}

func (s *storage) GetKnownMetrics() []string {
	return s.is.GetKnownMetrics()
}

func (s *storage) IsDBConnected() bool {
	return s.is.IsDBConnected()
}
