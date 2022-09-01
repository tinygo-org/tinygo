package pprof

// TinyGo does not implement pprof. However, a dummy shell is needed for the
// testing package (and testing/internal/pprof).

import (
	"errors"
	"io"
)

var ErrUnimplemented = errors.New("runtime/pprof: unimplemented")

type Profile struct{}

func StartCPUProfile(w io.Writer) error {
	return nil
}

func StopCPUProfile() {
}

func WriteHeapProfile(w io.Writer) error {
	return nil
}

func Lookup(name string) *Profile {
	return nil
}

func (p *Profile) Name() string {
	return ""
}

func (p *Profile) Count() int {
	return 0
}

func (p *Profile) WriteTo(w io.Writer, debug int) error {
	return ErrUnimplemented
}

func Profiles() []*Profile {
	return nil
}
