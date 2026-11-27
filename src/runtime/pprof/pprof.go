package pprof

// TinyGo implements the heap and allocs profiles, and CPU profiling on macOS.
// The other profiles are missing.

import (
	"errors"
	"io"
	"runtime"
)

var ErrUnimplemented = errors.New("runtime/pprof: unimplemented")

// A Profile is a collection of stack traces.
type Profile struct {
	name  string
	count func() int
	write func(io.Writer, int) error
}

var heapProfile = &Profile{
	name:  "heap",
	count: countHeap,
	write: func(w io.Writer, debug int) error {
		return writeHeap(w, debug, "")
	},
}

var allocsProfile = &Profile{
	name:  "allocs",
	count: countHeap,
	write: func(w io.Writer, debug int) error {
		return writeHeap(w, debug, "alloc_space")
	},
}

func countHeap() int {
	n, _ := runtime.MemProfile(nil, true)
	return n
}

// Lookup returns the profile with the given name, or nil if there is no such
// profile.
func Lookup(name string) *Profile {
	switch name {
	case "heap":
		return heapProfile
	case "allocs":
		return allocsProfile
	}
	return nil
}

// Name returns the profile's name.
func (p *Profile) Name() string {
	if p == nil {
		return ""
	}
	return p.name
}

// Count returns the number of records in the profile.
func (p *Profile) Count() int {
	if p == nil {
		return 0
	}
	return p.count()
}

// WriteTo writes the profile to w. Only debug=0, the gzipped protocol buffer
// format, is supported.
func (p *Profile) WriteTo(w io.Writer, debug int) error {
	if p == nil {
		return ErrUnimplemented
	}
	return p.write(w, debug)
}

// Profiles returns the profiles TinyGo implements.
func Profiles() []*Profile {
	return []*Profile{allocsProfile, heapProfile}
}
