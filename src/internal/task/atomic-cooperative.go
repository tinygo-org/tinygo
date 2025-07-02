//go:build tinygo.unicore

package task

// Atomics implementation for cooperative systems. The atomic types here aren't
// actually atomic, they assume that accesses cannot be interrupted by a
// different goroutine or interrupt happening at the same time.

type atomicIntegerType interface {
	uintptr | uint32 | uint64
}

type pseudoAtomic[T atomicIntegerType] struct {
	v T
}

func (x *pseudoAtomic[T]) Add(delta T) T { x.v += delta; return x.v }
func (x *pseudoAtomic[T]) Load() T       { return x.v }
func (x *pseudoAtomic[T]) Store(val T)   { x.v = val }
func (x *pseudoAtomic[T]) CompareAndSwap(old, new T) (swapped bool) {
	if x.v != old {
		return false
	}
	x.v = new
	return true
}
func (x *pseudoAtomic[T]) Swap(new T) (old T) {
	old = x.v
	x.v = new
	return
}

// Uintptr is an atomic uintptr when multithreading is enabled, and a plain old
// uintptr otherwise.
type Uintptr = pseudoAtomic[uintptr]

// Uint32 is an atomic uint32 when multithreading is enabled, and a plain old
// uint32 otherwise.
type Uint32 = pseudoAtomic[uint32]

// Uint64 is an atomic uint64 when multithreading is enabled, and a plain old
// uint64 otherwise.
type Uint64 = pseudoAtomic[uint64]
