//go:build !baremetal && !tinygo.wasm && !nintendoswitch

package os_test

import (
	"errors"
	"io"
	"math"
	"os"
	"syscall"
	"testing"
	"time"
)

func fdFlags(t *testing.T, fd uintptr) int {
	t.Helper()
	flags, _, errno := syscall.Syscall(syscall.SYS_FCNTL, fd, syscall.F_GETFL, 0)
	if errno != 0 {
		t.Fatalf("fcntl(F_GETFL) failed: %v", errno)
	}
	return int(flags)
}

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

	gotFD := ^uintptr(0)
	if err := rc.Control(func(fd uintptr) {
		gotFD = fd
	}); err != nil {
		t.Fatalf("RawConn.Control failed: %v", err)
	}
	if int(gotFD) < 0 {
		t.Fatalf("RawConn.Control fd = %d, want >= 0", int(gotFD))
	}

	if err := r.SetReadDeadline(time.Now().Add(-time.Millisecond)); err != nil {
		t.Fatalf("SetReadDeadline(past) failed: %v", err)
	}
	buf := make([]byte, 16)
	if _, err := r.Read(buf); !errors.Is(err, os.ErrDeadlineExceeded) {
		t.Fatalf("Read with past deadline = %v, want ErrDeadlineExceeded", err)
	}

	start := time.Now()
	if err := r.SetReadDeadline(start.Add(20 * time.Millisecond)); err != nil {
		t.Fatalf("SetReadDeadline(future) failed: %v", err)
	}
	if _, err := r.Read(buf); !errors.Is(err, os.ErrDeadlineExceeded) {
		t.Fatalf("Read with future deadline = %v, want ErrDeadlineExceeded", err)
	}
	if elapsed := time.Since(start); elapsed < 15*time.Millisecond {
		t.Fatalf("Read timed out too early after %v", elapsed)
	}

	if err := r.SetReadDeadline(time.Time{}); err != nil {
		t.Fatalf("SetReadDeadline(zero) failed: %v", err)
	}
	wrc, err := w.SyscallConn()
	if err != nil {
		t.Fatalf("writer SyscallConn failed: %v", err)
	}
	go func() {
		time.Sleep(10 * time.Millisecond)
		_ = wrc.Write(func(fd uintptr) bool {
			_, werr := syscall.Write(int(fd), []byte("ping"))
			return werr != syscall.EAGAIN
		})
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
		t.Errorf("RawConn.Read got %q, want %q", string(buf[:n]), "ping")
	}
}

func TestSetReadDeadlineWhileBlocked(t *testing.T) {
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()
	defer w.Close()

	done := make(chan error, 1)
	go func() {
		buf := make([]byte, 8)
		_, err := r.Read(buf)
		done <- err
	}()

	time.Sleep(15 * time.Millisecond)
	if err := r.SetReadDeadline(time.Now().Add(15 * time.Millisecond)); err != nil {
		t.Fatalf("SetReadDeadline failed: %v", err)
	}
	if err := <-done; !errors.Is(err, os.ErrDeadlineExceeded) {
		t.Fatalf("in-flight Read = %v, want ErrDeadlineExceeded", err)
	}
}

func TestFarFutureDeadlineNoOverflow(t *testing.T) {
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()
	defer w.Close()

	if err := r.SetReadDeadline(time.Now().Add(time.Duration(math.MaxInt64))); err != nil {
		t.Fatalf("SetReadDeadline(MaxInt64) failed: %v", err)
	}
	go func() {
		time.Sleep(10 * time.Millisecond)
		_, _ = w.Write([]byte("ok"))
	}()
	buf := make([]byte, 8)
	n, err := r.Read(buf)
	if err != nil {
		t.Fatalf("Read with MaxInt64 deadline failed: %v", err)
	}
	if string(buf[:n]) != "ok" {
		t.Fatalf("Read got %q, want %q", string(buf[:n]), "ok")
	}
}

func TestPipeFdRestoresBlockingAndNewFileKeepsBlocking(t *testing.T) {
	var raw [2]int
	if err := syscall.Pipe2(raw[:], syscall.O_CLOEXEC); err != nil {
		t.Fatal(err)
	}
	rf := os.NewFile(uintptr(raw[0]), "raw-r")
	wf := os.NewFile(uintptr(raw[1]), "raw-w")
	defer rf.Close()
	defer wf.Close()
	if fdFlags(t, uintptr(raw[0]))&syscall.O_NONBLOCK != 0 {
		t.Fatal("NewFile on blocking pipe unexpectedly set O_NONBLOCK")
	}

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()
	defer w.Close()

	rc, err := r.SyscallConn()
	if err != nil {
		t.Fatal(err)
	}
	var sysfd uintptr
	if err := rc.Control(func(fd uintptr) {
		sysfd = fd
	}); err != nil {
		t.Fatal(err)
	}
	if fdFlags(t, sysfd)&syscall.O_NONBLOCK == 0 {
		t.Fatal("os.Pipe reader expected O_NONBLOCK before Fd()")
	}
	if got := r.Fd(); got != sysfd {
		t.Fatalf("r.Fd() = %d, want %d", got, sysfd)
	}
	if fdFlags(t, sysfd)&syscall.O_NONBLOCK != 0 {
		t.Fatal("r.Fd() did not clear O_NONBLOCK")
	}
}

func TestRegularFileAndDevNullErrNoDeadline(t *testing.T) {
	f, err := os.CreateTemp(t.TempDir(), "poll-reg-*")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	rc, err := f.SyscallConn()
	if err != nil {
		t.Fatalf("SyscallConn on regular file failed: %v", err)
	}
	var called bool
	if err := rc.Control(func(fd uintptr) {
		called = fd == f.Fd()
	}); err != nil || !called {
		t.Fatalf("RawConn.Control = %v, called = %v", err, called)
	}
	if err := f.SetDeadline(time.Now().Add(time.Second)); !errors.Is(err, os.ErrNoDeadline) {
		t.Fatalf("SetDeadline on regular file = %v, want ErrNoDeadline", err)
	}

	dn, err := os.Open(os.DevNull)
	if err != nil {
		t.Fatal(err)
	}
	defer dn.Close()
	if err := dn.SetDeadline(time.Now().Add(time.Second)); !errors.Is(err, os.ErrNoDeadline) {
		t.Fatalf("SetDeadline on DevNull = %v, want ErrNoDeadline", err)
	}
}

func TestPipeCloseUnblocksRead(t *testing.T) {
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	defer w.Close()

	done := make(chan error, 1)
	go func() {
		buf := make([]byte, 8)
		_, err := r.Read(buf)
		done <- err
	}()

	time.Sleep(15 * time.Millisecond)
	if err := r.Close(); err != nil {
		t.Fatalf("Close failed: %v", err)
	}
	if err := <-done; !errors.Is(err, os.ErrClosed) {
		t.Fatalf("Read after Close = %v, want ErrClosed", err)
	}
	if got := r.Fd(); got != ^uintptr(0) {
		t.Fatalf("Fd after Close = %d, want ^uintptr(0)", got)
	}
}

func TestSocketpairConcurrentReadWriteWake(t *testing.T) {
	fds, err := syscall.Socketpair(syscall.AF_UNIX, syscall.SOCK_STREAM, 0)
	if err != nil {
		t.Fatal(err)
	}
	for _, fd := range fds {
		syscall.CloseOnExec(fd)
		if err := syscall.SetNonblock(fd, true); err != nil {
			t.Fatal(err)
		}
	}
	_ = syscall.SetsockoptInt(fds[0], syscall.SOL_SOCKET, syscall.SO_SNDBUF, 4096)

	sock0 := os.NewFile(uintptr(fds[0]), "sock0")
	sock1 := os.NewFile(uintptr(fds[1]), "sock1")
	defer sock0.Close()
	defer sock1.Close()

	readBuf := make([]byte, 5)
	readDone := make(chan error, 1)
	go func() {
		_, rerr := sock0.Read(readBuf)
		readDone <- rerr
	}()

	writeDone := make(chan error, 1)
	go func() {
		_, werr := sock0.Write(make([]byte, 512*1024))
		writeDone <- werr
	}()

	time.Sleep(20 * time.Millisecond)
	if err := sock0.SetWriteDeadline(time.Now().Add(-time.Millisecond)); err != nil {
		t.Fatal(err)
	}
	if err := <-writeDone; !errors.Is(err, os.ErrDeadlineExceeded) {
		t.Fatalf("blocked Write = %v, want ErrDeadlineExceeded", err)
	}
	select {
	case err := <-readDone:
		t.Fatalf("Read unexpectedly unblocked on SetWriteDeadline: %v", err)
	case <-time.After(10 * time.Millisecond):
	}

	if _, err := sock1.Write([]byte("hello")); err != nil {
		t.Fatal(err)
	}
	if err := <-readDone; err != nil {
		t.Fatalf("Read after peer write = %v, want nil", err)
	}
	if string(readBuf) != "hello" {
		t.Fatalf("Read got %q, want %q", string(readBuf), "hello")
	}

	if err := sock0.SetWriteDeadline(time.Time{}); err != nil {
		t.Fatal(err)
	}
	readDone2 := make(chan error, 1)
	go func() {
		_, rerr := sock0.Read(readBuf)
		readDone2 <- rerr
	}()
	writeDone2 := make(chan error, 1)
	go func() {
		_, werr := sock0.Write(make([]byte, 512*1024))
		writeDone2 <- werr
	}()

	time.Sleep(20 * time.Millisecond)
	if err := sock0.SetReadDeadline(time.Now().Add(-time.Millisecond)); err != nil {
		t.Fatal(err)
	}
	if err := <-readDone2; !errors.Is(err, os.ErrDeadlineExceeded) {
		t.Fatalf("blocked Read = %v, want ErrDeadlineExceeded", err)
	}
	select {
	case err := <-writeDone2:
		t.Fatalf("Write unexpectedly unblocked on SetReadDeadline: %v", err)
	case <-time.After(10 * time.Millisecond):
	}

	go func() {
		_, _ = io.Copy(io.Discard, sock1)
	}()
	if err := <-writeDone2; err != nil {
		t.Fatalf("Write after peer drain = %v, want nil", err)
	}
}
