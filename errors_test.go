package statful

import (
	"errors"
	"fmt"
	"testing"
)

func TestFlushErr_Error(t *testing.T) {
	var flushErr FlushErr

	flushErr = flushErr.appendErr(errors.New("err 1"))
	flushErr = flushErr.appendErr(errors.New("err 2"))

	if !flushErr.hasErrors() {
		t.Error("hasError() returned: false, expected: true")
	}

	expectedErrStr := fmt.Sprintf(
		"%s: %s",
		flushErrors,
		fmt.Sprintf("%s%s%s", "err 1", flushErrorsSep, "err 2"))
	flushErrStr := flushErr.Error()
	if flushErrStr != expectedErrStr {
		t.Errorf("Error() returned: %s, expected: %s", flushErrStr, expectedErrStr)
	}
}
