package builder

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

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
