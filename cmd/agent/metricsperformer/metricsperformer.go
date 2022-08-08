package metricsperformer

import (
	"math/rand"
	"runtime"

	"github.com/nivanov045/silver-octo-train/internal/metrics"
)

type metricsPerformer struct{}

func New() *metricsPerformer {
	return &metricsPerformer{}
}

func (*metricsPerformer) UpdateMetrics(m metrics.Metrics) {
	var memStat runtime.MemStats
	runtime.ReadMemStats(&memStat)
	for _, val := range metrics.KnownMetrics {
		switch val {
		case "Alloc":
			m.GaugeMetrics["Alloc"] = metrics.Gauge(memStat.Alloc)
		case "BuckHashSys":
			m.GaugeMetrics["BuckHashSys"] = metrics.Gauge(memStat.BuckHashSys)
		case "Frees":
			m.GaugeMetrics["Frees"] = metrics.Gauge(memStat.Frees)
		case "GCCPUFraction":
			m.GaugeMetrics["GCCPUFraction"] = metrics.Gauge(memStat.GCCPUFraction)
		case "GCSys":
			m.GaugeMetrics["GCSys"] = metrics.Gauge(memStat.GCSys)
		case "HeapAlloc":
			m.GaugeMetrics["HeapAlloc"] = metrics.Gauge(memStat.HeapAlloc)
		case "HeapIdle":
			m.GaugeMetrics["HeapIdle"] = metrics.Gauge(memStat.HeapIdle)
		case "HeapInuse":
			m.GaugeMetrics["HeapInuse"] = metrics.Gauge(memStat.HeapInuse)
		case "HeapObjects":
			m.GaugeMetrics["HeapObjects"] = metrics.Gauge(memStat.HeapObjects)
		case "HeapReleased":
			m.GaugeMetrics["HeapReleased"] = metrics.Gauge(memStat.HeapReleased)
		case "HeapSys":
			m.GaugeMetrics["HeapSys"] = metrics.Gauge(memStat.HeapSys)
		case "LastGC":
			m.GaugeMetrics["LastGC"] = metrics.Gauge(memStat.LastGC)
		case "Lookups":
			m.GaugeMetrics["Lookups"] = metrics.Gauge(memStat.Lookups)
		case "MCacheInuse":
			m.GaugeMetrics["MCacheInuse"] = metrics.Gauge(memStat.MCacheInuse)
		case "MCacheSys":
			m.GaugeMetrics["MCacheSys"] = metrics.Gauge(memStat.MCacheSys)
		case "MSpanInuse":
			m.GaugeMetrics["MSpanInuse"] = metrics.Gauge(memStat.MSpanInuse)
		case "MSpanSys":
			m.GaugeMetrics["MSpanSys"] = metrics.Gauge(memStat.MSpanSys)
		case "Mallocs":
			m.GaugeMetrics["Mallocs"] = metrics.Gauge(memStat.Mallocs)
		case "NextGC":
			m.GaugeMetrics["NextGC"] = metrics.Gauge(memStat.NextGC)
		case "NumForcedGC":
			m.GaugeMetrics["NumForcedGC"] = metrics.Gauge(memStat.NumForcedGC)
		case "NumGC":
			m.GaugeMetrics["NumGC"] = metrics.Gauge(memStat.NumGC)
		case "OtherSys":
			m.GaugeMetrics["OtherSys"] = metrics.Gauge(memStat.OtherSys)
		case "PauseTotalNs":
			m.GaugeMetrics["PauseTotalNs"] = metrics.Gauge(memStat.PauseTotalNs)
		case "StackInuse":
			m.GaugeMetrics["StackInuse"] = metrics.Gauge(memStat.StackInuse)
		case "StackSys":
			m.GaugeMetrics["StackSys"] = metrics.Gauge(memStat.StackSys)
		case "Sys":
			m.GaugeMetrics["Sys"] = metrics.Gauge(memStat.Sys)
		case "TotalAlloc":
			m.GaugeMetrics["TotalAlloc"] = metrics.Gauge(memStat.TotalAlloc)
		case "RandomValue":
			m.GaugeMetrics["RandomValue"] = metrics.Gauge(rand.Float64())
		case "PollCount":
			m.CounterMetrics["PollCount"]++
		}
	}
}
