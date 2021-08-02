package net

import "errors"

var (
	// copied from poll.ErrNetClosing
	errClosed = errors.New("use of closed network connection")

	ErrNotImplemented = errors.New("operation not implemented")
)
