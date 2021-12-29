// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package os_test

import (
	"io"
	. "os"
	"runtime"
	"strings"
	"testing"
)

// localTmp returns a local temporary directory not on NFS.
func localTmp() string {
	return TempDir()
}

func newFile(testName string, t *testing.T) (f *File) {
	f, err := CreateTemp("", testName)
	if err != nil {
		t.Fatalf("TempFile %s: %s", testName, err)
	}
	return
}

// Read with length 0 should not return EOF.
func TestRead0(t *testing.T) {
	f := newFile("TestRead0", t)
	defer Remove(f.Name())
	defer f.Close()

	const data = "hello, world\n"
	io.WriteString(f, data)
	f.Close()
	f, err := Open(f.Name())
	if err != nil {
		t.Errorf("failed to reopen")
	}

	b := make([]byte, 0)
	n, err := f.Read(b)
	if n != 0 || err != nil {
		t.Errorf("Read(0) = %d, %v, want 0, nil", n, err)
	}
	b = make([]byte, 5)
	n, err = f.Read(b)
	if n <= 0 || err != nil {
		t.Errorf("Read(5) = %d, %v, want >0, nil", n, err)
	}
}

// ReadAt with length 0 should not return EOF.
func TestReadAt0(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Log("TODO: implement Pread for Windows")
		return
	}
	f := newFile("TestReadAt0", t)
	defer Remove(f.Name())
	defer f.Close()

	const data = "hello, world\n"
	io.WriteString(f, data)

	b := make([]byte, 0)
	n, err := f.ReadAt(b, 0)
	if n != 0 || err != nil {
		t.Errorf("ReadAt(0,0) = %d, %v, want 0, nil", n, err)
	}
	b = make([]byte, 5)
	n, err = f.ReadAt(b, 0)
	if n <= 0 || err != nil {
		t.Errorf("ReadAt(5,0) = %d, %v, want >0, nil", n, err)
	}
}

func checkMode(t *testing.T, path string, mode FileMode) {
	dir, err := Stat(path)
	if err != nil {
		t.Fatalf("Stat %q (looking for mode %#o): %s", path, mode, err)
	}
	if dir.Mode()&ModePerm != mode {
		t.Errorf("Stat %q: mode %#o want %#o", path, dir.Mode(), mode)
	}
}

func TestReadAt(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Log("TODO: implement Pread for Windows")
		return
	}
	f := newFile("TestReadAt", t)
	defer Remove(f.Name())
	defer f.Close()

	const data = "hello, world\n"
	io.WriteString(f, data)

	b := make([]byte, 5)
	n, err := f.ReadAt(b, 7)
	if err != nil || n != len(b) {
		t.Fatalf("ReadAt 7: %d, %v", n, err)
	}
	if string(b) != "world" {
		t.Fatalf("ReadAt 7: have %q want %q", string(b), "world")
	}
}

// Verify that ReadAt doesn't affect seek offset.
// In the Plan 9 kernel, there used to be a bug in the implementation of
// the pread syscall, where the channel offset was erroneously updated after
// calling pread on a file.
func TestReadAtOffset(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Log("TODO: implement Pread for Windows")
		return
	}
	f := newFile("TestReadAtOffset", t)
	defer Remove(f.Name())
	defer f.Close()

	const data = "hello, world\n"
	io.WriteString(f, data)
	f.Close()
	f, err := Open(f.Name())
	if err != nil {
		t.Errorf("failed to reopen")
	}

	b := make([]byte, 5)

	n, err := f.ReadAt(b, 7)
	if err != nil || n != len(b) {
		t.Fatalf("ReadAt 7: %d, %v", n, err)
	}
	if string(b) != "world" {
		t.Fatalf("ReadAt 7: have %q want %q", string(b), "world")
	}

	n, err = f.Read(b)
	if err != nil || n != len(b) {
		t.Fatalf("Read: %d, %v", n, err)
	}
	if string(b) != "hello" {
		t.Fatalf("Read: have %q want %q", string(b), "hello")
	}
}

// Verify that ReadAt doesn't allow negative offset.
func TestReadAtNegativeOffset(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Log("TODO: implement Pread for Windows")
		return
	}
	f := newFile("TestReadAtNegativeOffset", t)
	defer Remove(f.Name())
	defer f.Close()

	const data = "hello, world\n"
	io.WriteString(f, data)
	f.Close()
	f, err := Open(f.Name())
	if err != nil {
		t.Errorf("failed to reopen")
	}

	b := make([]byte, 5)

	n, err := f.ReadAt(b, -10)

	const wantsub = "negative offset"
	if !strings.Contains(err.Error(), wantsub) || n != 0 {
		t.Errorf("ReadAt(-10) = %v, %v; want 0, ...%q...", n, err, wantsub)
	}
}

func TestReadAtEOF(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Log("TODO: implement Pread for Windows")
		return
	}
	f := newFile("TestReadAtEOF", t)
	defer Remove(f.Name())
	defer f.Close()

	_, err := f.ReadAt(make([]byte, 10), 0)
	switch err {
	case io.EOF:
		// all good
	case nil:
		t.Fatalf("ReadAt succeeded")
	default:
		t.Fatalf("ReadAt failed: %s", err)
	}
}
