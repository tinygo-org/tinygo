//go:build wasi
// +build wasi

package syscall

import (
	"internal/itoa"
	"unsafe"
)

// https://github.com/WebAssembly/wasi-libc/blob/main/expected/wasm32-wasi/predefined-macros.txt
// disagrees with ../../lib/wasi-libc/libc-top-half/musl/arch/wasm32/bits/signal.h for SIGCHLD?
// https://github.com/WebAssembly/wasi-libc/issues/271

type Signal int

const (
	SIGINT  Signal = 2
	SIGQUIT Signal = 3
	SIGILL  Signal = 4
	SIGTRAP Signal = 5
	SIGABRT Signal = 6
	SIGBUS  Signal = 7
	SIGFPE  Signal = 8
	SIGKILL Signal = 9
	SIGSEGV Signal = 11
	SIGPIPE Signal = 13
	SIGTERM Signal = 15
	SIGCHLD Signal = 17
)

func (s Signal) Signal() {}

func (s Signal) String() string {
	if 0 <= s && int(s) < len(signals) {
		str := signals[s]
		if str != "" {
			return str
		}
	}
	return "signal " + itoa.Itoa(int(s))
}

var signals = [...]string{}

const (
	Stdin  = 0
	Stdout = 1
	Stderr = 2
)

const (
	__WASI_OFLAGS_CREAT = 1
	__WASI_OFLAGS_EXCL  = 4
	__WASI_OFLAGS_TRUNC = 8

	__WASI_FDFLAGS_APPEND   = 1
	__WASI_FDFLAGS_SYNC     = 16

	O_RDONLY = 0x04000000
	O_WRONLY = 0x10000000
	O_RDWR   = O_RDONLY | O_WRONLY

	O_CREAT = __WASI_OFLAGS_CREAT << 12
	O_TRUNC = __WASI_OFLAGS_TRUNC << 12
	O_EXCL  = __WASI_OFLAGS_EXCL << 12

	O_APPEND   = __WASI_FDFLAGS_APPEND
	O_SYNC     = __WASI_FDFLAGS_SYNC

	O_CLOEXEC = 0

	// ../../lib/wasi-libc/sysroot/include/sys/mman.h
	MAP_FILE      = 0
	MAP_SHARED    = 0x01
	MAP_PRIVATE   = 0x02
	MAP_ANON      = 0x20
	MAP_ANONYMOUS = MAP_ANON

	// ../../lib/wasi-libc/sysroot/include/sys/mman.h
	PROT_NONE  = 0
	PROT_READ  = 1
	PROT_WRITE = 2
	PROT_EXEC  = 4
)

//go:extern errno
var libcErrno uintptr

func getErrno() error {
	return Errno(libcErrno)
}

func (e Errno) Is(target error) bool {
	switch target.Error() {
	case "permission denied":
		return e == EACCES || e == EPERM || e == ENOTCAPABLE // ENOTCAPABLE is unique in WASI
	case "file already exists":
		return e == EEXIST || e == ENOTEMPTY
	case "file does not exist":
		return e == ENOENT
	}
	return false
}

// https://github.com/WebAssembly/wasi-libc/blob/main/libc-bottom-half/headers/public/__errno.h
const (
	E2BIG           Errno = 1      /* Argument list too long */
	EACCES          Errno = 2      /* Permission denied */
	EADDRINUSE      Errno = 3      /* Address already in use */
	EADDRNOTAVAIL   Errno = 4      /* Address not available */
	EAFNOSUPPORT    Errno = 5      /* Address family not supported by protocol family */
	EAGAIN          Errno = 6      /* Try again */
	EWOULDBLOCK     Errno = EAGAIN /* Operation would block */
	EALREADY        Errno = 7      /* Socket already connected */
	EBADF           Errno = 8      /* Bad file number */
	EBADMSG         Errno = 9      /* Trying to read unreadable message */
	EBUSY           Errno = 10     /* Device or resource busy */
	ECANCELED       Errno = 11     /* Operation canceled. */
	ECHILD          Errno = 12     /* No child processes */
	ECONNABORTED    Errno = 13     /* Connection aborted */
	ECONNREFUSED    Errno = 14     /* Connection refused */
	ECONNRESET      Errno = 15     /* Connection reset by peer */
	EDEADLK         Errno = 16     /* Deadlock condition */
	EDESTADDRREQ    Errno = 17     /* Destination address required */
	EDOM            Errno = 18     /* Math arg out of domain of func */
	EDQUOT          Errno = 19     /* Quota exceeded */
	EEXIST          Errno = 20     /* File exists */
	EFAULT          Errno = 21     /* Bad address */
	EFBIG           Errno = 22     /* File too large */
	EHOSTUNREACH    Errno = 23     /* Host is unreachable */
	EIDRM           Errno = 24     /* Identifier removed */
	EILSEQ          Errno = 25
	EINPROGRESS     Errno = 26 /* Connection already in progress */
	EINTR           Errno = 27 /* Interrupted system call */
	EINVAL          Errno = 28 /* Invalid argument */
	EIO             Errno = 29 /* I/O error */
	EISCONN         Errno = 30 /* Socket is already connected */
	EISDIR          Errno = 31 /* Is a directory */
	ELOOP           Errno = 32 /* Too many symbolic links */
	EMFILE          Errno = 33 /* Too many open files */
	EMLINK          Errno = 34 /* Too many links */
	EMSGSIZE        Errno = 35 /* Message too long */
	EMULTIHOP       Errno = 36 /* Multihop attempted */
	ENAMETOOLONG    Errno = 37 /* File name too long */
	ENETDOWN        Errno = 38 /* Network interface is not configured */
	ENETRESET       Errno = 39
	ENETUNREACH     Errno = 40         /* Network is unreachable */
	ENFILE          Errno = 41         /* File table overflow */
	ENOBUFS         Errno = 42         /* No buffer space available */
	ENODEV          Errno = 43         /* No such device */
	ENOENT          Errno = 44         /* No such file or directory */
	ENOEXEC         Errno = 45         /* Exec format error */
	ENOLCK          Errno = 46         /* No record locks available */
	ENOLINK         Errno = 47         /* The link has been severed */
	ENOMEM          Errno = 48         /* Out of memory */
	ENOMSG          Errno = 49         /* No message of desired type */
	ENOPROTOOPT     Errno = 50         /* Protocol not available */
	ENOSPC          Errno = 51         /* No space left on device */
	ENOSYS          Errno = 52         /* Function not implemented */
	ENOTCONN        Errno = 53         /* Socket is not connected */
	ENOTDIR         Errno = 54         /* Not a directory */
	ENOTEMPTY       Errno = 55         /* Directory not empty */
	ENOTSOCK        Errno = 57         /* Socket operation on non-socket */
	ESOCKTNOSUPPORT Errno = 58         /* Socket type not supported */
	EOPNOTSUPP      Errno = 58         /* Operation not supported on transport endpoint */
	ENOTSUP         Errno = EOPNOTSUPP /* Not supported */
	ENOTTY          Errno = 59         /* Not a typewriter */
	ENXIO           Errno = 60         /* No such device or address */
	EOVERFLOW       Errno = 61         /* Value too large for defined data type */
	EPERM           Errno = 63         /* Operation not permitted */
	EPIPE           Errno = 64         /* Broken pipe */
	EPROTO          Errno = 65         /* Protocol error */
	EPROTONOSUPPORT Errno = 66         /* Unknown protocol */
	EPROTOTYPE      Errno = 67         /* Protocol wrong type for socket */
	ERANGE          Errno = 68         /* Math result not representable */
	EROFS           Errno = 69         /* Read-only file system */
	ESPIPE          Errno = 70         /* Illegal seek */
	ESRCH           Errno = 71         /* No such process */
	ESTALE          Errno = 72
	ETIMEDOUT       Errno = 73 /* Connection timed out */
	EXDEV           Errno = 75 /* Cross-device link */
	ENOTCAPABLE     Errno = 76 /* Extension: Capabilities insufficient. */
)

// https://github.com/WebAssembly/wasi-libc/blob/main/libc-bottom-half/headers/public/__struct_timespec.h
type Timespec struct {
	Sec  int32
	Nsec int64
}

// https://github.com/WebAssembly/wasi-libc/blob/main/libc-bottom-half/headers/public/__struct_stat.h
// https://github.com/WebAssembly/wasi-libc/blob/main/libc-bottom-half/headers/public/__typedef_ino_t.h
// etc.
// Go chose Linux's field names for Stat_t, see https://github.com/golang/go/issues/31735
type Stat_t struct {
	Dev       uint64
	Ino       uint64
	Nlink     uint64
	Mode      uint32
	Uid       uint32
	Gid       uint32
	Pad_cgo_0 [4]byte
	Rdev      uint64
	Size      int64
	Blksize   int32
	Blocks    int64

	Atim   Timespec
	Mtim   Timespec
	Ctim   Timespec
	Qspare [3]int64
}

// https://github.com/WebAssembly/wasi-libc/blob/main/libc-top-half/musl/include/sys/stat.h
const (
	S_IFBLK  = 0x6000
	S_IFCHR  = 0x2000
	S_IFDIR  = 0x4000
	S_IFIFO  = 0x1000
	S_IFLNK  = 0xa000
	S_IFMT   = 0xf000
	S_IFREG  = 0x8000
	S_IFSOCK = 0xc000
	S_IREAD  = 0x100
	S_IRGRP  = 0x20
	S_IROTH  = 0x4
	S_IRUSR  = 0x100
	S_IRWXG  = 0x38
	S_IRWXO  = 0x7
	S_IRWXU  = 0x1c0
	S_ISGID  = 0x400
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

// dummy
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

// dummy
type Dirent struct {
	Ino    uint64
	Reclen uint16
	Type   uint8
	Name   [1024]int8
}

func ReadDirent(fd int, buf []byte) (n int, err error) {
	return -1, ENOSYS
}

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

func Pipe2(p []int, flags int) (err error) {
	return ENOSYS // TODO
}

func Getpagesize() int {
	// per upstream
	return 65536
}

// int stat(const char *path, struct stat * buf);
//
//export stat
func libc_stat(pathname *byte, ptr unsafe.Pointer) int32

// int fstat(fd int, struct stat * buf);
//
//export fstat
func libc_fstat(fd int32, ptr unsafe.Pointer) int32

// int lstat(const char *path, struct stat * buf);
//
//export lstat
func libc_lstat(pathname *byte, ptr unsafe.Pointer) int32
