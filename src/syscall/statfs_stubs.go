//go:build baremetal || js || wasip1 || wasip2 || wasm_unknown || nintendoswitch

// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package syscall

// Statfs_t and Fsid have been copied from the Go source tree
// (syscall/ztypes_linux_amd64.go).

type Statfs_t struct {
	Type    int64
	Bsize   int64
	Blocks  uint64
	Bfree   uint64
	Bavail  uint64
	Files   uint64
	Ffree   uint64
	Fsid    Fsid
	Namelen int64
	Frsize  int64
	Flags   int64
	Spare   [4]int64
}

type Fsid struct {
	X__val [2]int32
}

// This is a stub, it is not functional.
func Statfs(path string, buf *Statfs_t) (err error) {
	return ENOSYS
}

// This is a stub, it is not functional.
func Fstatfs(fd int, buf *Statfs_t) (err error) {
	return ENOSYS
}
