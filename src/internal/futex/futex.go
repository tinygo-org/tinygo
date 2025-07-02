package futex

// Cross platform futex implementation.
// Futexes are supported on all major operating systems and on WebAssembly.
//
// For more information, see: https://outerproduct.net/futex-dictionary.html

import (
	"sync/atomic"
	"unsafe"
)

// A futex is a way for userspace to wait with the pointer as the key, and for
// another thread to wake one or all waiting threads keyed on the same pointer.
//
// A futex does not change the underlying value, it only reads it before going
// to sleep (atomically) to prevent lost wake-ups.
type Futex struct {
	atomic.Uint32
}

// Atomically check for cmp to still be equal to the futex value and if so, go
// to sleep. Return true if we were definitely awoken by a call to Wake or
// WakeAll, and false if we can't be sure of that.
func (f *Futex) Wait(cmp uint32) bool {
	tinygo_futex_wait((*uint32)(unsafe.Pointer(&f.Uint32)), cmp)

	// We *could* detect a zero return value from the futex system call which
	// would indicate we got awoken by a Wake or WakeAll call. However, this is
	// what the manual page has to say:
	//
	// > Note that a wake-up can also be caused by common futex usage patterns
	// > in unrelated code that happened to have previously used the futex
	// > word's memory location (e.g., typical futex-based implementations of
	// > Pthreads mutexes can cause this under some conditions). Therefore,
	// > callers should always conservatively assume that a return value of 0
	// > can mean a spurious wake-up, and use the futex word's value (i.e., the
	// > user-space synchronization scheme) to decide whether to continue to
	// > block or not.
	//
	// I'm not sure whether we do anything like pthread does, so to be on the
	// safe side we say we don't know whether the wakeup was spurious or not and
	// return false.
	return false
}

// Like Wait, but times out after the number of nanoseconds in timeout.
func (f *Futex) WaitUntil(cmp uint32, timeout uint64) {
	tinygo_futex_wait_timeout((*uint32)(unsafe.Pointer(&f.Uint32)), cmp, timeout)
}

// Wake a single waiter.
func (f *Futex) Wake() {
	tinygo_futex_wake((*uint32)(unsafe.Pointer(&f.Uint32)))
}

// Wake all waiters.
func (f *Futex) WakeAll() {
	tinygo_futex_wake_all((*uint32)(unsafe.Pointer(&f.Uint32)))
}

//export tinygo_futex_wait
func tinygo_futex_wait(addr *uint32, cmp uint32)

//export tinygo_futex_wait_timeout
func tinygo_futex_wait_timeout(addr *uint32, cmp uint32, timeout uint64)

//export tinygo_futex_wake
func tinygo_futex_wake(addr *uint32)

//export tinygo_futex_wake_all
func tinygo_futex_wake_all(addr *uint32)
