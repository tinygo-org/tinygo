//go:build (go1.16 && baremetal) || (go1.16 && js) || (go1.16 && wasi) || (go1.16 && windows)
// +build go1.16,baremetal go1.16,js go1.16,wasi go1.16,windows

// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package os

import (
	"syscall"
)

type dirInfo struct {
}

func (f *File) readdir(n int, mode readdirMode) (names []string, dirents []DirEntry, infos []FileInfo, err error) {
	return nil, nil, nil, &PathError{Op: "readdir unimplemented", Err: syscall.ENOTDIR}
}
