package statful

import (
	"fmt"
	"strings"
)

// metric[,tag1=value][,tag2=value] value unix_timestamp [aggregation1][,aggregation2][,aggregation_frequency]
func MetricToString(name string, value float64, tags Tags, timestamp int64, aggregations Aggregations, frequency AggregationFrequency) string {
	var b strings.Builder

	// metric_name
	b.WriteString(name)
	// tags
	for tk, tv := range tags {
		fmt.Fprintf(&b, ",%v=%v", tk, tv)
	}
	// value and timestamp
	fmt.Fprintf(&b, " %f %d", value, timestamp)
	// aggregations
	if len(aggregations) > 0 {
		fmt.Fprintf(&b, " ")
		for agg, _ := range aggregations {
			fmt.Fprintf(&b, "%v,", agg)
		}
		// aggregation frequency
		fmt.Fprintf(&b, "%d", frequency)
	}

	return b.String()
}
