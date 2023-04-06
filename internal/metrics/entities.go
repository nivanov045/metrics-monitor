package metrics

type Gauge float64
type Counter int64

type Metrics struct {
	GaugeMetrics   map[string]Gauge
	CounterMetrics map[string]Counter
}

type Metric struct {
	ID    string   `json:"id"`              // name of metrics
	MType string   `json:"type"`            // gauge or counter
	Delta *int64   `json:"delta,omitempty"` // value, if counter
	Value *float64 `json:"value,omitempty"` // vlaue, if gauge
	Hash  string   `json:"hash,omitempty"`  // value of hash
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
