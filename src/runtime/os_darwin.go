//go:build darwin

package runtime

import "unsafe"

const GOOS = "darwin"

const zeroSizeAllocPtr uintptr = 16 // part of the first protected page

const (
	// See https://github.com/golang/go/blob/master/src/syscall/zerrors_darwin_amd64.go
	flag_PROT_READ     = 0x1
	flag_PROT_WRITE    = 0x2
	flag_MAP_PRIVATE   = 0x2
	flag_MAP_ANONYMOUS = 0x1000 // MAP_ANON
)

// Source: https://opensource.apple.com/source/Libc/Libc-1439.100.3/include/time.h.auto.html
const (
	clock_REALTIME      = 0
	clock_MONOTONIC_RAW = 4
)

// Source:
// https://opensource.apple.com/source/xnu/xnu-7195.141.2/bsd/sys/signal.h.auto.html
const (
	sig_SIGBUS  = 10
	sig_SIGFPE  = 8
	sig_SIGILL  = 4
	sig_SIGSEGV = 11
)

func hardwareRand() (n uint64, ok bool) {
	n |= uint64(libc_arc4random())
	n |= uint64(libc_arc4random()) << 32
	return n, true
}

//go:linkname syscall_Getpagesize syscall.Getpagesize
func syscall_Getpagesize() int {
	return int(libc_getpagesize())
}

// uint32_t arc4random(void);
//
//export arc4random
func libc_arc4random() uint32

// int getpagesize(void);
//
//export getpagesize
func libc_getpagesize() int32

// This function returns the error location in the darwin ABI.
// Discovered by compiling the following code using Clang:
//
//	#include <errno.h>
//	int getErrno() {
//	    return errno;
//	}
//
//export __error
func libc_errno_location() *int32

//export tinygo_syscall
func call_syscall(fn, a1, a2, a3 uintptr) int32

//export tinygo_syscallX
func call_syscallX(fn, a1, a2, a3 uintptr) uintptr

//export tinygo_syscall6
func call_syscall6(fn, a1, a2, a3, a4, a5, a6 uintptr) int32

//export tinygo_syscall6X
func call_syscall6X(fn, a1, a2, a3, a4, a5, a6 uintptr) uintptr

// Go 1.26 replaced the individual Darwin syscall functions (syscall, syscallX,
// syscallPtr, syscall6, syscall6X) with two variadic entry points: syscalln and
// rawsyscalln. All the old wrappers now have Go bodies that call these and then
// interpret the result (errno, errnoX, errnoPtr). Because the callers decide
// how to check for errors, we must:
//   - always return the full register-width result (use call_syscallX /
//     call_syscall6X which return uintptr, not call_syscall / call_syscall6
//     which truncate to int32), and
//   - always read errno so that the Go wrappers have the value available when
//     the error condition is met.

// syscall_syscalln is a wrapper around the libc call with variable arguments.
//
//go:nosplit
//go:linkname syscall_syscalln syscall.syscalln
func syscall_syscalln(fn uintptr, args ...uintptr) (uintptr, uintptr, uintptr) {
	return syscall_rawsyscalln(fn, args...)
}

// syscall_rawsyscalln is a wrapper around the libc call with variable arguments.
//
//go:linkname syscall_rawsyscalln syscall.rawsyscalln
//go:nosplit
func syscall_rawsyscalln(fn uintptr, args ...uintptr) (uintptr, uintptr, uintptr) {
	var a1, a2, a3, a4, a5, a6 uintptr

	switch len(args) {
	case 3:
		a3 = args[2]
		fallthrough
	case 2:
		a2 = args[1]
		fallthrough
	case 1:
		a1 = args[0]
		fallthrough
	case 0:
		r1 := call_syscallX(fn, a1, a2, a3)
		err := uintptr(*libc_errno_location())
		return r1, 0, err

	case 6:
		a6 = args[5]
		fallthrough
	case 5:
		a5 = args[4]
		fallthrough
	case 4:
		a4 = args[3]

		a1, a2, a3 = args[0], args[1], args[2]
		r1 := call_syscall6X(fn, a1, a2, a3, a4, a5, a6)
		err := uintptr(*libc_errno_location())
		return r1, 0, err
	}

	panic("syscall args not handled")
}

//go:linkname os_runtime_executable_path os.runtime_executable_path
func os_runtime_executable_path() string {
	argv := (*unsafe.Pointer)(unsafe.Pointer(main_argv))

	// skip over argv
	argv = (*unsafe.Pointer)(unsafe.Add(unsafe.Pointer(argv), (uintptr(main_argc)+1)*unsafe.Sizeof(argv)))

	// skip over envv
	for (*argv) != nil {
		argv = (*unsafe.Pointer)(unsafe.Add(unsafe.Pointer(argv), unsafe.Sizeof(argv)))
	}

	// next string is exe path
	argv = (*unsafe.Pointer)(unsafe.Add(unsafe.Pointer(argv), unsafe.Sizeof(argv)))

	cstr := unsafe.Pointer(*argv)
	length := strlen(cstr)
	argString := _string{
		length: length,
		ptr:    (*byte)(cstr),
	}
	executablePath := *(*string)(unsafe.Pointer(&argString))

	// strip "executable_path=" prefix if available, it's added after OS X 10.11.
	executablePath = stringsTrimPrefix(executablePath, "executable_path=")
	return executablePath
}

func stringsTrimPrefix(s, prefix string) string {
	if len(s) >= len(prefix) && s[:len(prefix)] == prefix {
		return s[len(prefix):]
	}
	return s
}
