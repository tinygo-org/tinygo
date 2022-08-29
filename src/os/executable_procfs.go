// The following is copied from Go 1.17 official implementation.

// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build linux && !baremetal && !wasm_freestanding
// +build linux,!baremetal,!wasm_freestanding

package os

func Executable() (string, error) {
	path, err := Readlink("/proc/self/exe")

	// When the executable has been deleted then Readlink returns a
	// path appended with " (deleted)".
	return stringsTrimSuffix(path, " (deleted)"), err
}

// stringsTrimSuffix is the same as strings.TrimSuffix.
func stringsTrimSuffix(s, suffix string) string {
	if len(s) >= len(suffix) && s[len(s)-len(suffix):] == suffix {
		return s[:len(s)-len(suffix)]
	}
	return s
}
