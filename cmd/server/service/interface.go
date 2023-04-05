package service

import "github.com/nivanov045/silver-octo-train/internal/metrics"

type Storage interface {
	GetCounterMetrics(name string) (metrics.Counter, bool)
	GetGaugeMetrics(name string) (metrics.Gauge, bool)
	GetKnownMetrics() []string
	IsDBConnected() bool
	SetCounterMetrics(name string, val metrics.Counter) error
	SetGaugeMetrics(name string, val metrics.Gauge) error
}

type Crypto interface {
	CheckHash(m metrics.Metric) bool
	CreateHash(m metrics.Metric) []byte
}
