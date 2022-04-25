//go:build !baremetal && !js
// +build !baremetal,!js

package os_test

import (
	"io"
	. "os"
	"runtime"
	"testing"
)

// TestTempDir assumes that the test environment has a filesystem with a working temp dir.
func TestTempDir(t *testing.T) {
	name := TempDir() + "/_os_test_TestTempDir"
	Remove(name)
	f, err := OpenFile(name, O_RDWR|O_CREATE, 0644)
	if err != nil {
		t.Errorf("OpenFile %s: %s", name, err)
		return
	}
	err = f.Close()
	if err != nil {
		t.Errorf("Close %s: %s", name, err)
	}
	err = Remove(name)
	if err != nil {
		t.Errorf("Remove %s: %s", name, err)
	}
}

func TestChdir(t *testing.T) {
	// create and cd into a new directory
	oldDir, err := Getwd()
	if err != nil {
		t.Errorf("Getwd() returned %v", err)
		return
	}
	dir := "_os_test_TestChDir"
	Remove(dir)
	err = Mkdir(dir, 0755)
	defer Remove(dir) // even though not quite sure which directory it will execute in
	if err != nil {
		t.Errorf("Mkdir(%s, 0755) returned %v", dir, err)
	}
	err = Chdir(dir)
	if err != nil {
		t.Fatalf("Chdir %s: %s", dir, err)
		return
	}
	// create a file there
	file := "_os_test_TestTempDir.dat"
	f, err := OpenFile(file, O_RDWR|O_CREATE, 0644)
	if err != nil {
		t.Errorf("OpenFile %s: %s", file, err)
	}
	defer Remove(file) // even though not quite sure which directory it will execute in
	err = f.Close()
	if err != nil {
		t.Errorf("Close %s: %s", file, err)
	}
	// cd back to original directory
	// TODO: emulate "cd .." in wasi-libc better?
	//err = Chdir("..")
	err = Chdir(oldDir)
	if err != nil {
		t.Errorf("Chdir ..: %s", err)
	}
	// clean up file and directory explicitly so we can check for errors
	fullname := dir + "/" + file
	err = Remove(fullname)
	if err != nil {
		t.Errorf("Remove %s: %s", fullname, err)
	}
	err = Remove(dir)
	if err != nil {
		t.Errorf("Remove %s: %s", dir, err)
	}
}

func TestStandardFd(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Log("TODO: TestFd fails on Windows, skipping")
		return
	}
	if fd := Stdin.Fd(); fd != 0 {
		t.Errorf("Stdin.Fd() = %d, want 0", fd)
	}

	if fd := Stdout.Fd(); fd != 1 {
		t.Errorf("Stdout.Fd() = %d, want 1", fd)
	}

	if fd := Stderr.Fd(); fd != 2 {
		t.Errorf("Stderr.Fd() = %d, want 2", fd)
	}
}

func TestFd(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Log("TODO: TestFd fails on Windows, skipping")
		return
	}
	f := newFile("TestFd.txt", t)
	defer Remove(f.Name())
	defer f.Close()

	const data = "hello, world\n"
	io.WriteString(f, data)

	fd := NewFile(f.Fd(), "as-fd")
	defer fd.Close()

	b := make([]byte, 5)
	n, err := fd.ReadAt(b, 0)
	if n != 5 && err != nil {
		t.Errorf("Failed to read 5 bytes from file descriptor: %v", err)
	}

	if string(b) != data[:5] {
		t.Errorf("File descriptor contents not equal to file contents.")
	}
}
