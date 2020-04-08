package statful

import (
	"strings"
	"testing"
)

func TestAggregations_Add(t *testing.T) {
	scenarios := []struct {
		description string
		aggs        Aggregations
		agg         Aggregation
		expected    Aggregations
	}{
		{
			description: "tags{} + agg = tags{agg}",
			aggs:        Aggregations{},
			agg:         AggAvg,
			expected:    Aggregations{AggAvg: nothing},
		}, {
			description: "tags{agg0} + agg0 = tags{agg0}",
			aggs:        Aggregations{AggP90: nothing},
			agg:         AggP90,
			expected:    Aggregations{AggP90: nothing},
		}, {
			description: "tags{agg0} + agg1 = tags{agg0, agg1}",
			aggs:        Aggregations{AggCount: nothing},
			agg:         AggSum,
			expected:    Aggregations{AggCount: nothing, AggSum: nothing},
		}, {
			description: "tags{agg0, agg1} + agg2 = tags{agg0, agg1, agg2}",
			aggs:        Aggregations{AggFirst: nothing, AggLast: nothing},
			agg:         AggP95,
			expected:    Aggregations{AggFirst: nothing, AggLast: nothing, AggP95: nothing},
		}, {
			description: "tags{agg0, agg1, agg2} + agg1 = tags{agg0, agg1, agg2}",
			aggs:        Aggregations{AggMax: nothing, AggMin: nothing, AggP99: nothing},
			agg:         AggMin,
			expected:    Aggregations{AggMax: nothing, AggMin: nothing, AggP99: nothing},
		},
	}

	for _, s := range scenarios {
		t.Run(s.description, func(t *testing.T) {
			aggs := s.aggs.Add(s.agg)
			for e, _ := range s.expected {
				if _, ok := aggs[e]; !ok {
					t.Errorf("%v: missing aggregation '%v'", s.description, e)
				}
			}
		})
	}
}

func TestAggregations_Merge(t *testing.T) {
	scenarios := []struct {
		description string
		aggs        Aggregations
		otherAggs   Aggregations
		expected    Aggregations
	}{
		{
			description: "Empty aggregations merged with empty aggregations should give empty aggregations",
			aggs:        Aggregations{},
			otherAggs:   Aggregations{},
			expected:    Aggregations{},
		}, {
			description: "Empty aggregations merged with aggregations a1 should give aggregations a1",
			aggs:        Aggregations{},
			otherAggs:   Aggregations{AggAvg: nothing, AggCount: nothing, AggP90: nothing},
			expected:    Aggregations{AggAvg: nothing, AggCount: nothing, AggP90: nothing},
		}, {
			description: "aggregations a0 merged with empty aggregations should give aggregations a0",
			aggs:        Aggregations{AggCount: nothing, AggSum: nothing},
			otherAggs:   Aggregations{},
			expected:    Aggregations{AggCount: nothing, AggSum: nothing},
		}, {
			description: "aggregations a0 merged with aggregations a1 should give aggregations union(a0+a1)",
			aggs:        Aggregations{AggLast: nothing},
			otherAggs:   Aggregations{AggCount: nothing, AggSum: nothing},
			expected:    Aggregations{AggCount: nothing, AggSum: nothing, AggLast: nothing},
		},
	}

	for _, s := range scenarios {
		t.Run(s.description, func(t *testing.T) {
			aggs := s.aggs.Merge(s.otherAggs)
			for e, _ := range s.expected {
				if _, ok := aggs[e]; !ok {
					t.Errorf("%v: missing aggregation '%v'", s.description, e)
				}
			}
		})
	}
}

func TestAggregations_String(t *testing.T) {
	scenarios := []struct {
		description string
		aggs        Aggregations
		expected    []string
	}{
		{
			description: "Empty aggregations stringed should give empty string",
			aggs:        Aggregations{},
			expected:    []string{""},
		}, {
			description: "Single aggregation stringed should give single aggregation",
			aggs:        Aggregations{AggAvg: nothing},
			expected:    []string{AggAvg},
		}, {
			description: "Multiple aggregations stringed should give aggregations separated by ','",
			aggs:        Aggregations{AggCount: nothing, AggSum: nothing},
			expected:    []string{AggCount, AggSum},
		}, {
			description: "Multiple aggregations stringed should give aggregations separated by ','",
			aggs:        Aggregations{AggAvg: nothing, AggCount: nothing, AggP90: nothing},
			expected:    []string{AggAvg, AggCount, AggP90},
		},
	}

	for _, s := range scenarios {
		t.Run(s.description, func(t *testing.T) {
			sAggs := s.aggs.String()

			if len(s.expected) != len(strings.Split(sAggs, ",")) {
				t.Errorf("%v: expected \"%v\" got \"%v\"", s.description, len(s.expected), len(sAggs))
			}

			for _, agg := range s.expected {
				if !strings.Contains(sAggs, agg) {
					t.Errorf("%v: expected \"%v\" to be present in \"%v\"", s.description, agg, sAggs)
				}
			}
		})
	}
}
