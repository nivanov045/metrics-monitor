package storage

import "github.com/nivanov045/metrics-monitor/internal/metrics"

type InnerStorage interface {
	SetGaugeMetrics(name string, val metrics.Gauge) error
	GetGaugeMetrics(name string) (metrics.Gauge, bool)
	SetCounterMetrics(name string, val metrics.Counter) error
	GetCounterMetrics(name string) (metrics.Counter, bool)
	GetKnownMetrics() []string
	IsDBConnected() bool
}
