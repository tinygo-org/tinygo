// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package os_test

import (
	. "os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestMkdirAll(t *testing.T) {
	tmpDir := TempDir()
	path := tmpDir + "/_TestMkdirAll_/dir/./dir2"
	err := MkdirAll(path, 0777)
	if err != nil {
		t.Fatalf("MkdirAll %q: %s", path, err)
	}
	// TODO: revert to upstream code which uses RemoveAll
	defer Remove(tmpDir + "/_TestMkdirAll_/dir/dir2")
	defer Remove(tmpDir + "/_TestMkdirAll_/dir")
	defer Remove(tmpDir + "/_TestMkdirAll_")

	// Already exists, should succeed.
	err = MkdirAll(path, 0777)
	if err != nil {
		t.Fatalf("MkdirAll %q (second time): %s", path, err)
	}

	// Make file.
	fpath := path + "/file"
	f, err := Create(fpath)
	if err != nil {
		t.Fatalf("create %q: %s", fpath, err)
	}
	defer Remove(fpath)
	defer f.Close()

	// Can't make directory named after file.
	err = MkdirAll(fpath, 0777)
	if err == nil {
		t.Fatalf("MkdirAll %q: no error", fpath)
	}
	perr, ok := err.(*PathError)
	if !ok {
		t.Fatalf("MkdirAll %q returned %T, not *PathError", fpath, err)
	}
	if filepath.Clean(perr.Path) != filepath.Clean(fpath) {
		t.Fatalf("MkdirAll %q returned wrong error path: %q not %q", fpath, filepath.Clean(perr.Path), filepath.Clean(fpath))
	}

	// Can't make subdirectory of file.
	ffpath := fpath + "/subdir"
	err = MkdirAll(ffpath, 0777)
	if err == nil {
		t.Fatalf("MkdirAll %q: no error", ffpath)
	}
	perr, ok = err.(*PathError)
	if !ok {
		t.Fatalf("MkdirAll %q returned %T, not *PathError", ffpath, err)
	}
	if filepath.Clean(perr.Path) != filepath.Clean(fpath) {
		t.Fatalf("MkdirAll %q returned wrong error path: %q not %q", ffpath, filepath.Clean(perr.Path), filepath.Clean(fpath))
	}

	if runtime.GOOS == "windows" {
		path := tmpDir + `\_TestMkdirAll_\dir\.\dir2\`
		err := MkdirAll(path, 0777)
		if err != nil {
			t.Fatalf("MkdirAll %q: %s", path, err)
		}
	}
}
