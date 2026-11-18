//go:build linux && amd64 && !baremetal && !tinygo.wasm && !nintendoswitch

package runtime

// Linux/amd64 netpoller implementing the upstream internal/poll linkname
// contract (Go GOROOT src/internal/poll/fd_poll_runtime.go) on top of
// epoll(7). Works with both -scheduler=tasks (cooperative: the scheduler's
// waitForEvents/sleepTicks are hooked into epoll_wait) and
// -scheduler=threads (a dedicated netpoll goroutine runs epoll_wait).

import (
	"internal/task"
	"sync/atomic"
	"unsafe"
)

// Result codes returned by pollWait/pollReset; must match internal/poll's
// pollNoError/pollErrClosing/pollErrTimeout/pollErrNotPollable.
const (
	pollNoError        = 0
	pollErrClosing     = 1
	pollErrTimeout     = 2
	pollErrNotPollable = 3
)

// epoll constants (linux/amd64 kernel ABI).
const (
	epollCLOEXEC = 0x80000
	epollCTLAdd  = 1
	epollCTLDEL  = 2
	epollCTLMod  = 3

	epollIN    = 0x001
	epollOUT   = 0x004
	epollERR   = 0x008
	epollHUP   = 0x010
	epollRDHUP = 0x2000
	epollET    = 0x80000000
)

// wakeupEventData is the Data sentinel for the netpoll wakeup eventfd. It must
// not collide with any real fd number (fds are small non-negative ints).
const wakeupEventData = 1<<64 - 1

// Syscall numbers, linux/amd64.
const (
	sysRead         = 0
	sysWrite        = 1
	sysEpollCtl     = 233
	sysEpollPwait   = 281
	sysEventfd2     = 290
	sysEpollCreate1 = 291
)

// epollEvent mirrors struct epoll_event on linux/amd64. The kernel struct is
// __attribute__((packed)): 12 bytes with the 8-byte data at offset 4, so Data
// must be a byte array (as in upstream Go's internal/runtime/syscall/linux
// EpollEvent), not a uint64, which would put it at offset 8.
type epollEvent struct {
	Events uint32
	Data   [8]byte // unaligned uint64
}

// epollEventData packs a uint64 into the packed Data field.
func epollEventData(v uint64) [8]byte {
	var b [8]byte
	for i := 0; i < 8; i++ {
		b[i] = byte(v >> (8 * i))
	}
	return b
}

// epollEventGetData unpacks the packed Data field.
func epollEventGetData(b [8]byte) uint64 {
	var v uint64
	for i := 0; i < 8; i++ {
		v |= uint64(b[i]) << (8 * i)
	}
	return v
}

// The TinyGo compiler lowers any call whose linkname starts with
// "syscall.Syscall" to an inline-asm syscall instruction
// (compiler/compiler.go, createSyscall/createRawSyscall). Declaring a
// bodyless function with that linkname from the runtime package (which
// cannot import syscall) therefore gives us raw syscalls in pure Go.
//
//go:linkname syscall6 syscall.Syscall6
func syscall6(trap, a1, a2, a3, a4, a5, a6 uintptr) (r1, r2 uintptr, err uintptr)

func epollCreate1(flags uintptr) int32 {
	r1, _, e := syscall6(sysEpollCreate1, flags, 0, 0, 0, 0, 0)
	if e != 0 {
		return -int32(e)
	}
	return int32(r1)
}

func epollCtl(epfd int32, op int32, fd int32, ev *epollEvent) int32 {
	_, _, e := syscall6(sysEpollCtl, uintptr(epfd), uintptr(op), uintptr(fd), uintptr(unsafe.Pointer(ev)), 0, 0)
	if e != 0 {
		return -int32(e)
	}
	return 0
}

// epollPwait is epoll_pwait(2) with a NULL sigmask (arg5 == 0). epoll_pwait is
// used rather than epoll_wait because seccomp sandboxes commonly allow only
// epoll_pwait (and arm64 has no epoll_wait syscall at all, which matters when
// this netpoller is ported there).
func epollPwait(epfd int32, events []epollEvent, msec int32) int32 {
	var p unsafe.Pointer
	if len(events) > 0 {
		p = unsafe.Pointer(&events[0])
	}
	r1, _, e := syscall6(sysEpollPwait, uintptr(epfd), uintptr(p), uintptr(len(events)), uintptr(msec), 0, 0)
	if e != 0 {
		return -int32(e)
	}
	return int32(r1)
}

// pollDesc is the per-fd poller context; pollOpen returns it as the opaque ctx.
type pollDesc struct {
	fd int32

	mu task.PMutex // real mutex under scheduler.threads, no-op when cooperative

	closing bool

	readyR atomic.Bool // fd readable (or error/hup) reported by epoll
	readyW atomic.Bool // fd writable (or error/hup) reported by epoll

	waiterR atomic.Pointer[task.Task] // task parked in pollWait('r')
	waiterW atomic.Pointer[task.Task] // task parked in pollWait('w')

	rd int64 // read deadline, nanotime units; 0 = none, <0 = expired
	wd int64 // write deadline, same encoding

	rtimer *timerNode // armed read-deadline timer, nil when none
	wtimer *timerNode // armed write-deadline timer, nil when none
}

var (
	pollServerInitOnce atomic.Bool
	epollFd            int32 = -1
	netpollWakeFd      int32 = -1 // eventfd used to interrupt epoll_wait

	pollDescTable struct {
		mu task.PMutex
		m  map[int32]*pollDesc
	}
)

//go:linkname poll_runtime_pollServerInit internal/poll.runtime_pollServerInit
func poll_runtime_pollServerInit() {
	if pollServerInitOnce.Swap(true) {
		return
	}
	epfd := epollCreate1(epollCLOEXEC)
	if epfd < 0 {
		runtimeFatal("runtime: epoll_create1 failed")
	}
	epollFd = epfd
	// eventfd2(0, EFD_CLOEXEC|EFD_NONBLOCK) to interrupt a blocking epoll_wait.
	r1, _, e := syscall6(sysEventfd2, 0, 0x80000|0x800, 0, 0, 0, 0)
	if e != 0 {
		runtimeFatal("runtime: eventfd2 failed")
	}
	netpollWakeFd = int32(r1)
	var ev epollEvent
	ev.Events = epollIN
	ev.Data = epollEventData(wakeupEventData) // sentinel: marks the wakeup eventfd
	if epollCtl(epollFd, epollCTLAdd, netpollWakeFd, &ev) != 0 {
		runtimeFatal("runtime: epoll_ctl ADD eventfd failed")
	}
	// Under the threads scheduler there is no scheduler idle loop to hook, so
	// run a dedicated netpoll goroutine that blocks in epoll_wait forever and
	// wakes parked tasks. Under the cooperative scheduler the scheduler itself
	// calls netpollWait via the waitForEvents/sleepTicks hooks (see
	// poll_hook_unix.go), so no goroutine is needed.
	if !hasScheduler {
		go netpollLoop()
	}
}

// netpollLoop is the threads-scheduler netpoll goroutine.
func netpollLoop() {
	for {
		netpollWait(-1)
	}
}

//go:linkname poll_runtime_pollOpen internal/poll.runtime_pollOpen
func poll_runtime_pollOpen(fd uintptr) (uintptr, int) {
	poll_runtime_pollServerInit()
	pd := &pollDesc{fd: int32(fd)}
	var ev epollEvent
	// Edge-triggered (EPOLLET), as upstream netpollopen: readiness is
	// re-armed by pollReset clearing the ready flag and the fd re-triggering
	// only on a new edge, so an always-ready fd cannot spin epoll_wait.
	// Data carries the pd pointer, so a stale event for a closed-then-reused
	// fd number is dropped in netpollWait instead of marking the new pd ready.
	// (Upstream instead never frees pollDescs — persistentalloc — and compares
	// an fdseq tag packed into Data; this patch has neither: under threads an
	// event already sitting in the events[] buffer when pollClose runs can
	// still read a recycled pd.)
	pollDescTable.mu.Lock()
	if pollDescTable.m == nil {
		pollDescTable.m = make(map[int32]*pollDesc)
	}
	ev.Events = epollIN | epollOUT | epollRDHUP | epollET
	ev.Data = epollEventData(uint64(uintptr(unsafe.Pointer(pd))))
	if errno := epollCtl(epollFd, epollCTLAdd, int32(fd), &ev); errno != 0 {
		pollDescTable.mu.Unlock()
		return 0, int(-errno)
	}
	pollDescTable.m[pd.fd] = pd
	pollDescTable.mu.Unlock()
	return uintptr(unsafe.Pointer(pd)), 0
}

//go:linkname poll_runtime_pollClose internal/poll.runtime_pollClose
func poll_runtime_pollClose(ctx uintptr) {
	if ctx == 0 {
		return
	}
	pd := (*pollDesc)(unsafe.Pointer(ctx))
	var ev epollEvent
	epollCtl(epollFd, epollCTLDEL, pd.fd, &ev) // best effort
	pollDescTable.mu.Lock()
	delete(pollDescTable.m, pd.fd)
	pollDescTable.mu.Unlock()
	poll_runtime_pollUnblock(ctx)
}

//go:linkname poll_runtime_isPollServerDescriptor internal/poll.runtime_isPollServerDescriptor
func poll_runtime_isPollServerDescriptor(fd uintptr) bool {
	return int32(fd) == epollFd || int32(fd) == netpollWakeFd
}

//go:linkname poll_runtime_pollReset internal/poll.runtime_pollReset
//go:noinline
func poll_runtime_pollReset(ctx uintptr, mode int) int {
	pd := (*pollDesc)(unsafe.Pointer(ctx))
	if res := pollCheckErr(pd, mode); res != pollNoError {
		return res
	}
	if mode == 'r' {
		pd.readyR.Store(false)
	} else if mode == 'w' {
		pd.readyW.Store(false)
	}
	// With EPOLLET the fd only re-triggers on a new edge. Re-arm with
	// EPOLL_CTL_MOD so that an event that arrived between the epoll_pwait
	// return and this reset (a lost edge) is re-reported instead of lost.
	var ev epollEvent
	ev.Events = epollIN | epollOUT | epollRDHUP | epollET
	ev.Data = epollEventData(uint64(uintptr(unsafe.Pointer(pd))))
	epollCtl(epollFd, epollCTLMod, pd.fd, &ev) // best effort
	return pollNoError
}

//go:linkname poll_runtime_pollWait internal/poll.runtime_pollWait
//go:noinline
func poll_runtime_pollWait(ctx uintptr, mode int) int {
	pd := (*pollDesc)(unsafe.Pointer(ctx))
	for {
		if res := pollCheckErr(pd, mode); res != pollNoError {
			return res
		}
		if pollReady(pd, mode, true) {
			return pollNoError
		}
		pollWaitPark(pd, mode)
	}
}

//go:linkname poll_runtime_pollWaitCanceled internal/poll.runtime_pollWaitCanceled
func poll_runtime_pollWaitCanceled(ctx uintptr, mode int) {
	pd := (*pollDesc)(unsafe.Pointer(ctx))
	for !pollReady(pd, mode, true) {
		pd.mu.Lock()
		closing := pd.closing
		pd.mu.Unlock()
		if closing {
			return
		}
		pollWaitPark(pd, mode)
	}
}

//go:linkname poll_runtime_pollSetDeadline internal/poll.runtime_pollSetDeadline
//go:noinline
func poll_runtime_pollSetDeadline(ctx uintptr, d int64, mode int) {
	pd := (*pollDesc)(unsafe.Pointer(ctx))
	pd.mu.Lock()
	if pd.closing {
		pd.mu.Unlock()
		return
	}
	if d > 0 {
		d += nanotime()
		if d <= 0 {
			d = 1<<63 - 1 // overflow: clamp to max
		}
	}
	if mode == 'r' || mode == 'r'+'w' {
		pd.rd = d
	}
	if mode == 'w' || mode == 'r'+'w' {
		pd.wd = d
	}
	pollDeadlineArmLocked(pd, 'r')
	pollDeadlineArmLocked(pd, 'w')
	rd, wd := pd.rd, pd.wd
	pd.mu.Unlock()
	// A deadline set in the past must wake a currently parked waiter.
	if (mode == 'r' || mode == 'r'+'w') && rd < 0 {
		pollWake(pd, 'r')
		netpollBreak()
	}
	if (mode == 'w' || mode == 'r'+'w') && wd < 0 {
		pollWake(pd, 'w')
		netpollBreak()
	}
}

//go:linkname poll_runtime_pollUnblock internal/poll.runtime_pollUnblock
func poll_runtime_pollUnblock(ctx uintptr) {
	pd := (*pollDesc)(unsafe.Pointer(ctx))
	pd.mu.Lock()
	pd.closing = true
	// Stop the deadline timers, as upstream pollUnblock does
	// (runtime/netpoll.go): a callback already dequeued may still fire, but no
	// new deadline can be armed after closing.
	if pd.rtimer != nil {
		removeTimer(pd.rtimer.timer, true)
		pd.rtimer = nil
	}
	if pd.wtimer != nil {
		removeTimer(pd.wtimer.timer, true)
		pd.wtimer = nil
	}
	pd.mu.Unlock()
	pollWake(pd, 'r')
	pollWake(pd, 'w')
	netpollBreak()
}

// pollCheckErr mirrors upstream netpollcheckerr.
func pollCheckErr(pd *pollDesc, mode int) int {
	pd.mu.Lock()
	closing := pd.closing
	rd, wd := pd.rd, pd.wd
	pd.mu.Unlock()
	if closing {
		return pollErrClosing
	}
	if mode == 'r' && rd < 0 {
		return pollErrTimeout
	}
	if mode == 'w' && wd < 0 {
		return pollErrTimeout
	}
	return pollNoError
}

// pollReady reports whether the mode is ready. When consume is true the ready
// flag is cleared atomically (CAS true->false), so each ready signal satisfies
// exactly one wait — as upstream netpollblock does with pdReady->pdNil. Without
// the consume, a stale ready flag (fd drained by another reader, or a stale
// event after fd reuse) would make pollWait return at once in a busy spin
// under EPOLLET, where no new edge arrives to break it.
func pollReady(pd *pollDesc, mode int, consume bool) bool {
	if mode == 'r' {
		if consume {
			return pd.readyR.CompareAndSwap(true, false)
		}
		return pd.readyR.Load()
	}
	if consume {
		return pd.readyW.CompareAndSwap(true, false)
	}
	return pd.readyW.Load()
}

// pollWaitPark registers the current task as the waiter for mode and suspends
// it. The wakeup comes from netpollWait (epoll event), a deadline timer, or
// pollUnblock, all of which call pollWake -> scheduleTask.
func pollWaitPark(pd *pollDesc, mode int) {
	t := task.Current()
	if mode == 'r' {
		pd.waiterR.Store(t)
	} else {
		pd.waiterW.Store(t)
	}
	// Re-check after registering to avoid a lost wakeup. Consume the ready
	// flag here too: if it was set, this wait is the one it belongs to.
	if pollReady(pd, mode, true) || pollCheckErr(pd, mode) != pollNoError {
		// A concurrent pollWake may already have swapped the waiter out and
		// called scheduleTask(t); that pending resume must be absorbed, or
		// the next unrelated Pause would return early. If our Swap below
		// returns nil, a wake is in flight, so Pause once to absorb it.
		if mode == 'r' {
			if pd.waiterR.Swap(nil) == nil {
				task.Pause()
			}
		} else {
			if pd.waiterW.Swap(nil) == nil {
				task.Pause()
			}
		}
		return
	}
	task.Pause()
	pollWaitUnpark(pd, mode)
}

func pollWaitUnpark(pd *pollDesc, mode int) {
	if mode == 'r' {
		pd.waiterR.Store(nil)
	} else {
		pd.waiterW.Store(nil)
	}
}

// pollWake wakes the task parked for mode, if any.
func pollWake(pd *pollDesc, mode int) {
	var w *task.Task
	if mode == 'r' {
		w = pd.waiterR.Swap(nil)
	} else {
		w = pd.waiterW.Swap(nil)
	}
	if w != nil {
		scheduleTask(w)
	}
}

// pollDeadlineArmLocked (re)arms or disarms the deadline timer for one mode.
// pd.mu must be held.
func pollDeadlineArmLocked(pd *pollDesc, mode int) {
	var d int64
	var tn **timerNode
	if mode == 'r' {
		d = pd.rd
		tn = &pd.rtimer
	} else {
		d = pd.wd
		tn = &pd.wtimer
	}
	if *tn != nil {
		removeTimer((*tn).timer, true)
		*tn = nil
	}
	if d > 0 {
		t := &timer{
			when: d,
			f:    pollReadDeadlineCallback,
			arg:  pd,
		}
		if mode == 'w' {
			t.f = pollWriteDeadlineCallback
		}
		node := &timerNode{timer: t, callback: timerCallback}
		*tn = node
		addTimer(node)
	}
}

func pollReadDeadlineCallback(arg any, seq uintptr, delta int64) {
	pd := arg.(*pollDesc)
	pd.mu.Lock()
	pd.rd = -1
	pd.rtimer = nil
	pd.mu.Unlock()
	// Do not set readyR: pollCheckErr reports the timeout (rd<0) before
	// pollReady is consulted, and a stale ready flag would only make a later
	// pollWait return at once in a spin.
	pollWake(pd, 'r')
}

func pollWriteDeadlineCallback(arg any, seq uintptr, delta int64) {
	pd := arg.(*pollDesc)
	pd.mu.Lock()
	pd.wd = -1
	pd.wtimer = nil
	pd.mu.Unlock()
	// See pollReadDeadlineCallback: no readyW store.
	pollWake(pd, 'w')
}

// netpollWait blocks in epoll_wait for up to timeoutMs milliseconds
// (-1 = forever) and wakes the tasks whose fds became ready.
// Returns the number of events processed (<=0 on timeout/error).
func netpollWait(timeoutMs int32) int {
	if epollFd < 0 {
		return 0
	}
	var events [128]epollEvent
	n := epollPwait(epollFd, events[:], timeoutMs)
	if n <= 0 {
		return int(n)
	}
	for i := int32(0); i < n; i++ {
		ev := &events[i]
		data := epollEventGetData(ev.Data)
		if data == wakeupEventData {
			// Wakeup eventfd: drain the counter.
			var buf [8]byte
			syscall6(sysRead, uintptr(netpollWakeFd), uintptr(unsafe.Pointer(&buf[0])), 8, 0, 0, 0)
			continue
		}
		// Data is the pollDesc pointer (not the fd): a stale event for a
		// closed fd whose number was reused carries the old pd pointer, which
		// is no longer in the table, so it is dropped here. Residual (unverified):
		// under threads the event may already sit in events[] when pollClose
		// runs, and if the GC recycled the pd memory as a NEW pollDesc for the
		// same fd number, cur == pd matches and the new pd gets a spurious
		// wake (with the consume-CAS in pollWait that is a spurious wake, not
		// a hang). Upstream avoids this with persistentalloc + fdseq.
		pd := (*pollDesc)(unsafe.Pointer(uintptr(data)))
		pollDescTable.mu.Lock()
		cur := pollDescTable.m[pd.fd]
		pollDescTable.mu.Unlock()
		if cur != pd {
			continue // stale event (fd closed or reused)
		}
		if ev.Events&(epollIN|epollHUP|epollERR|epollRDHUP) != 0 {
			pd.readyR.Store(true)
			pollWake(pd, 'r')
		}
		if ev.Events&(epollOUT|epollHUP|epollERR) != 0 {
			pd.readyW.Store(true)
			pollWake(pd, 'w')
		}
	}
	return int(n)
}

// netpollBreak interrupts a blocking epoll_wait (deadline set in the past,
// fd unblocked) by writing to the eventfd.
func netpollBreak() {
	if netpollWakeFd < 0 {
		return
	}
	var one uint64 = 1
	syscall6(sysWrite, uintptr(netpollWakeFd), uintptr(unsafe.Pointer(&one)), 8, 0, 0, 0)
}

// pollLinknameKeepalive forces the SSA builder to keep the linkname'd
// functions that are only called from upstream internal/poll (via bodyless
// declarations). Without this, the SSA call graph does not see them as
// reachable from the runtime package and they are eliminated as dead code
// before the linkname takes effect at the LLVM IR level.
var pollLinknameKeepalive = []any{
	poll_runtime_pollReset,
	poll_runtime_pollWait,
	poll_runtime_pollWaitCanceled,
	poll_runtime_pollSetDeadline,
	poll_runtime_isPollServerDescriptor,
}

// init installs netpollWaitHook (declared in poll_hook_unix.go) so the
// cooperative scheduler's waitForEvents/sleepTicks route through epoll_wait.
func init() {
	netpollWaitHook = func(timeoutMs int32) bool {
		if epollFd < 0 || pollDescCount() == 0 {
			return false
		}
		netpollWait(timeoutMs)
		return true
	}
}

func pollDescCount() int {
	pollDescTable.mu.Lock()
	n := len(pollDescTable.m)
	pollDescTable.mu.Unlock()
	return n
}
