package main

import (
	"errors"
	"io"
	"os"
	"syscall"
)

func main() {
	data, err := os.ReadFile(os.DevNull)
	if err != nil || len(data) != 0 {
		panic("reading the null device should return an empty file")
	}
	if err := os.WriteFile(os.DevNull, []byte("discard"), 0666); err != nil {
		panic(err)
	}
	if f, err := os.OpenFile(os.DevNull, os.O_CREATE|os.O_EXCL, 0666); !errors.Is(err, os.ErrExist) {
		if f != nil {
			f.Close()
		}
		panic("exclusive creation of the null device should fail")
	}
	for _, flag := range []int{os.O_RDONLY, os.O_WRONLY, os.O_RDWR} {
		f, err := os.OpenFile(os.DevNull, flag, 0)
		if err != nil {
			panic(err)
		}
		if f.Name() != os.DevNull {
			panic("wrong null device name")
		}
		for _, stat := range []func() (os.FileInfo, error){
			f.Stat,
			func() (os.FileInfo, error) { return os.Stat(os.DevNull) },
			func() (os.FileInfo, error) { return os.Lstat(os.DevNull) },
		} {
			info, err := stat()
			if err != nil {
				panic(err)
			}
			if info.Name() != "null" || info.Size() != 0 || info.Mode()&(os.ModeDevice|os.ModeCharDevice) != os.ModeDevice|os.ModeCharDevice || info.IsDir() {
				panic("wrong null device metadata")
			}
		}
		if n, err := f.Read(nil); n != 0 || err != nil {
			panic("empty null device read should succeed")
		}
		if n, err := f.ReadAt(nil, 0); n != 0 || err != nil {
			panic("empty null device positional read should succeed")
		}
		n, err := f.Read(make([]byte, 1))
		if n != 0 || (flag == os.O_WRONLY && !errors.Is(err, syscall.EBADF)) || (flag != os.O_WRONLY && err != io.EOF) {
			panic("wrong null device read result")
		}
		n, err = f.ReadAt(make([]byte, 1), 123)
		if n != 0 || (flag == os.O_WRONLY && !errors.Is(err, syscall.EBADF)) || (flag != os.O_WRONLY && err != io.EOF) {
			panic("wrong null device positional read result")
		}
		n, err = f.Write([]byte("discard"))
		if (flag == os.O_RDONLY && (n != 0 || !errors.Is(err, syscall.EBADF))) || (flag != os.O_RDONLY && (n != 7 || err != nil)) {
			panic("wrong null device write result")
		}
		n, err = f.WriteAt([]byte("discard"), 123)
		if (flag == os.O_RDONLY && (n != 0 || !errors.Is(err, syscall.EBADF))) || (flag != os.O_RDONLY && (n != 7 || err != nil)) {
			panic("wrong null device positional write result")
		}
		n, err = f.Write(nil)
		if n != 0 || (flag == os.O_RDONLY && !errors.Is(err, syscall.EBADF)) || (flag != os.O_RDONLY && err != nil) {
			panic("wrong empty null device write result")
		}
		if n, err := f.ReadAt(make([]byte, 1), -1); n != 0 || err == nil {
			panic("negative null device read offset should fail")
		}
		if n, err := f.WriteAt([]byte("discard"), -1); n != 0 || err == nil {
			panic("negative null device write offset should fail")
		}
		for _, whence := range []int{io.SeekStart, io.SeekCurrent, io.SeekEnd} {
			if offset, err := f.Seek(123, whence); offset != 0 || err != nil {
				panic("wrong null device seek result")
			}
		}
		if _, err := f.Seek(0, -1); !errors.Is(err, syscall.EINVAL) {
			panic("invalid null device seek should fail")
		}
		if !errors.Is(f.Sync(), syscall.EINVAL) {
			panic("syncing the null device should fail")
		}
		if f.Fd() != ^uintptr(0) {
			panic("synthetic null device should not have a file descriptor")
		}
		if err := f.Close(); err != nil {
			panic(err)
		}
		for _, operation := range []func() error{
			f.Close,
			func() error { _, err := f.Read(nil); return err },
			func() error { _, err := f.ReadAt(nil, 0); return err },
			func() error { _, err := f.Write(nil); return err },
			func() error { _, err := f.WriteAt(nil, 0); return err },
			func() error { _, err := f.Seek(0, io.SeekStart); return err },
			f.Sync,
			func() error { _, err := f.Stat(); return err },
		} {
			if !errors.Is(operation(), os.ErrClosed) {
				panic("closed null device should return ErrClosed")
			}
		}
	}
	if !errors.Is(os.Mkdir(os.DevNull, 0777), os.ErrExist) {
		panic("the null device already exists")
	}
	if !errors.Is(os.Remove(os.DevNull), os.ErrPermission) {
		panic("the built-in null device cannot be removed")
	}
	if _, err := os.Open("/dev/missing"); !errors.Is(err, os.ErrNotExist) {
		panic("the null device should not provide other paths")
	}
	f, err := os.OpenFile(os.DevNull, os.O_WRONLY|os.O_APPEND, 0)
	if err != nil {
		panic(err)
	}
	if _, err := f.WriteAt([]byte("discard"), 0); err == nil {
		panic("positional writes in append mode should fail")
	}
	if err := f.Close(); err != nil {
		panic(err)
	}
	f, err = os.Open(os.DevNull)
	if err != nil {
		panic(err)
	}
	os.Mount("/dev/", denyFilesystem{})
	if _, err := os.Open(os.DevNull); !errors.Is(err, os.ErrPermission) {
		panic("mounted filesystems should override the null device")
	}
	if _, err := os.Stat(os.DevNull); !errors.Is(err, os.ErrNotImplemented) {
		panic("stat should not fall back through a mounted filesystem")
	}
	if _, err := os.Lstat(os.DevNull); !errors.Is(err, os.ErrNotImplemented) {
		panic("lstat should not fall back through a mounted filesystem")
	}
	if _, err := f.Stat(); err != nil {
		panic(err)
	}
	if err := f.Close(); err != nil {
		panic(err)
	}
	println("null device ok")
}

type denyFilesystem struct{}

func (denyFilesystem) OpenFile(name string, flag int, perm os.FileMode) (uintptr, error) {
	return 0, os.ErrPermission
}

func (denyFilesystem) Mkdir(name string, perm os.FileMode) error {
	return os.ErrPermission
}

func (denyFilesystem) Remove(name string) error {
	return os.ErrPermission
}
