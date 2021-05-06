package statful

import (
	"fmt"
	"strings"
)

// metric[,tag1=value][,tag2=value] value unix_timestamp [aggregation1][,aggregation2][,aggregation_frequency]
// metric[,tag1=value][,tag2=value] value,user unix_timestamp [aggregation1][,aggregation2][,aggregation_frequency]
func MetricToString(name string, value float64, user string, tags Tags, timestamp int64, aggregations Aggregations, frequency AggregationFrequency) string {
	var b strings.Builder

	// metric_name
	b.WriteString(name)
	// tags
	for tk, tv := range tags {
		fmt.Fprintf(&b, ",%v=%v", tk, tv)
	}

	if user == "" {
		// value and timestamp
		fmt.Fprintf(&b, " %f %d", value, timestamp)
	} else {
		fmt.Fprintf(&b, " %f,%s %d", value, user, timestamp)
	}
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
