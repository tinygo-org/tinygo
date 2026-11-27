//go:build !tinygo.memprofile

package pprof

import (
	"errors"
	"io"
)

// Heap profiling is off by default, so programs that do not ask for it pay
// nothing for it in the allocator.
var errNoMemProfile = errors.New("runtime/pprof: heap profiles need -tags=tinygo.memprofile")

// WriteHeapProfile is shorthand for Lookup("heap").WriteTo(w, 0).
func WriteHeapProfile(w io.Writer) error {
	return errNoMemProfile
}

func writeHeap(w io.Writer, debug int, defaultSampleType string) error {
	return errNoMemProfile
}
