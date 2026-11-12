//go:build !baremetal && !tinygo.wasm && !nintendoswitch

package os

import (
	"internal/poll"
	"runtime"
	"syscall"
	"time"
)

func init() {
	ErrDeadlineExceeded = poll.ErrDeadlineExceeded
}

type pollFD struct {
	fd *poll.FD
}

func (p pollFD) Close() error {
	return p.fd.Close()
}

func (p pollFD) Exist() bool {
	return p.fd != nil
}

func (f *File) ensurePollFD() (*poll.FD, error) {
	if f.handle == nil {
		return nil, ErrClosed
	}
	if f.pfd.fd != nil {
		return f.pfd.fd, nil
	}
	h, ok := f.handle.(fd)
	if !ok {
		return nil, ErrNotImplemented
	}
	pfd := &poll.FD{
		Sysfd:         int(h.Fd()),
		IsStream:      true,
		ZeroReadIsEOF: true,
	}
	_ = pfd.Init("file", true)
	f.pfd.fd = pfd
	return pfd, nil
}

func (f *File) read(b []byte) (int, error) {
	if f.pfd.fd != nil {
		return f.pfd.fd.Read(b)
	}
	n, err := f.handle.Read(b)
	if err == syscall.EAGAIN {
		if pfd, perr := f.ensurePollFD(); perr == nil {
			return pfd.Read(b)
		}
	}
	return n, err
}

func (f *File) write(b []byte) (int, error) {
	if f.pfd.fd != nil {
		return f.pfd.fd.Write(b)
	}
	n, err := f.handle.Write(b)
	if err == syscall.EAGAIN {
		if pfd, perr := f.ensurePollFD(); perr == nil {
			return pfd.Write(b)
		}
	}
	return n, err
}

type rawConn struct {
	file *File
}

func (c *rawConn) Control(fn func(uintptr)) error {
	pfd, err := c.file.ensurePollFD()
	if err != nil {
		return err
	}
	err = pfd.RawControl(fn)
	runtime.KeepAlive(c.file)
	return err
}

func (c *rawConn) Read(fn func(uintptr) bool) error {
	pfd, err := c.file.ensurePollFD()
	if err != nil {
		return err
	}
	err = pfd.RawRead(fn)
	runtime.KeepAlive(c.file)
	return err
}

func (c *rawConn) Write(fn func(uintptr) bool) error {
	pfd, err := c.file.ensurePollFD()
	if err != nil {
		return err
	}
	err = pfd.RawWrite(fn)
	runtime.KeepAlive(c.file)
	return err
}

func (f *File) syscallConn() (syscall.RawConn, error) {
	if _, err := f.ensurePollFD(); err != nil {
		return nil, err
	}
	return &rawConn{file: f}, nil
}

func (f *File) setDeadline(t time.Time) error {
	pfd, err := f.ensurePollFD()
	if err != nil {
		return err
	}
	return pfd.SetDeadline(t)
}

func (f *File) setReadDeadline(t time.Time) error {
	pfd, err := f.ensurePollFD()
	if err != nil {
		return err
	}
	return pfd.SetReadDeadline(t)
}

func (f *File) setWriteDeadline(t time.Time) error {
	pfd, err := f.ensurePollFD()
	if err != nil {
		return err
	}
	return pfd.SetWriteDeadline(t)
}
