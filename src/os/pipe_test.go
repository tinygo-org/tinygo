//go:build windows || darwin || (linux && !baremetal && !wasi)
// +build windows darwin linux,!baremetal,!wasi

// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Test pipes on Unix and Windows systems.

package os_test

import (
	"bytes"
	"os"
	"testing"
)

// TestSmokePipe is a simple smoke test for Pipe().
func TestSmokePipe(t *testing.T) {
	// Procedure:
	// 1. Get the bytes
	// 2. Light the bytes on fire
	// 3. Smoke the bytes

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
		return // TODO: remove once Fatal is fatal
	}
	defer r.Close()
	defer w.Close()

	msg := []byte("Sed nvlla nisi ardva virtvs")
	n, err := w.Write(msg)
	if err != nil {
		t.Errorf("Writing to fresh pipe failed, error %v", err)
	}
	want := len(msg)
	if n != want {
		t.Errorf("Writing to fresh pipe wrote %d bytes, expected %d", n, want)
	}

	buf := make([]byte, 2*len(msg))
	n, err = r.Read(buf)
	if err != nil {
		t.Errorf("Reading from pipe failed, error %v", err)
	}
	if n != want {
		t.Errorf("Reading from pipe got %d bytes, expected %d", n, want)
	}
	// Read() does not set len(buf), so do it here.
	buf = buf[:n]
	if !bytes.Equal(buf, msg) {
		t.Errorf("Reading from fresh pipe got wrong bytes")
	}
}
