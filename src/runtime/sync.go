package runtime

import (
	"internal/task"
	"sync/atomic"
	"unsafe"
)

const semTabSize = 16

var semtable [semTabSize]semaRoot

type semaWaiter struct {
	sema *uint32
	task *task.Task
	next *semaWaiter
}

type semaRoot struct {
	lock  task.PMutex
	head  *semaWaiter
	tail  *semaWaiter
	nwait atomic.Uint32
}

func semroot(sema *uint32) *semaRoot {
	return &semtable[(uintptr(unsafe.Pointer(sema))>>2)%semTabSize]
}

func canSemacquire(sema *uint32) bool {
	for {
		v := atomic.LoadUint32(sema)
		if v == 0 {
			return false
		}
		if atomic.CompareAndSwapUint32(sema, v, v-1) {
			return true
		}
	}
}

func (r *semaRoot) enqueue(w *semaWaiter) {
	w.next = nil
	if r.tail == nil {
		r.head = w
	} else {
		r.tail.next = w
	}
	r.tail = w
}

func (r *semaRoot) dequeue(sema *uint32) *semaWaiter {
	var prev *semaWaiter
	for w := r.head; w != nil; w = w.next {
		if w.sema == sema {
			if prev == nil {
				r.head = w.next
			} else {
				prev.next = w.next
			}
			if w.next == nil {
				r.tail = prev
			}
			w.next = nil
			return w
		}
		prev = w
	}
	return nil
}

//go:linkname semacquire internal/poll.runtime_Semacquire
func semacquire(sema *uint32) {
	if canSemacquire(sema) {
		return
	}
	if !hasScheduler && !hasParallelism {
		deadlock()
		return
	}
	root := semroot(sema)
	w := semaWaiter{
		sema: sema,
		task: task.Current(),
	}
	for {
		root.lock.Lock()
		root.nwait.Add(1)
		if canSemacquire(sema) {
			root.nwait.Add(^uint32(0))
			root.lock.Unlock()
			return
		}
		root.enqueue(&w)
		root.lock.Unlock()
		task.Pause()
		if canSemacquire(sema) {
			return
		}
	}
}

//go:linkname semrelease internal/poll.runtime_Semrelease
func semrelease(sema *uint32) {
	atomic.AddUint32(sema, 1)
	if !hasScheduler && !hasParallelism {
		return
	}
	root := semroot(sema)
	if root.nwait.Load() == 0 {
		return
	}
	root.lock.Lock()
	if root.nwait.Load() == 0 {
		root.lock.Unlock()
		return
	}
	var t *task.Task
	if w := root.dequeue(sema); w != nil {
		root.nwait.Add(^uint32(0))
		t = w.task
	}
	root.lock.Unlock()
	if t != nil {
		scheduleTask(t)
	}
}
