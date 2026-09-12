package loader

import (
	"os"
	"path/filepath"
	"testing"
)

func TestHTTPSubpackageMerge(t *testing.T) {
	goRoot := t.TempDir()
	tinyRoot := t.TempDir()
	files := map[string][]string{
		goRoot:   {"client.go", "cookiejar/jar.go", "httptest/server.go", "internal/ascii/print.go"},
		tinyRoot: {"client.go", "httptest/server.go", "internal/ascii/print.go"},
	}
	for root, names := range files {
		for _, name := range names {
			file := filepath.Join(root, "src/net/http", name)
			if err := os.MkdirAll(filepath.Dir(file), 0755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(file, []byte("package http\n"), 0644); err != nil {
				t.Fatal(err)
			}
		}
	}
	paths := pathsToOverride(26, false)
	overrides := make(map[string]bool)
	for _, path := range []string{"net/http/", "net/http/httptest/", "net/http/internal/"} {
		value, ok := paths[path]
		if !ok {
			t.Fatalf("missing override for %s", path)
		}
		overrides[path] = value
	}
	links, err := listGorootMergeLinks(goRoot, tinyRoot, overrides)
	if err != nil {
		t.Fatal(err)
	}
	for path, root := range map[string]string{
		"net/http/client.go": tinyRoot,
		"net/http/cookiejar": goRoot,
		"net/http/httptest":  tinyRoot,
		"net/http/internal":  tinyRoot,
	} {
		key := filepath.Join("src", path)
		if want := filepath.Join(root, key); links[key] != want {
			t.Errorf("%s links to %q, want %q", path, links[key], want)
		}
	}
}
