package builder

import (
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strconv"
	"testing"

	"tinygo.org/x/go-llvm"
)

func TestLookupCommandClangFallback(t *testing.T) {
	versioned := "clang-" + strconv.Itoa(llvm.VersionMajor)
	cands := commands["clang"]
	vIdx := slices.Index(cands, versioned)
	uIdx := slices.Index(cands, "clang")
	if vIdx < 0 || uIdx < 0 || vIdx >= uIdx {
		t.Fatalf("commands[\"clang\"] = %v, want %q before \"clang\"", cands, versioned)
	}

	if runtime.GOOS == "darwin" || runtime.GOOS == "freebsd" {
		return
	}

	dir := t.TempDir()
	t.Setenv("PATH", dir)

	ext := ""
	if runtime.GOOS == "windows" {
		ext = ".exe"
	}
	unversionedPath := filepath.Join(dir, "clang"+ext)
	if err := os.WriteFile(unversionedPath, []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatal(err)
	}

	got, err := LookupCommand("clang")
	if err != nil {
		t.Fatalf("LookupCommand(\"clang\") failed: %v", err)
	}
	if got != "clang" {
		t.Errorf("LookupCommand(\"clang\") = %q, want \"clang\"", got)
	}

	versionedPath := filepath.Join(dir, versioned+ext)
	if err := os.WriteFile(versionedPath, []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatal(err)
	}

	got, err = LookupCommand("clang")
	if err != nil {
		t.Fatalf("LookupCommand(\"clang\") failed: %v", err)
	}
	if got != versioned {
		t.Errorf("LookupCommand(\"clang\") = %q, want %q", got, versioned)
	}
}
