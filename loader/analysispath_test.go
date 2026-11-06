package loader

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/tinygo-org/tinygo/compileopts"
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

func TestAnalysisPatterns(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("TINYGOROOT", dir)
	files := map[string]string{
		"src/crypto/default.go":                 "package crypto\n",
		"src/crypto/tagged.go":                  "//go:build custom\n\npackage crypto\n",
		"src/device/arm/arm.go":                 "//go:build tinygo_root_test_marker\n\npackage arm\n",
		"src/internal/itoa/itoa.go":             "package itoa\n",
		"src/machine/machine.go":                "package machine\n",
		"src/machine/usb/usb.go":                "package usb\n",
		"src/machine/vendor/hidden/hidden.go":   "package hidden\n",
		"src/machine/testdata/hidden/hidden.go": "package hidden\n",
		"src/machine/_private/private.go":       "package private\n",
		"src/.hidden/hidden.go":                 "package hidden\n",
		"src/runtime/internal/sys/zversion.go":  "//go:build tinygo_root_test_marker\n\npackage sys\n",
	}
	for name, contents := range files {
		path := filepath.Join(dir, name)
		if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(contents), 0600); err != nil {
			t.Fatal(err)
		}
	}
	config := &compileopts.Config{
		Options:        &compileopts.Options{Directory: dir},
		Target:         &compileopts.TargetSpec{Triple: "x86_64-unknown-linux", GOOS: "linux", GOARCH: "amd64"},
		GoMinorVersion: 25,
	}
	for _, test := range []struct {
		patterns []string
		want     []string
	}{
		{
			patterns: []string{"./src/..."},
			want:     []string{"crypto", "internal/itoa", "machine", "machine/usb"},
		},
		{
			patterns: []string{"./src/machine/..."},
			want:     []string{"machine", "machine/usb"},
		},
		{
			patterns: []string{"./src/machine"},
			want:     []string{"machine"},
		},
	} {
		got, err := analysisPatterns(config, test.patterns)
		if err != nil {
			t.Fatal(err)
		}
		if !slices.Equal(got, test.want) {
			t.Errorf("analysisPatterns(%q) = %q, want %q", test.patterns, got, test.want)
		}
	}
	if _, err := analysisPatterns(config, []string{"./src"}); err == nil ||
		!strings.Contains(err.Error(), "use ./src/...") {
		t.Fatalf("analysisPatterns(./src) error = %v, want a ./src/... hint", err)
	}
}
