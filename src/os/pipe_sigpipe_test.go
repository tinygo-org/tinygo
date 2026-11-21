//go:build darwin || (linux && !baremetal && !wasip1 && !wasip2)

package os_test

import (
	"errors"
	"os"
	"syscall"
	"testing"
)

// A write to a pipe without a reader returns EPIPE instead of killing the
// process with SIGPIPE, as in Go.
func TestPipeWriteWithoutReader(t *testing.T) {
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	defer w.Close()
	r.Close()
	if _, err := w.Write([]byte("x")); !errors.Is(err, syscall.EPIPE) {
		t.Fatalf("got %v, want EPIPE", err)
	}
}
