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
	f := newFile("TestChown", t)
	defer Remove(f.Name())
	defer f.Close()

	if runtime.GOOS != "windows" {
		if err := Chown(f.Name(), 0, 0); err != nil {
			t.Fatalf("chown %s 0 0: %s", f.Name(), err)
		}

		fi, err := Stat(f.Name())
		if err != nil {
			t.Fatalf("stat %s: %s", f.Name(), err)
		}

		if fi.Sys().(*syscall.Stat_t).Uid != 0 || fi.Sys().(*syscall.Stat_t).Gid != 0 {
			t.Fatalf("chown %s 0 0: uid=%d gid=%d", f.Name(), fi.Sys().(*syscall.Stat_t).Uid, fi.Sys().(*syscall.Stat_t).Gid)
		}
	}
}
