package runtime

import (
	"internal/task"
	"sync/atomic"
	"unsafe"
)

const semTableSize = 64

var semTable [semTableSize]task.Futex

func semBucket(addr *uint32) *task.Futex {
	return &semTable[(uintptr(unsafe.Pointer(addr))>>3)%semTableSize]
}

//go:linkname semacquire internal/poll.runtime_Semacquire
func semacquire(sema *uint32) {
	bucket := semBucket(sema)
	for {
		v := atomic.LoadUint32(sema)
		if v > 0 {
			if atomic.CompareAndSwapUint32(sema, v, v-1) {
				return
			}
			continue
		}
		gen := bucket.Load()
		if atomic.LoadUint32(sema) != 0 {
			continue
		}
		bucket.Wait(gen)
	}
}

//go:linkname semrelease internal/poll.runtime_Semrelease
func semrelease(sema *uint32) {
	atomic.AddUint32(sema, 1)
	bucket := semBucket(sema)
	bucket.Add(1)
	bucket.WakeAll()
}
