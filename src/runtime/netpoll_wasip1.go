//go:build wasip1 && (scheduler.tasks || scheduler.asyncify)

package runtime

import (
	"internal/task"
	"unsafe"
)

// pollMode identifies the I/O direction a goroutine is waiting on.
// Zero is intentionally invalid so an uninitialized pollDesc cannot
// silently look like a read waiter.
type pollMode uint8

const (
	pollRead  pollMode = 1
	pollWrite pollMode = 2
)

// pollDesc tracks one parked goroutine waiting for an FD to become ready.
// It is created by netpollAddWait, kept alive by activePolls, and freed
// (eventually GC'd) once unlinked.
type pollDesc struct {
	fd    uint32
	mode  pollMode
	fired bool // set by pollIO when the wait is satisfied; netpollDone uses this for idempotency
	task  *task.Task
	bnxt  *pollDesc // chain in activePolls
}

var (
	// activePolls is the singly-linked list of all currently-parked FD
	// waiters. wasip1 is single-threaded — every mutation happens from
	// the running goroutine or the scheduler loop, never both.
	activePolls *pollDesc
	pollCount   int

	// Scratch buffers for poll_oneoff. Grown on demand, never shrunk —
	// the working set settles on a stable max.
	pollSubs   []__wasi_subscription_t
	pollEvents []__wasi_event_t
)

// netpollAddWait registers the calling goroutine's interest in fd / mode
// and returns a descriptor identifying the wait. The caller must:
//
//  1. call task.Pause() to suspend until the FD is ready (or the task is
//     woken for some other reason — timer, manual scheduleTask), and
//  2. call netpollDone(pd) after Pause returns to deregister.
//
// Multiple waiters on the same (fd, mode) pair are supported; each gets
// its own pollDesc and its own subscription in the next poll_oneoff call.
func netpollAddWait(fd uint32, mode pollMode) *pollDesc {
	pd := &pollDesc{
		fd:   fd,
		mode: mode,
		task: task.Current(),
		bnxt: activePolls,
	}
	activePolls = pd
	pollCount++
	return pd
}

// netpollDone removes pd from activePolls if it is still registered.
// Idempotent — if pollIO has already woken the waiter, pd.fired is true
// and this is a no-op.
func netpollDone(pd *pollDesc) {
	if pd.fired {
		return
	}
	pp := &activePolls
	for *pp != nil {
		if *pp == pd {
			*pp = pd.bnxt
			pd.bnxt = nil
			pollCount--
			return
		}
		pp = &(*pp).bnxt
	}
}

// pollIO is the cooperative scheduler's blocking wait on wasip1. It
// invokes poll_oneoff with one subscription per pollDesc currently in
// activePolls, plus optionally a clock subscription.
//
//	timeoutNs >  0 : add a clock subscription with this nanosecond timeout.
//	timeoutNs == 0 : non-blocking poll, no clock sub. (Forward-looking; the
//	                 v1 scheduler does not invoke this path.)
//	timeoutNs <  0 : block until any FD is ready, no clock sub. Caller must
//	                 ensure pollCount > 0 — calling poll_oneoff with zero
//	                 subscriptions returns EINVAL.
//
// Tasks whose subscriptions fire are pushed onto runqueue. The caller
// (the scheduler) re-walks the sleep / timer queues on its next loop
// iteration to handle clock fires.
func pollIO(timeoutNs int64) {
	addClock := timeoutNs > 0
	nsubs := pollCount
	if addClock {
		nsubs++
	}
	if nsubs == 0 {
		// Caller is responsible for not invoking pollIO with nothing to
		// wait on; bail out rather than calling poll_oneoff with zero
		// subscriptions.
		return
	}
	if cap(pollSubs) < nsubs {
		pollSubs = make([]__wasi_subscription_t, nsubs)
		pollEvents = make([]__wasi_event_t, nsubs)
	} else {
		pollSubs = pollSubs[:nsubs]
		pollEvents = pollEvents[:nsubs]
	}

	i := 0
	for pd := activePolls; pd != nil; pd = pd.bnxt {
		var et __wasi_eventtype_t
		if pd.mode == pollRead {
			et = __wasi_eventtype_t_fd_read
		} else {
			et = __wasi_eventtype_t_fd_write
		}
		pollSubs[i].userData = uint64(uintptr(unsafe.Pointer(pd)))
		pollSubs[i].u.setFDReadWrite(et, pd.fd)
		i++
	}

	if addClock {
		pollSubs[i].userData = 0
		pollSubs[i].u.setClock(0, uint64(timeoutNs), timePrecisionNanoseconds, 0)
		i++
	}

	var nevents uint32
	poll_oneoff(&pollSubs[0], &pollEvents[0], uint32(nsubs), &nevents)

	for k := uint32(0); k < nevents; k++ {
		ev := &pollEvents[k]
		if ev.userData == 0 {
			continue
		}
		pd := (*pollDesc)(unsafe.Pointer(uintptr(ev.userData)))
		if pd.fired {
			continue
		}
		pd.fired = true
		pp := &activePolls
		for *pp != nil {
			if *pp == pd {
				*pp = pd.bnxt
				pd.bnxt = nil
				pollCount--
				break
			}
			pp = &(*pp).bnxt
		}
		runqueue.Push(pd.task)
	}
}

// runtime_netpoll_addwait is the linkname target used by package syscall
// (and any future package using //go:linkname into runtime) to register
// a wait on an FD without sharing the runtime's pollDesc / pollMode
// types. The returned uintptr is an opaque pollDesc pointer; callers
// must pass it back to runtime_netpoll_done.
//
// mode must be one of pollRead (1) or pollWrite (2).
//
//go:linkname runtime_netpoll_addwait
func runtime_netpoll_addwait(fd uint32, mode uint8) uintptr {
	return uintptr(unsafe.Pointer(netpollAddWait(fd, pollMode(mode))))
}

// runtime_netpoll_done is the linkname target used by package syscall to
// release a pollDesc previously returned by runtime_netpoll_addwait.
// Idempotent; safe to call whether or not pollIO has already woken the
// waiter.
//
//go:linkname runtime_netpoll_done
func runtime_netpoll_done(pd uintptr) {
	if pd == 0 {
		return
	}
	netpollDone((*pollDesc)(unsafe.Pointer(pd)))
}

// runtime_netpoll_pdfired reports whether the given pollDesc has already
// been woken (either by a poll_oneoff event or by a manual wake). Used
// by deadline-driven cancellation paths to avoid double-waking a task.
//
//go:linkname runtime_netpoll_pdfired
func runtime_netpoll_pdfired(pd uintptr) bool {
	if pd == 0 {
		return true
	}
	return (*pollDesc)(unsafe.Pointer(pd)).fired
}

// runtime_netpoll_wake wakes the task parked on pd from outside the
// poll_oneoff event loop — for example, from a deadline timer's
// callback. Idempotent: a second call (or a race with pollIO firing
// the same pd) is a no-op thanks to the pd.fired flag.
//
// wasip1 is single-threaded so we don't need atomic ops here.
//
//go:linkname runtime_netpoll_wake
func runtime_netpoll_wake(pd uintptr) {
	if pd == 0 {
		return
	}
	p := (*pollDesc)(unsafe.Pointer(pd))
	if p.fired {
		return
	}
	p.fired = true
	pp := &activePolls
	for *pp != nil {
		if *pp == p {
			*pp = p.bnxt
			p.bnxt = nil
			pollCount--
			break
		}
		pp = &(*pp).bnxt
	}
	runqueue.Push(p.task)
}
