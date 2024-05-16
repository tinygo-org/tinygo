// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package os_test

import (
	. "os"
	"testing"
)

func TestFindProcess(t *testing.T) {
	// NOTE: For now, we only test the Linux case since only exec_posix.go is currently the only implementation.
	// Linux guarantees that there is pid 0
	proc, err := FindProcess(0)
	if err != nil {
		t.Error("FindProcess(0): wanted err == nil, got %v:", err)
	}

	if proc.Pid != 0 {
		t.Error("Expected pid 0, got: ", proc.Pid)
	}

	pid0 := Process{Pid: 0}
	if *proc != pid0 {
		t.Error("Expected &Process{Pid: 0}, got", *proc)
	}
}
