//go:build darwin || (linux && !baremetal)
// +build darwin linux,!baremetal

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

// statNolog stats a file with no test logging.
func statNolog(name string) (FileInfo, error) {
	var fs fileStat
	err := ignoringEINTR(func() error {
		return syscall.Stat(name, &fs.sys)
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
		return syscall.Lstat(name, &fs.sys)
	})
	if err != nil {
		return nil, &PathError{Op: "lstat", Path: name, Err: err}
	}
	fillFileStatFromSys(&fs, name)
	return &fs, nil
}

// fstatNolog fstats a file handle with no test logging
func fstatNolog(f *File) (FileInfo, error) {
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
