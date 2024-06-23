// Package unique implements the upstream Go unique package for TinyGo.
//
// It is not a full implementation: while it should behave the same way, it
// doesn't free unreferenced uniqued objects.
package unique

import (
	"sync"
	"unsafe"
)

var (
	// We use a two-level map because that way it's easier to store and retrieve
	// values.
	globalMap map[unsafe.Pointer]any // map value type is always map[T]Handle[T]

	globalMapMutex sync.Mutex
)

// Unique handle for the given value. Comparing two handles is cheap.
type Handle[T comparable] struct {
	value *T
}

// Value returns a shallow copy of the T value that produced the Handle.
func (h Handle[T]) Value() T {
	return *h.value
}

// Make a new unqique handle for the given value.
func Make[T comparable](value T) Handle[T] {
	// Very simple implementation of the unique package. This is much, *much*
	// simpler than the upstream implementation. Sadly it's not possible to
	// reuse the upstream version because it relies on implementation details of
	// the upstream runtime.
	// It probably isn't as efficient as the upstream version, but the first
	// goal here is compatibility. If the performance is a problem, it can be
	// optimized later.

	globalMapMutex.Lock()

	// The map isn't initialized at program startup (and after a test run), so
	// create it.
	if globalMap == nil {
		globalMap = make(map[unsafe.Pointer]any)
	}

	// Retrieve the type-specific map, creating it if not yet present.
	typeptr, _ := decomposeInterface(value)
	var typeSpecificMap map[T]Handle[T]
	if typeSpecificMapValue, ok := globalMap[typeptr]; !ok {
		typeSpecificMap = make(map[T]Handle[T])
		globalMap[typeptr] = typeSpecificMap
	} else {
		typeSpecificMap = typeSpecificMapValue.(map[T]Handle[T])
	}

	// Retrieve the handle for the value, creating it if it isn't created yet.
	var handle Handle[T]
	if h, ok := typeSpecificMap[value]; !ok {
		var clone T = value
		handle.value = &clone
		typeSpecificMap[value] = handle
	} else {
		handle = h
	}

	globalMapMutex.Unlock()

	return handle
}

//go:linkname decomposeInterface runtime.decomposeInterface
func decomposeInterface(i interface{}) (unsafe.Pointer, unsafe.Pointer)
