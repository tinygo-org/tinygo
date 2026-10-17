//go:build wasm

package pass_test

import (
	"os"
	"path"
	"runtime"
	"testing"
)

var initialWorkingDir string
var initialWorkingDirError error
var initialFileError error

func init() {
	if runtime.GOOS == "wasip1" {
		initialWorkingDir, initialWorkingDirError = os.Getwd()
		_, initialFileError = os.ReadFile("pass_wasm_test.go")
	}
}

func TestWASIWorkingDirectory(t *testing.T) {
	if runtime.GOOS != "wasip1" {
		t.Skip("WASIp1 working-directory initialization")
	}
	if initialWorkingDirError != nil {
		t.Fatal(initialWorkingDirError)
	}
	if initialFileError != nil {
		t.Fatal(initialFileError)
	}
	if pwd := os.Getenv("PWD"); pwd != "" && initialWorkingDir != path.Join("/", pwd) {
		t.Fatalf("initial cwd = %q, want %q", initialWorkingDir, path.Join("/", pwd))
	}
}

func TestDeepSuspendWithDefer(t *testing.T) {
	ready := make(chan struct{})
	release := make(chan struct{})
	done := make(chan struct{})
	go func() {
		deepSuspendWithDefer(512, ready, release)
		close(done)
	}()
	<-ready
	close(release)
	<-done
}

func deepSuspendWithDefer(depth int, ready chan<- struct{}, release <-chan struct{}) {
	defer func() {}()
	if depth == 0 {
		ready <- struct{}{}
		<-release
		return
	}
	deepSuspendWithDefer(depth-1, ready, release)
}
