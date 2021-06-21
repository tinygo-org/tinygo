package net

import (
	"context"
	"time"
)

type Dialer struct {
	Timeout   time.Duration
	Deadline  time.Time
	DualStack bool
	KeepAlive time.Duration
}

func Dial(network, address string) (Conn, error) {
	return nil, ErrNotImplemented
}

func Listen(network, address string) (Listener, error) {
	return nil, ErrNotImplemented
}

func (d *Dialer) DialContext(ctx context.Context, network, address string) (Conn, error) {
	return nil, ErrNotImplemented
}
