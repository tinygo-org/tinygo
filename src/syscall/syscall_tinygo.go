package syscall

// Stub out Syscall() functions.  Args will be Posix/Unix encoded for the arch.
// I guess?  The idea is to stub out the Syscall() functions for now.  And see
// which ones get hit when running examples/net examples. This will define
// (suggest?) the interface to the networking stack.  Hmmm...probably will work
// for more than networking, for example, file system calls could interface
// here also.

import (
	"internal/itoa"
	"unsafe"
)

func RawSyscall6(trap, a1, a2, a3, a4, a5, a6 uintptr) (r1, r2 uintptr, err Errno) {
	println("RawSyscall6 not implemented", trap, a1, a2, a3, a4, a5, a6)
	return r1, r2, EOPNOTSUPP
}

func RawSyscall(trap, a1, a2, a3 uintptr) (r1, r2 uintptr, err Errno) {
	println("RawSyscall not implemented", trap, a1, a2, a3)
	return RawSyscall6(trap, a1, a2, a3, 0, 0, 0)
}

func Syscall(trap, a1, a2, a3 uintptr) (r1, r2 uintptr, err Errno) {
	println("Syscall not implemented", trap, a1, a2, a3)
	return r1, r2, EOPNOTSUPP
}

func Syscall6(trap, a1, a2, a3, a4, a5, a6 uintptr) (r1, r2 uintptr, err Errno) {
	println("Syscall6 not implemented", trap, a1, a2, a3, a4, a5, a6)
	return r1, r2, EOPNOTSUPP
}

func rawSyscallNoError(trap, a1, a2, a3 uintptr) (r1, r2 uintptr) {
	println("rawSyscallNoError not implemented", trap, a1, a2, a3)
	return r1, r2
}

func Faccessat(dirfd int, path string, mode uint32, flags int) (err error) {
	println("Faccesat not implemented", dirfd, path, mode, flags)
	return EOPNOTSUPP
}

func Accept4(fd int, flags int) (nfd int, sa Sockaddr, err error) {
	println("Accept4 not implemented", fd, flags)
	return nfd, sa, EOPNOTSUPP
}

func prlimit(pid int, resource int, newlimit *Rlimit, old *Rlimit) (err error) {
	println("prlimit not implemented", pid, resource, newlimit, old)
	return EOPNOTSUPP
}

func Pread(fd int, p []byte, offset int64) (n int, err error) {
	println("Pread not implemented", fd, p, offset)
	return n, EOPNOTSUPP
}

func Pwrite(fd int, p []byte, offset int64) (n int, err error) {
	println("Pwrite not implemented", fd, p, offset)
	return n, EOPNOTSUPP
}

func Write(fd int, p []byte) (n int, err error) {
	println("Write not implemented", fd, p)
	return n, EOPNOTSUPP
}

func Read(fd int, p []byte) (n int, err error) {
	println("Read not implemented", fd, p)
	return n, EOPNOTSUPP
}

func GetsockoptInt(fd, level, opt int) (value int, err error) {
	println("GetsockoptInt not implemented", fd, level, opt)
	return value, EOPNOTSUPP
}

func Recvfrom(fd int, p []byte, flags int) (n int, from Sockaddr, err error) {
	println("Recvfrom not implemented", fd, p, flags)
	return n, from, EOPNOTSUPP
}

func Recvmsg(fd int, p, oob []byte, flags int) (n, oobn int, recvflags int, from Sockaddr, err error) {
	println("Recvmsg not implemented", fd, p, oob, flags)
	return n, oobn, recvflags, from, EOPNOTSUPP
}

func SendmsgN(fd int, p, oob []byte, to Sockaddr, flags int) (n int, err error) {
	println("SendmsgN not implemented", fd, p, oob, to, flags)
	return n, EOPNOTSUPP
}

func Sendto(fd int, p []byte, flags int, to Sockaddr) (err error) {
	println("Sendto not implemented", fd, p, flags)
	return EOPNOTSUPP
}

func ReadDirent(fd int, buf []byte) (n int, err error) {
	println("ReadDirent not implemented", fd, buf)
	return n, EOPNOTSUPP
}

type Sockaddr interface {
	sockaddr() (ptr unsafe.Pointer, len _Socklen, err error) // lowercase; only we can define Sockaddrs
}

type SockaddrInet4 struct {
	Port int
	Addr [4]byte
	raw  RawSockaddrInet4
}

type SockaddrInet6 struct {
	Port   int
	ZoneId uint32
	Addr   [16]byte
	raw    RawSockaddrInet6
}

func (sa *SockaddrInet4) sockaddr() (unsafe.Pointer, _Socklen, error) {
	if sa.Port < 0 || sa.Port > 0xFFFF {
		return nil, 0, EINVAL
	}
	sa.raw.Family = AF_INET
	p := (*[2]byte)(unsafe.Pointer(&sa.raw.Port))
	p[0] = byte(sa.Port >> 8)
	p[1] = byte(sa.Port)
	sa.raw.Addr = sa.Addr
	return unsafe.Pointer(&sa.raw), SizeofSockaddrInet4, nil
}

func (sa *SockaddrInet6) sockaddr() (unsafe.Pointer, _Socklen, error) {
	if sa.Port < 0 || sa.Port > 0xFFFF {
		return nil, 0, EINVAL
	}
	sa.raw.Family = AF_INET6
	p := (*[2]byte)(unsafe.Pointer(&sa.raw.Port))
	p[0] = byte(sa.Port >> 8)
	p[1] = byte(sa.Port)
	sa.raw.Scope_id = sa.ZoneId
	sa.raw.Addr = sa.Addr
	return unsafe.Pointer(&sa.raw), SizeofSockaddrInet6, nil
}

type SockaddrUnix struct {
	Name string
	raw  RawSockaddrUnix
}

func (sa *SockaddrUnix) sockaddr() (unsafe.Pointer, _Socklen, error) {
	name := sa.Name
	n := len(name)
	if n > len(sa.raw.Path) {
		return nil, 0, EINVAL
	}
	if n == len(sa.raw.Path) && name[0] != '@' {
		return nil, 0, EINVAL
	}
	sa.raw.Family = AF_UNIX
	for i := 0; i < n; i++ {
		sa.raw.Path[i] = int8(name[i])
	}
	// length is family (uint16), name, NUL.
	sl := _Socklen(2)
	if n > 0 {
		sl += _Socklen(n) + 1
	}
	if sa.raw.Path[0] == '@' || (sa.raw.Path[0] == 0 && sl > 3) {
		// Check sl > 3 so we don't change unnamed socket behavior.
		sa.raw.Path[0] = 0
		// Don't count trailing NUL for abstract address.
		sl--
	}

	return unsafe.Pointer(&sa.raw), sl, nil
}

type SockaddrLinklayer struct {
	Protocol uint16
	Ifindex  int
	Hatype   uint16
	Pkttype  uint8
	Halen    uint8
	Addr     [8]byte
	raw      RawSockaddrLinklayer
}

func (sa *SockaddrLinklayer) sockaddr() (unsafe.Pointer, _Socklen, error) {
	if sa.Ifindex < 0 || sa.Ifindex > 0x7fffffff {
		return nil, 0, EINVAL
	}
	sa.raw.Family = AF_PACKET
	sa.raw.Protocol = sa.Protocol
	sa.raw.Ifindex = int32(sa.Ifindex)
	sa.raw.Hatype = sa.Hatype
	sa.raw.Pkttype = sa.Pkttype
	sa.raw.Halen = sa.Halen
	sa.raw.Addr = sa.Addr
	return unsafe.Pointer(&sa.raw), SizeofSockaddrLinklayer, nil
}

type SockaddrNetlink struct {
	Family uint16
	Pad    uint16
	Pid    uint32
	Groups uint32
	raw    RawSockaddrNetlink
}

func (sa *SockaddrNetlink) sockaddr() (unsafe.Pointer, _Socklen, error) {
	sa.raw.Family = AF_NETLINK
	sa.raw.Pad = sa.Pad
	sa.raw.Pid = sa.Pid
	sa.raw.Groups = sa.Groups
	return unsafe.Pointer(&sa.raw), SizeofSockaddrNetlink, nil
}

func anyToSockaddr(rsa *RawSockaddrAny) (Sockaddr, error) {
	switch rsa.Addr.Family {
	case AF_NETLINK:
		pp := (*RawSockaddrNetlink)(unsafe.Pointer(rsa))
		sa := new(SockaddrNetlink)
		sa.Family = pp.Family
		sa.Pad = pp.Pad
		sa.Pid = pp.Pid
		sa.Groups = pp.Groups
		return sa, nil

	case AF_PACKET:
		pp := (*RawSockaddrLinklayer)(unsafe.Pointer(rsa))
		sa := new(SockaddrLinklayer)
		sa.Protocol = pp.Protocol
		sa.Ifindex = int(pp.Ifindex)
		sa.Hatype = pp.Hatype
		sa.Pkttype = pp.Pkttype
		sa.Halen = pp.Halen
		sa.Addr = pp.Addr
		return sa, nil

	case AF_UNIX:
		pp := (*RawSockaddrUnix)(unsafe.Pointer(rsa))
		sa := new(SockaddrUnix)
		if pp.Path[0] == 0 {
			// "Abstract" Unix domain socket.
			// Rewrite leading NUL as @ for textual display.
			// (This is the standard convention.)
			// Not friendly to overwrite in place,
			// but the callers below don't care.
			pp.Path[0] = '@'
		}

		// Assume path ends at NUL.
		// This is not technically the Linux semantics for
		// abstract Unix domain sockets--they are supposed
		// to be uninterpreted fixed-size binary blobs--but
		// everyone uses this convention.
		n := 0
		for n < len(pp.Path) && pp.Path[n] != 0 {
			n++
		}
		sa.Name = string(unsafe.Slice((*byte)(unsafe.Pointer(&pp.Path[0])), n))
		return sa, nil

	case AF_INET:
		pp := (*RawSockaddrInet4)(unsafe.Pointer(rsa))
		sa := new(SockaddrInet4)
		p := (*[2]byte)(unsafe.Pointer(&pp.Port))
		sa.Port = int(p[0])<<8 + int(p[1])
		sa.Addr = pp.Addr
		return sa, nil

	case AF_INET6:
		pp := (*RawSockaddrInet6)(unsafe.Pointer(rsa))
		sa := new(SockaddrInet6)
		p := (*[2]byte)(unsafe.Pointer(&pp.Port))
		sa.Port = int(p[0])<<8 + int(p[1])
		sa.ZoneId = pp.Scope_id
		sa.Addr = pp.Addr
		return sa, nil
	}
	return nil, EAFNOSUPPORT
}

func SetsockoptByte(fd, level, opt int, value byte) (err error) {
	println("SetsockoptByte not implemented", fd, level, opt, value)
	return EOPNOTSUPP
}

func SetsockoptInt(fd, level, opt int, value int) (err error) {
	println("SetsockoptInt not implemented", fd, level, opt, value)
	return EOPNOTSUPP
}

func SetsockoptInet4Addr(fd, level, opt int, value [4]byte) (err error) {
	println("SetsockoptInet4Addr not implemented", fd, level, opt)
	return EOPNOTSUPP
}

func SetsockoptIPMreq(fd, level, opt int, mreq *IPMreq) (err error) {
	println("SetsockoptIPMreq not implemented", fd, level, opt, mreq)
	return EOPNOTSUPP
}

func SetsockoptIPv6Mreq(fd, level, opt int, mreq *IPv6Mreq) (err error) {
	println("SetsockoptIPv6Mreq not implemented", fd, level, opt, mreq)
	return EOPNOTSUPP
}

func SetsockoptLinger(fd, level, opt int, l *Linger) (err error) {
	println("SetsockoptLinger not implemented", fd, level, opt, l)
	return EOPNOTSUPP
}

func SetsockoptIPMreqn(fd, level, opt int, mreq *IPMreqn) (err error) {
	println("SetsockoptIPMreqn not implemented", fd, level, opt, mreq)
	return EOPNOTSUPP
}

func Socket(domain, typ, proto int) (fd int, err error) {
	println("Socket not implemented", domain, typ, proto)
	return fd, EOPNOTSUPP
}

func Sendfile(outfd int, infd int, offset *int64, count int) (written int, err error) {
	println("Sendfile not implemented", outfd, infd, offset, count)
	return written, EOPNOTSUPP
}

func Bind(fd int, sa Sockaddr) (err error) {
	println("Bind not implemented", fd, sa)
	return EOPNOTSUPP
}

func Connect(fd int, sa Sockaddr) (err error) {
	println("Connect not implemented", fd, sa)
	return EOPNOTSUPP
}

func Getpeername(fd int) (sa Sockaddr, err error) {
	println("Getpeername not implemented", fd)
	return sa, EOPNOTSUPP
}

func Getsockname(fd int) (sa Sockaddr, err error) {
	println("Getsockname not implemented", fd)
	return sa, EOPNOTSUPP
}

func Pipe2(p []int, flags int) error {
	println("Pipe2 not implemented", p, flags)
	return EOPNOTSUPP
}

func Unlink(path string) error {
	println("Unlink not implemented", path)
	return EOPNOTSUPP
}

func Open(path string, mode int, perm uint32) (fd int, err error) {
	println("Open not implemented", path, mode, perm)
	return fd, EOPNOTSUPP
}

// Do the interface allocations only once for common
// Errno values.
var (
	errEAGAIN error = EAGAIN
	errEINVAL error = EINVAL
	errENOENT error = ENOENT
)

// errnoErr returns common boxed Errno values, to prevent
// allocations at runtime.
func errnoErr(e Errno) error {
	switch e {
	case 0:
		return nil
	case EAGAIN:
		return errEAGAIN
	case EINVAL:
		return errEINVAL
	case ENOENT:
		return errENOENT
	}
	return e
}

// A Signal is a number describing a process signal.
// It implements the os.Signal interface.
type Signal int

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
