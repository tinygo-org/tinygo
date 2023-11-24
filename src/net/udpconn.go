package net

import (
	"syscall"
	"time"
)

// UDPConn implements UDPConn by always failing.
type UDPConn struct {
}

func ListenUDP(_ string, _ *UDPAddr) (*UDPConn, error) {
	return nil, ErrNotImplemented
}

func ResolveUDPAddr(network, address string) (*UDPAddr, error) {
	return nil, ErrNotImplemented
}

// Read implements the Conn Read method
func (c *UDPConn) Read(b []byte) (n int, err error) {
	return 0, ErrNotImplemented
}

// Write implements the Conn Write method
func (c *UDPConn) Write(b []byte) (n int, err error) {
	return 0, ErrNotImplemented
}

// Close implements the Conn Close method
func (c *UDPConn) Close() error {
	return ErrNotImplemented
}

// LocalAddr implements the Conn LocalAddr method
func (c *UDPConn) LocalAddr() Addr {
	return nil
}

// RemoteAddr implements the Conn RemoteAddr method
func (c *UDPConn) RemoteAddr() Addr {
	return nil
}

// SetReadDeadline implements the Conn SetReadDeadline method
func (c *UDPConn) SetReadDeadline(t time.Time) error {
	return ErrNotImplemented
}

// SetWriteDeadline implements the Conn SetWriteDeadline method
func (c *UDPConn) SetWriteDeadline(t time.Time) error {
	return ErrNotImplemented
}

// SetDeadline implements the Conn SetDeadline method
func (c *UDPConn) SetDeadline(t time.Time) error {
	return ErrNotImplemented
}

// SetReadBuffer implements the Conn SetReadBuffer method
func (c *UDPConn) SetReadBuffer(bytes int) {
	return
}

// SetWriteBuffer implements the Conn SetWriteBuffer method
func (c *UDPConn) SetWriteBuffer(bytes int) {
	return
}

// SyscallConn implements the Conn SyscallConn method
func (c *UDPConn) SyscallConn() (syscall.RawConn, error) {
	return nil, ErrNotImplemented
}

func (c *UDPConn) WriteTo(b []byte, addr Addr) (n int, err error) {
	return -1, ErrNotImplemented
}

func (c *UDPConn) ReadFrom(p []byte) (n int, addr Addr, err error) {
	return -1, nil, ErrNotImplemented
}

func (c *UDPConn) ReadFromUDP(p []byte) (n int, addr *UDPAddr, err error) {
	return -1, nil, ErrNotImplemented
}
