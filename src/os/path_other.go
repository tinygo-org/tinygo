// +build baremetal js

// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package os

import (
	"syscall"
)

// MkdirAll creates a directory named path,
// along with any necessary parents, and returns nil,
// or else returns an error.
// The permission bits perm (before umask) are used for all
// directories that MkdirAll creates.
// If path is already a directory, MkdirAll does nothing
// and returns nil.
//func MkdirAll(path string, perm FileMode) error {
//	return &PathError{Op: "MkdirAll", Path: path, Err: syscall.ENOSYS}
//}

// RemoveAll removes path and any children it contains.
// It removes everything it can but returns the first error
// it encounters. If the path does not exist, RemoveAll
// returns nil (no error).
// If there is an error, it will be of type *PathError.
func RemoveAll(path string) error {
	return &PathError{Op: "RemoveAll", Path: path, Err: syscall.ENOSYS}
}
