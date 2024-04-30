//go:build !baremetal && !js && !wasip1 && !wasip2

// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// TODO: Move this back into os_test.go (as upstream has it) when wasi supports chmod

package os_test

import (
	. "os"
	"runtime"
	"syscall"
	"testing"
)

func TestChmod(t *testing.T) {
	// Chmod
	f := newFile("TestChmod", t)
	defer Remove(f.Name())
	defer f.Close()
	// Creation mode is read write

	fm := FileMode(0456)
	if runtime.GOOS == "windows" {
		fm = FileMode(0444) // read-only file
	}
	if err := Chmod(f.Name(), fm); err != nil {
		t.Fatalf("chmod %s %#o: %s", f.Name(), fm, err)
	}
	checkMode(t, f.Name(), fm)

}

// Since testing syscalls requires a static, predictable environment that has to be controlled
// by the CI, we don't test for success but for failures and verify that the error messages are as expected.
// EACCES is returned when the user does not have the required permissions to change the ownership of the file
// ENOENT is returned when the file does not exist
// ENOTDIR is returned when the file is not a directory
func TestChownErr(t *testing.T) {
	if runtime.GOOS == "windows" || runtime.GOOS == "plan9" {
		t.Skip("skipping on " + runtime.GOOS)
	}

	var (
		TEST_UID_ROOT = 0
		TEST_GID_ROOT = 0
	)

	f := newFile("TestChown", t)
	// technically, newFile should never return nil, but just in case
	if f == nil {
		t.Fatalf("newFile failed")
	}

	defer Remove(f.Name())
	defer f.Close()

	// EACCES
	if err := Chown(f.Name(), TEST_UID_ROOT, TEST_GID_ROOT); err != nil {
		if err.(syscall.Errno) != syscall.EPERM {
			t.Fatalf("chown(%s, uid=%v, gid=%v): got '%v', want %v", f.Name(), TEST_UID_ROOT, TEST_GID_ROOT, err, syscall.EPERM)
		}
	}

	// ENOENT
	if err := Chown("invalid", Geteuid(), Getgid()); err != nil {
		if err.(syscall.Errno) != syscall.ENOENT {
			t.Fatalf("chown(%s, uid=%v, gid=%v): got '%v', want %v", f.Name(), Geteuid(), Getegid(), err, syscall.ENOENT)
		}
	}
}
