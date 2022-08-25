package metrics

type Gauge float64
type Counter int64

type Metrics struct {
	GaugeMetrics   map[string]Gauge
	CounterMetrics map[string]Counter
}

func (m Metrics) Clone() Metrics {
	var res Metrics
	res.GaugeMetrics = make(map[string]Gauge)
	if m.GaugeMetrics != nil {
		for k, v := range m.GaugeMetrics {
			res.GaugeMetrics[k] = v
		}
	}
	res.CounterMetrics = make(map[string]Counter)
	if m.CounterMetrics != nil {
		for k, v := range m.CounterMetrics {
			res.CounterMetrics[k] = v
		}
	}
	return res
}

type Interface struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
	Hash  string   `json:"hash,omitempty"`  // значение хеш-функции
}

var KnownMetrics = [...]string{
	"Alloc",
	"BuckHashSys",
	"Frees",
	"GCCPUFraction",
	"GCSys",
	"HeapAlloc",
	"HeapIdle",
	"HeapInuse",
	"HeapObjects",
	"HeapReleased",
	"HeapSys",
	"LastGC",
	"Lookups",
	"MCacheInuse",
	"MCacheSys",
	"MSpanInuse",
	"MSpanSys",
	"Mallocs",
	"NextGC",
	"NumForcedGC",
	"NumGC",
	"OtherSys",
	"PauseTotalNs",
	"StackInuse",
	"StackSys",
	"Sys",
	"TotalAlloc",
	"RandomValue",
	"PollCount",
	"TotalMemory",
	"FreeMemory",
	"CPUutilization1",
}
