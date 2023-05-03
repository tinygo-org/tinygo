package os

import (
	"strings"
)

// mounts lists the mount points currently mounted in the filesystem provided by
// the os package. To resolve a path to a mount point, it is scanned from top to
// bottom looking for the first prefix match.
var mounts []mountPoint

type mountPoint struct {
	// prefix is a filesystem prefix, that always starts and ends with a forward
	// slash. To denote the root filesystem, use a single slash: "/".
	// This allows fast checking whether a path lies within a mount point.
	prefix string

	// filesystem is the Filesystem implementation that is mounted at this mount
	// point.
	filesystem Filesystem
}

// Filesystem provides an interface for generic filesystem drivers mounted in
// the os package. The errors returned must be one of the os.Err* errors, or a
// custom error if one doesn't exist. It should not be a *PathError because
// errors will be wrapped with a *PathError by the filesystem abstraction.
//
// WARNING: this interface is not finalized and may change in a future version.
type Filesystem interface {
	// OpenFile opens the named file.
	OpenFile(name string, flag int, perm FileMode) (uintptr, error)

	// Mkdir creates a new directoy with the specified permission (before
	// umask). Some filesystems may not support directories or permissions.
	Mkdir(name string, perm FileMode) error

	// Remove removes the named file or (empty) directory.
	Remove(name string) error
}

// FileHandle is an interface that should be implemented by filesystems
// implementing the Filesystem interface.
//
// WARNING: this interface is not finalized and may change in a future version.
type FileHandle interface {
	// Read reads up to len(b) bytes from the file.
	Read(b []byte) (n int, err error)

	// ReadAt reads up to len(b) bytes from the file starting at the given absolute offset
	ReadAt(b []byte, offset int64) (n int, err error)

	// Seek resets the file pointer relative to start, current position, or end
	Seek(offset int64, whence int) (newoffset int64, err error)

	// Sync blocks until buffered writes have been written to persistent storage
	Sync() (err error)

	// Truncate adjusts the file to the given size
	Truncate(size int64) (err error)

	// Write writes up to len(b) bytes to the file.
	Write(b []byte) (n int, err error)

	// Close closes the file, making it unusable for further writes.
	Close() (err error)
}

// findMount returns the appropriate (mounted) filesystem to use for a given
// filename plus the path relative to that filesystem.
func findMount(path string) (Filesystem, string) {
	for i := len(mounts) - 1; i >= 0; i-- {
		mount := mounts[i]
		if strings.HasPrefix(path, mount.prefix) {
			return mount.filesystem, path[len(mount.prefix)-1:]
		}
	}
	if isOS {
		// Assume that the first entry in the mounts slice is the OS filesystem
		// at the root of the directory tree. Use it as-is, to support relative
		// paths.
		return mounts[0].filesystem, path
	}
	return nil, path
}

// Mount mounts the given filesystem in the filesystem abstraction layer of the
// os package. It is not possible to unmount filesystems. Filesystems added
// later will override earlier filesystems.
//
// The provided prefix must start and end with a forward slash. This is true for
// the root directory ("/") for example.
func Mount(prefix string, filesystem Filesystem) {
	if prefix[0] != '/' || prefix[len(prefix)-1] != '/' {
		panic("os.Mount: invalid prefix")
	}
	mounts = append(mounts, mountPoint{prefix, filesystem})
}
