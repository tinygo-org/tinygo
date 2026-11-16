//go:build darwin || tinygo.wasi

package os

import (
	"syscall"
	_ "unsafe" // for go:linkname
)

func NewFile(fd uintptr, name string) *File {
	return &File{&file{handle: unixFileHandle(fd), name: name}}
}

func Pipe() (r *File, w *File, err error) {
	var p [2]int
	err = handleSyscallError(pipe(p[:]))
	if err != nil {
		return
	}
	r = NewFile(uintptr(p[0]), "|0")
	w = NewFile(uintptr(p[1]), "|1")
	return
}

func (f *File) Truncate(size int64) error {
	if f == nil {
		return ErrInvalid
	}
	if f.handle == nil {
		return ErrClosed
	}
	return Truncate(f.name, size)
}

func (f *File) Chown(uid, gid int) error {
	if f == nil {
		return ErrInvalid
	}
	if f.handle == nil {
		return ErrClosed
	}
	return Chown(f.name, uid, gid)
}

func (f *File) chmod(mode FileMode) error {
	longName := fixLongPath(f.name)
	e := ignoringEINTR(func() error {
		return syscall.Chmod(longName, syscallMode(mode))
	})
	if e != nil {
		return &PathError{Op: "chmod", Path: f.name, Err: e}
	}
	return nil
}

func (f *File) chdir() error {
	longName := fixLongPath(f.name)
	e := ignoringEINTR(func() error {
		return syscall.Chdir(longName)
	})
	if e != nil {
		return &PathError{Op: "chdir", Path: f.name, Err: e}
	}
	return nil
}

func (f *File) fstat(s *syscall.Stat_t) error {
	if f.handle == nil {
		return ErrClosed
	}
	return ignoringEINTR(func() error {
		return syscall.Fstat(int(f.handle.(unixFileHandle)), s)
	})
}

//go:linkname net_newUnixFile net.newUnixFile
func net_newUnixFile(fd int, name string) *File {
	if fd < 0 {
		panic("invalid FD")
	}
	return NewFile(uintptr(fd), name)
}
