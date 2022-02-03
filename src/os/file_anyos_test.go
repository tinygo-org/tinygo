//go:build !baremetal && !js
// +build !baremetal,!js

package os_test

import (
	. "os"
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
	dir := "_os_test_TestChDir"
	Remove(dir)
	err := Mkdir(dir, 0755)
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
	err = Chdir("..")
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
