package statful

import (
	"bytes"
	"fmt"
)

type Tags map[string]string

func (t Tags) Merge(t2 Tags) Tags {
	merged := Tags{}
	for k, v := range t {
		if _, ok := merged[k]; !ok {
			merged[k] = v
		}
	}

	for k, v := range t2 {
		if _, ok := merged[k]; !ok {
			merged[k] = v
		}
	}
	return merged
}

func (t Tags) String() string {
	if len(t) == 0 {
		return ""
	}

	b := new (bytes.Buffer)
	b.Grow(len(t) + len(t)-1)
	first := true
	for tk, tv := range t {
		if !first {
			b.WriteString(",")
		}
		fmt.Fprintf(b, "%v=%v", tk, tv)
		first = false
	}

	return b.String()
}