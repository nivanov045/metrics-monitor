package storage

import (
	_ "github.com/lib/pq"

	"github.com/nivanov045/silver-octo-train/cmd/server/config"
	"github.com/nivanov045/silver-octo-train/cmd/server/storage/dbstorage"
	"github.com/nivanov045/silver-octo-train/cmd/server/storage/inmemorystorage"
	"github.com/nivanov045/silver-octo-train/internal/metrics"
)

type InnerStorage interface {
	SetGaugeMetrics(name string, val metrics.Gauge) error
	GetGaugeMetrics(name string) (metrics.Gauge, bool)
	SetCounterMetrics(name string, val metrics.Counter) error
	GetCounterMetrics(name string) (metrics.Counter, bool)
	GetKnownMetrics() []string
	IsDBConnected() bool
}

type storage struct {
	innerStorage InnerStorage
}

func New(config config.Config) (*storage, error) {
	var res = &storage{}
	var err error
	if len(config.Database) > 0 {
		res.innerStorage, err = dbstorage.New(config.Database)
	} else {
		res.innerStorage = inmemorystorage.New(config.StoreInterval, config.StoreFile, config.Restore)
	}
	return res, err
}

func (s *storage) SetCounterMetrics(name string, val metrics.Counter) error {
	return s.innerStorage.SetCounterMetrics(name, val)
}

func (s *storage) GetCounterMetrics(name string) (metrics.Counter, bool) {
	return s.innerStorage.GetCounterMetrics(name)
}

func (s *storage) SetGaugeMetrics(name string, val metrics.Gauge) error {
	return s.innerStorage.SetGaugeMetrics(name, val)
}

func (s *storage) GetGaugeMetrics(name string) (metrics.Gauge, bool) {
	return s.innerStorage.GetGaugeMetrics(name)
}

func (s *storage) GetKnownMetrics() []string {
	return s.innerStorage.GetKnownMetrics()
}

func (s *storage) IsDBConnected() bool {
	return s.innerStorage.IsDBConnected()
}

func NewForcedInMemory(config config.Config) *storage {
	var res = &storage{}
	res.innerStorage = inmemorystorage.New(config.StoreInterval, config.StoreFile, config.Restore)
	return res
}
