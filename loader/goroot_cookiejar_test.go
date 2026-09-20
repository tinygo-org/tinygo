package loader

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"
)

func TestHTTPSubpackageOverrides(t *testing.T) {
	paths := pathsToOverride(26, false)
	src := filepath.Join("..", "src")
	err := filepath.WalkDir(filepath.Join(src, "net", "http"), func(dir string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !entry.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(src, dir)
		if err != nil {
			return err
		}
		path := filepath.ToSlash(rel) + "/"
		merge, ok := paths[path]
		if !ok {
			t.Errorf("missing override for %s", path)
		}
		if !merge {
			return filepath.SkipDir
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestHTTPSubpackageMerge(t *testing.T) {
	goRoot := t.TempDir()
	tinyRoot := t.TempDir()
	files := map[string][]string{
		goRoot:   {"client.go", "cgi/child.go", "fcgi/child.go", "cookiejar/jar.go", "httptest/server.go", "internal/ascii/print.go"},
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
	for _, path := range []string{"net/http/", "net/http/cgi/", "net/http/fcgi/", "net/http/httptest/", "net/http/internal/"} {
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
		"net/http/cgi":       tinyRoot,
		"net/http/fcgi":      tinyRoot,
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
