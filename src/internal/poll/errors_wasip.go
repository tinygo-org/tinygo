//go:build wasip1 || wasip2

// Shared error sentinels for the wasip1 and wasip2 internal/poll
// implementations. The error values are part of the public API surface
// upstream net relies on; sharing them keeps fd_wasip1.go and
// fd_wasip2.go free of redundant declarations.

package poll

import "errors"

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
