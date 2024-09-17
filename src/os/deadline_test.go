//go:build posix && !baremetal && !js

package os_test

import (
	"errors"
	. "os"
	"testing"
)

func TestDeadlines(t *testing.T) {
	// Create a handle to a known-good, existing file
	f, err := Open("/dev/null")
	if err != nil {
		t.Fatal(err)
	}

	if err := f.SetDeadline(0); err == nil {
		if err != nil {
			t.Errorf("wanted nil, got %v", err)
		}
	}

	if err := f.SetDeadline(1); err == nil {
		if !errors.Is(err, ErrNotImplemented) {
			t.Errorf("wanted ErrNotImplemented, got %v", err)
		}
	}

	if err := f.SetReadDeadline(1); err == nil {
		if !errors.Is(err, ErrNotImplemented) {
			t.Errorf("wanted ErrNotImplemented, got %v", err)
		}
	}

	if err := f.SetWriteDeadline(1); err == nil {
		if !errors.Is(err, ErrNotImplemented) {
			t.Errorf("wanted ErrNotImplemented, got %v", err)
		}
	}

	// Closed files must return an error
	f.Close()

	if err := f.SetDeadline(0); err == nil {
		if err != ErrClosed {
			t.Errorf("wanted ErrClosed, got %v", err)
		}
	}
}
