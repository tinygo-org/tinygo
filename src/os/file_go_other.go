//go:build !go1.16
// +build !go1.16

package os

import "time"

// A FileInfo describes a file and is returned by Stat and Lstat.
type FileInfo interface {
	Name() string       // base name of the file
	Size() int64        // length in bytes for regular files; system-dependent for others
	Mode() FileMode     // file mode bits
	ModTime() time.Time // modification time
	IsDir() bool        // abbreviation for Mode().IsDir()
	Sys() interface{}   // underlying data source (can return nil)
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

// IsRegular is a stub, always returning false
func (m FileMode) IsRegular() bool {
	return false
}

// Perm is a stub, always returning 0.
func (m FileMode) Perm() FileMode {
	return 0
}
