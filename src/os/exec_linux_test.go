//go:build linux && !baremetal && !tinygo.wasm

package os_test

import (
	"errors"
	"io"
	. "os"
	"path/filepath"
	"strconv"
	"syscall"
	"testing"
)

// Test the functionality of the forkExec function, which is used to fork and exec a new process.
// This test is not run on Windows, as forkExec is not supported on Windows.
// This test is not run on Plan 9, as forkExec is not supported on Plan 9.
func TestForkExec(t *testing.T) {
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
	dir := t.TempDir()
	proc, err := StartProcess("/bin/sh", []string{"sh", "-c", "pwd -P > out"}, &ProcAttr{Dir: dir})
	if err != nil {
		t.Fatal(err)
	}
	ps, err := proc.Wait()
	if err != nil {
		t.Fatal(err)
	}
	if !ps.Success() {
		t.Fatalf("child did not exit cleanly: %v", ps)
	}
	data, err := ReadFile(filepath.Join(dir, "out"))
	if err != nil {
		t.Fatal(err)
	}
	want, err := filepath.EvalSymlinks(dir)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != want+"\n" {
		t.Fatalf("child dir = %q, want %q", data, want+"\n")
	}
}

func TestForkExecProcDirNotExist(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "missing")
	proc, err := StartProcess("/bin/echo", []string{"hello", "world"}, &ProcAttr{Dir: dir})
	var pe *PathError
	if !errors.As(err, &pe) || pe.Op != "chdir" || !errors.Is(err, ErrNotExist) {
		t.Fatalf("wanted chdir PathError with ErrNotExist, got %v", err)
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

// startWithFiles runs script with files and returns what it writes to each pipe.
func startWithFiles(t *testing.T, script string, files []*File, writers []*File, readers []*File) []string {
	t.Helper()
	proc, err := StartProcess("/bin/sh", []string{"sh", "-c", script}, &ProcAttr{Files: files})
	for _, w := range writers {
		w.Close()
	}
	if err != nil {
		t.Fatal(err)
	}
	var out []string
	for _, r := range readers {
		data, err := io.ReadAll(r)
		if err != nil {
			t.Fatal(err)
		}
		out = append(out, string(data))
	}
	if _, err := proc.Wait(); err != nil {
		t.Fatal(err)
	}
	return out
}

func TestProcAttrFiles(t *testing.T) {
	r, w, err := Pipe()
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()
	out := startWithFiles(t, "echo hello", []*File{Stdin, w, Stderr}, []*File{w}, []*File{r})
	if out[0] != "hello\n" {
		t.Fatalf("stdout = %q, want %q", out[0], "hello\n")
	}
}

// A nil entry closes that file descriptor in the new process.
func TestProcAttrFilesNil(t *testing.T) {
	r, w, err := Pipe()
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()
	script := "if [ -e /proc/$$/fd/2 ]; then echo open; else echo closed; fi"
	out := startWithFiles(t, script, []*File{Stdin, w, nil}, []*File{w}, []*File{r})
	if out[0] != "closed\n" {
		t.Fatalf("fd 2 is %q, want %q", out[0], "closed\n")
	}
}

// A source descriptor that is also a slot must not be replaced before it is used.
func TestProcAttrFilesOverlap(t *testing.T) {
	r1, w1, err := Pipe()
	if err != nil {
		t.Fatal(err)
	}
	defer r1.Close()
	r2, w2, err := Pipe()
	if err != nil {
		t.Fatal(err)
	}
	defer r2.Close()
	a := int(w1.Fd())
	if a >= 9 {
		t.Skipf("pipe fd %d is too high for sh redirections", a)
	}
	files := make([]*File, 10)
	files[0], files[1], files[2] = Stdin, Stdout, Stderr
	files[a] = w2
	files[9] = w1
	script := "echo one >&9; echo two >&" + strconv.Itoa(a)
	out := startWithFiles(t, script, files, []*File{w1, w2}, []*File{r1, r2})
	if out[0] != "one\n" || out[1] != "two\n" {
		t.Fatalf("pipe 1 got %q, pipe 2 got %q, want %q and %q", out[0], out[1], "one\n", "two\n")
	}
}

// A file that is already at its target descriptor stays open, even with FD_CLOEXEC set.
func TestProcAttrFilesSameFd(t *testing.T) {
	r, w, err := Pipe()
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()
	fd := int(w.Fd())
	if fd > 9 {
		t.Skipf("pipe fd %d is too high for sh redirections", fd)
	}
	if _, _, errno := syscall.Syscall(syscall.SYS_FCNTL, uintptr(fd), syscall.F_SETFD, syscall.FD_CLOEXEC); errno != 0 {
		t.Fatal(errno)
	}
	files := make([]*File, fd+1)
	files[0], files[1], files[2] = Stdin, Stdout, Stderr
	files[fd] = w
	out := startWithFiles(t, "echo same >&"+strconv.Itoa(fd), files, []*File{w}, []*File{r})
	if out[0] != "same\n" {
		t.Fatalf("fd %d got %q, want %q", fd, out[0], "same\n")
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
	ps, err := proc.Wait()
	if err != nil {
		t.Fatal(err)
	}
	if !ps.Success() {
		t.Fatalf("child did not exit cleanly: %v", ps)
	}
	data, err := ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "child\n" {
		t.Fatalf("child output = %q, want %q", data, "child\n")
	}
}

func TestProcessWait(t *testing.T) {
	for _, tc := range []struct {
		script  string
		code    int
		success bool
		str     string
	}{
		{"exit 0", 0, true, "exit status 0"},
		{"exit 3", 3, false, "exit status 3"},
	} {
		proc, err := StartProcess("/bin/sh", []string{"sh", "-c", tc.script}, &ProcAttr{})
		if err != nil {
			t.Fatal(err)
		}
		ps, err := proc.Wait()
		if err != nil {
			t.Fatal(err)
		}
		if ps.Pid() != proc.Pid {
			t.Errorf("%s: Pid() = %d, want %d", tc.script, ps.Pid(), proc.Pid)
		}
		if !ps.Exited() || ps.ExitCode() != tc.code || ps.Success() != tc.success || ps.String() != tc.str {
			t.Errorf("%s: Exited %v, ExitCode %d, Success %v, String %q", tc.script, ps.Exited(), ps.ExitCode(), ps.Success(), ps.String())
		}
	}
}

func TestProcessWaitSignal(t *testing.T) {
	proc, err := StartProcess("/bin/sh", []string{"sh", "-c", "kill -KILL $$"}, &ProcAttr{})
	if err != nil {
		t.Fatal(err)
	}
	ps, err := proc.Wait()
	if err != nil {
		t.Fatal(err)
	}
	if ps.Exited() || ps.ExitCode() != -1 || ps.Success() || ps.String() != "signal: killed" {
		t.Errorf("Exited %v, ExitCode %d, Success %v, String %q", ps.Exited(), ps.ExitCode(), ps.Success(), ps.String())
	}
}

// Wait on a zero Process must not reap another child.
func TestProcessWaitZeroValue(t *testing.T) {
	proc, err := StartProcess("/bin/sh", []string{"sh", "-c", "exit 7"}, &ProcAttr{})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := new(Process).Wait(); err == nil {
		t.Fatal("wanted an error from Wait on a zero Process")
	}
	ps, err := proc.Wait()
	if err != nil {
		t.Fatal(err)
	}
	if ps.ExitCode() != 7 {
		t.Fatalf("ExitCode = %d, want 7", ps.ExitCode())
	}
}
