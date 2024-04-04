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

func TestChown(t *testing.T) {
	var (
		TEST_UID_1           = 1001
		TEST_GID_1           = 127
		TEST_UID_2           = 0
		TEST_GID_2           = 0
		ERR_PERM_DENIED      = "permission denied"
		ERR_OP_NOT_PERMITTED = "operation not permitted"
	)

	if runtime.GOOS == "windows" || runtime.GOOS == "plan9" {
		t.Skip()
	}

	f := newFile("TestChown", t)
	defer Remove(f.Name())
	defer f.Close()

	// User with CI permissions run chown on owned file
	// This test does not change the initial permissions of the file
	// It only checks if the chown operation was successful and consistent
	if err := Chown(f.Name(), TEST_UID_1, TEST_GID_1); err != nil {
		if err.Error() != ERR_OP_NOT_PERMITTED {
			t.Fatalf("chown(%s, uid=%v, gid=%v): got %v, want != nil", f.Name(), TEST_UID_1, TEST_GID_1, err)
		}
	}
	fi, err := Stat(f.Name())
	if err != nil {
		t.Fatalf("stat %s: got %v, want nil", f.Name(), err)
	}

	if fi.Sys() == nil {
		t.Fatalf("stat %s: fi.Sys(): got nil", f.Name())
	}

	s, ok := fi.Sys().(*syscall.Stat_t)
	if !ok {
		t.Fatalf("stat %s: fi.Sys(): is not *syscall.Stat_t", f.Name())
	}

	uid, gid := s.Uid, s.Gid
	if uid != uint32(TEST_UID_1) || gid != uint32(TEST_GID_1) {
		t.Fatalf("chown(%s, %d, %d): want (%d,%d), got (%d, %d)", f.Name(), TEST_UID_1, TEST_GID_1, TEST_UID_1, TEST_GID_1, uid, gid)
	}

	// Root permission denied
	if err = Chown(f.Name(), TEST_UID_2, TEST_GID_2); err.Error() != ERR_PERM_DENIED {
		t.Fatalf("chown(%s, uid=%v, gid=%v): got %v, want != nil", f.Name(), TEST_UID_2, TEST_GID_2, err)
	}

	fi, err = Stat(f.Name())
	if err != nil {
		t.Fatalf("stat %s: got %v, want nil", f.Name(), err)
	}

	if fi.Sys() == nil {
		t.Fatalf("stat %s: fi.Sys(): got nil", f.Name())
	}

	s, ok = fi.Sys().(*syscall.Stat_t)
	if !ok {
		t.Fatalf("stat %s: fi.Sys(): is not *syscall.Stat_t", f.Name())
	}

	uid, gid = s.Uid, s.Gid
	if uid != uint32(TEST_UID_2) || gid != uint32(TEST_GID_2) {
		// Chown will fail due to permissions denied, so the UID and GID should stay the same
		if uid != uint32(TEST_UID_1) && gid != uint32(TEST_GID_1) {
			t.Fatalf("chown(%s, %d, %d): want (%d,%d), got (%d, %d)", f.Name(), TEST_UID_2, TEST_GID_2, TEST_UID_2, TEST_GID_2, uid, gid)
		}
	}
}
