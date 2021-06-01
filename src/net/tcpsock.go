package net

// TCPConn is an implementation of the Conn interface for TCP network
// connections.
type TCPConn struct {
	conn
}

func (c *TCPConn) CloseWrite() error {
	return &OpError{"close", "", nil, nil, ErrNotImplemented}
}
