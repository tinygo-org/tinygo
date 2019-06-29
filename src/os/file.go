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
	ErrUnsupported = errors.New("operation not supported")
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

// Readdir is a stub (for now), always returning nil for both the slice of FileInfo and the error
func (f *File) Readdir(n int) ([]FileInfo, error) {
	return nil, nil
}

// Readdirnames is a stub (for now), always returning nil for both the slice of names and the error
func (f *File) Readdirnames(n int) (names []string, err error) {
	return nil, nil
}

// Stat is a stub (for now), always returning nil for both the FileInfo and the error
func (f *File) Stat() (FileInfo, error) {
	return nil, nil
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
		return nil, errors.New("unknown file: '" + name + "'")
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

// Mode constants
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
	Name() string       // base name of the file
	Size() int64        // length in bytes for regular files; system-dependent for others
	Mode() FileMode     // file mode bits
	// ModTime() time.Time // modification time
	IsDir() bool        // abbreviation for Mode().IsDir()
	Sys() interface{}   // underlying data source (can return nil)
}

// Stat is a stub (for now), always returning nil
func Stat(name string) (FileInfo, error) {
	return nil, nil
}

// Lstat is a stub (for now), always returning nil
func Lstat(name string) (FileInfo, error) {
	return nil, nil
}

// Getwd is a stub (for now), always returning the string it was given
func Getwd() (dir string, err error) {
	return dir, nil
}

// Readlink is a stub (for now), always returning the string it was given
func Readlink(name string) (string, error) {
	return name, nil
}

// TempDir is a stub (for now), always returning the string "/tmp/tinygo-tmp"
func TempDir() string {
	return "/tmp/tinygo-tmp"
}

// Mkdir is a stub (for now), always returning nil
func Mkdir(name string, perm FileMode) error {
	return nil
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
