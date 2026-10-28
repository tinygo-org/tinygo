//go:build windows

package syscall

import (
	"io/fs"
	"testing"
)

func TestErrnoIsNotExist(t *testing.T) {
	for _, errno := range []Errno{2, 3, 53} {
		if !errno.Is(fs.ErrNotExist) {
			t.Errorf("Errno(%d).Is(fs.ErrNotExist) = false, want true", errno)
		}
	}
	if (Errno(5)).Is(fs.ErrNotExist) {
		t.Error("Errno(5).Is(fs.ErrNotExist) = true, want false")
	}
	if (Errno(2)).Is(fs.ErrExist) {
		t.Error("Errno(2).Is(fs.ErrExist) = true, want false")
	}
}
