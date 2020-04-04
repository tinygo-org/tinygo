// +build darwin

package syscall

import (
	"unsafe"
)

func Close(fd int) (err error) {
	return ENOSYS // TODO
}

func Write(fd int, p []byte) (n int, err error) {
	buf, count := splitSlice(p)
	n = libc_write(int32(fd), buf, uint(count))
	if n < 0 {
		err = getErrno()
	}
	return
}

func Read(fd int, p []byte) (n int, err error) {
	return 0, ENOSYS // TODO
}

func Seek(fd int, offset int64, whence int) (off int64, err error) {
	return 0, ENOSYS // TODO
}

func Open(path string, mode int, perm uint32) (fd int, err error) {
	return 0, ENOSYS // TODO
}

func Kill(pid int, sig Signal) (err error) {
	return ENOSYS // TODO
}

func Getpid() (pid int) {
	panic("unimplemented: getpid") // TODO
}

func Getenv(key string) (value string, found bool) {
	return "", false // TODO
}

func splitSlice(p []byte) (buf *byte, len uintptr) {
	slice := (*struct {
		buf *byte
		len uintptr
		cap uintptr
	})(unsafe.Pointer(&p))
	return slice.buf, slice.len
}

// ssize_t write(int fd, const void *buf, size_t count)
//export write
func libc_write(fd int32, buf *byte, count uint) int
