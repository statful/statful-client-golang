package statful

type MetricsSender interface {
	Put(metric []*Metric) error
	Flush()
}
