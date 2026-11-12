//go:build !baremetal && !tinygo.wasm && !nintendoswitch

package os_test

import (
	"errors"
	"os"
	"syscall"
	"testing"
	"time"
)

func TestPipeReadDeadlineAndSyscallConn(t *testing.T) {
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()
	defer w.Close()

	rc, err := r.SyscallConn()
	if err != nil {
		t.Fatalf("SyscallConn failed: %v", err)
	}

	var gotFD uintptr
	if err := rc.Control(func(fd uintptr) {
		gotFD = fd
	}); err != nil {
		t.Fatalf("RawConn.Control failed: %v", err)
	}
	if gotFD == 0 {
		t.Fatal("RawConn.Control reported zero fd")
	}

	// Set a past read deadline and verify Read returns ErrDeadlineExceeded.
	if err := r.SetReadDeadline(time.Now().Add(-time.Millisecond)); err != nil {
		t.Fatalf("SetReadDeadline failed: %v", err)
	}
	buf := make([]byte, 16)
	_, err = r.Read(buf)
	if !errors.Is(err, os.ErrDeadlineExceeded) {
		t.Fatalf("Read with expired deadline = %v, want ErrDeadlineExceeded", err)
	}

	// Clear the deadline and verify RawConn.Read waits until data arrives.
	if err := r.SetReadDeadline(time.Time{}); err != nil {
		t.Fatalf("SetReadDeadline(zero) failed: %v", err)
	}
	go func() {
		time.Sleep(10 * time.Millisecond)
		_, _ = w.Write([]byte("ping"))
	}()

	var n int
	if err := rc.Read(func(fd uintptr) bool {
		var rerr error
		n, rerr = syscall.Read(int(fd), buf)
		return rerr != syscall.EAGAIN
	}); err != nil {
		t.Fatalf("RawConn.Read failed: %v", err)
	}
	if string(buf[:n]) != "ping" {
		t.Errorf("RawConn.Read got %q, want \"ping\"", string(buf[:n]))
	}
}
