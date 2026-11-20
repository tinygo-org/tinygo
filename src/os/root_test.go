package os_test

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRootOperationsAndTraversal(t *testing.T) {
	outer := t.TempDir()
	secretFile := filepath.Join(outer, "secret.txt")
	if err := os.WriteFile(secretFile, []byte("secret"), 0o600); err != nil {
		t.Fatalf("WriteFile secret: %v", err)
	}

	rootDir := filepath.Join(outer, "root")
	if err := os.Mkdir(rootDir, 0o755); err != nil {
		t.Fatalf("Mkdir root: %v", err)
	}

	r, err := os.OpenRoot(rootDir)
	if err != nil {
		t.Fatalf("OpenRoot: %v", err)
	}
	defer r.Close()

	if err := r.MkdirAll("a/b", 0o755); err != nil {
		t.Fatalf("r.MkdirAll: %v", err)
	}
	if err := r.WriteFile("a/b/hello.txt", []byte("world"), 0o644); err != nil {
		t.Fatalf("r.WriteFile: %v", err)
	}
	got, err := r.ReadFile("a/b/../b/hello.txt")
	if err != nil {
		t.Fatalf("r.ReadFile: %v", err)
	}
	if string(got) != "world" {
		t.Fatalf("r.ReadFile = %q, want %q", got, "world")
	}

	f, err := os.OpenInRoot(rootDir, "a/b/hello.txt")
	if err != nil {
		t.Fatalf("OpenInRoot: %v", err)
	}
	f.Close()

	for _, bad := range []string{"", "/", "/etc/passwd", "../secret.txt", "a/../../secret.txt"} {
		if _, err := r.ReadFile(bad); err == nil {
			t.Fatalf("r.ReadFile(%q) succeeded, expected traversal error", bad)
		}
	}

	if err := r.Symlink("../../secret.txt", "a/escape_rel"); err != nil {
		t.Fatalf("r.Symlink relative: %v", err)
	}
	if _, err := r.ReadFile("a/escape_rel"); err == nil {
		t.Fatal("r.ReadFile through relative escaping symlink succeeded, expected error")
	}

	if err := r.Symlink(secretFile, "a/escape_abs"); err != nil {
		t.Fatalf("r.Symlink absolute: %v", err)
	}
	if _, err := r.ReadFile("a/escape_abs"); err == nil {
		t.Fatal("r.ReadFile through absolute symlink succeeded, expected error")
	}

	if err := r.Symlink("b/hello.txt", "a/valid_link"); err != nil {
		t.Fatalf("r.Symlink valid: %v", err)
	}
	got, err = r.ReadFile("a/valid_link")
	if err != nil || string(got) != "world" {
		t.Fatalf("r.ReadFile valid symlink = (%q, %v), want (%q, nil)", got, err, "world")
	}
}
