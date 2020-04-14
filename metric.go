package statful

import (
	"fmt"
	"strings"
)

type Metric struct {
	Name      string
	Value     float64
	Timestamp int64
	Tags      Tags
	Aggs      Aggregations
	Freq      AggregationFrequency
}

// metric[,tag1=value][,tag2=value] value unix_timestamp [aggregation1][,aggregation2][,aggregation_frequency]
func (m *Metric) String() string {
	var b strings.Builder

	// metric_name
	b.WriteString(m.Name)
	// tags
	for tk, tv := range m.Tags {
		fmt.Fprintf(&b, ",%v=%v", tk, tv)
	}
	// value and timestamp
	fmt.Fprintf(&b, " %f %d", m.Value, m.Timestamp)
	// aggregations
	if len(m.Aggs) > 0 {
		fmt.Fprintf(&b, " ")
		for agg, _ := range m.Aggs {
			fmt.Fprintf(&b, "%v,", agg)
		}
		// aggregation frequency
		fmt.Fprintf(&b, "%d", m.Freq)
	}

	fmt.Fprintln(&b, "")
	return b.String()
}
