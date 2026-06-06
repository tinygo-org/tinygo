//go:build wasip1

package os

import "internal/poll"

// PollFD returns the *poll.FD wrapping this file's underlying syscall
// FD. The first call lazily allocates and caches the *poll.FD on the
// File; subsequent calls return the same pointer so that refcount
// semantics shared with net.FileListener / net.FileConn (via
// poll.FD.Copy) work correctly:
//
//   - net.FileListener(f) calls f.PollFD().Copy(); the Copy increments
//     the refcount via the cached *poll.FD's SysFile.
//   - f.Close() routes through the cached *poll.FD's Close (see
//     file_unix.go's file.close), which decrements the refcount and
//     only releases the syscall FD when the count reaches zero.
//   - The eventual Listener.Close / Conn.Close decrements the refcount
//     from the other side.
//
// PollFD is intended for use by upstream Go's net/file_wasip1.go (which
// reaches it via a //go:linkname-style type assertion in this package).
func (f *File) PollFD() *poll.FD {
	if f.handle == nil {
		return nil
	}
	if f.pfd != nil {
		return f.pfd
	}
	pfd := &poll.FD{
		Sysfd:    int(f.handle.(interface{ Fd() uintptr }).Fd()),
		IsStream: true,
	}
	pfd.SysFile.RefCount = 1
	pfd.SysFile.RefCountPtr = &pfd.SysFile.RefCount
	f.pfd = pfd
	return pfd
}
