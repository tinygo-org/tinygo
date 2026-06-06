//go:build wasip1

package syscall

import "unsafe"

// Sockaddr is the wasip1 socket-address sentinel. wasip1 socket syscalls
// don't surface peer addresses (sock_accept doesn't return one), so any
// Sockaddr-typed return is always nil. Defined as `any` to match
// upstream Go's net_fake.go.
type Sockaddr = any

// Concrete sockaddr types exist so upstream Go's net package compiles;
// none of them are ever populated (Accept always returns nil).
type SockaddrInet4 struct {
	Port int
	Addr [4]byte
}

type SockaddrInet6 struct {
	Port   int
	ZoneId uint32
	Addr   [16]byte
}

type SockaddrUnix struct {
	Name string
}

// Address-family / socket-type / protocol constants. AF_INET / AF_INET6
// are already defined in syscall.go (Linux values, 0x2 / 0xa); we add
// the rest here. wasip1's host never reads these — they exist so
// upstream Go's net builds.
const (
	AF_UNSPEC = 0
	AF_UNIX   = 1
)

const (
	SOCK_STREAM = 1 + iota
	SOCK_DGRAM
	SOCK_RAW
	SOCK_SEQPACKET
)

const (
	IPPROTO_IP   = 0
	IPPROTO_IPV4 = 4
	IPPROTO_IPV6 = 0x29
	IPPROTO_TCP  = 6
	IPPROTO_UDP  = 0x11
)

const SOMAXCONN = 0x80

// Socket-option / fcntl constants used by upstream net but unsupported
// on wasip1; they exist so the build compiles.
const (
	IPV6_V6ONLY = 1
	SO_ERROR    = 2
)

const F_DUPFD_CLOEXEC = 1

// RLIMIT_NOFILE is referenced by net's rlimit_unix.go. Rlimit /
// Setrlimit are defined in syscall.go; we add the missing constant and
// a Getrlimit stub here.
const RLIMIT_NOFILE = 0

func Getrlimit(which int, lim *Rlimit) error { return ENOSYS }

const (
	SHUT_RD   = 0x1
	SHUT_WR   = 0x2
	SHUT_RDWR = SHUT_RD | SHUT_WR
)

// sock_recv ri_flags / sock_send si_flags. Currently only the receive
// flags have public counterparts in wasi-libc; we expose them for
// callers that want MSG_PEEK-style behaviour. internal/poll's hot-path
// Read/Write pass 0.
const (
	MSG_PEEK    = 0x1
	MSG_WAITALL = 0x2
)

// wasi flag types. fdflags is shared with syscall_libc_wasi.go's O_*
// constants (e.g. O_NONBLOCK = __WASI_FDFLAGS_NONBLOCK = 4).
type (
	fdflags = uint16
	sdflags = uint32
	riflags = uint16
	roflags = uint16
	siflags = uint16
)

//go:wasmimport wasi_snapshot_preview1 sock_accept
//go:noescape
func sock_accept(fd int32, flags fdflags, newfd unsafe.Pointer) uint32

//go:wasmimport wasi_snapshot_preview1 sock_shutdown
//go:noescape
func sock_shutdown(fd int32, flags sdflags) uint32

// Accept wraps wasi sock_accept. The returned Sockaddr is always nil
// because wasi preview1 doesn't surface the peer address. The accepted
// FD inherits the listener's flags, including O_NONBLOCK — pass
// __WASI_FDFLAGS_NONBLOCK explicitly so we don't depend on inheritance
// semantics that vary between hosts.
func Accept(fd int) (int, Sockaddr, error) {
	var newfd int32
	errno := sock_accept(int32(fd), __WASI_FDFLAGS_NONBLOCK, unsafe.Pointer(&newfd))
	if errno != 0 {
		return -1, nil, Errno(errno)
	}
	return int(newfd), nil, nil
}

// Shutdown wraps wasi sock_shutdown. how is one of SHUT_RD, SHUT_WR,
// SHUT_RDWR.
func Shutdown(fd int, how int) error {
	if errno := sock_shutdown(int32(fd), sdflags(how)); errno != 0 {
		return Errno(errno)
	}
	return nil
}

// The remaining socket-related entry points exist as stubs because
// upstream Go's net package references them on the wasip1 build path,
// even though the FileConn / FileListener flow we care about doesn't
// reach them. Each one returns ENOSYS so callers see a clean error.

func Socket(proto, sotype, unused int) (int, error) { return -1, ENOSYS }

func Bind(fd int, sa Sockaddr) error { return ENOSYS }

func Listen(fd int, backlog int) error { return ENOSYS }

func Connect(fd int, sa Sockaddr) error { return ENOSYS }

func Recvfrom(fd int, p []byte, flags int) (int, Sockaddr, error) {
	return 0, nil, ENOSYS
}

func Sendto(fd int, p []byte, flags int, to Sockaddr) error { return ENOSYS }

func Recvmsg(fd int, p, oob []byte, flags int) (n, oobn, recvflags int, from Sockaddr, err error) {
	return 0, 0, 0, nil, ENOSYS
}

func SendmsgN(fd int, p, oob []byte, to Sockaddr, flags int) (int, error) {
	return 0, ENOSYS
}

func GetsockoptInt(fd, level, opt int) (int, error) { return 0, ENOSYS }

func SetsockoptInt(fd, level, opt int, value int) error { return ENOSYS }

func SetReadDeadline(fd int, t int64) error { return ENOSYS }

func SetWriteDeadline(fd int, t int64) error { return ENOSYS }

func StopIO(fd int) error { return ENOSYS }
