// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build linux && !baremetal && !tinygo.wasm

package os_test

import (
	"fmt"
	"internal/testenv"
	"io"
	. "os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func testStartProcess(dir, cmd string, args []string, expect string) func(t *testing.T) {
	return func(t *testing.T) {
		t.Parallel()

		r, w, err := Pipe()
		if err != nil {
			t.Fatalf("Pipe: %v", err)
		}
		defer r.Close()
		attr := &ProcAttr{Dir: dir, Files: []*File{nil, w, Stderr}}
		p, err := StartProcess(cmd, args, attr)
		if err != nil {
			t.Fatalf("StartProcess: %v", err)
		}
		w.Close()

		var b strings.Builder
		io.Copy(&b, r)
		output := b.String()

		fi1, _ := Stat(strings.TrimSpace(output))
		fi2, _ := Stat(expect)
		if !SameFile(fi1, fi2) {
			t.Errorf("exec %q returned %q wanted %q",
				strings.Join(append([]string{cmd}, args...), " "), output, expect)
		}
		p.Wait()
	}
}

func TestStartProcess(t *testing.T) {
	testenv.MustHaveExec(t)
	t.Parallel()

	var dir, cmd string
	var args []string
	switch runtime.GOOS {
	case "android":
		t.Skip("android doesn't have /bin/pwd")
	case "windows":
		cmd = Getenv("COMSPEC")
		dir = Getenv("SystemRoot")
		args = []string{"/c", "cd"}
	default:
		var err error
		cmd, err = exec.LookPath("pwd")
		if err != nil {
			t.Fatalf("Can't find pwd: %v", err)
		}
		dir = "/"
		args = []string{}
		t.Logf("Testing with %v", cmd)
	}
	cmddir, cmdbase := filepath.Split(cmd)
	args = append([]string{cmdbase}, args...)
	t.Run("absolute", testStartProcess(dir, cmd, args, dir))
	t.Run("relative", testStartProcess(cmddir, cmdbase, args, cmddir))
}

func TestGetppid(t *testing.T) {
	if runtime.GOOS == "plan9" {
		// TODO: golang.org/issue/8206
		t.Skipf("skipping test on plan9; see issue 8206")
	}

	if Getenv("GO_WANT_HELPER_PROCESS") == "1" {
		fmt.Print(Getppid())
		Exit(0)
	}

	t.Parallel()

	cmd := testenv.Command(t, testenv.Executable(t), "-test.run=^TestGetppid$")
	cmd.Env = append(Environ(), "GO_WANT_HELPER_PROCESS=1")

	// verify that Getppid() from the forked process reports our process id
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to spawn child process: %v %q", err, string(output))
	}

	childPpid := string(output)
	ourPid := fmt.Sprintf("%d", Getpid())
	if childPpid != ourPid {
		t.Fatalf("Child process reports parent process id '%v', expected '%v'", childPpid, ourPid)
	}
}
