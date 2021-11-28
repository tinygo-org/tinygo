// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package os_test

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"
)

// testStatAndLstat verifies that all os.Stat, os.Lstat os.File.Stat and os.Readdir work.
func testStatAndLstat(t *testing.T, path string, isLink bool, statCheck, lstatCheck func(*testing.T, string, fs.FileInfo)) {
	// TODO: revert to upstream test once fstat and readdir are implemented
	// test os.Stat
	sfi, err := os.Stat(path)
	if err != nil {
		t.Error(err)
		return
	}
	statCheck(t, path, sfi)

	// test os.Lstat
	lsfi, err := os.Lstat(path)
	if err != nil {
		t.Error(err)
		return
	}
	lstatCheck(t, path, lsfi)

	if isLink {
		if os.SameFile(sfi, lsfi) {
			t.Errorf("stat and lstat of %q should not be the same", path)
		}
	} else {
		if !os.SameFile(sfi, lsfi) {
			t.Errorf("stat and lstat of %q should be the same", path)
		}
	}
}

// testIsDir verifies that fi refers to directory.
func testIsDir(t *testing.T, path string, fi fs.FileInfo) {
	t.Helper()
	if !fi.IsDir() {
		t.Errorf("%q should be a directory", path)
	}
	if fi.Mode()&fs.ModeSymlink != 0 {
		t.Errorf("%q should not be a symlink", path)
	}
}

// testIsFile verifies that fi refers to file.
func testIsFile(t *testing.T, path string, fi fs.FileInfo) {
	t.Helper()
	if fi.IsDir() {
		t.Errorf("%q should not be a directory", path)
	}
	if fi.Mode()&fs.ModeSymlink != 0 {
		t.Errorf("%q should not be a symlink", path)
	}
}

func testDirStats(t *testing.T, path string) {
	testStatAndLstat(t, path, false, testIsDir, testIsDir)
}

func testFileStats(t *testing.T, path string) {
	testStatAndLstat(t, path, false, testIsFile, testIsFile)
}

func TestDirAndSymlinkStats(t *testing.T) {
	// TODO: revert to upstream test once symlinks and t.TempDir are implemented
	tmpdir := os.TempDir()
	dir := filepath.Join(tmpdir, "dir")
	os.Remove(dir)
	if err := os.Mkdir(dir, 0777); err != nil {
		t.Fatal(err)
		return
	}
	testDirStats(t, dir)

}

func TestFileAndSymlinkStats(t *testing.T) {
	// TODO: revert to upstream test once symlinks and t.TempDir are implemented
	tmpdir := os.TempDir()
	file := filepath.Join(tmpdir, "file")
	if err := os.WriteFile(file, []byte("abcdefg"), 0644); err != nil {
		t.Fatal(err)
		return
	}
	testFileStats(t, file)
}
