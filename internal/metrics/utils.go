package metrics

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
