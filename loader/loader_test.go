package loader

import (
	"path/filepath"
	"slices"
	"testing"

	"github.com/tinygo-org/tinygo/compileopts"
	"github.com/tinygo-org/tinygo/goenv"
)

func TestRecordedPath(t *testing.T) {
	config := &compileopts.Config{
		Options: &compileopts.Options{TrimPath: true},
		Target:  &compileopts.TargetSpec{GOOS: "linux"},
	}
	program := &Program{
		config:   config,
		goroot:   filepath.FromSlash("/goroot"),
		Packages: make(map[string]*Package),
	}
	pkg := &Package{
		program: program,
		PackageJSON: PackageJSON{
			Dir:        filepath.FromSlash("/tmp/module"),
			ImportPath: "example.com/main",
		},
	}
	pkg.Module.Path = "example.com/main"
	pkg.Module.Dir = filepath.FromSlash("/tmp/module")
	dependency := &Package{
		program: program,
		PackageJSON: PackageJSON{
			Dir:        filepath.FromSlash("/tmp/dependency/subpackage"),
			ImportPath: "example.com/dependency/subpackage",
		},
	}
	dependency.Module.Path = "example.com/dependency"
	dependency.Module.Version = "v1.2.3"
	dependency.Module.Dir = filepath.FromSlash("/tmp/dependency")
	dependency.CFlags = []string{
		"-I" + filepath.FromSlash("/tmp/dependency/include"),
		"-isystem",
		filepath.FromSlash("/tmp/dependency/system"),
		"-iquote" + filepath.FromSlash("/tmp/dependency/include"),
		"-ffile-prefix-map=" + filepath.FromSlash("/tmp/dependency") + "=/mapped",
		`-DCONFIG_PATH="/tmp/dependency"`,
	}
	program.Packages[pkg.ImportPath] = pkg
	program.Packages[dependency.ImportPath] = dependency

	if got, want := dependency.RecordedDir(), "example.com/dependency@v1.2.3/subpackage"; got != want {
		t.Fatalf("RecordedDir() = %q, want %q", got, want)
	}
	filename := filepath.FromSlash("/tmp/dependency/subpackage/file.go")
	if got, want := pkg.RecordedPath(filename), "example.com/dependency@v1.2.3/subpackage/file.go"; got != want {
		t.Fatalf("RecordedPath() = %q, want %q", got, want)
	}
	header := filepath.FromSlash("/tmp/dependency/include/shared.h")
	if got, want := pkg.RecordedPath(header), "example.com/dependency@v1.2.3/include/shared.h"; got != want {
		t.Fatalf("RecordedPath() = %q, want %q", got, want)
	}
	if got, want := dependency.DebugPrefixMap(), "-ffile-prefix-map="+filepath.FromSlash("/tmp/dependency")+"=/_/example.com/dependency@v1.2.3"; got != want {
		t.Fatalf("DebugPrefixMap() = %q, want %q", got, want)
	}
	if got, want := dependency.RecordedCFlags(), []string{
		"-Iexample.com/dependency@v1.2.3/include",
		"-isystem",
		"example.com/dependency@v1.2.3/system",
		"-iquoteexample.com/dependency@v1.2.3/include",
		"-ffile-prefix-map=example.com/dependency@v1.2.3=/mapped",
		`-DCONFIG_PATH="/tmp/dependency"`,
	}; !slices.Equal(got, want) {
		t.Fatalf("RecordedCFlags() = %q, want %q", got, want)
	}

	for _, root := range []string{program.goroot, goenv.Get("GOROOT"), goenv.Get("TINYGOROOT")} {
		filename := filepath.Join(root, "src", "runtime", "header.h")
		if got := pkg.RecordedPath(filename); got != "runtime/header.h" {
			t.Errorf("RecordedPath(%q) = %q", filename, got)
		}
		if got := pkg.recordedCFlagPath(filename); got != "runtime/header.h" {
			t.Errorf("recordedCFlagPath(%q) = %q", filename, got)
		}
	}
	outside := filepath.FromSlash("/tmp/module-other/include")
	if got := pkg.recordedCFlagPath(outside); got != outside {
		t.Errorf("rewrote a path outside the module: %q", got)
	}

	vendored := &Package{
		program: program,
		PackageJSON: PackageJSON{
			Dir:        filepath.FromSlash("/tmp/main/vendor/example.com/dependency/subpackage"),
			ImportPath: "example.com/dependency/subpackage",
		},
	}
	vendored.Module.Path = "example.com/dependency"
	vendored.Module.Version = "v1.2.3"
	program.Packages[vendored.ImportPath] = vendored
	if got, want := vendored.OriginalModuleDir(), filepath.FromSlash("/tmp/main/vendor/example.com/dependency"); got != want {
		t.Fatalf("vendored OriginalModuleDir() = %q, want %q", got, want)
	}
	vendoredHeader := filepath.FromSlash("/tmp/main/vendor/example.com/dependency/include/shared.h")
	if got, want := pkg.RecordedPath(vendoredHeader), "example.com/dependency@v1.2.3/include/shared.h"; got != want {
		t.Fatalf("vendored RecordedPath() = %q, want %q", got, want)
	}
	if got, want := vendored.DebugPrefixMap(), "-ffile-prefix-map="+filepath.FromSlash("/tmp/main/vendor/example.com/dependency")+"=/_/example.com/dependency@v1.2.3"; got != want {
		t.Fatalf("vendored DebugPrefixMap() = %q, want %q", got, want)
	}

	gopath := &Package{
		program: program,
		PackageJSON: PackageJSON{
			Dir:        filepath.FromSlash("/gopath/src/example.com/dependency/subpackage"),
			ImportPath: "example.com/dependency/subpackage",
			Root:       filepath.FromSlash("/gopath"),
		},
	}
	program.Packages[gopath.ImportPath] = gopath
	if got, want := gopath.OriginalModuleDir(), filepath.FromSlash("/gopath/src"); got != want {
		t.Fatalf("GOPATH OriginalModuleDir() = %q, want %q", got, want)
	}
	gopathHeader := filepath.FromSlash("/gopath/src/example.com/dependency/include/shared.h")
	if got, want := pkg.RecordedPath(gopathHeader), "example.com/dependency/include/shared.h"; got != want {
		t.Fatalf("GOPATH RecordedPath() = %q, want %q", got, want)
	}
	if got, want := gopath.DebugPrefixMap(), "-ffile-prefix-map="+filepath.FromSlash("/gopath/src")+"=/_"; got != want {
		t.Fatalf("GOPATH DebugPrefixMap() = %q, want %q", got, want)
	}
}
