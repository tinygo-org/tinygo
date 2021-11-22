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
