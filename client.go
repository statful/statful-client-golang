package statful

import (
	"io"
)

type Client interface {
	Put(data io.Reader) error
	PutAggregated(data io.Reader, agg Aggregation, frequency AggregationFrequency) error
}
