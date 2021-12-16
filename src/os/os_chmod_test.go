// +build !baremetal,!js,!wasi

// TODO: Move this into os_test.go (as upstream has it) when wasi supports chmod

package os_test

import (
	. "os"
	"runtime"
	"testing"
)

func TestChmod(t *testing.T) {
	f := newFile("TestChmod", t)
	defer Remove(f.Name())
	defer f.Close()
	// Creation mode is read write

	fm := FileMode(0456)
	if runtime.GOOS == "windows" {
		fm = FileMode(0444) // read-only file
	}
	if err := Chmod(f.Name(), fm); err != nil {
		t.Fatalf("chmod %s %#o: %s", f.Name(), fm, err)
	}
	checkMode(t, f.Name(), fm)
}
