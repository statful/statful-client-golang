package statful

import (
	"fmt"
	"strings"
)

type FlushErr struct {
	errors []error
}

const (
	flushErrors    = "flush errors"
	flushErrorsSep = "; "
)

func (f FlushErr) Error() string {
	var errStrs []string
	for _, err := range f.errors {
		errStrs = append(errStrs, err.Error())
	}
	return fmt.Sprintf("%s: %s", flushErrors, strings.Join(errStrs, flushErrorsSep))
}

func (f FlushErr) appendErr(err error) FlushErr {
	f.errors = append(f.errors, err)
	return f
}

func (f FlushErr) hasErrors() bool {
	return len(f.errors) > 0
}
