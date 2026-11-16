//go:build !linux || baremetal || tinygo.wasm || nintendoswitch

package os

import (
	"syscall"
	"time"
)

func openNewFile(handle uintptr, name string, flag int) *File {
	f := NewFile(handle, name)
	f.appendMode = (flag & O_APPEND) != 0
	return f
}

func (f *File) checkClosed() error {
	if f.handle == nil {
		return ErrClosed
	}
	return nil
}

func (f *File) closeFile() error {
	if f.handle == nil {
		return ErrClosed
	}
	err := f.file.close()
	if err == nil {
		f.handle = nil
	}
	return err
}

func (f *File) read(b []byte) (int, error) {
	if f.handle == nil {
		return 0, ErrClosed
	}
	return f.handle.Read(b)
}

func (f *File) pread(b []byte, offset int64) (int, error) {
	return f.handle.ReadAt(b, offset)
}

func (f *File) write(b []byte) (int, error) {
	if f.handle == nil {
		return 0, ErrClosed
	}
	return f.handle.Write(b)
}

func (f *File) pwrite(b []byte, offset int64) (int, error) {
	return f.handle.WriteAt(b, offset)
}

func (f *File) seek(offset int64, whence int) (int64, error) {
	if f.handle == nil {
		return 0, ErrClosed
	}
	return f.handle.Seek(offset, whence)
}

func (f *File) sync() error {
	if f.handle == nil {
		return ErrClosed
	}
	return f.handle.Sync()
}

func (f *File) fd() uintptr {
	handle, ok := f.handle.(fd)
	if ok {
		return handle.Fd()
	}
	return ^uintptr(0)
}

func (f *File) syscallConn() (syscall.RawConn, error) {
	if f.handle == nil {
		return nil, ErrClosed
	}
	return nil, ErrNotImplemented
}

func (f *File) setDeadline(t time.Time) error {
	if f.handle == nil {
		return ErrClosed
	}
	if t.IsZero() {
		return nil
	}
	return ErrNotImplemented
}

func (f *File) setReadDeadline(t time.Time) error {
	if f.handle == nil {
		return ErrClosed
	}
	if t.IsZero() {
		return nil
	}
	return ErrNotImplemented
}

func (f *File) setWriteDeadline(t time.Time) error {
	if f.handle == nil {
		return ErrClosed
	}
	if t.IsZero() {
		return nil
	}
	return ErrNotImplemented
}
