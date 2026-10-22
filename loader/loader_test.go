package loader

import (
	"errors"
	"fmt"
	"go/token"
	"go/types"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"syscall"
	"testing"

	"github.com/tinygo-org/tinygo/compileopts"
	"github.com/tinygo-org/tinygo/goenv"
)

func TestCGoCompilerFlags(t *testing.T) {
	dir := t.TempDir()
	for name, content := range map[string]string{
		"pkginc/order.h": "#define ORDER 1\n",
		"envinc/order.h": "#define ORDER 2\n",
		"probe.go": `package probe
/*
#cgo CPPFLAGS: -Ipkginc -DSHARED_DIRECTIVE=1
#cgo CFLAGS: -DC_DIRECTIVE=1
#cgo CXXFLAGS: -std=c++17 -DCXX_DIRECTIVE=1
#include "order.h"
#if ORDER != 1
#error wrong include order
#endif
#if defined(CXX_ENV) || defined(CXX_DIRECTIVE)
#error C++ flags passed to C parser
#endif
*/
import "C"
`,
	} {
		path := filepath.Join(dir, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	envInc := "-I" + filepath.Join(dir, "envinc")
	t.Setenv("CGO_CPPFLAGS", "-DSHARED_ENV=1")
	t.Setenv("CGO_CFLAGS", envInc+" -DC_ENV=1")
	t.Setenv("CGO_CXXFLAGS", "-DCXX_ENV=1")
	p := &Package{
		PackageJSON: PackageJSON{Dir: dir, ImportPath: "probe", CgoFiles: []string{"probe.go"}},
		FileHashes:  make(map[string][]byte),
		program: &Program{
			workingDir: dir,
			fset:       token.NewFileSet(),
			config: &compileopts.Config{
				Options: &compileopts.Options{Opt: "0"},
				Target: &compileopts.TargetSpec{
					Triple: "armv6m-unknown-unknown-eabi", GOOS: "linux", GOARCH: "arm",
				},
			},
		},
	}
	if _, err := p.parseFiles(); err != nil {
		t.Fatal(err)
	}
	pkgInc := "-I" + filepath.Join(dir, "pkginc")
	for _, test := range []struct {
		flags []string
		want  []string
	}{
		{p.CFlags, []string{"-DSHARED_ENV=1", pkgInc, "-DSHARED_DIRECTIVE=1", envInc, "-DC_ENV=1", "-DC_DIRECTIVE=1"}},
		{p.CXXFlags, []string{"-DSHARED_ENV=1", pkgInc, "-DSHARED_DIRECTIVE=1", "-DCXX_ENV=1", "-std=c++17", "-DCXX_DIRECTIVE=1"}},
	} {
		if len(test.flags) < len(test.want) || !slices.Equal(test.flags[len(test.flags)-len(test.want):], test.want) {
			t.Errorf("flags %q do not end with %q", test.flags, test.want)
		}
	}
}

func TestPackageTestContext(t *testing.T) {
	t.Setenv("GOFLAGS", "")
	dir := t.TempDir()
	for name, content := range map[string]string{
		"go.mod":        "module example.com/testcontext\n\ngo 1.25.0\n",
		"go.work":       "go 1.25.0\n\nuse .\n",
		"probe.go":      "package probe\n",
		"probe_test.go": "package probe\nimport \"testing\"\nfunc TestProbe(t *testing.T) {}\n",
	} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	t.Setenv("GOWORK", filepath.Join(dir, "go.work"))
	_, minor, err := goenv.GetGorootVersion()
	if err != nil {
		t.Fatal(err)
	}
	config := &compileopts.Config{
		Options:        &compileopts.Options{Directory: dir},
		Target:         &compileopts.TargetSpec{GOOS: runtime.GOOS, GOARCH: runtime.GOARCH},
		GoMinorVersion: minor,
		TestConfig:     compileopts.TestConfig{CompileTestBinary: true},
	}
	goroot, err := GetCachedGoroot(config)
	if err != nil {
		t.Fatal(err)
	}
	stdlibDir := filepath.Join(goroot, "src")
	for _, test := range []struct {
		name     string
		pkg      string
		wantDir  string
		standard bool
	}{
		{"stdlib tests", "fmt", stdlibDir, true},
		{"module tests", ".", dir, false},
	} {
		t.Run(test.name, func(t *testing.T) {
			config.TestConfig.StandardPackage = test.standard
			program, err := Load(config, test.pkg, types.Config{})
			if err != nil {
				t.Fatal(err)
			}
			if program.workingDir != test.wantDir {
				t.Errorf("working directory = %q, want %q", program.workingDir, test.wantDir)
			}
			if config.Options.Directory != dir {
				t.Fatal("Load changed the caller's directory")
			}
		})
	}
}

func TestPackageWorkingDirectory(t *testing.T) {
	t.Setenv("GOFLAGS", "")
	_, minor, err := goenv.GetGorootVersion()
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"ordinary", "symlink"} {
		t.Run(name, func(t *testing.T) {
			dir, err := filepath.EvalSymlinks(t.TempDir())
			if err != nil {
				t.Fatal(err)
			}
			if name == "symlink" {
				link := filepath.Join(t.TempDir(), "workspace")
				if err := os.Symlink(dir, link); err != nil {
					if errors.Is(err, os.ErrPermission) || errors.Is(err, syscall.ENOSYS) || errors.Is(err, syscall.ENOTSUP) {
						t.Skipf("cannot create a symlink: %v", err)
					}
					t.Fatal(err)
				}
				dir = link
			}
			for name, content := range map[string]string{
				"go.mod":   "module example.com/workdir\n\ngo 1.25.0\n",
				"go.work":  "go 1.25.0\n\nuse .\n",
				"probe.go": "package probe\n",
			} {
				if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o600); err != nil {
					t.Fatal(err)
				}
			}
			t.Setenv("GOWORK", filepath.Join(dir, "go.work"))
			config := &compileopts.Config{
				Options:        &compileopts.Options{Directory: dir},
				Target:         &compileopts.TargetSpec{GOOS: runtime.GOOS, GOARCH: runtime.GOARCH},
				GoMinorVersion: minor,
			}
			cmd, err := List(config, []string{"-f", "{{.Dir}}"}, []string{"."})
			if err != nil {
				t.Fatal(err)
			}
			if config.Options.Directory != dir {
				t.Fatal("List changed the caller's directory")
			}
			output, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("go list failed: %v\n%s", err, output)
			}
			if got := strings.TrimSpace(string(output)); got != dir {
				t.Fatalf("package directory = %q, want %q", got, dir)
			}
			program, err := Load(config, ".", types.Config{})
			if err != nil {
				t.Fatal(err)
			}
			if program.workingDir != dir {
				t.Errorf("working directory = %q, want %q", program.workingDir, dir)
			}
			pkg := program.Packages["example.com/workdir"]
			if pkg == nil {
				t.Fatal("Load did not find the workspace package")
			}
			if pkg.Dir != dir {
				t.Errorf("loaded package directory = %q, want %q", pkg.Dir, dir)
			}
			if config.Options.Directory != dir {
				t.Fatal("Load changed the caller's directory")
			}
		})
	}
}

func BenchmarkRecordedPath(b *testing.B) {
	for _, count := range []int{10, 100, 1000} {
		b.Run(fmt.Sprintf("packages=%d", count), func(b *testing.B) {
			program := &Program{
				config: &compileopts.Config{
					Options: &compileopts.Options{TrimPath: true},
				},
				goroot:   filepath.FromSlash("/goroot"),
				Packages: make(map[string]*Package),
			}
			var pkg *Package
			for i := range count {
				pkg = &Package{
					program: program,
					PackageJSON: PackageJSON{
						Dir:        filepath.FromSlash(fmt.Sprintf("/tmp/module/pkg%d", i)),
						ImportPath: fmt.Sprintf("example.com/module/pkg%d", i),
					},
				}
				pkg.Module.Path = "example.com/module"
				pkg.Module.Dir = filepath.FromSlash("/tmp/module")
				program.Packages[pkg.ImportPath] = pkg
			}
			program.initRecordedPaths()
			for _, name := range []string{"package", "module"} {
				b.Run(name, func(b *testing.B) {
					filename := filepath.Join(pkg.Dir, "file.go")
					want := pkg.ImportPath + "/file.go"
					if name == "module" {
						filename = filepath.FromSlash("/tmp/module/include/shared.h")
						want = "example.com/module/include/shared.h"
					}
					b.ReportAllocs()
					for b.Loop() {
						if got := pkg.RecordedPath(filename); got != want {
							b.Fatalf("RecordedPath() = %q, want %q", got, want)
						}
					}
				})
			}
		})
	}
}

func TestRecordedPathPrecedence(t *testing.T) {
	root := t.TempDir()
	testPath := func(path string) string {
		return filepath.Join(root, filepath.FromSlash(path))
	}
	program := &Program{
		config: &compileopts.Config{
			Options: &compileopts.Options{TrimPath: true},
		},
		goroot:   testPath("/goroot"),
		Packages: make(map[string]*Package),
	}
	var pkg *Package
	for _, entry := range []struct{ dir, name, moduleDir, modulePath string }{
		{"/tmp/root/pkg", "example.com/z", "/tmp/root", "example.com/root"},
		{"/tmp/root/pkg", "example.com/a", "/tmp/root", "example.com/root"},
		{"/tmp/root/pkg/nested", "example.com/nested", "/tmp/root", "example.com/root"},
		{"/tmp/root/pkg/include/sub", "example.com/includes/sub", "/tmp/root/pkg/include", "example.com/includes"},
		{"/tmp/root/extra/sub", "example.com/extra/sub", "/tmp/root/extra", "example.com/extra"},
	} {
		pkg = &Package{
			program: program,
			PackageJSON: PackageJSON{
				Dir:        testPath(entry.dir),
				ImportPath: entry.name,
			},
		}
		pkg.Module.Dir = testPath(entry.moduleDir)
		pkg.Module.Path = entry.modulePath
		program.Packages[pkg.ImportPath] = pkg
	}
	program.initRecordedPaths()
	for _, tc := range []struct{ filename, want string }{
		{"/tmp/root/pkg/file.go", "example.com/a/file.go"},
		{"/tmp/root/pkg/nested/file.go", "example.com/nested/file.go"},
		{"/tmp/root/pkg/include/shared.h", "example.com/a/include/shared.h"},
		{"/tmp/root/extra/include/shared.h", "example.com/extra/include/shared.h"},
		{"/tmp/root/include/shared.h", "example.com/root/include/shared.h"},
		{"/tmp/root/pkg-other/file.go", "example.com/root/pkg-other/file.go"},
		{"/tmp/root-other/file.go", testPath("/tmp/root-other/file.go")},
	} {
		if got := pkg.RecordedPath(testPath(tc.filename)); got != tc.want {
			t.Errorf("RecordedPath(%q) = %q, want %q", tc.filename, got, tc.want)
		}
	}
}

func TestRecordedPath(t *testing.T) {
	root := t.TempDir()
	testPath := func(path string) string {
		return filepath.Join(root, filepath.FromSlash(path))
	}
	config := &compileopts.Config{
		Options: &compileopts.Options{TrimPath: true},
		Target:  &compileopts.TargetSpec{GOOS: "linux"},
	}
	program := &Program{
		config:   config,
		goroot:   testPath("/goroot"),
		Packages: make(map[string]*Package),
	}
	pkg := &Package{
		program: program,
		PackageJSON: PackageJSON{
			Dir:        testPath("/tmp/module"),
			ImportPath: "example.com/main",
		},
	}
	pkg.Module.Path = "example.com/main"
	pkg.Module.Dir = testPath("/tmp/module")
	dependency := &Package{
		program: program,
		PackageJSON: PackageJSON{
			Dir:        testPath("/tmp/dependency/subpackage"),
			ImportPath: "example.com/dependency/subpackage",
		},
	}
	dependency.Module.Path = "example.com/dependency"
	dependency.Module.Version = "v1.2.3"
	dependency.Module.Dir = testPath("/tmp/dependency")
	dependency.CFlags = []string{
		"-I" + testPath("/tmp/dependency/include"),
		"-isystem",
		testPath("/tmp/dependency/system"),
		"-iquote" + testPath("/tmp/dependency/include"),
		"-ffile-prefix-map=" + testPath("/tmp/dependency") + "=/mapped",
		`-DCONFIG_PATH="/tmp/dependency"`,
	}
	program.Packages[pkg.ImportPath] = pkg
	program.Packages[dependency.ImportPath] = dependency
	program.initRecordedPaths()

	if got, want := dependency.RecordedDir(), "example.com/dependency@v1.2.3/subpackage"; got != want {
		t.Fatalf("RecordedDir() = %q, want %q", got, want)
	}
	filename := testPath("/tmp/dependency/subpackage/file.go")
	if got, want := pkg.RecordedPath(filename), "example.com/dependency@v1.2.3/subpackage/file.go"; got != want {
		t.Fatalf("RecordedPath() = %q, want %q", got, want)
	}
	header := testPath("/tmp/dependency/include/shared.h")
	if got, want := pkg.RecordedPath(header), "example.com/dependency@v1.2.3/include/shared.h"; got != want {
		t.Fatalf("RecordedPath() = %q, want %q", got, want)
	}
	if got, want := dependency.DebugPrefixMap(), "-ffile-prefix-map="+testPath("/tmp/dependency")+"=/_/example.com/dependency@v1.2.3"; got != want {
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
	outside := testPath("/tmp/module-other/include")
	if got := pkg.recordedCFlagPath(outside); got != outside {
		t.Errorf("rewrote a path outside the module: %q", got)
	}

	vendored := &Package{
		program: program,
		PackageJSON: PackageJSON{
			Dir:        testPath("/tmp/main/vendor/example.com/dependency/subpackage"),
			ImportPath: "example.com/dependency/subpackage",
		},
	}
	vendored.Module.Path = "example.com/dependency"
	vendored.Module.Version = "v1.2.3"
	program.Packages[vendored.ImportPath] = vendored
	program.initRecordedPaths()
	if got, want := vendored.OriginalModuleDir(), testPath("/tmp/main/vendor/example.com/dependency"); got != want {
		t.Fatalf("vendored OriginalModuleDir() = %q, want %q", got, want)
	}
	vendoredHeader := testPath("/tmp/main/vendor/example.com/dependency/include/shared.h")
	if got, want := pkg.RecordedPath(vendoredHeader), "example.com/dependency@v1.2.3/include/shared.h"; got != want {
		t.Fatalf("vendored RecordedPath() = %q, want %q", got, want)
	}
	if got, want := vendored.DebugPrefixMap(), "-ffile-prefix-map="+testPath("/tmp/main/vendor/example.com/dependency")+"=/_/example.com/dependency@v1.2.3"; got != want {
		t.Fatalf("vendored DebugPrefixMap() = %q, want %q", got, want)
	}

	gopath := &Package{
		program: program,
		PackageJSON: PackageJSON{
			Dir:        testPath("/gopath/src/example.com/dependency/subpackage"),
			ImportPath: "example.com/dependency/subpackage",
			Root:       testPath("/gopath"),
		},
	}
	program.Packages[gopath.ImportPath] = gopath
	program.initRecordedPaths()
	if got, want := gopath.OriginalModuleDir(), testPath("/gopath/src"); got != want {
		t.Fatalf("GOPATH OriginalModuleDir() = %q, want %q", got, want)
	}
	gopathHeader := testPath("/gopath/src/example.com/dependency/include/shared.h")
	if got, want := pkg.RecordedPath(gopathHeader), "example.com/dependency/include/shared.h"; got != want {
		t.Fatalf("GOPATH RecordedPath() = %q, want %q", got, want)
	}
	if got, want := gopath.DebugPrefixMap(), "-ffile-prefix-map="+testPath("/gopath/src")+"=/_"; got != want {
		t.Fatalf("GOPATH DebugPrefixMap() = %q, want %q", got, want)
	}
}
