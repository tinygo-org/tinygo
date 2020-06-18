// Package os implements a subset of the Go "os" package. See
// https://godoc.org/os for details.
//
// Note that the current implementation is blocking. This limitation should be
// removed in a future version.
package os

import (
	"errors"
)

// Portable analogs of some common system call errors.
// Note that these are exported for use in the Filesystem interface.
var (
	ErrUnsupported    = errors.New("operation not supported")
	ErrNotImplemented = errors.New("operation not implemented")
	ErrNotExist       = errors.New("file not found")
	ErrExist          = errors.New("file exists")
)

// Mkdir creates a directory. If the operation fails, it will return an error of
// type *PathError.
func Mkdir(path string, perm FileMode) error {
	fs, suffix := findMount(path)
	if fs == nil {
		return &PathError{"mkdir", path, ErrNotExist}
	}
	err := fs.Mkdir(suffix, perm)
	if err != nil {
		return &PathError{"mkdir", path, err}
	}
	return nil
}

// Remove removes a file or (empty) directory. If the operation fails, it will
// return an error of type *PathError.
func Remove(path string) error {
	fs, suffix := findMount(path)
	if fs == nil {
		return &PathError{"remove", path, ErrNotExist}
	}
	err := fs.Remove(suffix)
	if err != nil {
		return &PathError{"remove", path, err}
	}
	return nil
}

// File represents an open file descriptor.
type File struct {
	handle FileHandle
	name   string
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
		return nil, &PathError{"open", name, ErrNotExist}
	}
	handle, err := fs.OpenFile(suffix, flag, perm)
	if err != nil {
		return nil, &PathError{"open", name, err}
	}
	return &File{name: name, handle: handle}, nil
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
	n, err = f.handle.Read(b)
	if err != nil {
		err = &PathError{"read", f.name, err}
	}
	return
}

// Write writes len(b) bytes to the File. It returns the number of bytes written
// and an error, if any. Write returns a non-nil error when n != len(b).
func (f *File) Write(b []byte) (n int, err error) {
	n, err = f.handle.Write(b)
	if err != nil {
		err = &PathError{"write", f.name, err}
	}
	return
}

// Close closes the File, rendering it unusable for I/O.
func (f *File) Close() (err error) {
	err = f.handle.Close()
	if err != nil {
		err = &PathError{"close", f.name, err}
	}
	return
}

// Readdir is a stub, not yet implemented
func (f *File) Readdir(n int) ([]FileInfo, error) {
	return nil, &PathError{"readdir", f.name, ErrNotImplemented}
}

// Readdirnames is a stub, not yet implemented
func (f *File) Readdirnames(n int) (names []string, err error) {
	return nil, &PathError{"readdirnames", f.name, ErrNotImplemented}
}

// Stat is a stub, not yet implemented
func (f *File) Stat() (FileInfo, error) {
	return nil, &PathError{"stat", f.name, ErrNotImplemented}
}

const (
	PathSeparator     = '/' // OS-specific path separator
	PathListSeparator = ':' // OS-specific path list separator
)

// IsPathSeparator reports whether c is a directory separator character.
func IsPathSeparator(c uint8) bool {
	return PathSeparator == c
}

// PathError records an error and the operation and file path that caused it.
type PathError struct {
	Op   string
	Path string
	Err  error
}

func (e *PathError) Error() string {
	return e.Op + " " + e.Path + ": " + e.Err.Error()
}

type FileMode uint32

// Mode constants, copied from the mainline Go source
// https://github.com/golang/go/blob/4ce6a8e89668b87dce67e2f55802903d6eb9110a/src/os/types.go#L35-L63
const (
	// The single letters are the abbreviations used by the String method's formatting.
	ModeDir        FileMode = 1 << (32 - 1 - iota) // d: is a directory
	ModeAppend                                     // a: append-only
	ModeExclusive                                  // l: exclusive use
	ModeTemporary                                  // T: temporary file; Plan 9 only
	ModeSymlink                                    // L: symbolic link
	ModeDevice                                     // D: device file
	ModeNamedPipe                                  // p: named pipe (FIFO)
	ModeSocket                                     // S: Unix domain socket
	ModeSetuid                                     // u: setuid
	ModeSetgid                                     // g: setgid
	ModeCharDevice                                 // c: Unix character device, when ModeDevice is set
	ModeSticky                                     // t: sticky
	ModeIrregular                                  // ?: non-regular file; nothing else is known about this file

	// Mask for the type bits. For regular files, none will be set.
	ModeType = ModeDir | ModeSymlink | ModeNamedPipe | ModeSocket | ModeDevice | ModeCharDevice | ModeIrregular

	ModePerm FileMode = 0777 // Unix permission bits
)

// IsDir is a stub, always returning false
func (m FileMode) IsDir() bool {
	return false
}

// Stub constants
const (
	O_RDONLY int = 1
	O_WRONLY int = 2
	O_RDWR   int = 4
	O_APPEND int = 8
	O_CREATE int = 16
	O_EXCL   int = 32
	O_SYNC   int = 64
	O_TRUNC  int = 128
)

// A FileInfo describes a file and is returned by Stat and Lstat.
type FileInfo interface {
	Name() string   // base name of the file
	Size() int64    // length in bytes for regular files; system-dependent for others
	Mode() FileMode // file mode bits
	// ModTime() time.Time // modification time
	IsDir() bool      // abbreviation for Mode().IsDir()
	Sys() interface{} // underlying data source (can return nil)
}

// Stat is a stub, not yet implemented
func Stat(name string) (FileInfo, error) {
	return nil, &PathError{"stat", name, ErrNotImplemented}
}

// Lstat is a stub, not yet implemented
func Lstat(name string) (FileInfo, error) {
	return nil, &PathError{"lstat", name, ErrNotImplemented}
}

// Getwd is a stub (for now), always returning an empty string
func Getwd() (string, error) {
	return "", nil
}

// Readlink is a stub (for now), always returning the string it was given
func Readlink(name string) (string, error) {
	return name, nil
}

// TempDir is a stub (for now), always returning the string "/tmp"
func TempDir() string {
	return "/tmp"
}

// IsExist is a stub (for now), always returning false
func IsExist(err error) bool {
	return false
}

// IsNotExist is a stub (for now), always returning false
func IsNotExist(err error) bool {
	return false
}

// Getpid is a stub (for now), always returning 1
func Getpid() int {
	return 1
}
