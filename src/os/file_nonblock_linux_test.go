//go:build linux && !baremetal && !wasip1 && !wasip2 && !tinygo.wasm

package os_test

import (
	"errors"
	"os"
	"syscall"
	"testing"
	"time"
)

func nonblockPipe(t *testing.T) (*os.File, int) {
	t.Helper()
	var p [2]int
	if err := syscall.Pipe2(p[:], syscall.O_NONBLOCK|syscall.O_CLOEXEC); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { syscall.Close(p[1]) })
	return os.NewFile(uintptr(p[0]), "nonblock"), p[1]
}

func TestReadNonblockWaits(t *testing.T) {
	f, w := nonblockPipe(t)
	defer f.Close()
	time.AfterFunc(100*time.Millisecond, func() { syscall.Write(w, []byte("x")) })
	b := make([]byte, 4)
	if n, err := f.Read(b); err != nil || n != 1 {
		t.Fatalf("Read = %d, %v; want 1, nil", n, err)
	}
}

func TestCloseWakesNonblockRead(t *testing.T) {
	f, _ := nonblockPipe(t)
	done := make(chan error, 1)
	go func() {
		_, err := f.Read(make([]byte, 4))
		done <- err
	}()
	time.Sleep(100 * time.Millisecond)
	select {
	case err := <-done:
		t.Fatalf("Read returned %v before Close", err)
	default:
	}
	f.Close()
	select {
	case err := <-done:
		if !errors.Is(err, os.ErrClosed) {
			t.Fatalf("Read after Close = %v, want ErrClosed", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Read still blocked 2s after Close")
	}
}

func TestWriteNonblockWaits(t *testing.T) {
	var p [2]int
	if err := syscall.Pipe2(p[:], syscall.O_NONBLOCK|syscall.O_CLOEXEC); err != nil {
		t.Fatal(err)
	}
	r := os.NewFile(uintptr(p[0]), "r")
	w := os.NewFile(uintptr(p[1]), "w")
	defer r.Close()
	got := make(chan int, 1)
	go func() {
		n := 0
		b := make([]byte, 64<<10)
		for {
			m, err := r.Read(b)
			n += m
			if err != nil {
				break
			}
		}
		got <- n
	}()
	const size = 1 << 20
	if n, err := w.Write(make([]byte, size)); err != nil || n != size {
		t.Fatalf("Write = %d, %v; want %d, nil", n, err, size)
	}
	w.Close()
	if n := <-got; n != size {
		t.Fatalf("read %d bytes, want %d", n, size)
	}
}
