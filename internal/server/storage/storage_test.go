package storage

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/nivanov045/silver-octo-train/internal/metrics"
	"github.com/nivanov045/silver-octo-train/internal/server/storage/inmemorystorage"
)

func Test_storage_SetGetCounterMetrics(t *testing.T) {
	type args struct {
		name string
		val  metrics.Counter
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "set and get",
			args: args{
				name: "someMetric",
				val:  123,
			},
		},
	}
	s := storage{
		innerStorage: &inmemorystorage.InMemoryStorage{Metrics: metrics.Metrics{
			GaugeMetrics:   map[string]metrics.Gauge{},
			CounterMetrics: map[string]metrics.Counter{},
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := s.SetCounterMetrics(tt.args.name, tt.args.val)
			assert.NoError(t, err)
			val, ok := s.GetCounterMetrics(tt.args.name)
			if !ok || tt.args.val != val {
				t.Errorf("storage.SetCounterMetrics() error with name %v, val %v", tt.args.name, tt.args.val)
			}
		})
	}
}

func Test_storage_SetGetGaugeMetrics(t *testing.T) {
	type args struct {
		name string
		val  metrics.Gauge
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "set and get",
			args: args{
				name: "someMetric",
				val:  123.456,
			},
		},
	}
	s := storage{
		innerStorage: &inmemorystorage.InMemoryStorage{Metrics: metrics.Metrics{
			GaugeMetrics:   map[string]metrics.Gauge{},
			CounterMetrics: map[string]metrics.Counter{},
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := s.SetGaugeMetrics(tt.args.name, tt.args.val)
			assert.NoError(t, err)
			val, ok := s.GetGaugeMetrics(tt.args.name)
			if !ok || tt.args.val != val {
				t.Errorf("storage.SetCounterMetrics() error with name %v, val %v", tt.args.name, tt.args.val)
			}
		})
	}
}

func Test_storage_GetKnownMetrics(t *testing.T) {
	tests := []struct {
		name string
		s    *storage
		want []string
	}{
		{
			name: "no metrics",
			s: &storage{
				innerStorage: &inmemorystorage.InMemoryStorage{Metrics: metrics.Metrics{
					GaugeMetrics:   map[string]metrics.Gauge{},
					CounterMetrics: map[string]metrics.Counter{},
				}},
			},
			want: []string(nil),
		},
		{
			name: "one metric",
			s: &storage{
				innerStorage: &inmemorystorage.InMemoryStorage{Metrics: metrics.Metrics{
					GaugeMetrics:   map[string]metrics.Gauge{"TestMetric": 0.0},
					CounterMetrics: map[string]metrics.Counter{},
				}},
			},
			want: []string{"TestMetric"},
		},
		{
			name: "several metrics",
			s: &storage{
				innerStorage: &inmemorystorage.InMemoryStorage{Metrics: metrics.Metrics{
					GaugeMetrics:   map[string]metrics.Gauge{"TestMetricG": 0.5},
					CounterMetrics: map[string]metrics.Counter{"TestMetricC": 10},
				}},
			},
			want: []string{"TestMetricC", "TestMetricG"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.s.GetKnownMetrics(), tt.want)
		})
	}
}
