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
var (
	errUnsupported = errors.New("operation not supported")
	notImplemented = errors.New("os: not implemented")
)

// Stdin, Stdout, and Stderr are open Files pointing to the standard input,
// standard output, and standard error file descriptors.
var (
	Stdin  = &File{0, "/dev/stdin"}
	Stdout = &File{1, "/dev/stdout"}
	Stderr = &File{2, "/dev/stderr"}
)

// File represents an open file descriptor.
type File struct {
	fd   uintptr
	name string
}

// Readdir is a stub, not yet implemented
func (f *File) Readdir(n int) ([]FileInfo, error) {
	return nil, notImplemented
}

// Readdirnames is a stub, not yet implemented
func (f *File) Readdirnames(n int) (names []string, err error) {
	return nil, notImplemented
}

// Stat is a stub, not yet implemented
func (f *File) Stat() (FileInfo, error) {
	return nil, notImplemented
}

// NewFile returns a new File with the given file descriptor and name.
func NewFile(fd uintptr, name string) *File {
	return &File{fd, name}
}

// Fd returns the integer Unix file descriptor referencing the open file. The
// file descriptor is valid only until f.Close is called.
func (f *File) Fd() uintptr {
	return f.fd
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

func (e *PathError) Error() string { return e.Op + " " + e.Path + ": " + e.Err.Error() }

// Open is a super simple stub function (for now), only capable of opening stdin, stdout, and stderr
func Open(name string) (*File, error) {
	fd := uintptr(999)
	switch name {
	case "/dev/stdin":
		fd = 0
	case "/dev/stdout":
		fd = 1
	case "/dev/stderr":
		fd = 2
	default:
		return nil, &PathError{"open", name, notImplemented}
	}
	return &File{fd, name}, nil
}

// OpenFile is a stub, passing through to the stub Open() call
func OpenFile(name string, flag int, perm FileMode) (*File, error) {
	return Open(name)
}

// Create is a stub, passing through to the stub Open() call
func Create(name string) (*File, error) {
	return Open(name)
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
	return nil, notImplemented
}

// Lstat is a stub, not yet implemented
func Lstat(name string) (FileInfo, error) {
	return nil, notImplemented
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

// Mkdir is a stub, not yet implemented
func Mkdir(name string, perm FileMode) error {
	return notImplemented
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
