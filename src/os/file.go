// Portions copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This file was originally copied from Go, see:
// https://github.com/golang/go/blob/master/src/os/file.go
//
// Some of the code inherited from Go is not used anymore in Tinygo, but we keep
// changes to a minimum to help simplify bringing changes (e.g. the lstat global
// is not used here anymore, but we might need it if we add tests from Go in
// this package).

// Package os implements a subset of the Go "os" package. See
// https://godoc.org/os for details.
//
// Note that the current implementation is blocking. This limitation should be
// removed in a future version.
package os

import (
	"errors"
	"io"
	"io/fs"
	"runtime"
	"syscall"
)

// Seek whence values.
//
// Deprecated: Use io.SeekStart, io.SeekCurrent, and io.SeekEnd.
const (
	SEEK_SET int = io.SeekStart
	SEEK_CUR int = io.SeekCurrent
	SEEK_END int = io.SeekEnd
)

// lstat is overridden in tests.
var lstat = Lstat

// Mkdir creates a directory. If the operation fails, it will return an error of
// type *PathError.
func Mkdir(path string, perm FileMode) error {
	fs, suffix := findMount(path)
	if fs == nil {
		return &PathError{Op: "mkdir", Path: path, Err: ErrNotExist}
	}
	err := fs.Mkdir(suffix, perm)
	if err != nil {
		return &PathError{Op: "mkdir", Path: path, Err: err}
	}
	return nil
}

// Many functions in package syscall return a count of -1 instead of 0.
// Using fixCount(call()) instead of call() corrects the count.
func fixCount(n int, err error) (int, error) {
	if n < 0 {
		n = 0
	}
	return n, err
}

// Remove removes a file or (empty) directory. If the operation fails, it will
// return an error of type *PathError.
func Remove(path string) error {
	fs, suffix := findMount(path)
	if fs == nil {
		return &PathError{Op: "remove", Path: path, Err: ErrNotExist}
	}
	err := fs.Remove(suffix)
	if err != nil {
		return err
	}
	return nil
}

// Name returns the name of the file with which it was opened.
func (f *File) Name() string {
	return f.name
}

// OpenFile opens the named file. If the operation fails, the returned error
// will be of type *PathError.
func OpenFile(name string, flag int, perm FileMode) (*File, error) {
	fs, suffix := findMount(name)
	if fs == nil {
		return nil, &PathError{Op: "open", Path: name, Err: ErrNotExist}
	}
	handle, err := fs.OpenFile(suffix, flag, perm)
	if err != nil {
		return nil, &PathError{Op: "open", Path: name, Err: err}
	}
	return NewFile(handle, name), nil
}

// Open opens the file named for reading.
func Open(name string) (*File, error) {
	return OpenFile(name, O_RDONLY, 0)
}

// Create creates the named file, overwriting it if it already exists.
func Create(name string) (*File, error) {
	return OpenFile(name, O_RDWR|O_CREATE|O_TRUNC, 0666)
}

// Read reads up to len(b) bytes from the File. It returns the number of bytes
// read and any error encountered. At end of file, Read returns 0, io.EOF.
func (f *File) Read(b []byte) (n int, err error) {
	if f.handle == nil {
		err = ErrClosed
	} else {
		n, err = f.handle.Read(b)
	}
	// TODO: want to always wrap, like upstream, but ReadFile() compares against exactly io.EOF?
	if err != nil && err != io.EOF {
		err = &PathError{Op: "read", Path: f.name, Err: err}
	}
	return
}

var errNegativeOffset = errors.New("negative offset")

// ReadAt reads up to len(b) bytes from the File at the given absolute offset.
// It returns the number of bytes read and any error encountered, possible io.EOF.
// At end of file, Read returns 0, io.EOF.
func (f *File) ReadAt(b []byte, offset int64) (n int, err error) {
	if offset < 0 {
		return 0, &PathError{Op: "readat", Path: f.name, Err: errNegativeOffset}
	}
	if f.handle == nil {
		return 0, &PathError{Op: "readat", Path: f.name, Err: ErrClosed}
	}

	for len(b) > 0 {
		m, e := f.handle.ReadAt(b, offset)
		if e != nil {
			// TODO: want to always wrap, like upstream, but TestReadAtEOF compares against exactly io.EOF?
			if e != io.EOF {
				err = &PathError{Op: "readat", Path: f.name, Err: e}
			} else {
				err = e
			}
			break
		}
		n += m
		b = b[m:]
		offset += int64(m)
	}

	return
}

// Write writes len(b) bytes to the File. It returns the number of bytes written
// and an error, if any. Write returns a non-nil error when n != len(b).
func (f *File) Write(b []byte) (n int, err error) {
	if f.handle == nil {
		err = ErrClosed
	} else {
		n, err = f.handle.Write(b)
	}
	if err != nil {
		err = &PathError{Op: "write", Path: f.name, Err: err}
	}
	return
}

// WriteString is like Write, but writes the contents of string s rather than a
// slice of bytes.
func (f *File) WriteString(s string) (n int, err error) {
	return f.Write([]byte(s))
}

func (f *File) WriteAt(b []byte, off int64) (n int, err error) {
	if f.handle == nil {
		err = ErrClosed
	} else {
		err = ErrNotImplemented
	}
	err = &PathError{Op: "writeat", Path: f.name, Err: err}
	return
}

// Close closes the File, rendering it unusable for I/O.
func (f *File) Close() (err error) {
	if f.handle == nil {
		err = ErrClosed
	} else {
		// Some platforms manage extra state other than the system handle which
		// needs to be released when the file is closed. For example, darwin
		// files have a DIR object holding a dup of the file descriptor.
		//
		// These platform-specific logic is provided by the (*file).close method
		// which is why we do not call the handle's Close method directly.
		err = f.file.close()
		if err == nil {
			f.handle = nil
		}
	}
	if err != nil {
		err = &PathError{Op: "close", Path: f.name, Err: err}
	}
	return
}

// Seek sets the offset for the next Read or Write on file to offset, interpreted
// according to whence: 0 means relative to the origin of the file, 1 means
// relative to the current offset, and 2 means relative to the end.
// It returns the new offset and an error, if any.
// The behavior of Seek on a file opened with O_APPEND is not specified.
//
// If f is a directory, the behavior of Seek varies by operating
// system; you can seek to the beginning of the directory on Unix-like
// operating systems, but not on Windows.
func (f *File) Seek(offset int64, whence int) (ret int64, err error) {
	if f.handle == nil {
		err = ErrClosed
	} else {
		ret, err = f.handle.Seek(offset, whence)
	}
	if err != nil {
		err = &PathError{Op: "seek", Path: f.name, Err: err}
	}
	return
}

func (f *File) SyscallConn() (conn syscall.RawConn, err error) {
	if f.handle == nil {
		err = ErrClosed
	} else {
		err = ErrNotImplemented
	}
	return
}

// fd is an internal interface that is used to try a type assertion in order to
// call the Fd() method of the underlying file handle if it is implemented.
type fd interface {
	Fd() uintptr
}

// Fd returns the file handle referencing the open file.
func (f *File) Fd() uintptr {
	handle, ok := f.handle.(fd)
	if ok {
		return handle.Fd()
	}
	return ^uintptr(0)
}

// Truncate is a stub, not yet implemented
func (f *File) Truncate(size int64) (err error) {
	if f.handle == nil {
		err = ErrClosed
	} else {
		err = ErrNotImplemented
	}
	return &PathError{Op: "truncate", Path: f.name, Err: err}
}

// LinkError records an error during a link or symlink or rename system call and
// the paths that caused it.
type LinkError struct {
	Op  string
	Old string
	New string
	Err error
}

func (e *LinkError) Error() string {
	return e.Op + " " + e.Old + " " + e.New + ": " + e.Err.Error()
}

func (e *LinkError) Unwrap() error {
	return e.Err
}

const (
	O_RDONLY int = syscall.O_RDONLY
	O_WRONLY int = syscall.O_WRONLY
	O_RDWR   int = syscall.O_RDWR
	O_APPEND int = syscall.O_APPEND
	O_CREATE int = syscall.O_CREAT
	O_EXCL   int = syscall.O_EXCL
	O_SYNC   int = syscall.O_SYNC
	O_TRUNC  int = syscall.O_TRUNC
)

func Getwd() (string, error) {
	return syscall.Getwd()
}

// TempDir returns the default directory to use for temporary files.
//
// On Unix systems, it returns $TMPDIR if non-empty, else /tmp.
// On Windows, it uses GetTempPath, returning the first non-empty
// value from %TMP%, %TEMP%, %USERPROFILE%, or the Windows directory.
//
// The directory is neither guaranteed to exist nor have accessible
// permissions.
func TempDir() string {
	return tempDir()
}

// UserHomeDir returns the current user's home directory.
//
// On Unix, including macOS, it returns the $HOME environment variable.
// On Windows, it returns %USERPROFILE%.
// On Plan 9, it returns the $home environment variable.
func UserHomeDir() (string, error) {
	env, enverr := "HOME", "$HOME"
	switch runtime.GOOS {
	case "windows":
		env, enverr = "USERPROFILE", "%userprofile%"
	case "plan9":
		env, enverr = "home", "$home"
	}
	if v := Getenv(env); v != "" {
		return v, nil
	}
	// On some geese the home directory is not always defined.
	switch runtime.GOOS {
	case "android":
		return "/sdcard", nil
	case "ios":
		return "/", nil
	}
	return "", errors.New(enverr + " is not defined")
}

type (
	FileMode = fs.FileMode
	FileInfo = fs.FileInfo
)

// The followings are copied from Go 1.16 or 1.17 official implementation:
// https://github.com/golang/go/blob/go1.16/src/os/file.go

// DirFS returns a file system (an fs.FS) for the tree of files rooted at the directory dir.
//
// Note that DirFS("/prefix") only guarantees that the Open calls it makes to the
// operating system will begin with "/prefix": DirFS("/prefix").Open("file") is the
// same as os.Open("/prefix/file"). So if /prefix/file is a symbolic link pointing outside
// the /prefix tree, then using DirFS does not stop the access any more than using
// os.Open does. DirFS is therefore not a general substitute for a chroot-style security
// mechanism when the directory tree contains arbitrary content.
func DirFS(dir string) fs.FS {
	return dirFS(dir)
}

func containsAny(s, chars string) bool {
	for i := 0; i < len(s); i++ {
		for j := 0; j < len(chars); j++ {
			if s[i] == chars[j] {
				return true
			}
		}
	}
	return false
}

type dirFS string

func (dir dirFS) Open(name string) (fs.File, error) {
	if !fs.ValidPath(name) || runtime.GOOS == "windows" && containsAny(name, `\:`) {
		return nil, &PathError{Op: "open", Path: name, Err: ErrInvalid}
	}
	f, err := Open(string(dir) + "/" + name)
	if err != nil {
		return nil, err // nil fs.File
	}
	return f, nil
}

func (dir dirFS) Stat(name string) (fs.FileInfo, error) {
	if !fs.ValidPath(name) || runtime.GOOS == "windows" && containsAny(name, `\:`) {
		return nil, &PathError{Op: "stat", Path: name, Err: ErrInvalid}
	}
	f, err := Stat(string(dir) + "/" + name)
	if err != nil {
		return nil, err
	}
	return f, nil
}

// ReadFile reads the named file and returns the contents.
// A successful call returns err == nil, not err == EOF.
// Because ReadFile reads the whole file, it does not treat an EOF from Read
// as an error to be reported.
func ReadFile(name string) ([]byte, error) {
	f, err := Open(name)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var size int
	if info, err := f.Stat(); err == nil {
		size64 := info.Size()
		if int64(int(size64)) == size64 {
			size = int(size64)
		}
	}
	size++ // one byte for final read at EOF

	// If a file claims a small size, read at least 512 bytes.
	// In particular, files in Linux's /proc claim size 0 but
	// then do not work right if read in small pieces,
	// so an initial read of 1 byte would not work correctly.
	if size < 512 {
		size = 512
	}

	data := make([]byte, 0, size)
	for {
		if len(data) >= cap(data) {
			d := append(data[:cap(data)], 0)
			data = d[:len(data)]
		}
		n, err := f.Read(data[len(data):cap(data)])
		data = data[:len(data)+n]
		if err != nil {
			if err == io.EOF {
				err = nil
			}
			return data, err
		}
	}
}

// WriteFile writes data to the named file, creating it if necessary.
// If the file does not exist, WriteFile creates it with permissions perm (before umask);
// otherwise WriteFile truncates it before writing, without changing permissions.
func WriteFile(name string, data []byte, perm FileMode) error {
	f, err := OpenFile(name, O_WRONLY|O_CREATE|O_TRUNC, perm)
	if err != nil {
		return err
	}
	_, err = f.Write(data)
	if err1 := f.Close(); err1 != nil && err == nil {
		err = err1
	}
	return err
}

// The defined file mode bits are the most significant bits of the FileMode.
// The nine least-significant bits are the standard Unix rwxrwxrwx permissions.
// The values of these bits should be considered part of the public API and
// may be used in wire protocols or disk representations: they must not be
// changed, although new bits might be added.
const (
	// The single letters are the abbreviations
	// used by the String method's formatting.
	ModeDir        = fs.ModeDir        // d: is a directory
	ModeAppend     = fs.ModeAppend     // a: append-only
	ModeExclusive  = fs.ModeExclusive  // l: exclusive use
	ModeTemporary  = fs.ModeTemporary  // T: temporary file; Plan 9 only
	ModeSymlink    = fs.ModeSymlink    // L: symbolic link
	ModeDevice     = fs.ModeDevice     // D: device file
	ModeNamedPipe  = fs.ModeNamedPipe  // p: named pipe (FIFO)
	ModeSocket     = fs.ModeSocket     // S: Unix domain socket
	ModeSetuid     = fs.ModeSetuid     // u: setuid
	ModeSetgid     = fs.ModeSetgid     // g: setgid
	ModeCharDevice = fs.ModeCharDevice // c: Unix character device, when ModeDevice is set
	ModeSticky     = fs.ModeSticky     // t: sticky
	ModeIrregular  = fs.ModeIrregular  // ?: non-regular file; nothing else is known about this file

	// Mask for the type bits. For regular files, none will be set.
	ModeType = fs.ModeType

	ModePerm = fs.ModePerm // Unix permission bits, 0o777
)
