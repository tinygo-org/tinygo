//go:build js || nintendoswitch || wasip2 || (wasip1 && !scheduler.tasks && !scheduler.asyncify)

package syscall

import "unsafe"

// These are the default Read/Write/Pread/Pwrite implementations for
// libc-backed wasm targets that do NOT have the cooperative scheduler
// + wasip1 netpoll integration. They are simple pass-throughs to the
// underlying libc syscalls and block the entire wasm module if the FD
// is in blocking mode.
//
// The wasip1 + cooperative-scheduler build replaces these with versions
// that park the goroutine on EAGAIN; see syscall_libc_wasip1.go.

func Write(fd int, p []byte) (n int, err error) {
	n = libc_write(int32(fd), unsafe.SliceData(p), uint(len(p)))
	if n < 0 {
		err = getErrno()
	}
	return
}

func Read(fd int, p []byte) (n int, err error) {
	n = libc_read(int32(fd), unsafe.SliceData(p), uint(len(p)))
	if n < 0 {
		err = getErrno()
	}
	return
}

func Pread(fd int, p []byte, offset int64) (n int, err error) {
	n = libc_pread(int32(fd), unsafe.SliceData(p), uint(len(p)), offset)
	if n < 0 {
		err = getErrno()
	}
	return
}

func Pwrite(fd int, p []byte, offset int64) (n int, err error) {
	n = libc_pwrite(int32(fd), unsafe.SliceData(p), uint(len(p)), offset)
	if n < 0 {
		err = getErrno()
	}
	return
}
