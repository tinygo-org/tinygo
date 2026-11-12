//go:build linux && !baremetal && !nintendoswitch && !wasip1 && !wasm_unknown && !wasip2

package runtime

// This file is for systems that are _actually_ Linux (not systems that pretend
// to be Linux, like baremetal systems).

import (
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
