//go:build !tinygo.unicore

package task

// Atomics implementation for non-cooperative systems (multithreaded, etc).
// These atomic types use real atomic instructions.

import "sync/atomic"

type (
	Uintptr = atomic.Uintptr
	Uint32  = atomic.Uint32
	Uint64  = atomic.Uint64
)
