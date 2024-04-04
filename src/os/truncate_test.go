//go:build darwin || (linux && !baremetal && !js && !wasi)

// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package os_test

import (
	. "os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestTruncate(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip()
	}

	tmpDir := t.TempDir()
	file := filepath.Join(tmpDir, "truncate_test")

	fd, err := Create(file)
	if err != nil {
		t.Fatalf("create %q: got %v, want nil", file, err)
	}

	// truncate up to 0x100
	if err := fd.Truncate(0x100); err != nil {
		t.Fatalf("truncate %q: got %v, want nil", file, err)
	}

	// check if size is 0x100
	fi, err := Stat(file)
	if err != nil {
		t.Fatalf("stat %q: got %v, want nil", file, err)
	}

	if fi.Size() != 0x100 {
		t.Fatalf("size of %q is %d; want 0x100", file, fi.Size())
	}

	// truncate down to 0x80
	if err := fd.Truncate(0x80); err != nil {
		t.Fatalf("truncate %q: got %v, want nil", file, err)
	}

	// check if size is 0x80
	fi, err = Stat(file)
	if err != nil {
		t.Fatalf("stat %q: got %v, want nil", file, err)
	}

	if fi.Size() != 0x80 {
		t.Fatalf("size of %q is %d; want 0x80", file, fi.Size())
	}
}
