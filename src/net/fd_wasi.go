//go:build wasi
// +build wasi

package net

import (
	"os"
	"syscall"
	"time"
)

// Network file descriptor.
type netFD struct {
	*os.File

	net   string
	laddr Addr
	raddr Addr
}

// Read implements the Conn Read method.
func (c *netFD) Read(b []byte) (int, error) {
	// TODO: Handle EAGAIN and perform poll
	return c.File.Read(b)
}

// Write implements the Conn Write method.
func (c *netFD) Write(b []byte) (int, error) {
	// TODO: Handle EAGAIN and perform poll
	return c.File.Write(b)
}

// LocalAddr implements the Conn LocalAddr method.
func (*netFD) LocalAddr() Addr {
	return nil
}

// RemoteAddr implements the Conn RemoteAddr method.
func (*netFD) RemoteAddr() Addr {
	return nil
}

// SetDeadline implements the Conn SetDeadline method.
func (*netFD) SetDeadline(t time.Time) error {
	return ErrNotImplemented
}

// SetReadDeadline implements the Conn SetReadDeadline method.
func (*netFD) SetReadDeadline(t time.Time) error {
	return ErrNotImplemented
}

// SetWriteDeadline implements the Conn SetWriteDeadline method.
func (*netFD) SetWriteDeadline(t time.Time) error {
	return ErrNotImplemented
}

type listener struct {
	*os.File
}

// Accept implements the Listener Accept method.
func (l *listener) Accept() (Conn, error) {
	fd, err := syscall.SockAccept(int(l.File.Fd()), syscall.O_NONBLOCK)
	if err != nil {
		return nil, wrapSyscallError("sock_accept", err)
	}
	return &netFD{File: os.NewFile(uintptr(fd), "conn"), net: "file+net", laddr: nil, raddr: nil}, nil
}

// Addr implements the Listener Addr method.
func (*listener) Addr() Addr {
	return nil
}

func fileListener(f *os.File) (Listener, error) {
	return &listener{File: f}, nil
}
