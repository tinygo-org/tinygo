// +build !windows,!baremetal,!js,!wasi

// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package os_test

import (
	. "os"
	"testing"
)

// TODO: Move this back into os_anyos_test.go when wasi supports symlink

func TestSymlink(t *testing.T) {
	//testenv.MustHaveSymlink(t)

	defer chtmpdir(t)()
	from, to := "symlinktestfrom", "symlinktestto"
	file, err := Create(to)
	if err != nil {
		t.Fatalf("Create(%q) failed: %v", to, err)
	}
	if err = file.Close(); err != nil {
		t.Errorf("Close(%q) failed: %v", to, err)
	}
	err = Symlink(to, from)
	if err != nil {
		t.Fatalf("Symlink(%q, %q) failed: %v", to, from, err)
	}
	tostat, err := Lstat(to)
	if err != nil {
		t.Fatalf("Lstat(%q) failed: %v", to, err)
	}
	if tostat.Mode()&ModeSymlink != 0 {
		t.Fatalf("Lstat(%q).Mode()&ModeSymlink = %v, want 0", to, tostat.Mode()&ModeSymlink)
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
	if fromstat.Mode()&ModeSymlink == 0 {
		t.Fatalf("Lstat(%q).Mode()&ModeSymlink = 0, want %v", from, ModeSymlink)
	}
	fromstat, err = Stat(from)
	if err != nil {
		t.Fatalf("Stat(%q) failed: %v", from, err)
	}
	if fromstat.Name() != from {
		t.Errorf("Stat(%q).Name() = %q, want %q", from, fromstat.Name(), from)
	}
	if fromstat.Mode()&ModeSymlink != 0 {
		t.Fatalf("Stat(%q).Mode()&ModeSymlink = %v, want 0", from, fromstat.Mode()&ModeSymlink)
	}
	s, err := Readlink(from)
	if err != nil {
		t.Fatalf("Readlink(%q) failed: %v", from, err)
	}
	if s != to {
		t.Fatalf("Readlink(%q) = %q, want %q", from, s, to)
	}
	file, err = Open(from)
	if err != nil {
		t.Fatalf("Open(%q) failed: %v", from, err)
	}
	file.Close()
}
