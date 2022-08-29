//go:build darwin || (linux && !baremetal && !wasm_freestanding)
// +build darwin linux,!baremetal,!wasm_freestanding

// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package os

import (
	"syscall"
)

// Sync is a stub, not yet implemented
func (f *File) Sync() error {
	return ErrNotImplemented
}

// Stat returns the FileInfo structure describing file.
// If there is an error, it will be of type *PathError.
func (f *File) Stat() (FileInfo, error) {
	var fs fileStat
	err := ignoringEINTR(func() error {
		return syscall.Fstat(int(f.handle.(unixFileHandle)), &fs.sys)
	})
	if err != nil {
		return nil, &PathError{Op: "fstat", Path: f.name, Err: err}
	}
	fillFileStatFromSys(&fs, f.name)
	return &fs, nil
}

// statNolog stats a file with no test logging.
func statNolog(name string) (FileInfo, error) {
	var fs fileStat
	err := ignoringEINTR(func() error {
		return handleSyscallError(syscall.Stat(name, &fs.sys))
	})
	if err != nil {
		return nil, &PathError{Op: "stat", Path: name, Err: err}
	}
	fillFileStatFromSys(&fs, name)
	return &fs, nil
}

// lstatNolog lstats a file with no test logging.
func lstatNolog(name string) (FileInfo, error) {
	var fs fileStat
	err := ignoringEINTR(func() error {
		return handleSyscallError(syscall.Lstat(name, &fs.sys))
	})
	if err != nil {
		return nil, &PathError{Op: "lstat", Path: name, Err: err}
	}
	fillFileStatFromSys(&fs, name)
	return &fs, nil
}
