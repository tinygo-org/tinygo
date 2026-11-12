//go:build linux && !baremetal && !tinygo.wasm

package os_test

import (
	"errors"
	"io"
	. "os"
	"path/filepath"
	"runtime"
	"syscall"
	"testing"
)

func TestForkExec(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Logf("skipping test on %s", runtime.GOOS)
		return
	}

	proc, err := StartProcess("/bin/true", []string{"true"}, &ProcAttr{})
	if err != nil {
		t.Fatalf("StartProcess failed: %v", err)
	}
	if proc == nil || proc.Pid <= 0 {
		t.Fatalf("invalid process: %+v", proc)
	}

	pid := proc.Pid
	state, err := proc.Wait()
	if err != nil {
		t.Fatalf("Wait failed: %v", err)
	}
	if !state.Exited() || !state.Success() || state.ExitCode() != 0 {
		t.Errorf("unexpected state: exited=%v success=%v code=%d str=%q", state.Exited(), state.Success(), state.ExitCode(), state.String())
	}
	if state.Pid() != pid {
		t.Errorf("state.Pid() = %d, want %d", state.Pid(), pid)
	}
}

func TestForkExecErrNotExist(t *testing.T) {
	proc, err := StartProcess("invalid", []string{"invalid"}, &ProcAttr{})
	if !errors.Is(err, ErrNotExist) {
		t.Fatalf("wanted ErrNotExist, got %v", err)
	}
	if proc != nil {
		t.Fatalf("wanted nil, got %v", proc)
	}
}

func TestForkExecProcDir(t *testing.T) {
	dir := t.TempDir()
	proc, err := StartProcess("/bin/true", []string{"true"}, &ProcAttr{Dir: dir})
	if err != nil {
		t.Fatalf("StartProcess with Dir failed: %v", err)
	}
	state, err := proc.Wait()
	if err != nil || !state.Success() {
		t.Fatalf("Wait failed: state=%v err=%v", state, err)
	}

	missing := filepath.Join(dir, "does-not-exist")
	proc, err = StartProcess("/bin/true", []string{"true"}, &ProcAttr{Dir: missing})
	if !errors.Is(err, ErrNotExist) {
		t.Fatalf("wanted ErrNotExist for missing Dir, got %v", err)
	}
	if proc != nil {
		t.Fatalf("wanted nil proc, got %v", proc)
	}
}

func TestForkExecProcSys(t *testing.T) {
	proc, err := StartProcess("/bin/true", []string{"true"}, &ProcAttr{Sys: &syscall.SysProcAttr{}})
	if err != nil {
		t.Fatalf("StartProcess with Sys failed: %v", err)
	}
	state, err := proc.Wait()
	if err != nil || !state.Success() {
		t.Fatalf("Wait failed: state=%v err=%v", state, err)
	}
}

func TestForkExecProcFiles(t *testing.T) {
	r, w, err := Pipe()
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()

	proc, err := StartProcess("/bin/echo", []string{"echo", "hello", "world"}, &ProcAttr{
		Files: []*File{nil, w, nil},
	})
	w.Close()
	if err != nil {
		t.Fatalf("StartProcess with Files failed: %v", err)
	}

	out, err := io.ReadAll(r)
	if err != nil {
		t.Fatal(err)
	}
	state, err := proc.Wait()
	if err != nil || !state.Success() {
		t.Fatalf("Wait failed: state=%v err=%v", state, err)
	}
	if string(out) != "hello world\n" {
		t.Errorf("pipe output = %q, want %q", string(out), "hello world\n")
	}
}

func TestProcessSignalKill(t *testing.T) {
	proc, err := StartProcess("/bin/sleep", []string{"sleep", "60"}, &ProcAttr{})
	if err != nil {
		t.Fatalf("StartProcess failed: %v", err)
	}
	if err := proc.Kill(); err != nil {
		t.Fatalf("Kill failed: %v", err)
	}
	state, err := proc.Wait()
	if err != nil {
		t.Fatalf("Wait failed: %v", err)
	}
	if state.Success() || state.ExitCode() != -1 {
		t.Errorf("killed process state: success=%v code=%d str=%q", state.Success(), state.ExitCode(), state.String())
	}
}
