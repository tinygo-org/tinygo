// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build !baremetal && !tinygo.wasm

package os_test

import (
	"os"
	"os/signal"
	"testing"
	"time"
)

func TestProcessLiteral(t *testing.T) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	defer signal.Stop(c)

	p := &os.Process{Pid: os.Getpid()}
	if err := p.Signal(os.Interrupt); err != nil {
		t.Fatalf("Signal got err %v, want nil", err)
	}

	// Verify we actually received the signal.
	select {
	case <-time.After(1 * time.Second):
		t.Error("timeout waiting for signal")
	case <-c:
		// Good
	}
}
