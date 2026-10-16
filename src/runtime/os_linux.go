//go:build linux && !baremetal && !nintendoswitch && !wasip1 && !wasm_unknown && !wasip2

package runtime

// This file is for systems that are _actually_ Linux (not systems that pretend
// to be Linux, like baremetal systems).

import (
	"internal/task"
	"sync/atomic"
	"unsafe"
)

const GOOS = "linux"

const zeroSizeAllocPtr uintptr = 16 // part of the first protected page

const (
	// See https://github.com/torvalds/linux/blob/master/include/uapi/asm-generic/mman-common.h
	flag_PROT_READ     = 0x1
	flag_PROT_WRITE    = 0x2
	flag_MAP_PRIVATE   = 0x2
	flag_MAP_ANONYMOUS = linux_MAP_ANONYMOUS // different on alpha, hppa, mips, xtensa
)

// Source: https://github.com/torvalds/linux/blob/master/include/uapi/linux/time.h
const (
	clock_REALTIME      = 0
	clock_MONOTONIC_RAW = 4
)

const (
	sig_SIGBUS  = linux_SIGBUS
	sig_SIGFPE  = linux_SIGFPE
	sig_SIGILL  = linux_SIGILL
	sig_SIGSEGV = linux_SIGSEGV
)

// int *__errno_location(void);
//
//export __errno_location
func libc_errno_location() *int32

//export getpagesize
func libc_getpagesize() int

//go:linkname syscall_Getpagesize syscall.Getpagesize
func syscall_Getpagesize() int {
	return libc_getpagesize()
}

func hardwareRand() (n uint64, ok bool) {
	read := libc_getrandom(unsafe.Pointer(&n), 8, 0)
	if read != 8 {
		return 0, false
	}
	return n, true
}

// ssize_t getrandom(void buf[.buflen], size_t buflen, unsigned int flags);
//
//export getrandom
func libc_getrandom(buf unsafe.Pointer, buflen uintptr, flags uint32) uint32

// int fcntl(int fd, int cmd, int arg);
//
//export fcntl
func libc_fcntl(fd int, cmd int, arg int) (ret int)

func fcntl(fd int32, cmd int32, arg int32) (ret int32, errno int32) {
	ret = int32(libc_fcntl(int(fd), int(cmd), int(arg)))
	errno = *libc_errno_location()
	return
}

const (
	pollNoError        = 0
	pollErrClosing     = 1
	pollErrTimeout     = 2
	pollErrNotPollable = 3
)

type pollNode struct {
	fd        int32
	wakefd    int32
	closing   atomic.Bool
	rdeadline atomic.Int64
	wdeadline atomic.Int64
	nextFree  uintptr
}

var (
	pollMu       task.PMutex
	pollNodes    []*pollNode
	pollFreeHead uintptr
)

//export tinygo_poll_eventfd
func tinygo_poll_eventfd() int32

//export tinygo_poll_wake
func tinygo_poll_wake(wakefd int32)

//export tinygo_poll_close
func tinygo_poll_close(wakefd int32)

//export tinygo_poll_wait
func tinygo_poll_wait(fd int32, wakefd int32, mode int32, timeoutNs int64) int32

func getPollNode(ctx uintptr) *pollNode {
	if ctx == 0 {
		return nil
	}
	pollMu.Lock()
	node := pollNodes[ctx-1]
	pollMu.Unlock()
	return node
}

//go:linkname poll_runtime_pollServerInit internal/poll.runtime_pollServerInit
func poll_runtime_pollServerInit() {
}

//go:linkname poll_runtime_pollOpen internal/poll.runtime_pollOpen
func poll_runtime_pollOpen(fd uintptr) (uintptr, int) {
	wakefd := tinygo_poll_eventfd()
	if wakefd < 0 {
		return 0, int(*libc_errno_location())
	}
	pollMu.Lock()
	var ctx uintptr
	var node *pollNode
	if pollFreeHead != 0 {
		ctx = pollFreeHead
		node = pollNodes[ctx-1]
		pollFreeHead = node.nextFree
	} else {
		node = &pollNode{}
		pollNodes = append(pollNodes, node)
		ctx = uintptr(len(pollNodes))
	}
	node.fd = int32(fd)
	node.wakefd = wakefd
	node.closing.Store(false)
	node.rdeadline.Store(0)
	node.wdeadline.Store(0)
	node.nextFree = 0
	pollMu.Unlock()
	return ctx, 0
}

//go:linkname poll_runtime_pollClose internal/poll.runtime_pollClose
func poll_runtime_pollClose(ctx uintptr) {
	node := getPollNode(ctx)
	if node == nil {
		return
	}
	if node.wakefd >= 0 {
		tinygo_poll_close(node.wakefd)
		node.wakefd = -1
	}
	pollMu.Lock()
	node.nextFree = pollFreeHead
	pollFreeHead = ctx
	pollMu.Unlock()
}

func pollCheckDeadline(node *pollNode, mode int) (int64, bool) {
	var d int64
	if mode == 'w' {
		d = node.wdeadline.Load()
	} else {
		d = node.rdeadline.Load()
	}
	if d < 0 {
		return 0, true
	}
	if d > 0 {
		rem := d - nanotime()
		if rem <= 0 {
			return 0, true
		}
		return rem, false
	}
	return -1, false
}

//go:linkname poll_runtime_pollReset internal/poll.runtime_pollReset
func poll_runtime_pollReset(ctx uintptr, mode int) int {
	node := getPollNode(ctx)
	if node == nil {
		return pollNoError
	}
	if node.closing.Load() {
		return pollErrClosing
	}
	if _, expired := pollCheckDeadline(node, mode); expired {
		return pollErrTimeout
	}
	return pollNoError
}

//go:linkname poll_runtime_pollWait internal/poll.runtime_pollWait
func poll_runtime_pollWait(ctx uintptr, mode int) int {
	node := getPollNode(ctx)
	if node == nil {
		return pollNoError
	}
	for {
		if node.closing.Load() {
			return pollErrClosing
		}
		timeoutNs, expired := pollCheckDeadline(node, mode)
		if expired {
			return pollErrTimeout
		}
		res := tinygo_poll_wait(node.fd, node.wakefd, int32(mode), timeoutNs)
		if res == 0 {
			if node.closing.Load() {
				return pollErrClosing
			}
			return pollNoError
		}
	}
}

//go:linkname poll_runtime_pollWaitCanceled internal/poll.runtime_pollWaitCanceled
func poll_runtime_pollWaitCanceled(ctx uintptr, mode int) {
}

//go:linkname poll_runtime_pollSetDeadline internal/poll.runtime_pollSetDeadline
func poll_runtime_pollSetDeadline(ctx uintptr, d int64, mode int) {
	node := getPollNode(ctx)
	if node == nil {
		return
	}
	if d > 0 {
		d += nanotime()
	}
	if mode == 'r' || mode == 'r'+'w' {
		node.rdeadline.Store(d)
	}
	if mode == 'w' || mode == 'r'+'w' {
		node.wdeadline.Store(d)
	}
	tinygo_poll_wake(node.wakefd)
}

//go:linkname poll_runtime_pollUnblock internal/poll.runtime_pollUnblock
func poll_runtime_pollUnblock(ctx uintptr) {
	node := getPollNode(ctx)
	if node == nil {
		return
	}
	node.closing.Store(true)
	tinygo_poll_wake(node.wakefd)
}

//go:linkname poll_runtime_isPollServerDescriptor internal/poll.runtime_isPollServerDescriptor
func poll_runtime_isPollServerDescriptor(fd uintptr) bool {
	return false
}

//go:linkname syscall_RawSyscall syscall.RawSyscall
func syscall_RawSyscall(trap, a1, a2, a3 uintptr) (r1, r2, err uintptr)

//go:linkname syscall_rawVforkSyscall syscall.rawVforkSyscall
func syscall_rawVforkSyscall(trap, a1, a2, a3 uintptr) (r1, err uintptr) {
	// Strip CLONE_VFORK (0x4000) | CLONE_VM (0x0100) so the kernel performs a
	// standard copy-on-write clone instead of sharing the parent's stack.
	const cloneVforkVM = 0x4100
	if trap == 435 && a1 != 0 { // _SYS_clone3
		flagsPtr := (*uint64)(unsafe.Pointer(a1))
		*flagsPtr &^= cloneVforkVM
	} else { // SYS_CLONE
		if GOARCH == "s390x" {
			a2 &^= cloneVforkVM
		} else {
			a1 &^= cloneVforkVM
		}
	}
	r1, _, err = syscall_RawSyscall(trap, a1, a2, a3)
	return r1, err
}
