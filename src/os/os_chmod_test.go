//go:build !baremetal && !js && !wasip1 && !wasip2

// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// TODO: Move this back into os_test.go (as upstream has it) when wasi supports chmod

package os_test

import (
	"errors"
	"io/fs"
	. "os"
	"runtime"
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
		t.Log("skipping on " + runtime.GOOS)
		return
	}

	var (
		TEST_UID_ROOT = 0
		TEST_GID_ROOT = 0
	)

	f := newFile("TestChown", t)
	defer Remove(f.Name())
	defer f.Close()

	// EACCES
	if err := Chown(f.Name(), TEST_UID_ROOT, TEST_GID_ROOT); err != nil {
		errCmp := fs.PathError{Op: "chown", Path: f.Name(), Err: errors.New("operation not permitted")}
		if errors.Is(err, &errCmp) {
			t.Fatalf("chown(%s, uid=%v, gid=%v): got '%v', want 'operation not permitted'", f.Name(), TEST_UID_ROOT, TEST_GID_ROOT, err)
		}
	}

	// ENOENT
	if err := Chown("invalid", Geteuid(), Getgid()); err != nil {
		errCmp := fs.PathError{Op: "chown", Path: "invalid", Err: errors.New("no such file or directory")}
		if errors.Is(err, &errCmp) {
			t.Fatalf("chown(%s, uid=%v, gid=%v): got '%v', want 'no such file or directory'", f.Name(), Geteuid(), Getegid(), err)
		}
	}
}
