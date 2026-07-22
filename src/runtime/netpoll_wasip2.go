//go:build wasip2 && (scheduler.tasks || scheduler.asyncify)

package runtime

import (
	"internal/cm"
	"internal/task"
	monotonicclock "internal/wasi/clocks/v0.2.0/monotonic-clock"
	"internal/wasi/io/v0.2.0/poll"
	"unsafe"
)

// pollMode is unused on wasip2 (each pollable encodes its own direction at
// subscribe time — InputStream.Subscribe vs OutputStream.Subscribe vs
// TcpSocket.Subscribe). Kept for API parity with the wasip1 netpoll
// callers that pass a mode constant.
type pollMode uint8

const (
	pollRead  pollMode = 1
	pollWrite pollMode = 2
)

// pollDesc tracks one parked goroutine waiting for a wasi pollable to
// become ready. It owns the pollable handle and is responsible for
// dropping it (either inside pollIO when the pollable fires, or via
// netpollDone for an unfired desc).
type pollDesc struct {
	pollable uint32 // poll.Pollable resource handle (cm.Resource = uint32)
	fired    bool
	task     *task.Task
	bnxt     *pollDesc
}

var (
	activePolls *pollDesc
	pollCount   int

	// Scratch buffers for the next Poll() call. Grown on demand, never
	// shrunk — the working set settles on a stable max.
	pollList     []poll.Pollable // pollables passed to Poll
	pollDescList []*pollDesc     // parallel scratch: same indices as pollList; nil entry == clock pollable
	pollResult   cm.List[uint32] // index list returned by Poll
)

// netpollAddPollable registers the calling goroutine's interest in a
// pollable and returns a descriptor identifying the wait. The caller
// transfers ownership of the pollable to the runtime: pollIO drops it
// when it fires; netpollDone drops it when the caller gives up.
//
// The caller must:
//
//  1. call task.Pause() to suspend until the pollable is ready (or the
//     task is woken for some other reason — timer, manual scheduleTask),
//     and
//  2. call netpollDone(pd) after Pause returns.
func netpollAddPollable(p uint32) *pollDesc {
	pd := &pollDesc{
		pollable: p,
		task:     task.Current(),
		bnxt:     activePolls,
	}
	activePolls = pd
	pollCount++
	return pd
}

// netpollDone deregisters pd. If the pollable hasn't fired yet, its
// resource is dropped. Idempotent.
func netpollDone(pd *pollDesc) {
	if pd.fired {
		// pollIO already dropped the pollable and unlinked the desc.
		return
	}
	pp := &activePolls
	for *pp != nil {
		if *pp == pd {
			*pp = pd.bnxt
			pd.bnxt = nil
			pollCount--
			pd.fired = true
			(poll.Pollable)(cm.Reinterpret[poll.Pollable](pd.pollable)).ResourceDrop()
			return
		}
		pp = &(*pp).bnxt
	}
}

// pollIO is the cooperative scheduler's blocking wait on wasip2. It
// invokes wasi:io/poll.Poll with one pollable per active pollDesc, plus
// optionally a clock pollable.
//
//	timeoutNs >  0 : subscribe a fresh monotonic-clock pollable with this
//	                 duration; include it in the Poll list.
//	timeoutNs == 0 : non-blocking poll — use Pollable.Ready() on each
//	                 registered pollable. Wake any that are ready;
//	                 return without calling Poll.
//	timeoutNs <  0 : block until any FD pollable fires. Caller must ensure
//	                 pollCount > 0 — Poll with zero pollables would
//	                 block forever with no way out.
func pollIO(timeoutNs int64) {
	addClock := timeoutNs > 0
	if !addClock && pollCount == 0 {
		// Non-blocking poll with nothing to check, or block-forever with
		// nothing to block on — caller should have caught this.
		return
	}

	if timeoutNs == 0 {
		// Non-blocking fast path: Ready() each registered pollable and
		// wake any that are already ready.
		pp := &activePolls
		for *pp != nil {
			pd := *pp
			pollable := cm.Reinterpret[poll.Pollable](pd.pollable)
			if pollable.Ready() {
				*pp = pd.bnxt
				pd.bnxt = nil
				pollCount--
				pd.fired = true
				pollable.ResourceDrop()
				runqueue.Push(pd.task)
				continue
			}
			pp = &(*pp).bnxt
		}
		return
	}

	// Build pollList in this order: [clock?, active pollables...]. The
	// pollDescList parallel slice maps each index back to its pd (nil for
	// the clock).
	n := pollCount
	if addClock {
		n++
	}
	if cap(pollList) < n {
		pollList = make([]poll.Pollable, n)
		pollDescList = make([]*pollDesc, n)
	} else {
		pollList = pollList[:n]
		pollDescList = pollDescList[:n]
	}

	i := 0
	var clockPollable poll.Pollable
	if addClock {
		clockPollable = monotonicclock.SubscribeDuration(monotonicclock.Duration(timeoutNs))
		pollList[i] = clockPollable
		pollDescList[i] = nil
		i++
	}
	for pd := activePolls; pd != nil; pd = pd.bnxt {
		pollList[i] = cm.Reinterpret[poll.Pollable](pd.pollable)
		pollDescList[i] = pd
		i++
	}

	pollResult = poll.Poll(cm.ToList(pollList))

	// Walk the returned indices. Any pd that fired is unlinked + woken;
	// its pollable is dropped. The clock pollable is always dropped at
	// the end of this call, fired or not.
	for _, idx := range pollResult.Slice() {
		if int(idx) >= len(pollDescList) {
			continue // defensive
		}
		pd := pollDescList[idx]
		if pd == nil {
			// Clock pollable fired — drop happens unconditionally below.
			continue
		}
		if pd.fired {
			continue
		}
		pd.fired = true
		// Unlink from activePolls.
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
		cm.Reinterpret[poll.Pollable](pd.pollable).ResourceDrop()
		runqueue.Push(pd.task)
	}
	if addClock {
		// Drop the clock pollable whether or not it fired — it was
		// freshly subscribed for this call only.
		clockPollable.ResourceDrop()
	}

	// Clear pointers in the scratch slices so we don't pin pollDescs in
	// memory between calls.
	for i := range pollDescList {
		pollDescList[i] = nil
	}
}

// runtime_netpoll_addpollable_wasip2 is the linkname target used by
// internal/poll and other stdlib callers that hold a poll.Pollable handle
// (as a raw uint32) and want to park the current goroutine until it
// becomes ready. Returns an opaque uintptr (the pollDesc pointer);
// pass it back to runtime_netpoll_done_wasip2.
//
//go:linkname runtime_netpoll_addpollable_wasip2
func runtime_netpoll_addpollable_wasip2(pollable uint32) uintptr {
	return uintptr(unsafe.Pointer(netpollAddPollable(pollable)))
}

// runtime_netpoll_done_wasip2 releases a pollDesc previously returned by
// runtime_netpoll_addpollable_wasip2. Idempotent.
//
//go:linkname runtime_netpoll_done_wasip2
func runtime_netpoll_done_wasip2(pd uintptr) {
	if pd == 0 {
		return
	}
	netpollDone((*pollDesc)(unsafe.Pointer(pd)))
}

// runtime_netpoll_pdfired_wasip2 reports whether the given pollDesc has
// already been woken. Used by deadline-driven cancellation paths to
// avoid double-waking a task.
//
//go:linkname runtime_netpoll_pdfired_wasip2
func runtime_netpoll_pdfired_wasip2(pd uintptr) bool {
	if pd == 0 {
		return true
	}
	return (*pollDesc)(unsafe.Pointer(pd)).fired
}

// runtime_netpoll_wake_wasip2 wakes the task parked on pd from outside
// the Poll event loop — for example, from a deadline timer's callback.
// Idempotent: a second call (or a race with pollIO firing the same pd)
// is a no-op thanks to the pd.fired flag.
//
// wasip2 is single-threaded so we don't need atomic ops here.
//
//go:linkname runtime_netpoll_wake_wasip2
func runtime_netpoll_wake_wasip2(pd uintptr) {
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
	cm.Reinterpret[poll.Pollable](p.pollable).ResourceDrop()
	runqueue.Push(p.task)
}
