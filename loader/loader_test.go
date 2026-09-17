package loader

import (
	"path/filepath"
	"testing"

	"github.com/tinygo-org/tinygo/compileopts"
)

func TestRecordedPath(t *testing.T) {
	config := &compileopts.Config{
		Options: &compileopts.Options{TrimPath: true},
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
	dependency := &Package{
		program: program,
		PackageJSON: PackageJSON{
			Dir:        filepath.FromSlash("/tmp/module/dependency"),
			ImportPath: "example.com/dependency/subpackage",
		},
	}
	dependency.Module.Path = "example.com/dependency"
	dependency.Module.Version = "v1.2.3"
	program.Packages[pkg.ImportPath] = pkg
	program.Packages[dependency.ImportPath] = dependency

	if got, want := dependency.RecordedDir(), "example.com/dependency@v1.2.3/subpackage"; got != want {
		t.Fatalf("RecordedDir() = %q, want %q", got, want)
	}
	filename := filepath.FromSlash("/tmp/module/dependency/file.go")
	if got, want := pkg.RecordedPath(filename), "example.com/dependency@v1.2.3/subpackage/file.go"; got != want {
		t.Fatalf("RecordedPath() = %q, want %q", got, want)
	}
}
