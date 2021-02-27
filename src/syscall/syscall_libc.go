// +build darwin nintendoswitch wasi

package syscall

import (
	"unsafe"
)

type sliceHeader struct {
	buf *byte
	len uintptr
	cap uintptr
}

func Close(fd int) (err error) {
	n := libc_close(int32(fd))

	if n < 0 {
		return getErrno()
	}
	return nil
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
	buf, count := splitSlice(p)
	n = libc_read(int32(fd), buf, uint(count))
	if n < 0 {
		err = getErrno()
	}
	return
}

func Seek(fd int, offset int64, whence int) (off int64, err error) {
	return 0, ENOSYS // TODO
}

func Open(path string, mode int, perm uint32) (fd int, err error) {
	buf, count := splitSlice([]byte(path))
	cstr := libc_strndup(buf, count)
	defer libc_free(cstr)
	fd = libc_open(cstr, mode, perm)
	if fd < 0 {
		err = getErrno()
		return -1, err
	}

	return fd, nil
}

func Mkdir(path string, mode uint32) (err error) {
	return ENOSYS // TODO
}

func Unlink(path string) (err error) {
	return ENOSYS // TODO
}

func Kill(pid int, sig Signal) (err error) {
	return ENOSYS // TODO
}

func Getpid() (pid int) {
	panic("unimplemented: getpid") // TODO
}

func Getenv(key string) (value string, found bool) {
	data := append([]byte(key), 0)
	raw := libc_getenv(&data[0])
	if raw == nil {
		return "", false
	}

	ptr := uintptr(unsafe.Pointer(raw))
	for size := uintptr(0); ; size++ {
		v := *(*byte)(unsafe.Pointer(ptr))
		if v == 0 {
			src := *(*[]byte)(unsafe.Pointer(&sliceHeader{buf: raw, len: size, cap: size}))
			return string(src), true
		}
		ptr += unsafe.Sizeof(byte(0))
	}
}

func splitSlice(p []byte) (buf *byte, len uintptr) {
	slice := (*sliceHeader)(unsafe.Pointer(&p))
	return slice.buf, slice.len
}

// int open(const char *path, int oflag, int perm)
//export open
func libc_open(path *byte, oflag int, perm uint32) int

// int close(int fd)
//export close
func libc_close(fd int32) int

// ssize_t write(int fd, const void *buf, size_t count)
//export write
func libc_write(fd int32, buf *byte, count uint) int

// ssize_t read(int fd, void *buf, size_t nbyte);
//export read
func libc_read(fd int32, buf *byte, count uint) int

// char *getenv(const char *name);
//export getenv
func libc_getenv(name *byte) *byte

//export free
func libc_free(buf *byte)

//export strndup
func libc_strndup(buf *byte, size uintptr) *byte
