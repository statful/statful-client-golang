package statful

import (
	"bytes"
)

type void struct{}
var nothing void
type Aggregation string
type Aggregations map[Aggregation]void
type AggregationFrequency int

const (
	AggAvg = "avg"
	AggSum = "sum"
	AggCount = "count"
	AggFirst = "first"
	AggLast = "last"
	AggP90 = "p90"
	AggP95 = "p95"
	AggP99 = "p99"
	AggMin = "min"
	AggMax = "max"

	Freq10s = 10
	Freq30s = 30
	Freq60s = 60
	Freq120s = 120
	Freq180s = 180
	Freq300s = 300
)

func (a Aggregations) Add(agg Aggregation) Aggregations {
	a[agg] = nothing
	return a
}

func (a Aggregations) Merge(aggs Aggregations) Aggregations {
	for agg, _ := range aggs {
		a[agg] = nothing
	}

	return a
}

func (a Aggregations) String() string {
	if len(a) == 0 {
		return ""
	}

	sep := ","
	total := (len(a)-1) * len(sep)
	for agg, _ := range a {
		total += len(agg)
	}
	b := new(bytes.Buffer)

	b.Grow(total)

	first := true
	for agg, _ := range a {
		if !first {
			b.WriteString(sep)
		}
		b.WriteString(string(agg))
		first = false
	}

	return b.String()
}
