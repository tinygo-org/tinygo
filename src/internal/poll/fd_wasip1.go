//go:build wasip1

// Package poll is a minimal subset of upstream Go's internal/poll, scoped
// to what is needed to back a wasip1 net implementation on top of
// TinyGo's cooperative-scheduler netpoll integration.
//
// On wasip1 the cooperative scheduler integrates poll_oneoff with FD
// waiters (see runtime/netpoll_wasip1.go and syscall/syscall_libc_wasip1.go).
// This package wraps the syscall layer to:
//
//   - own the O_NONBLOCK policy decision (set on Init for pollable FDs),
//     unblocking the EAGAIN→park retry loop that syscall.Read already has;
//   - provide a Go-shaped FD type that net.* can use without reaching
//     into syscall directly;
//   - thread per-FD read/write deadlines through a runtime helper that
//     lets a time.AfterFunc callback wake the parked goroutine;
//   - dispatch socket FDs through wasi sock_recv / sock_send / sock_accept
//     / sock_shutdown so net.Conn/net.Listener (via upstream Go's
//     net/file_wasip1.go) work end-to-end.
package poll

import (
	"errors"
	"internal/task"
	"syscall"
	"time"
	"unsafe"
)

// ErrFileClosing is returned when a Read or Write is started on a closed FD.
var ErrFileClosing = errors.New("use of closed file")

// ErrNetClosing is returned for network operations on a closed FD.
var ErrNetClosing = errors.New("use of closed network connection")

// ErrDeadlineExceeded is returned by Read/Write when a deadline expired.
// Matches the error returned by os.IsTimeout-style helpers.
var ErrDeadlineExceeded = errors.New("i/o timeout")

// ErrNoDeadline is returned if SetDeadline is called on an FD whose
// underlying type does not support deadlines.
var ErrNoDeadline = errors.New("file type does not support deadline")

// pollMode constants must mirror runtime/netpoll_wasip1.go's pollRead/
// pollWrite values.
const (
	pollModeRead  uint8 = 1
	pollModeWrite uint8 = 2
)

//go:linkname runtime_netpoll_addwait runtime.runtime_netpoll_addwait
func runtime_netpoll_addwait(fd uint32, mode uint8) uintptr

//go:linkname runtime_netpoll_done runtime.runtime_netpoll_done
func runtime_netpoll_done(pd uintptr)

//go:linkname runtime_netpoll_pdfired runtime.runtime_netpoll_pdfired
func runtime_netpoll_pdfired(pd uintptr) bool

//go:linkname runtime_netpoll_wake runtime.runtime_netpoll_wake
func runtime_netpoll_wake(pd uintptr)

//go:linkname fd_fdstat_get_type syscall.fd_fdstat_get_type
func fd_fdstat_get_type(fd int) (syscall.Filetype, error)

// wasiIovec / wasiCiovec mirror wasi-snapshot-preview1's iovec / ciovec
// records: a buffer pointer plus a 32-bit length. Marshalled inline with
// each fd_read / fd_write / sock_recv / sock_send call.
type wasiIovec struct {
	buf    *byte
	bufLen uint32
}

// fd_read / fd_write are bound directly here rather than going through
// wasi-libc so the deadline-aware Read/Write loop can observe EAGAIN
// cleanly. Outside package runtime, TinyGo's wasmimport binding
// restricts us to unsafe.Pointer for struct args and uint32 for the
// errno result; the wasi spec uses *iovec and uint16 respectively but
// the wire layout is identical.
//
//go:wasmimport wasi_snapshot_preview1 fd_read
func wasi_fd_read(fd int32, iovs unsafe.Pointer, iovsLen uint32, nread unsafe.Pointer) uint32

//go:wasmimport wasi_snapshot_preview1 fd_write
func wasi_fd_write(fd int32, iovs unsafe.Pointer, iovsLen uint32, nwritten unsafe.Pointer) uint32

// sock_recv / sock_send are the socket-data wasi syscalls used when the
// FD's Filetype indicates a socket. We bind them here for the same
// reason as fd_read / fd_write: the deadline-aware loop wants direct
// access to the EAGAIN signal without libc's translation layer.
//
//go:wasmimport wasi_snapshot_preview1 sock_recv
func wasi_sock_recv(fd int32, riData unsafe.Pointer, riDataLen uint32, riFlags uint32, roDatalen unsafe.Pointer, roFlags unsafe.Pointer) uint32

//go:wasmimport wasi_snapshot_preview1 sock_send
func wasi_sock_send(fd int32, siData unsafe.Pointer, siDataLen uint32, siFlags uint32, soDatalen unsafe.Pointer) uint32

// SysFile carries per-FD bookkeeping that upstream Go's poll.FD uses to
// share an underlying syscall FD between an os.File and a net.Conn (see
// Copy below). RefCountPtr / RefCount handle the shared-ownership case;
// Filetype caches the wasi filetype so socket-vs-file dispatch in Read /
// Write is a single integer compare on the hot path.
//
// wasip1 is single-threaded, so the refcount is a plain int — no atomics
// needed. Match upstream's field naming for source-level compatibility
// with code that constructs FDs via struct literal (e.g. upstream's
// net/fd_fake.go).
type SysFile struct {
	RefCountPtr *int32
	RefCount    int32
	Filetype    uint32
}

// init lazily allocates the refcount the first time it's needed (i.e.
// the first Init or Copy on this FD). A zero SysFile starts at refcount
// 1 — the FD's sole owner is the caller.
func (s *SysFile) init() {
	if s.RefCountPtr == nil {
		s.RefCount = 1
		s.RefCountPtr = &s.RefCount
	}
}

// ref increments the shared refcount and returns a SysFile that points
// at the same counter. Used by FD.Copy.
func (s *SysFile) ref() SysFile {
	s.init()
	*s.RefCountPtr++
	return SysFile{RefCountPtr: s.RefCountPtr, Filetype: s.Filetype}
}

// destroy decrements the refcount and reports whether the underlying
// syscall FD should now be closed (i.e. this was the last owner).
func (s *SysFile) destroy() bool {
	if s.RefCountPtr == nil {
		return true
	}
	*s.RefCountPtr--
	return *s.RefCountPtr <= 0
}

// FD is the wasip1 file/socket descriptor wrapped with the bookkeeping
// that net and os rely on. It owns the lifecycle of the underlying
// syscall FD (modulo the Copy / refcount handoff between os.File and
// net.Conn).
//
// The struct mirrors upstream Go's internal/poll.FD field naming
// (Sysfd, IsStream, ZeroReadIsEOF, SysFile) so that upstream's
// net/file_wasip1.go and net/fd_fake.go can construct one via struct
// literal without modification.
type FD struct {
	Sysfd         int
	SysFile       SysFile
	IsStream      bool
	ZeroReadIsEOF bool
	closed        bool

	// Per-FD deadlines, zero means "no deadline". Subsequent Read/Write
	// calls observe whatever the current value is at call time; an
	// in-flight call uses the deadline it captured at its start.
	rDeadline time.Time
	wDeadline time.Time
}

// Init readies the FD for use. When pollable is true (i.e. the FD might
// block — sockets, pipes, FIFOs), Init sets O_NONBLOCK so that
// Read/Write enter the EAGAIN→park retry loop instead of blocking the
// entire wasm module.
//
// Init also caches the wasi filetype so the Read/Write hot path can
// dispatch socket vs file with a single integer compare.
//
// The net argument is currently ignored but kept for parity with
// upstream Go.
func (fd *FD) Init(net string, pollable bool) error {
	_ = net
	fd.SysFile.init()
	if ft, err := fd_fdstat_get_type(fd.Sysfd); err == nil {
		fd.SysFile.Filetype = uint32(ft)
	}
	if !pollable {
		return nil
	}
	flags, err := syscall.Fcntl(fd.Sysfd, syscall.F_GETFL, 0)
	if err != nil {
		return err
	}
	if flags&syscall.O_NONBLOCK != 0 {
		return nil
	}
	if _, err := syscall.Fcntl(fd.Sysfd, syscall.F_SETFL, flags|syscall.O_NONBLOCK); err != nil {
		return err
	}
	return nil
}

// Copy returns a duplicate FD that shares the underlying Sysfd through
// the SysFile refcount. The original and the copy can independently
// call Close — only the last one actually issues the syscall. Used by
// upstream net/file_wasip1.go to hand a socket FD off from an os.File
// to a net.Listener / net.Conn.
func (fd *FD) Copy() FD {
	return FD{
		Sysfd:         fd.Sysfd,
		SysFile:       fd.SysFile.ref(),
		IsStream:      fd.IsStream,
		ZeroReadIsEOF: fd.ZeroReadIsEOF,
	}
}

// Close marks the FD closed. The underlying syscall FD is only released
// when the refcount drops to zero — earlier Close calls (e.g. on the
// os.File side after a successful net.FileConn handoff) just decrement.
func (fd *FD) Close() error {
	if fd.closed {
		return ErrFileClosing
	}
	fd.closed = true
	if !fd.SysFile.destroy() {
		return nil
	}
	return syscall.Close(fd.Sysfd)
}

// Exist reports whether fd points to an actual FD wrapper (non-nil).
// Callers in package os hold a *FD via a per-target alias and need to
// nil-check without depending on the concrete type being a pointer.
func (fd *FD) Exist() bool { return fd != nil }

// CloseFunc is the hook upstream net's poll package exposes so tests
// can intercept Close. We just point at syscall.Close.
var CloseFunc func(int) error = syscall.Close

// String is an internal string definition for methods/functions that
// shouldn't be used outside the stdlib. Upstream net references it
// from rawconn.go to mark methods as not-for-external-use.
type String string

// RawControl / RawRead / RawWrite back syscall.RawConn's three callback
// methods. They invoke f with the underlying FD; the bool return of
// RawRead / RawWrite controls retry-on-EAGAIN, which we implement by
// parking the goroutine on the netpoll registry and retrying until f
// returns true (the same loop upstream uses).
func (fd *FD) RawControl(f func(uintptr)) error {
	if fd.closed {
		return ErrFileClosing
	}
	f(uintptr(fd.Sysfd))
	return nil
}

func (fd *FD) RawRead(f func(uintptr) bool) error {
	if fd.closed {
		return ErrFileClosing
	}
	for {
		if f(uintptr(fd.Sysfd)) {
			return nil
		}
		wait(fd.Sysfd, pollModeRead)
	}
}

func (fd *FD) RawWrite(f func(uintptr) bool) error {
	if fd.closed {
		return ErrFileClosing
	}
	for {
		if f(uintptr(fd.Sysfd)) {
			return nil
		}
		wait(fd.Sysfd, pollModeWrite)
	}
}

// Shutdown calls wasi sock_shutdown. how is one of syscall.SHUT_RD,
// SHUT_WR, SHUT_RDWR.
func (fd *FD) Shutdown(how int) error {
	return syscall.Shutdown(fd.Sysfd, how)
}

// Accept loops over wasi sock_accept, parking the goroutine on EAGAIN
// (waiting for a new connection) through the netpoll registry. Returns
// (newfd, sockaddr=nil, errcall, error) — sockaddr is always nil because
// wasi sock_accept doesn't return one.
func (fd *FD) Accept() (int, syscall.Sockaddr, string, error) {
	deadline := fd.rDeadline
	for {
		if !deadline.IsZero() && !time.Now().Before(deadline) {
			return -1, nil, "accept", ErrDeadlineExceeded
		}
		nfd, _, err := syscall.Accept(fd.Sysfd)
		if err == nil {
			return nfd, nil, "", nil
		}
		if err == syscall.EINTR {
			continue
		}
		if err != syscall.EAGAIN {
			return -1, nil, "accept", err
		}
		if deadline.IsZero() {
			wait(fd.Sysfd, pollModeRead)
		} else {
			if perr := fd.parkUntil(pollModeRead, deadline); perr != nil {
				return -1, nil, "accept", perr
			}
		}
	}
}

// isSocket reports whether the cached Filetype is a stream or datagram
// socket. False if Init was never called (Filetype stays 0 = UNKNOWN).
func (fd *FD) isSocket() bool {
	ft := syscall.Filetype(fd.SysFile.Filetype)
	return ft == syscall.FILETYPE_SOCKET_STREAM || ft == syscall.FILETYPE_SOCKET_DGRAM
}

// Read reads from the FD into p. Sockets dispatch to sock_recv (which
// honours fd.rDeadline directly); regular files/pipes go through
// syscall.Read for the no-deadline fast path or readWithDeadline when
// a deadline is set.
func (fd *FD) Read(p []byte) (int, error) {
	if fd.closed {
		return 0, ErrFileClosing
	}
	if len(p) == 0 {
		return 0, nil
	}
	if fd.isSocket() {
		return fd.sockRecv(p)
	}
	if fd.rDeadline.IsZero() {
		return syscall.Read(fd.Sysfd, p)
	}
	return fd.readWithDeadline(p)
}

// Write writes p to the FD. Sockets dispatch to sock_send. Regular
// files / pipes go through syscall.Write or writeWithDeadline.
func (fd *FD) Write(p []byte) (int, error) {
	if fd.closed {
		return 0, ErrFileClosing
	}
	if fd.isSocket() {
		return fd.sockSend(p)
	}
	if fd.wDeadline.IsZero() {
		return syscall.Write(fd.Sysfd, p)
	}
	return fd.writeWithDeadline(p)
}

// Pread reads from the FD at the given offset. Always file semantics —
// sockets aren't seekable so this never goes through sock_recv.
func (fd *FD) Pread(p []byte, off int64) (int, error) {
	if fd.closed {
		return 0, ErrFileClosing
	}
	return syscall.Pread(fd.Sysfd, p, off)
}

// Pwrite writes to the FD at the given offset.
func (fd *FD) Pwrite(p []byte, off int64) (int, error) {
	if fd.closed {
		return 0, ErrFileClosing
	}
	return syscall.Pwrite(fd.Sysfd, p, off)
}

// SetDeadline sets both the read and write deadlines.
func (fd *FD) SetDeadline(t time.Time) error {
	fd.rDeadline = t
	fd.wDeadline = t
	return nil
}

// SetReadDeadline sets the deadline for future Read calls. A zero t
// clears the deadline.
func (fd *FD) SetReadDeadline(t time.Time) error {
	fd.rDeadline = t
	return nil
}

// SetWriteDeadline sets the deadline for future Write calls.
func (fd *FD) SetWriteDeadline(t time.Time) error {
	fd.wDeadline = t
	return nil
}

// readWithDeadline implements the EAGAIN→park retry loop with deadline
// cancellation for non-socket FDs. The deadline is captured at function
// entry; later SetReadDeadline calls don't affect this in-flight Read.
func (fd *FD) readWithDeadline(p []byte) (int, error) {
	deadline := fd.rDeadline
	iov := wasiIovec{buf: &p[0], bufLen: uint32(len(p))}
	for {
		if !time.Now().Before(deadline) {
			return 0, ErrDeadlineExceeded
		}
		var n uint32
		errno := wasi_fd_read(int32(fd.Sysfd), unsafe.Pointer(&iov), 1, unsafe.Pointer(&n))
		switch errno {
		case 0:
			return int(n), nil
		case wasiErrnoIntr:
			continue
		case wasiErrnoAgain:
			if err := fd.parkUntil(pollModeRead, deadline); err != nil {
				return 0, err
			}
		default:
			return 0, syscall.Errno(errno)
		}
	}
}

func (fd *FD) writeWithDeadline(p []byte) (int, error) {
	deadline := fd.wDeadline
	var nn int
	for {
		if !time.Now().Before(deadline) {
			return nn, ErrDeadlineExceeded
		}
		if nn == len(p) {
			return nn, nil
		}
		buf := p[nn:]
		iov := wasiIovec{buf: &buf[0], bufLen: uint32(len(buf))}
		var n uint32
		errno := wasi_fd_write(int32(fd.Sysfd), unsafe.Pointer(&iov), 1, unsafe.Pointer(&n))
		switch errno {
		case 0:
			nn += int(n)
		case wasiErrnoIntr:
			// retry
		case wasiErrnoAgain:
			if err := fd.parkUntil(pollModeWrite, deadline); err != nil {
				return nn, err
			}
		default:
			return nn, syscall.Errno(errno)
		}
	}
}

// sockRecv is the socket sibling of readWithDeadline / syscall.Read.
// Always issues sock_recv, regardless of deadline state, so callers
// that don't want fd_read on a socket get a guaranteed sock_recv path.
func (fd *FD) sockRecv(p []byte) (int, error) {
	deadline := fd.rDeadline
	iov := wasiIovec{buf: &p[0], bufLen: uint32(len(p))}
	for {
		if !deadline.IsZero() && !time.Now().Before(deadline) {
			return 0, ErrDeadlineExceeded
		}
		var n uint32
		var roFlags uint32
		errno := wasi_sock_recv(int32(fd.Sysfd), unsafe.Pointer(&iov), 1, 0, unsafe.Pointer(&n), unsafe.Pointer(&roFlags))
		switch errno {
		case 0:
			return int(n), nil
		case wasiErrnoIntr:
			continue
		case wasiErrnoAgain:
			if deadline.IsZero() {
				wait(fd.Sysfd, pollModeRead)
			} else if err := fd.parkUntil(pollModeRead, deadline); err != nil {
				return 0, err
			}
		default:
			return 0, syscall.Errno(errno)
		}
	}
}

func (fd *FD) sockSend(p []byte) (int, error) {
	deadline := fd.wDeadline
	var nn int
	for {
		if !deadline.IsZero() && !time.Now().Before(deadline) {
			return nn, ErrDeadlineExceeded
		}
		if nn == len(p) {
			return nn, nil
		}
		buf := p[nn:]
		iov := wasiIovec{buf: &buf[0], bufLen: uint32(len(buf))}
		var n uint32
		errno := wasi_sock_send(int32(fd.Sysfd), unsafe.Pointer(&iov), 1, 0, unsafe.Pointer(&n))
		switch errno {
		case 0:
			nn += int(n)
		case wasiErrnoIntr:
			// retry
		case wasiErrnoAgain:
			if deadline.IsZero() {
				wait(fd.Sysfd, pollModeWrite)
			} else if err := fd.parkUntil(pollModeWrite, deadline); err != nil {
				return nn, err
			}
		default:
			return nn, syscall.Errno(errno)
		}
	}
}

// wasi snapshot preview1 errno values, kept here so the read/write loops
// can compare without dragging in the full syscall errno table for a
// switch on three constants. These match the wasi spec exactly; see also
// syscall_libc_wasi.go in package syscall.
const (
	wasiErrnoAgain uint32 = 6
	wasiErrnoIntr  uint32 = 27
)

// wait parks the current goroutine until the FD becomes ready in the
// given direction. No deadline; mirrors the helper of the same name in
// package syscall (intentionally duplicated rather than linknamed
// across the package boundary — see project memory on shim avoidance).
func wait(fd int, mode uint8) {
	pd := runtime_netpoll_addwait(uint32(fd), mode)
	task.Pause()
	runtime_netpoll_done(pd)
}

// parkUntil parks the current goroutine on (fd, mode) with a deadline.
// Returns nil if the FD became ready or the timer fired (caller's loop
// re-checks the deadline at top); ErrDeadlineExceeded if the deadline
// was already in the past.
//
// Race handling: the deadline timer's callback and pollIO's event walk
// can both target the same pollDesc. The pd.fired flag guards against
// double-pushing the task to the run queue; whichever arrives second
// is a no-op.
func (fd *FD) parkUntil(mode uint8, deadline time.Time) error {
	d := time.Until(deadline)
	if d <= 0 {
		return ErrDeadlineExceeded
	}
	pd := runtime_netpoll_addwait(uint32(fd.Sysfd), mode)
	timer := time.AfterFunc(d, func() {
		runtime_netpoll_wake(pd)
	})
	task.Pause()
	timer.Stop()
	runtime_netpoll_done(pd)
	return nil
}
