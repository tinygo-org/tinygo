//go:build wasi
// +build wasi

package net

import (
	"io"
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
