package main

import (
	"os"
	"syscall"
	"unsafe"
)

func main() {
	pid, _, errno := syscall.RawSyscall(syscall.SYS_GETPID, 0, 0, 0)
	println("RawSyscall getpid:", int(pid) == os.Getpid(), errno == 0)

	pid, _, errno = syscall.RawSyscall6(syscall.SYS_GETPID, 0, 0, 0, 0, 0, 0)
	println("RawSyscall6 getpid:", int(pid) == os.Getpid(), errno == 0)

	r1, _, errno := syscall.Syscall(syscall.SYS_CLOSE, 12345, 0, 0)
	println("Syscall close:", errno == syscall.EBADF)
	if r1 != ^uintptr(0) {
		println("unexpected r1:", r1)
	}

	msg := []byte("Syscall6 write\n")
	n, _, errno := syscall.Syscall6(syscall.SYS_WRITE, 1, uintptr(unsafe.Pointer(&msg[0])), uintptr(len(msg)), 0, 0, 0)
	println("wrote:", int(n) == len(msg), errno == 0)
}
