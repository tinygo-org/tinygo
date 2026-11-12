//go:build darwin || (linux && !baremetal && !wasip1 && !wasm_unknown && !wasip2 && !nintendoswitch)

package os_test

import (
	"os"
	"path/filepath"
	"syscall"
	"testing"
	_ "unsafe"
)

//go:linkname syscall_fcntl syscall.fcntl
func syscall_fcntl(fd int, cmd int, arg int) (val int, err error)

func TestOpenFileCloexec(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.txt")
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	flags, err := syscall_fcntl(int(f.Fd()), syscall.F_GETFD, 0)
	if err != nil {
		t.Fatalf("fcntl(F_GETFD) failed: %v", err)
	}
	if flags&syscall.FD_CLOEXEC == 0 {
		t.Errorf("fcntl(F_GETFD) = %#x, want FD_CLOEXEC (%#x) set", flags, syscall.FD_CLOEXEC)
	}
}
