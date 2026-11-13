//go:build !baremetal && !nintendoswitch && !wasm_unknown && !wasip2

package os_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestChtimes(t *testing.T) {
	path := filepath.Join(t.TempDir(), "file.txt")
	if err := os.WriteFile(path, []byte("hello"), 0o600); err != nil {
		t.Fatal(err)
	}

	wantMtime := time.Date(2025, 3, 15, 12, 30, 45, 0, time.UTC)
	if err := os.Chtimes(path, wantMtime, wantMtime); err != nil {
		t.Fatalf("Chtimes failed: %v", err)
	}

	fi, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if !fi.ModTime().Equal(wantMtime) {
		t.Errorf("ModTime() = %v, want %v", fi.ModTime(), wantMtime)
	}

	// Zero time.Time should leave the modification time unchanged.
	if err := os.Chtimes(path, time.Time{}, time.Time{}); err != nil {
		t.Fatalf("Chtimes with zero times failed: %v", err)
	}
	fi, err = os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if !fi.ModTime().Equal(wantMtime) {
		t.Errorf("ModTime() after zero Chtimes = %v, want %v", fi.ModTime(), wantMtime)
	}
}
