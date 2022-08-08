package metricsperformer

import (
	"testing"

	"github.com/nivanov045/silver-octo-train/internal/metrics"
	"github.com/stretchr/testify/assert"
)

func Test_metricsPerformer_UpdateMetrics(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "update",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := metrics.Metrics{
				GaugeMetrics:   map[string]metrics.Gauge{},
				CounterMetrics: map[string]metrics.Counter{},
			}
			mp := metricsPerformer{}
			assert.Equal(t, len(m.GaugeMetrics), 0)
			assert.Equal(t, len(m.CounterMetrics), 0)
			mp.UpdateMetrics(m)
			assert.Equal(t, len(m.GaugeMetrics), 28)
			assert.Equal(t, len(m.CounterMetrics), 1)
			assert.Equal(t, m.CounterMetrics["PollCount"], metrics.Counter(1))
			mp.UpdateMetrics(m)
			assert.Equal(t, m.CounterMetrics["PollCount"], metrics.Counter(2))
			mp.UpdateMetrics(m)
			assert.Equal(t, m.CounterMetrics["PollCount"], metrics.Counter(3))
		})
	}
}
