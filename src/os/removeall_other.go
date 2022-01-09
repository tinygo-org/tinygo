//go:build baremetal || js || wasi
// +build baremetal js wasi

// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package os

import (
	"syscall"
)

func removeAll(path string) error {
	return &PathError{Op: "RemoveAll", Path: path, Err: syscall.ENOSYS}
}
