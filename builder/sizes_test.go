package builder

import (
	"fmt"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/tinygo-org/tinygo/compileopts"
	"github.com/tinygo-org/tinygo/goenv"
	"github.com/tinygo-org/tinygo/interp"
)

var sema = make(chan struct{}, runtime.NumCPU())

func TestTrimPathPackageSize(t *testing.T) {
	root := goenv.Get("TINYGOROOT")
	packages := map[string]string{
		filepath.FromSlash("example.com/main"):                         "main",
		filepath.FromSlash("example.com/dependency@v1.2.3/subpackage"): "example.com/dependency/subpackage",
		filepath.Join(root, "src", "runtime"):                          "runtime",
	}
	for _, tc := range []struct{ path, pkg, file string }{
		{"example.com/main/main.go", "main", "main.go"},
		{"/_/example.com/main/main.c", "main", "main.c"},
		{"//_/_/example.com/main/main.c", "main", "main.c"},
		{"/_/example.com/dependency@v1.2.3/subpackage/dependency.c", "example.com/dependency/subpackage", "dependency.c"},
		{"/_/github.com/tinygo-org/tinygo/src/runtime/runtime_unix.c", "runtime", "runtime_unix.c"},
		{"/_/github.com/tinygo-org/tinygo/lib/musl/src/stdio/puts.c", "C musl", filepath.FromSlash("src/stdio/puts.c")},
		{"//_/_/github.com/tinygo-org/tinygo/lib/bdwgc/alloc.c", "C bdwgc", "alloc.c"},
	} {
		pkg, file := findPackagePath(filepath.FromSlash(tc.path), packages)
		if pkg != tc.pkg || file != tc.file {
			t.Errorf("findPackagePath(%q) = %q, %q; want %q, %q", tc.path, pkg, file, tc.pkg, tc.file)
		}
	}
}

// Check that the -size=full flag attributes binary size to the correct package.
func TestSizeFull(t *testing.T) {
	tests := []struct {
		target   string
		trimPath bool
	}{
		{"microbit", false},
		{"microbit", true},
		{"wasip1", false},
		{"wasip1", true},
	}

	libMatch := regexp.MustCompile(`^C [a-z -]+$`) // example: "C interrupt vector"
	root := goenv.Get("TINYGOROOT")

	for _, test := range tests {
		t.Run(fmt.Sprintf("%s/trimpath=%v", test.target, test.trimPath), func(t *testing.T) {
			t.Parallel()

			// Build the binary.
			result := buildBinary(t, test.target, "examples/serial", test.trimPath)

			// Check whether the binary doesn't contain any unexpected package
			// names.
			sizes, err := loadProgramSize(result.Executable, result.PackagePathMap)
			if err != nil {
				t.Fatal("could not read program size:", err)
			}
			for _, pkg := range sizes.sortedPackageNames() {
				if pkg == "(padding)" || pkg == "(unknown)" || pkg == "Go types" {
					// TODO: correctly attribute all unknown binary size.
					continue
				}
				if libMatch.MatchString(pkg) {
					continue
				}
				if isSizePackagePath(pkg, root, test.trimPath) {
					continue
				}
				t.Error("unexpected package name in size output:", pkg)
			}
		})
	}
}

var sizePackagePathMatch = regexp.MustCompile(`^[a-z/]+$`)

func isSizePackagePath(pkg, root string, trimPath bool) bool {
	if filepath.IsAbs(pkg) {
		if trimPath {
			return false
		}
		rel, err := filepath.Rel(root, pkg)
		return err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))
	}
	return sizePackagePathMatch.MatchString(pkg)
}

func TestSizePackagePath(t *testing.T) {
	root := filepath.Join(t.TempDir(), "TinyGo Root")
	lowercaseRoot := filepath.Join(filepath.VolumeName(root)+string(filepath.Separator), "home", "runner", "work", "tinygo", "tinygo")
	tests := []struct {
		name      string
		root      string
		pkg       string
		untrimmed bool
		trimmed   bool
	}{
		{"package", root, "internal/task", true, true},
		{"empty package", root, "", false, false},
		{"uppercase package", root, "Internal/task", false, false},
		{"package with spaces", root, "internal/task name", false, false},
		{"native root", root, root, true, false},
		{"native inside root", root, filepath.Join(root, "lib", "nrfx", "mdk"), true, false},
		{"native uppercase and spaces", root, filepath.Join(root, "lib", "Vendor SDK", "mdk"), true, false},
		{"native outside root", root, filepath.Join(filepath.Dir(root), "outside", "lib", "nrfx", "mdk"), false, false},
		{"native parent", root, filepath.Dir(root), false, false},
		{"native sibling prefix", root, filepath.Join(root+"other", "lib", "nrfx", "mdk"), false, false},
		{"native dot prefix inside root", root, filepath.Join(root, "..vendor", "mdk"), true, false},
		{"lowercase root", lowercaseRoot, lowercaseRoot, true, false},
		{"lowercase inside root", lowercaseRoot, filepath.Join(lowercaseRoot, "lib", "nrfx", "mdk"), true, false},
		{"lowercase outside root", lowercaseRoot, filepath.Join(filepath.Dir(lowercaseRoot), "other", "lib", "nrfx", "mdk"), false, false},
		{"lowercase parent", lowercaseRoot, filepath.Dir(lowercaseRoot), false, false},
		{"lowercase sibling prefix", lowercaseRoot, filepath.Join(lowercaseRoot+"other", "lib", "nrfx", "mdk"), false, false},
	}
	for _, test := range tests {
		for _, trimPath := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/trimpath=%v", test.name, trimPath), func(t *testing.T) {
				want := test.untrimmed
				if trimPath {
					want = test.trimmed
				}
				if got := isSizePackagePath(test.pkg, test.root, trimPath); got != want {
					t.Errorf("isSizePackagePath(%q, %q, %v) = %v; want %v", test.pkg, test.root, trimPath, got, want)
				}
			})
		}
	}
}

func buildBinary(t *testing.T, targetString, pkgName string, trimPath bool) *BuildResult {
	t.Helper()
	result, err := buildBinaryInDir(targetString, pkgName, "z", t.TempDir(), trimPath)
	if err != nil {
		t.Fatal(err)
	}
	return result
}

func buildBinaryInDir(targetString, pkgName, opt, tmpdir string, trimPath bool) (*BuildResult, error) {
	options := compileopts.Options{
		Target:                  targetString,
		Opt:                     opt,
		Semaphore:               sema,
		InterpTimeout:           60 * time.Second,
		InterpMaxLoopIterations: interp.DefaultMaxInterpBlockEntries,
		Debug:                   true,
		VerifyIR:                true,
		TrimPath:                trimPath,
		// The GOROOT string is stored in data, so its length must not depend on the host.
		GlobalValues: map[string]map[string]string{"runtime": {"goroot": "/usr/local/go"}},
	}
	target, err := compileopts.LoadTarget(&options)
	if err != nil {
		return nil, fmt.Errorf("could not load target: %w", err)
	}
	config := &compileopts.Config{
		Options: &options,
		Target:  target,
	}
	result, err := Build(pkgName, "", tmpdir, config)
	if err != nil {
		return nil, fmt.Errorf("could not build: %w", err)
	}
	return result, nil
}
