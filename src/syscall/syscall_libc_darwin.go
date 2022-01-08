// +build darwin

package syscall

import (
	"unsafe"
)

// This file defines errno and constants to match the darwin libsystem ABI.
// Values have been copied from src/syscall/zerrors_darwin_amd64.go.

// This function returns the error location in the darwin ABI.
// Discovered by compiling the following code using Clang:
//
//     #include <errno.h>
//     int getErrno() {
//         return errno;
//     }
//
//export __error
func libc___error() *int32

// getErrno returns the current C errno. It may not have been caused by the last
// call, so it should only be relied upon when the last call indicates an error
// (for example, by returning -1).
func getErrno() Errno {
	errptr := libc___error()
	return Errno(uintptr(*errptr))
}

func (e Errno) Is(target error) bool {
	switch target.Error() {
	case "permission denied":
		return e == EACCES || e == EPERM
	case "file already exists":
		return e == EEXIST
	case "file does not exist":
		return e == ENOENT
	}
	return false
}

// Source: upstream zerrors_darwin_amd64.go
const (
	DT_BLK     = 0x6
	DT_CHR     = 0x2
	DT_DIR     = 0x4
	DT_FIFO    = 0x1
	DT_LNK     = 0xa
	DT_REG     = 0x8
	DT_SOCK    = 0xc
	DT_UNKNOWN = 0x0
	DT_WHT     = 0xe
)

// Source: https://opensource.apple.com/source/xnu/xnu-7195.81.3/bsd/sys/errno.h.auto.html
const (
	EPERM       Errno = 1
	ENOENT      Errno = 2
	EACCES      Errno = 13
	EEXIST      Errno = 17
	EINTR       Errno = 4
	ENOTDIR     Errno = 20
	EINVAL      Errno = 22
	EMFILE      Errno = 24
	EPIPE       Errno = 32
	EAGAIN      Errno = 35
	ETIMEDOUT   Errno = 60
	ENOSYS      Errno = 78
	EWOULDBLOCK Errno = EAGAIN
)

type Signal int

const (
	SIGCHLD Signal = 0x14
	SIGINT  Signal = 0x2
	SIGKILL Signal = 0x9
	SIGTRAP Signal = 0x5
	SIGQUIT Signal = 0x3
	SIGTERM Signal = 0xf
)

const (
	Stdin  = 0
	Stdout = 1
	Stderr = 2
)

const (
	O_RDONLY = 0x0
	O_WRONLY = 0x1
	O_RDWR   = 0x2
	O_APPEND = 0x8
	O_SYNC   = 0x80
	O_CREAT  = 0x200
	O_TRUNC  = 0x400
	O_EXCL   = 0x800

	O_CLOEXEC = 0x01000000
)

// Source: https://opensource.apple.com/source/xnu/xnu-7195.81.3/bsd/sys/mman.h.auto.html
const (
	PROT_NONE  = 0x00 // no permissions
	PROT_READ  = 0x01 // pages can be read
	PROT_WRITE = 0x02 // pages can be written
	PROT_EXEC  = 0x04 // pages can be executed

	MAP_SHARED  = 0x0001 // share changes
	MAP_PRIVATE = 0x0002 // changes are private

	MAP_FILE      = 0x0000 // map from file (default)
	MAP_ANON      = 0x1000 // allocated from memory, swap space
	MAP_ANONYMOUS = MAP_ANON
)

type Timespec struct {
	Sec  int64
	Nsec int64
}

// Source: upstream ztypes_darwin_amd64.go
type Dirent struct {
	Ino       uint64
	Seekoff   uint64
	Reclen    uint16
	Namlen    uint16
	Type      uint8
	Name      [1024]int8
	Pad_cgo_0 [3]byte
}

// Go chose Linux's field names for Stat_t, see https://github.com/golang/go/issues/31735
type Stat_t struct {
	Dev       int32
	Mode      uint16
	Nlink     uint16
	Ino       uint64
	Uid       uint32
	Gid       uint32
	Rdev      int32
	Pad_cgo_0 [4]byte
	Atim      Timespec
	Mtim      Timespec
	Ctim      Timespec
	Btim      Timespec
	Size      int64
	Blocks    int64
	Blksize   int32
	Flags     uint32
	Gen       uint32
	Lspare    int32
	Qspare    [2]int64
}

// Source: https://github.com/apple/darwin-xnu/blob/main/bsd/sys/_types/_s_ifmt.h
const (
	S_IEXEC  = 0x40
	S_IFBLK  = 0x6000
	S_IFCHR  = 0x2000
	S_IFDIR  = 0x4000
	S_IFIFO  = 0x1000
	S_IFLNK  = 0xa000
	S_IFMT   = 0xf000
	S_IFREG  = 0x8000
	S_IFSOCK = 0xc000
	S_IFWHT  = 0xe000
	S_IREAD  = 0x100
	S_IRGRP  = 0x20
	S_IROTH  = 0x4
	S_IRUSR  = 0x100
	S_IRWXG  = 0x38
	S_IRWXO  = 0x7
	S_IRWXU  = 0x1c0
	S_ISGID  = 0x400
	S_ISTXT  = 0x200
	S_ISUID  = 0x800
	S_ISVTX  = 0x200
	S_IWGRP  = 0x10
	S_IWOTH  = 0x2
	S_IWRITE = 0x80
	S_IWUSR  = 0x80
	S_IXGRP  = 0x8
	S_IXOTH  = 0x1
	S_IXUSR  = 0x40
)

func Stat(path string, p *Stat_t) (err error) {
	data := cstring(path)
	n := libc_stat(&data[0], unsafe.Pointer(p))

	if n < 0 {
		err = getErrno()
	}
	return
}

func Fstat(fd int, p *Stat_t) (err error) {
	n := libc_fstat(int32(fd), unsafe.Pointer(p))

	if n < 0 {
		err = getErrno()
	}
	return
}

func Lstat(path string, p *Stat_t) (err error) {
	data := cstring(path)
	n := libc_lstat(&data[0], unsafe.Pointer(p))
	if n < 0 {
		err = getErrno()
	}
	return
}

func Fdopendir(fd int) (dir uintptr, err error) {
	r0, e1 := libc_fdopendir(int32(fd))
	dir = uintptr(r0)
	if e1 != 0 {
		err = getErrno()
	}
	return
}

func readdir_r(dir uintptr, entry *Dirent, result **Dirent) (err error) {
	e1 := libc_readdir_r(unsafe.Pointer(dir), unsafe.Pointer(entry), unsafe.Pointer(result))
	if e1 != 0 {
		err = getErrno()
	}
	return
}

// The odd $INODE64 suffix is an Apple compatibility feature, see
// https://assert.cc/posts/darwin_use_64_bit_inode_vs_ctypes/
// and https://github.com/golang/go/issues/35269
// Without it, you get the old, smaller struct stat from mac os 10.2 or so.

// int fdopendir(int fd, struct DIR * buf);
//export fdopendir$INODE64
func libc_fdopendir(fd int32) (unsafe.Pointer, int32)

// int readdir_r(struct DIR * buf, struct dirent *entry, struct dirent **result);
//export readdir_r$INODE64
func libc_readdir_r(unsafe.Pointer, unsafe.Pointer, unsafe.Pointer) int32

// int stat(const char *path, struct stat * buf);
//export stat$INODE64
func libc_stat(pathname *byte, ptr unsafe.Pointer) int32

// int fstat(int fd, struct stat * buf);
//export fstat$INODE64
func libc_fstat(fd int32, ptr unsafe.Pointer) int32

// int lstat(const char *path, struct stat * buf);
//export lstat$INODE64
func libc_lstat(pathname *byte, ptr unsafe.Pointer) int32
