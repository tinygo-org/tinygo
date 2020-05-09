// +build darwin nintendoswitch

package syscall

import (
	"errors"
	"unsafe"
)

var (
	notImplemented = errors.New("syscall: not implemented")
)

func Close(fd int) (err error) {
	errN := libc_close(int32(fd))
	if errN != 0 {
		return errors.New("error closing fd") // TODO: Return errno
	}
	return nil
}

func Write(fd int, p []byte) (n int, err error) {
	buf, count := splitSlice(p)
	n = libc_write(int32(fd), buf, uint(count))
	if n < 0 {
		err = errors.New("error writing to fd") // TODO: Return errno
	}
	return
}

func Read(fd int, p []byte) (n int, err error) {
	buf, count := splitSlice(p)
	n = libc_read(int32(fd), buf, uint(count))
	if n < 0 {
		err = errors.New("error reading from fd") // TODO: Return errno
	}
	return
}

func Seek(fd int, offset int64, whence int) (off int64, err error) {
	return libc_lseek(int32(fd), offset, whence), nil
}

func Open(path string, mode int, perm uint32) (fd int, err error) {
	buf, _ := splitString(path)
	fd = libc_open(buf, uint(mode), uint(perm))
	if fd < 0 {
		err = errors.New("error opening path") // TODO: Return errno
	}

	return
}

func Mkdir(path string, mode uint32) (err error) {
	return notImplemented // TODO
}

func Unlink(path string) (err error) {
	return notImplemented // TODO
}

func Kill(pid int, sig Signal) (err error) {
	return notImplemented // TODO
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

func splitString(p string) (buf *byte, len uintptr) {
	slice := (*struct {
		ptr    *byte
		length uintptr
	})(unsafe.Pointer(&p))
	return slice.ptr, slice.length
}

// int write(int fd, const void *buf, size_t cnt)
//export write
func libc_write(fd int32, buffer *byte, size uint) int

// int read(int fd, void *buf, size_t count);
//export read
func libc_read(fd int32, buffer *byte, size uint) int

// int close(int fd);
//export close
func libc_close(fd int32) int

// int open(const char *pathname, int flags, mode_t mode);
//export open
func libc_open(pathname *byte, flags uint, mode uint) int

//export lseek
func libc_lseek(fd int32, offset int64, whence int) int64
