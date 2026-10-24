package runtime

import (
	"internal/futex"
	"sync/atomic"
	"unsafe"
)

// This file contains semaphore implementations for internal/poll.

//go:linkname semacquire internal/poll.runtime_Semacquire
func semacquire(sema *uint32) {
	for {
		v := atomic.LoadUint32(sema)
		if v == 0 {
			if !hasScheduler {
				(*futex.Futex)(unsafe.Pointer(sema)).Wait(0)
			} else {
				Gosched()
			}
			continue
		}
		if atomic.CompareAndSwapUint32(sema, v, v-1) {
			return
		}
	}
}

//go:linkname semrelease internal/poll.runtime_Semrelease
func semrelease(sema *uint32) {
	atomic.AddUint32(sema, 1)
	if !hasScheduler {
		(*futex.Futex)(unsafe.Pointer(sema)).Wake()
	}
}
