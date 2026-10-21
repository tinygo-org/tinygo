package loader

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/tinygo-org/tinygo/goenv"
)

func TestAnalysisPath(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "target")
	if err := os.Mkdir(target, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(target, "probe.go"), []byte("package probe\n"), 0600); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(dir, "link")
	if err := symlink(target, link); err != nil {
		t.Fatal(err)
	}
	for _, suffix := range []string{"", "probe.go"} {
		want, err := analysisPath(filepath.Join(target, suffix))
		if err != nil {
			t.Fatal(err)
		}
		got, err := analysisPath(filepath.Join(link, suffix))
		if err != nil {
			t.Fatal(err)
		}
		if got != want {
			t.Fatalf("resolved path = %q, want %q", got, want)
		}
	}
	if _, err := analysisPath(filepath.Join(dir, "missing")); !os.IsNotExist(err) {
		t.Fatalf("missing path: %v", err)
	}
}

func TestAnalysisGoRootPath(t *testing.T) {
	root := filepath.Join(goenv.Get("GOROOT"), "src")
	resolved, err := analysisPath(root)
	if err != nil {
		t.Fatal(err)
	}
	original, err := os.Stat(root)
	if err != nil {
		t.Fatal(err)
	}
	actual, err := os.Stat(resolved)
	if err != nil {
		t.Fatal(err)
	}
	if !os.SameFile(original, actual) {
		t.Fatalf("%q and %q refer to different directories", root, resolved)
	}
}
