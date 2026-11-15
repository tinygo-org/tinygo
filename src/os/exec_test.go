// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package os_test

import (
	"internal/testenv"
	"os"
	"runtime"
	"syscall"
	"testing"
)

func TestFindProcess(t *testing.T) {
	// NOTE: For now, we only test the Linux case since only exec_posix.go is currently the only implementation.
	// Linux guarantees that there is pid 0
	proc, err := os.FindProcess(0)
	if err != nil {
		t.Error("FindProcess(0): wanted err == nil, got %v:", err)
	}

	if proc.Pid != 0 {
		t.Error("Expected pid 0, got: ", proc.Pid)
	}

	pid0 := os.Process{Pid: 0}
	if *proc != pid0 {
		t.Error("Expected &Process{Pid: 0}, got", *proc)
	}
}

func TestProcessReleaseTwice(t *testing.T) {
	testenv.MustHaveGoBuild(t)
	t.Parallel()

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("Pipe() got err %v, want nil", err)
	}
	defer r.Close()
	defer w.Close()

	p, err := os.StartProcess(testenv.GoToolPath(t), []string{"go"}, &os.ProcAttr{
		// N.B. On Windows, StartProcess requires exactly 3 Files. Pass
		// in a dummy pipe to avoid irrelevant output on the test stdout.
		Files: []*os.File{r, w, w},
	})
	if err != nil {
		t.Fatalf("starting test process: %v", err)
	}
	if err := p.Release(); err != nil {
		t.Fatalf("first Release: got err %v, want nil", err)
	}

	err = p.Release()

	// We want EINVAL from a second Release call only on Windows.
	var want error
	if runtime.GOOS == "windows" {
		want = syscall.EINVAL
	}

	if err != want {
		t.Fatalf("second Release: got err %v, want %v", err, want)
	}
}
