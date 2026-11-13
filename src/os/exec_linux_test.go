//go:build linux && !baremetal && !tinygo.wasm

package os_test

import (
	"errors"
	. "os"
	"path/filepath"
	"runtime"
	"syscall"
	"testing"
)

// Test the functionality of the forkExec function, which is used to fork and exec a new process.
// This test is not run on Windows, as forkExec is not supported on Windows.
// This test is not run on Plan 9, as forkExec is not supported on Plan 9.
func TestForkExec(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Logf("skipping test on %s", runtime.GOOS)
		return
	}

	proc, err := StartProcess("/bin/echo", []string{"hello", "world"}, &ProcAttr{})
	if !errors.Is(err, nil) {
		t.Fatalf("forkExec failed: %v", err)
	}

	if proc == nil {
		t.Fatalf("proc is nil")
	}

	if proc.Pid == 0 {
		t.Fatalf("forkExec failed: new process has pid 0")
	}
}

func TestForkExecErrNotExist(t *testing.T) {
	proc, err := StartProcess("invalid", []string{"invalid"}, &ProcAttr{})
	if !errors.Is(err, ErrNotExist) {
		t.Fatalf("wanted ErrNotExist, got %s\n", err)
	}

	if proc != nil {
		t.Fatalf("wanted nil, got %v\n", proc)
	}
}

func TestForkExecProcDir(t *testing.T) {
	proc, err := StartProcess("/bin/echo", []string{"hello", "world"}, &ProcAttr{Dir: "dir"})
	if !errors.Is(err, ErrNotImplementedDir) {
		t.Fatalf("wanted ErrNotImplementedDir, got %v\n", err)
	}

	if proc != nil {
		t.Fatalf("wanted nil, got %v\n", proc)
	}
}

func TestForkExecProcSys(t *testing.T) {
	proc, err := StartProcess("/bin/echo", []string{"hello", "world"}, &ProcAttr{Sys: &syscall.SysProcAttr{}})
	if !errors.Is(err, ErrNotImplementedSys) {
		t.Fatalf("wanted ErrNotImplementedSys, got %v\n", err)
	}

	if proc != nil {
		t.Fatalf("wanted nil, got %v\n", proc)
	}
}

func TestForkExecProcFiles(t *testing.T) {
	proc, err := StartProcess("/bin/echo", []string{"hello", "world"}, &ProcAttr{Files: []*File{Stdin, Stdout, Stderr}})
	if !errors.Is(err, ErrNotImplementedFiles) {
		t.Fatalf("wanted ErrNotImplementedFiles, got %v\n", err)
	}

	if proc != nil {
		t.Fatalf("wanted nil, got %v\n", proc)
	}
}

func TestForkExecNilAttr(t *testing.T) {
	proc, err := StartProcess("/bin/echo", []string{"hello", "world"}, nil)
	if err == nil {
		t.Fatal("wanted an error for a nil ProcAttr")
	}

	if proc != nil {
		t.Fatalf("wanted nil, got %v\n", proc)
	}
}

// The caller keeps running and the command runs once.
func TestForkExecParentContinues(t *testing.T) {
	out := filepath.Join(t.TempDir(), "out")
	proc, err := StartProcess("/bin/sh", []string{"sh", "-c", "echo child >> " + out}, &ProcAttr{})
	if err != nil {
		t.Fatal(err)
	}
	var ws syscall.WaitStatus
	if _, err := syscall.Wait4(proc.Pid, &ws, 0, nil); err != nil {
		t.Fatal(err)
	}
	if !ws.Exited() || ws.ExitStatus() != 0 {
		t.Fatalf("child did not exit cleanly: %v", ws)
	}
	data, err := ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "child\n" {
		t.Fatalf("child output = %q, want %q", data, "child\n")
	}
}
