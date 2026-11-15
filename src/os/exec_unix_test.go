// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build linux && !baremetal && !tinygo.wasm

package os_test

import (
	"errors"
	"math"
	. "os"
	"runtime"
	"syscall"
	"testing"
)

// Lookup of a process that does not exist at time of lookup.
func TestProcessAlreadyDone(t *testing.T) {
	// Theoretically MaxInt32 is a valid PID, but the chance of it actually
	// being used is extremely unlikely.
	pid := math.MaxInt32
	if runtime.GOOS == "solaris" || runtime.GOOS == "illumos" {
		// Solaris/Illumos have a lower limit, above which wait returns
		// EINVAL (see waitid in usr/src/uts/common/os/exit.c in
		// illumos). This is configurable via sysconf(_SC_MAXPID), but
		// we'll just take the default.
		pid = 30000 - 1
	}

	p, err := FindProcess(pid)
	if err != nil {
		t.Fatalf("FindProcess(math.MaxInt32) got err %v, want nil", err)
	}

	if ps, err := p.Wait(); !errors.Is(err, syscall.ECHILD) {
		t.Errorf("Wait() got err %v (ps %+v), want %v", err, ps, syscall.ECHILD)
	}

	if err := p.Release(); err != nil {
		t.Errorf("Release() got err %v, want nil", err)
	}
}

func TestProcessBadPID(t *testing.T) {
	p, err := FindProcess(-1)
	if err != nil {
		t.Fatalf("unexpected FindProcess error: %v", err)
	}
	err = p.Signal(syscall.Signal(0))
	if err == nil {
		t.Error("p.Signal succeeded unexpectedly")
	}
}
