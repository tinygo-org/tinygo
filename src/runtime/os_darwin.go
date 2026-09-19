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

//go:linkname os_runtime_executable_path os.runtime_executable_path
func os_runtime_executable_path() string {
	return executablePath
}

var executablePath string

func platform_argv(argc int32, argv *unsafe.Pointer) {
	executablePath = executable_path(argc, argv)
}

func executable_path(argc int32, argv *unsafe.Pointer) string {
	// skip over argv
	argv = (*unsafe.Pointer)(unsafe.Add(unsafe.Pointer(argv), (uintptr(argc)+1)*unsafe.Sizeof(argv)))

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
	path := *(*string)(unsafe.Pointer(&argString))

	// strip "executable_path=" prefix if available, it's added after OS X 10.11.
	path = stringsTrimPrefix(path, "executable_path=")

	if path != "" && path[0] == '/' {
		// absolute path
		return path
	}

	cwd := getcwd()
	if cwd != "" {
		path = joinPath(cwd, path)
	}
	return path
}

func stringsTrimPrefix(s, prefix string) string {
	if len(s) >= len(prefix) && s[:len(prefix)] == prefix {
		return s[len(prefix):]
	}
	return s
}

func joinPath(dir, name string) string {
	if len(dir) > 0 && dir[len(dir)-1] == '/' {
		return dir + name
	}
	return dir + "/" + name
}

// from syscall
func getcwd() string {
	const pathMax = 1024
	var buf [4 * pathMax]byte
	s := libc_getcwd(&buf[0], uint(len(buf)))
	if s == nil {
		return ""
	}
	n := clen(buf[:])
	if n < 1 {
		return ""
	}
	return string(buf[:n])
}

// clen returns the index of the first NULL byte in n or len(n) if n contains no NULL byte.
func clen(n []byte) int {
	for i := 0; i < len(n); i++ {
		if n[i] == 0 {
			return i
		}
	}
	return len(n)
}

// char *getcwd(char *buf, size_t size)
//
//export getcwd
func libc_getcwd(buf *byte, size uint) *byte
