package builder

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/tinygo-org/tinygo/compileopts"
	"github.com/tinygo-org/tinygo/goenv"
)

func TestLibraryCacheImplicitIncludePaths(t *testing.T) {
	dir := t.TempDir()
	include1 := t.TempDir()
	include2 := filepath.Join(t.TempDir(), "implicit headers")
	if err := os.Mkdir(include2, 0o777); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "test.c"), []byte("#include <value.h>\nint value(void) { return VALUE; }\n"), 0o666); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(include2, "value.h"), []byte("#define VALUE 1\n"), 0o666); err != nil {
		t.Fatal(err)
	}
	config := &compileopts.Config{
		Options: &compileopts.Options{Opt: "z", Semaphore: sema},
		Target: &compileopts.TargetSpec{
			Triple: "x86_64-unknown-linux-gnu",
			GOOS:   "linux",
			GOARCH: "amd64",
			Libc:   "glibc",
		},
	}
	library := Library{
		name:      "implicit-headers",
		needsLibc: true,
		sourceDir: func() string { return dir },
		librarySources: func(string, bool) ([]string, error) {
			return []string{"test.c"}, nil
		},
		cflags: func(string, string) []string {
			return []string{
				"-nostdinc", "-Werror",
				"-Xclang", "-internal-isystem", "-Xclang", include1,
				"-Xclang", "-internal-isystem", "-Xclang", include2,
			}
		},
	}
	before, err := library.cacheInput(config)
	if err != nil {
		t.Fatal(err)
	}
	unchanged, err := library.cacheInput(config)
	if err != nil {
		t.Fatal(err)
	}
	if before.key() != unchanged.key() {
		t.Fatal("unchanged implicit headers changed the library cache key")
	}

	if err := os.WriteFile(filepath.Join(include2, "value.h"), []byte("#define VALUE 2\n"), 0o666); err != nil {
		t.Fatal(err)
	}
	changed, err := library.cacheInput(config)
	if err != nil {
		t.Fatal(err)
	}
	if before.key() == changed.key() {
		t.Fatal("editing an implicit header reused the library cache key")
	}

	if err := os.WriteFile(filepath.Join(include1, "value.h"), []byte("#define VALUE 3\n"), 0o666); err != nil {
		t.Fatal(err)
	}
	shadowed, err := library.cacheInput(config)
	if err != nil {
		t.Fatal(err)
	}
	if changed.key() == shadowed.key() {
		t.Fatal("shadowing an implicit header reused the library cache key")
	}
}

func TestHashLibraryInputFilesIncludesNonHeaders(t *testing.T) {
	dir := t.TempDir()
	for name, contents := range map[string]string{
		"header.h": "header",
		"source.c": "source",
		"data":     "data",
	} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(contents), 0o666); err != nil {
			t.Fatal(err)
		}
	}

	hashes, err := hashLibraryInputFiles(dir)
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"header.h", "source.c", "data"} {
		if _, ok := hashes[name]; !ok {
			t.Errorf("input file %q was not hashed", name)
		}
	}
}

func TestCompilerInputPaths(t *testing.T) {
	args := []string{
		"-I", "include",
		"-Iinclude2",
		"-isystem", "system",
		"-iquotequoted",
		"-idirafter", "after",
		"-include", "config.h",
		"-imacrosmacros.h",
		"-resource-dir=resource",
		"--sysroot", "sysroot",
		"-isysrootsdk",
		"-DVALUE=1",
	}

	want := []string{
		"include",
		"include2",
		"system",
		"quoted",
		"after",
		"config.h",
		"macros.h",
		"resource",
		"sysroot",
		"sdk",
	}
	if got := compilerInputPaths(args); !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected compiler input paths:\nwant: %#v\ngot:  %#v", want, got)
	}
}

func TestExpandLibraryCompileArgs(t *testing.T) {
	cacheDir := t.TempDir()
	t.Setenv("GOCACHE", cacheDir)
	args := []string{
		"-I" + libraryHeaderPathPlaceholder,
		"-o=" + libraryBuildDirPlaceholder,
		"-ffile-prefix-map=" + libraryCacheDirPlaceholder + "=/cache",
		"-I/literal/$HEADER/$BUILDDIR/$GOCACHE",
	}
	want := []string{
		"-I/header/$BUILDDIR",
		"-o=/build/$GOCACHE",
		"-ffile-prefix-map=" + goenv.Get("GOCACHE") + "=/cache",
		"-I/literal/$HEADER/$BUILDDIR/$GOCACHE",
	}
	if got := expandLibraryCompileArgs(args, "/header/$BUILDDIR", "/build/$GOCACHE"); !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected expanded compiler arguments:\nwant: %#v\ngot:  %#v", want, got)
	}
}

func TestHashLibraryInputFilesAncestorSymlinks(t *testing.T) {
	dir := t.TempDir()
	sub := filepath.Join(dir, "sub")
	if err := os.Mkdir(sub, 0o777); err != nil {
		t.Fatal(err)
	}
	for path, contents := range map[string]string{
		filepath.Join(dir, "value.h"): "root",
		filepath.Join(sub, "value.h"): "sub",
	} {
		if err := os.WriteFile(path, []byte(contents), 0o666); err != nil {
			t.Fatal(err)
		}
	}
	link := filepath.Join(sub, "up")
	if err := os.Symlink("..", link); err != nil {
		t.Skipf("cannot create symlink: %v", err)
	}
	before, err := hashLibraryInputFiles(dir)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(link); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(".", link); err != nil {
		t.Fatal(err)
	}
	after, err := hashLibraryInputFiles(dir)
	if err != nil {
		t.Fatal(err)
	}
	if reflect.DeepEqual(before, after) {
		t.Fatal("changing an ancestor symlink target did not change the fingerprint")
	}
}

func TestGeneratedHeaderSymlinkFingerprints(t *testing.T) {
	for _, mode := range []string{"direct", "aliased-temp"} {
		t.Run(mode, func(t *testing.T) {
			if mode == "aliased-temp" {
				dir := t.TempDir()
				alias := filepath.Join(t.TempDir(), "alias")
				if err := os.Symlink(dir, alias); err != nil {
					t.Skipf("cannot create symlink: %v", err)
				}
				for _, name := range []string{"TMPDIR", "TMP", "TEMP"} {
					t.Setenv(name, alias)
				}
			}
			library := Library{
				makeHeaders: func(_ string, dir string) error {
					header := filepath.Join(dir, "value.h")
					if err := os.WriteFile(header, []byte("header"), 0o666); err != nil {
						return err
					}
					for name, target := range map[string]string{
						"alias.h":    header,
						"dangling.h": filepath.Join(dir, "missing", "value.h"),
						"self":       dir,
					} {
						if err := os.Symlink(target, filepath.Join(dir, name)); err != nil {
							t.Skipf("cannot create symlink: %v", err)
						}
					}
					return nil
				},
			}
			before, err := library.hashGeneratedHeaders("")
			if err != nil {
				t.Fatal(err)
			}
			after, err := library.hashGeneratedHeaders("")
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(before, after) {
				t.Fatalf("temporary generated-header paths changed the fingerprint:\n%v\n%v", before, after)
			}
		})
	}
}

func TestHashLibraryCompileInputsTracksIncludeDirectories(t *testing.T) {
	includeDir := t.TempDir()
	args := []string{"-I", includeDir}

	before, err := hashLibraryCompileInputs(args)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(includeDir, "transitive.c"), []byte("input"), 0o666); err != nil {
		t.Fatal(err)
	}
	after, err := hashLibraryCompileInputs(args)
	if err != nil {
		t.Fatal(err)
	}
	if reflect.DeepEqual(before, after) {
		t.Fatal("adding a file to an include directory did not change its cache input")
	}
}

func TestLibraryCacheInputsUnderCache(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_CACHE_HOME", home)
	t.Setenv("LocalAppData", home)
	config := &compileopts.Config{
		Options:     &compileopts.Options{Opt: "z"},
		Target:      &compileopts.TargetSpec{Triple: "x86_64-unknown-linux-gnu", GOOS: "linux", GOARCH: "amd64"},
		LibraryKeys: map[string]string{"dependency": "content-key", "unkeyed": ""},
	}
	dependencyDir := config.LibraryPath("dependency")
	dependencyIncludes := filepath.Join(dependencyDir, "include")
	includeDirs := []struct {
		name string
		path string
	}{
		{"resource-headers", filepath.Join(goenv.Get("GOCACHE"), "installation", "lib", "clang", "include")},
		{"dependency-sibling", filepath.Join(dependencyDir+"-sdk", "include")},
		{"unkeyed-library", filepath.Join(config.LibraryPath("unkeyed"), "include")},
	}
	writeHeader := func(t *testing.T, dir, value string) {
		t.Helper()
		if err := os.MkdirAll(dir, 0o777); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "probe.h"), []byte("#define VALUE "+value+"\n"), 0o666); err != nil {
			t.Fatal(err)
		}
	}
	for _, dir := range includeDirs {
		writeHeader(t, dir.path, "1")
	}
	sourceDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(sourceDir, "probe.c"), []byte("int probe(void) { return 1; }\n"), 0o666); err != nil {
		t.Fatal(err)
	}
	library := Library{
		name:           "probe",
		sourceDir:      func() string { return sourceDir },
		librarySources: func(string, bool) ([]string, error) { return []string{"probe.c"}, nil },
		cflags: func(string, string) []string {
			args := []string{"-isystem", dependencyIncludes}
			for _, dir := range includeDirs {
				args = append(args, "-I"+dir.path)
			}
			return args
		},
	}
	key := func(t *testing.T) string {
		t.Helper()
		input, err := library.cacheInput(config)
		if err != nil {
			t.Fatal(err)
		}
		return input.key()
	}
	before := key(t)
	writeHeader(t, dependencyIncludes, "1")
	if before != key(t) {
		t.Fatal("materializing a content-keyed dependency changed the library cache key")
	}
	for _, dir := range includeDirs {
		t.Run(dir.name, func(t *testing.T) {
			before := key(t)
			writeHeader(t, dir.path, "2")
			if before == key(t) {
				t.Fatal("editing an unrelated cached header reused the library cache key")
			}
		})
	}
}

func TestHashLibraryCompileInputsDanglingSymlinks(t *testing.T) {
	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, "nested"), 0o777); err != nil {
		t.Fatal(err)
	}
	header := filepath.Join(dir, "value.h")
	if err := os.WriteFile(header, []byte("#define VALUE 1\n"), 0o666); err != nil {
		t.Fatal(err)
	}
	target := filepath.Join(t.TempDir(), "target.h")
	for _, name := range []string{"dangling.h", filepath.Join("nested", "dangling.h")} {
		if err := os.Symlink(target, filepath.Join(dir, name)); err != nil {
			t.Skipf("cannot create symlink: %v", err)
		}
	}
	args := []string{"-I", dir}
	before, err := hashLibraryCompileInputs(args)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := before[dir]["value.h"]; !ok {
		t.Fatal("dangling symlink discarded an existing header's hash")
	}
	if err := os.WriteFile(header, []byte("#define VALUE 2\n"), 0o666); err != nil {
		t.Fatal(err)
	}
	changed, err := hashLibraryCompileInputs(args)
	if err != nil {
		t.Fatal(err)
	}
	if reflect.DeepEqual(before, changed) {
		t.Fatal("editing a header did not change inputs with dangling symlinks present")
	}
	if err := os.WriteFile(target, []byte("#define TARGET 1\n"), 0o666); err != nil {
		t.Fatal(err)
	}
	resolved, err := hashLibraryCompileInputs(args)
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"dangling.h", "nested/dangling.h"} {
		if _, ok := resolved[dir][name]; !ok {
			t.Errorf("resolved symlink %q was not hashed", name)
		}
	}
}

func TestHashLibraryInputFilesFollowsSymlinkDirectories(t *testing.T) {
	dir := t.TempDir()
	target := t.TempDir()
	if err := os.WriteFile(filepath.Join(target, "input.c"), []byte("input"), 0o666); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(target, filepath.Join(dir, "linked")); err != nil {
		t.Skipf("cannot create symlink: %v", err)
	}

	hashes, err := hashLibraryInputFiles(dir)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := hashes["linked/input.c"]; !ok {
		t.Fatal("file in symlinked input directory was not hashed")
	}
}
