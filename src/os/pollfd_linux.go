//go:build !baremetal && !tinygo.wasm && !nintendoswitch

package os

import (
	"internal/poll"
	"internal/syscall/unix"
	"runtime"
	"syscall"
	"time"
	_ "unsafe" // for go:linkname
)

func init() {
	ErrNoDeadline = poll.ErrNoDeadline
	ErrDeadlineExceeded = poll.ErrDeadlineExceeded
}

type pollFD struct {
	fd       poll.FD
	nonblock bool
}

func mapPollErr(err error) error {
	if err == poll.ErrFileClosing {
		return ErrClosed
	}
	if errno, ok := err.(syscall.Errno); ok {
		return handleSyscallError(errno)
	}
	return err
}

func (p *pollFD) Close() error {
	return mapPollErr(p.fd.Close())
}

func (p *pollFD) Exist() bool {
	return true
}

type newFileKind int

const (
	kindNewFile newFileKind = iota
	kindOpenFile
	kindPipe
	kindSock
)

func newLinuxFile(fd int, name string, kind newFileKind, nonBlocking bool) *File {
	f := &File{&file{
		name: name,
		pfd: pollFD{
			fd: poll.FD{
				Sysfd:         fd,
				IsStream:      true,
				ZeroReadIsEOF: true,
			},
		},
	}}

	pollable := kind == kindOpenFile || kind == kindPipe || kind == kindSock || nonBlocking
	if pollable && (kind == kindOpenFile || kind == kindNewFile) {
		var st syscall.Stat_t
		err := ignoringEINTR(func() error {
			return syscall.Fstat(fd, &st)
		})
		if err != nil {
			pollable = false
		} else {
			switch st.Mode & syscall.S_IFMT {
			case syscall.S_IFREG, syscall.S_IFDIR, syscall.S_IFBLK:
				pollable = false
			}
		}
	}

	clearNonBlock := false
	if pollable {
		if nonBlocking {
			if kind == kindSock {
				f.pfd.nonblock = true
			}
		} else if syscall.SetNonblock(fd, true) == nil {
			f.pfd.nonblock = true
			clearNonBlock = true
		} else {
			pollable = false
		}
	}

	if err := f.pfd.fd.Init("file", pollable); err != nil && clearNonBlock {
		if syscall.SetNonblock(fd, false) == nil {
			f.pfd.nonblock = false
		}
	}
	return f
}

func openNewFile(handle uintptr, name string, flag int) *File {
	f := newLinuxFile(int(handle), name, kindOpenFile, flag&syscall.O_NONBLOCK != 0)
	f.appendMode = (flag & O_APPEND) != 0
	return f
}

func NewFile(fd uintptr, name string) *File {
	fdi := int(fd)
	if fdi < 0 {
		return nil
	}
	flags, err := unix.Fcntl(fdi, syscall.F_GETFL, 0)
	if err != nil {
		flags = 0
	}
	f := newLinuxFile(fdi, name, kindNewFile, flags&syscall.O_NONBLOCK != 0)
	f.appendMode = flags&syscall.O_APPEND != 0
	return f
}

func Pipe() (r *File, w *File, err error) {
	var p [2]int
	err = handleSyscallError(pipe(p[:]))
	if err != nil {
		return nil, nil, err
	}
	return newLinuxFile(p[0], "|0", kindPipe, false), newLinuxFile(p[1], "|1", kindPipe, false), nil
}

//go:linkname net_newUnixFile net.newUnixFile
func net_newUnixFile(fd int, name string) *File {
	if fd < 0 {
		panic("invalid FD")
	}
	return newLinuxFile(fd, name, kindSock, true)
}

func (f *File) checkClosed() error {
	err := mapPollErr(f.pfd.fd.RawControl(func(uintptr) {}))
	runtime.KeepAlive(f)
	return err
}

func (f *File) closeFile() error {
	return f.pfd.Close()
}

func (f *File) read(b []byte) (int, error) {
	n, err := f.pfd.fd.Read(b)
	runtime.KeepAlive(f)
	return n, mapPollErr(err)
}

func (f *File) pread(b []byte, offset int64) (int, error) {
	n, err := f.pfd.fd.Pread(b, offset)
	runtime.KeepAlive(f)
	return n, mapPollErr(err)
}

func (f *File) write(b []byte) (int, error) {
	n, err := f.pfd.fd.Write(b)
	runtime.KeepAlive(f)
	return n, mapPollErr(err)
}

func (f *File) pwrite(b []byte, offset int64) (int, error) {
	n, err := f.pfd.fd.Pwrite(b, offset)
	runtime.KeepAlive(f)
	return n, mapPollErr(err)
}

func (f *File) seek(offset int64, whence int) (int64, error) {
	if f.dirinfo != nil {
		f.dirinfo = nil
	}
	ret, err := f.pfd.fd.Seek(offset, whence)
	runtime.KeepAlive(f)
	return ret, mapPollErr(err)
}

func (f *File) sync() error {
	err := f.pfd.fd.Fsync()
	runtime.KeepAlive(f)
	if err != nil {
		return &PathError{Op: "sync", Path: f.name, Err: mapPollErr(err)}
	}
	return nil
}

func (f *File) fd() uintptr {
	if f.pfd.nonblock {
		_ = f.pfd.fd.SetBlocking()
	}
	sysfd := ^uintptr(0)
	_ = f.pfd.fd.RawControl(func(fd uintptr) {
		sysfd = fd
	})
	runtime.KeepAlive(f)
	return sysfd
}

func (f *File) Truncate(size int64) error {
	if f == nil {
		return ErrInvalid
	}
	err := f.pfd.fd.Ftruncate(size)
	runtime.KeepAlive(f)
	if err != nil {
		return &PathError{Op: "truncate", Path: f.name, Err: mapPollErr(err)}
	}
	return nil
}

func (f *File) Chown(uid, gid int) error {
	if f == nil {
		return ErrInvalid
	}
	err := f.pfd.fd.Fchown(uid, gid)
	runtime.KeepAlive(f)
	if err != nil {
		return &PathError{Op: "chown", Path: f.name, Err: mapPollErr(err)}
	}
	return nil
}

func (f *File) chmod(mode FileMode) error {
	err := f.pfd.fd.Fchmod(syscallMode(mode))
	runtime.KeepAlive(f)
	if err != nil {
		return &PathError{Op: "chmod", Path: f.name, Err: mapPollErr(err)}
	}
	return nil
}

func (f *File) chdir() error {
	err := f.pfd.fd.Fchdir()
	runtime.KeepAlive(f)
	if err != nil {
		return &PathError{Op: "chdir", Path: f.name, Err: mapPollErr(err)}
	}
	return nil
}

func (f *File) fstat(s *syscall.Stat_t) error {
	err := f.pfd.fd.Fstat(s)
	runtime.KeepAlive(f)
	return mapPollErr(err)
}

type rawConn struct {
	file *File
}

func (c *rawConn) Control(fn func(uintptr)) error {
	if c.file == nil {
		return ErrInvalid
	}
	err := mapPollErr(c.file.pfd.fd.RawControl(fn))
	runtime.KeepAlive(c.file)
	return err
}

func (c *rawConn) Read(fn func(uintptr) bool) error {
	if c.file == nil {
		return ErrInvalid
	}
	err := mapPollErr(c.file.pfd.fd.RawRead(fn))
	runtime.KeepAlive(c.file)
	return err
}

func (c *rawConn) Write(fn func(uintptr) bool) error {
	if c.file == nil {
		return ErrInvalid
	}
	err := mapPollErr(c.file.pfd.fd.RawWrite(fn))
	runtime.KeepAlive(c.file)
	return err
}

func (f *File) syscallConn() (syscall.RawConn, error) {
	if err := f.checkClosed(); err != nil {
		return nil, err
	}
	return &rawConn{file: f}, nil
}

func (f *File) setDeadline(t time.Time) error {
	err := mapPollErr(f.pfd.fd.SetDeadline(t))
	runtime.KeepAlive(f)
	return err
}

func (f *File) setReadDeadline(t time.Time) error {
	err := mapPollErr(f.pfd.fd.SetReadDeadline(t))
	runtime.KeepAlive(f)
	return err
}

func (f *File) setWriteDeadline(t time.Time) error {
	err := mapPollErr(f.pfd.fd.SetWriteDeadline(t))
	runtime.KeepAlive(f)
	return err
}
