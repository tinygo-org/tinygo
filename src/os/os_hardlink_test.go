//go:build !windows && !baremetal && !js && !wasip1 && !wasm_unknown

// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package os_test

import (
	. "os"
	"syscall"
	"testing"
)

func TestHardlink(t *testing.T) {
	defer chtmpdir(t)()
	from, to := "hardlinktestfrom", "hardlinktestto"

	file, err := Create(to)
	if err != nil {
		t.Fatalf("Create(%q) failed: %v", to, err)
	}
	if err = file.Close(); err != nil {
		t.Errorf("Close(%q) failed: %v", to, err)
	}
	err = Link(to, from)
	if err != nil {
		t.Fatalf("Link(%q, %q) failed: %v", to, from, err)
	}

	tostat, err := Lstat(to)
	if err != nil {
		t.Fatalf("Lstat(%q) failed: %v", to, err)
	}
	fromstat, err := Stat(from)
	if err != nil {
		t.Fatalf("Stat(%q) failed: %v", from, err)
	}
	if !SameFile(tostat, fromstat) {
		t.Errorf("Symlink(%q, %q) did not create symlink", to, from)
	}

	fromstat, err = Lstat(from)
	if err != nil {
		t.Fatalf("Lstat(%q) failed: %v", from, err)
	}
	// if they have the same inode, they are hard links
	if fromstat.Sys().(*syscall.Stat_t).Ino != tostat.Sys().(*syscall.Stat_t).Ino {
		t.Fatalf("Lstat(%q).Sys().Ino = %v, Lstat(%q).Sys().Ino = %v, want the same", to, tostat.Sys().(*syscall.Stat_t).Ino, from, fromstat.Sys().(*syscall.Stat_t).Ino)
	}

	file.Close()
}
