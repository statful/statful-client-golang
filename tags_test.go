package statful

import (
	"strings"
	"testing"
)

func TestTags_Merge(t *testing.T) {
	scenarios := []struct {
		description string
		tags        Tags
		otherTags   Tags
		expected    Tags
	}{
		{
			description: "Empty tags merged with empty tags should give empty tags",
			tags:        Tags{},
			otherTags:   Tags{},
			expected:    Tags{},
		}, {
			description: "Empty tags merged with tags t1 should give tags t1",
			tags:        Tags{},
			otherTags:   Tags{},
			expected:    Tags{},
		}, {
			description: "tags t0 merged with empty tags should give tags t0",
			tags:        Tags{
				"foo": "bar",
				"bar": "baz",
			},
			otherTags: Tags{},
			expected:    Tags{
				"foo": "bar",
				"bar": "baz",
			},
		}, {
			description: "tags t0 merged with tags t1 should give tags union(t0+t1)",
			tags:        Tags{
				"foo": "bar",
			},
			otherTags:   Tags{
				"2": "abc",
				"3": "def",
			},
			expected:    Tags{
				"foo": "bar",
				"2": "abc",
				"3": "def",
			},
		},
	}

	for _, s := range scenarios {
		t.Run(s.description, func(t *testing.T) {

			tags := s.tags.Merge(s.otherTags)
			for e, _ := range s.expected {
				if _, ok := tags[e]; !ok {
					t.Errorf("%v: missing tag '%v'", s.description, e)
				}
			}
		})
	}
}

func TestTags_String(t *testing.T) {
	scenarios := []struct {
		description string
		tags        Tags
		expected    []string
	}{
		{
			description: "Empty tags stringed should give empty string",
			tags:        Tags{},
			expected:    []string{""},
		}, {
			description: "Single aggregation stringed should give single aggregation",
			tags:        Tags{"foo": "bar"},
			expected:    []string{"foo=bar"},
		}, {
			description: "Multiple tags stringed should give tags separated by ','",
			tags:        Tags{
				"foo": "bar",
				"bar": "baz",
			},
			expected:    []string{"foo=bar", "bar=baz"},
		}, {
			description: "Multiple tags stringed should give tags separated by ','",
			tags:        Tags{
				"foo": "bar",
				"2": "abc",
				"3": "def",
			},
			expected: []string{"foo=bar", "2=abc", "3=def"},
		},
	}

	for _, s := range scenarios {
		t.Run(s.description, func(t *testing.T) {
			sTags := s.tags.String()

			if len(s.expected) != len(strings.Split(sTags, ",")) {
				t.Errorf("%v: expected \"%v\" got \"%v\"", s.description, len(s.expected), len(s.tags))
			}

			for _, tag := range s.expected {
				if !strings.Contains(sTags, tag) {
					t.Errorf("%v: expected \"%v\" to be present in \"%v\"", s.description, tag, sTags)
				}
			}
		})
	}
}
