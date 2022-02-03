// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build go1.16 && !baremetal && !js && !wasi
// +build go1.16,!baremetal,!js,!wasi

// DirFS tests copied verbatim from upstream os_test.go, and adjusted minimally to fit tinygo.

package os_test

import (
	"io/fs"
	"os"
	. "os"
	"path/filepath"
	"runtime"
	"testing"
	"testing/fstest"
)

func TestDirFS(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Log("TODO: implement Readdir for Windows")
		return
	}
	if err := fstest.TestFS(DirFS("./testdata/dirfs"), "a", "b", "dir/x"); err != nil {
		t.Fatal(err)
	}

	// Test that Open does not accept backslash as separator.
	d := DirFS(".")
	_, err := d.Open(`testdata\dirfs`)
	if err == nil {
		t.Fatalf(`Open testdata\dirfs succeeded`)
	}
}

func TestDirFSPathsValid(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Log("skipping on Windows")
		return
	}

	// TODO: switch back to t.TempDir once it's implemented
	d, err := MkdirTemp("", "TestDirFSPathsValid")
	if err != nil {
		t.Fatal(err)
	}
	defer Remove(d)
	if err := os.WriteFile(filepath.Join(d, "control.txt"), []byte(string("Hello, world!")), 0644); err != nil {
		t.Fatal(err)
	}
	defer Remove(filepath.Join(d, "control.txt"))
	if err := os.WriteFile(filepath.Join(d, `e:xperi\ment.txt`), []byte(string("Hello, colon and backslash!")), 0644); err != nil {
		t.Fatal(err)
	}
	defer Remove(filepath.Join(d, `e:xperi\ment.txt`))

	fsys := os.DirFS(d)
	err = fs.WalkDir(fsys, ".", func(path string, e fs.DirEntry, err error) error {
		if fs.ValidPath(e.Name()) {
			t.Logf("%q ok", e.Name())
		} else {
			t.Errorf("%q INVALID", e.Name())
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
