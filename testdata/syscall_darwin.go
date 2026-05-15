//go:build darwin

package main

import (
	"fmt"
	"os"
	"syscall"
)

func main() {
	// Happy path: SYS_GETPID returns the current pid; no errno.
	r1, _, errno := syscall.Syscall(syscall.SYS_GETPID, 0, 0, 0)
	if errno != 0 {
		fmt.Println("getpid errno:", errno)
		return
	}
	if int(r1) != os.Getpid() {
		fmt.Println("getpid mismatch:", r1, os.Getpid())
		return
	}
	fmt.Println("getpid ok")

	// Error path: close(99999) should return EBADF.
	_, _, errno = syscall.Syscall(syscall.SYS_CLOSE, 99999, 0, 0)
	if errno == syscall.EBADF {
		fmt.Println("close ebadf ok")
	} else {
		fmt.Println("close errno unexpected:", errno)
	}

	// Syscall6 path: anonymous mmap exercises all 6 argument
	// registers (RDI/RSI/RDX/R10/R8/R9 on amd64; X0..X5 on arm64).
	// A swap of any register slot would make the kernel reject the
	// call, so successful return is meaningful coverage.
	const (
		PROT_READ_WRITE = 0x3    // PROT_READ | PROT_WRITE
		MAP_PRIVATE     = 0x0002
		MAP_ANON        = 0x1000
	)
	addr, _, errno := syscall.Syscall6(syscall.SYS_MMAP, 0, 4096, PROT_READ_WRITE, MAP_PRIVATE|MAP_ANON, ^uintptr(0), 0)
	if errno != 0 || addr == 0 {
		fmt.Println("mmap6 failed:", errno, "addr=", addr)
		return
	}
	fmt.Println("mmap6 ok")
}
