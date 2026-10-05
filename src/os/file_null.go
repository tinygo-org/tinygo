//go:build baremetal || (tinygo.wasm && !wasip1 && !wasip2) || nintendoswitch

package os

import (
	"io"
	"syscall"
	"time"
)

func openFileNoMount(name string, flag int) (*File, error) {
	if name != DevNull {
		return nil, &PathError{Op: "open", Path: name, Err: ErrNotExist}
	}
	if flag&(O_CREATE|O_EXCL) == O_CREATE|O_EXCL {
		return nil, &PathError{Op: "open", Path: name, Err: ErrExist}
	}
	return &File{&file{
		handle:     nullFileHandle(flag),
		name:       name,
		appendMode: flag&O_APPEND != 0,
	}}, nil
}

type nullFileHandle int

func (f nullFileHandle) accessMode() int {
	return int(f) & (O_RDONLY | O_WRONLY | O_RDWR)
}

func (f nullFileHandle) Read(b []byte) (int, error) {
	if len(b) == 0 {
		return 0, nil
	}
	mode := f.accessMode()
	if mode != O_RDONLY && mode != O_RDWR {
		return 0, syscall.EBADF
	}
	return 0, io.EOF
}

func (f nullFileHandle) ReadAt(b []byte, offset int64) (int, error) {
	return f.Read(b)
}

func (f nullFileHandle) Write(b []byte) (int, error) {
	mode := f.accessMode()
	if mode != O_WRONLY && mode != O_RDWR {
		return 0, syscall.EBADF
	}
	return len(b), nil
}

func (f nullFileHandle) WriteAt(b []byte, offset int64) (int, error) {
	return f.Write(b)
}

func (f nullFileHandle) Seek(offset int64, whence int) (int64, error) {
	if whence < io.SeekStart || whence > io.SeekEnd {
		return 0, syscall.EINVAL
	}
	return 0, nil
}

func (f nullFileHandle) Sync() error {
	return syscall.EINVAL
}

func (f nullFileHandle) Close() error {
	return nil
}

func isNullDevice(name string) bool {
	fs, _ := findMount(name)
	return fs == nil && name == DevNull
}

type nullFileInfo struct{}

func (nullFileInfo) Name() string       { return "null" }
func (nullFileInfo) Size() int64        { return 0 }
func (nullFileInfo) Mode() FileMode     { return ModeDevice | ModeCharDevice | 0666 }
func (nullFileInfo) ModTime() time.Time { return time.Time{} }
func (nullFileInfo) IsDir() bool        { return false }
func (nullFileInfo) Sys() interface{}   { return nil }
