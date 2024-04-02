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
	if runtime.GOOS == "windows" || runtime.GOOS == "plan9" {
		t.Skip()
	}

	testCases := map[string]struct {
		uid     int
		gid     int
		wantErr bool
	}{
		"root": {
			uid:     0,
			gid:     0,
			wantErr: true,
		},
		"user-" + runtime.GOOS: {
			uid:     1001,
			gid:     127,
			wantErr: false,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			f := newFile("TestChown", t)
			defer Remove(f.Name())
			defer f.Close()

			err := Chown(f.Name(), tc.uid, tc.gid)
			if (tc.wantErr && err == nil) || (!tc.wantErr && err != nil) {
				t.Fatalf("chown(%s, uid=%v, gid=%v): got %v, want error: %v", f.Name(), tc.uid, tc.gid, err, tc.wantErr)
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
			if uid != uint32(tc.uid) || gid != uint32(tc.gid) {
				t.Fatalf("chown(%s, %d, %d): want (%d,%d), got (%d, %d)", f.Name(), tc.uid, tc.gid, tc.uid, tc.gid, uid, gid)
			}
		})
	}
}
