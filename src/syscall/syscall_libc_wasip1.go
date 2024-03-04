//go:build wasip1

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
	__WASI_OFLAGS_CREAT     = 1
	__WASI_OFLAGS_DIRECTORY = 2
	__WASI_OFLAGS_EXCL      = 4
	__WASI_OFLAGS_TRUNC     = 8

	__WASI_FDFLAGS_APPEND   = 1
	__WASI_FDFLAGS_DSYNC    = 2
	__WASI_FDFLAGS_NONBLOCK = 4
	__WASI_FDFLAGS_RSYNC    = 8
	__WASI_FDFLAGS_SYNC     = 16

	__WASI_FILETYPE_UNKNOWN          = 0
	__WASI_FILETYPE_BLOCK_DEVICE     = 1
	__WASI_FILETYPE_CHARACTER_DEVICE = 2
	__WASI_FILETYPE_DIRECTORY        = 3
	__WASI_FILETYPE_REGULAR_FILE     = 4
	__WASI_FILETYPE_SOCKET_DGRAM     = 5
	__WASI_FILETYPE_SOCKET_STREAM    = 6
	__WASI_FILETYPE_SYMBOLIC_LINK    = 7

	// ../../lib/wasi-libc/libc-bottom-half/headers/public/__header_fcntl.h
	O_RDONLY = 0x04000000
	O_WRONLY = 0x10000000
	O_RDWR   = O_RDONLY | O_WRONLY

	O_CREAT     = __WASI_OFLAGS_CREAT << 12
	O_TRUNC     = __WASI_OFLAGS_TRUNC << 12
	O_EXCL      = __WASI_OFLAGS_EXCL << 12
	O_DIRECTORY = __WASI_OFLAGS_DIRECTORY << 12

	O_APPEND   = __WASI_FDFLAGS_APPEND
	O_DSYNC    = __WASI_FDFLAGS_DSYNC
	O_NONBLOCK = __WASI_FDFLAGS_NONBLOCK
	O_RSYNC    = __WASI_FDFLAGS_RSYNC
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

	// ../../lib/wasi-libc/expected/wasm32-wasi/predefined-macros.txt
	F_GETFL = 3
	F_SETFL = 4
)

// These values are needed as a stub until Go supports WASI as a full target.
// The constant values don't have a meaning and don't correspond to anything
// real.
const (
	_ = iota
	SYS_FCNTL
	SYS_FCNTL64
	SYS_FSTATAT64
	SYS_OPENAT
	SYS_UNLINKAT
	PATH_MAX = 4096
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

// Unix returns the time stored in ts as seconds plus nanoseconds.
func (ts *Timespec) Unix() (sec int64, nsec int64) {
	return int64(ts.Sec), int64(ts.Nsec)
}

// https://github.com/WebAssembly/wasi-libc/blob/main/libc-bottom-half/headers/public/__struct_stat.h
// https://github.com/WebAssembly/wasi-libc/blob/main/libc-bottom-half/headers/public/__typedef_ino_t.h
// etc.
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

// https://github.com/WebAssembly/wasi-libc/blob/main/libc-bottom-half/headers/public/__header_dirent.h
const (
	DT_BLK     = __WASI_FILETYPE_BLOCK_DEVICE
	DT_CHR     = __WASI_FILETYPE_CHARACTER_DEVICE
	DT_DIR     = __WASI_FILETYPE_DIRECTORY
	DT_FIFO    = __WASI_FILETYPE_SOCKET_STREAM
	DT_LNK     = __WASI_FILETYPE_SYMBOLIC_LINK
	DT_REG     = __WASI_FILETYPE_REGULAR_FILE
	DT_UNKNOWN = __WASI_FILETYPE_UNKNOWN
)

// Dirent is returned by pointer from Readdir to iterate over directory entries.
//
// The pointer is managed by wasi-libc and is only valid until the next call to
// Readdir or Fdclosedir.
//
// https://github.com/WebAssembly/wasi-libc/blob/main/libc-bottom-half/headers/public/__struct_dirent.h
type Dirent struct {
	Ino  uint64
	Type uint8
}

func (dirent *Dirent) Name() []byte {
	// The dirent C struct uses a flexible array member to indicate that the
	// directory name is laid out in memory right after the struct data:
	//
	// struct dirent {
	//   ino_t d_ino;
	//   unsigned char d_type;
	//   char d_name[];
	// };
	name := (*[PATH_MAX]byte)(unsafe.Add(unsafe.Pointer(dirent), 9))
	for i, c := range name {
		if c == 0 {
			return name[:i:i]
		}
	}
	return name[:]
}

func Fdopendir(fd int) (dir uintptr, err error) {
	d := libc_fdopendir(int32(fd))

	if d == nil {
		err = getErrno()
	}
	return uintptr(d), err
}

func Fdclosedir(dir uintptr) (err error) {
	// Unlike on other unix platform where only closedir exists, wasi-libc has
	// fdclosedir which releases resources and returns the file descriptor but
	// does not close it. This is useful for us since we want to be able to keep
	// using it.
	n := libc_fdclosedir(unsafe.Pointer(dir))

	if n < 0 {
		err = getErrno()
	}
	return
}

func Readdir(dir uintptr) (dirent *Dirent, err error) {
	// There might be a leftover errno value in the global variable, so we have
	// to clear it before calling readdir because we cannot know whether a nil
	// return means that we reached EOF or that an error occured.
	libcErrno = 0

	dirent = libc_readdir(unsafe.Pointer(dir))

	if dirent == nil && libcErrno != 0 {
		err = getErrno()
	}
	return
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

func Chmod(path string, mode uint32) (err error) {
	// wasi does not have chmod, but there are tests that validate that calling
	// os.Chmod does not error (e.g. io/fs.TestIssue51617).
	//
	// We make a call to Lstat instead so we detect conditions like the path not
	// existing, but we don't honnor the request to modify the file permissions.
	stat := Stat_t{}
	return Lstat(path, &stat)
}

// TODO: should this return runtime.wasmPageSize?
func Getpagesize() int {
	return libc_getpagesize()
}

type Utsname struct {
	Sysname    [65]int8
	Nodename   [65]int8
	Release    [65]int8
	Version    [65]int8
	Machine    [65]int8
	Domainname [65]int8
}

// Stub Utsname, needed because WASI pretends to be linux/arm.
func Uname(buf *Utsname) (err error)

type RawSockaddrInet4 struct {
	// stub
}

type RawSockaddrInet6 struct {
	// stub
}

// This is a stub, it is not functional.
func Syscall(trap, a1, a2, a3 uintptr) (r1, r2 uintptr, err Errno)

// This is a stub, it is not functional.
func Syscall6(trap, a1, a2, a3, a4, a5, a6 uintptr) (r1, r2 uintptr, err Errno)

// int getpagesize(void);
//
//export getpagesize
func libc_getpagesize() int

// int stat(const char *path, struct stat * buf);
//
//export stat
func libc_stat(pathname *byte, ptr unsafe.Pointer) int32

// int fstat(int fd, struct stat * buf);
//
//export fstat
func libc_fstat(fd int32, ptr unsafe.Pointer) int32

// int lstat(const char *path, struct stat * buf);
//
//export lstat
func libc_lstat(pathname *byte, ptr unsafe.Pointer) int32

// int open(const char *pathname, int flags, mode_t mode);
//
//export open
func libc_open(pathname *byte, flags int32, mode uint32) int32

// DIR *fdopendir(int);
//
//export fdopendir
func libc_fdopendir(fd int32) unsafe.Pointer

// int fdclosedir(DIR *);
//
//export fdclosedir
func libc_fdclosedir(unsafe.Pointer) int32

// struct dirent *readdir(DIR *);
//
//export readdir
func libc_readdir(unsafe.Pointer) *Dirent
