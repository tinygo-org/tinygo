package builder

import (
	"bytes"
	"debug/dwarf"
	"debug/elf"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/tinygo-org/tinygo/compileopts"
	"github.com/tinygo-org/tinygo/goenv"
)

func TestTrimPathGeneratedHeaders(t *testing.T) {
	root := goenv.Get("TINYGOROOT")
	work := t.TempDir()
	t.Setenv("HOME", work)
	t.Setenv("XDG_CACHE_HOME", work)
	t.Setenv("LocalAppData", work)
	roots := []string{filepath.Join(work, "root-a-$HEADER"), filepath.Join(work, "root-b-$BUILDDIR-$GOCACHE")}
	for _, alias := range roots {
		if err := os.Symlink(root, alias); err != nil {
			t.Skipf("cannot create symlink: %v", err)
		}
	}
	source := filepath.Join(work, "probe.c")
	if err := os.WriteFile(source, []byte("#include <time.h>\nstruct timespec probe;\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	dependent := Library{
		name: "dependent", needsLibc: true,
		cflags: func(string, string) []string { return nil },
	}
	var before *libraryCacheInput
	objects := make(map[string][]byte)
	for i, alias := range roots {
		t.Setenv("TINYGOROOT", alias)
		config := &compileopts.Config{
			Options: &compileopts.Options{TrimPath: true, Opt: "0"},
			Target: &compileopts.TargetSpec{
				Triple: "x86_64-unknown-linux-musl",
				GOOS:   "linux", GOARCH: "amd64", CPU: "x86-64", Libc: "musl",
			},
		}
		input, err := libMusl.cacheInput(config)
		if err != nil {
			t.Fatal(err)
		}
		if before != nil {
			if before.key() == input.key() {
				t.Fatal("installation paths did not change the cache key")
			}
			if !reflect.DeepEqual(before.Sources, input.Sources) || !reflect.DeepEqual(before.GeneratedHeaders, input.GeneratedHeaders) {
				t.Fatal("installation paths changed source or generated-header contents")
			}
		}
		before = input
		config.LibraryKeys = map[string]string{"musl": input.key()}
		dir := config.LibraryPath("musl")
		if err := libMusl.prepareHeaders(config.Triple(), dir); err != nil {
			t.Fatal(err)
		}
		headerPath := filepath.Join(work, "include")
		buildDir := filepath.Join(work, "build")
		flags := map[string][]string{
			"cgo": config.CFlags(false),
			"library": expandLibraryCompileArgs(input.CompileArgs,
				filepath.Join(dir, "include"), buildDir),
			"dependent": expandLibraryCompileArgs(dependent.compileArgs(config, config.Triple(),
				headerPath, buildDir, input.ResourceDir), headerPath, buildDir),
		}
		for name, flags := range flags {
			object := filepath.Join(work, name+".o")
			if err := runCCompiler(append(flags, "-c", source, "-o", object)...); err != nil {
				t.Fatal(err)
			}
			data, err := os.ReadFile(object)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Contains(data, []byte("/_/tinygo-cache/musl/include/bits")) {
				t.Errorf("%s: missing stable generated-header path", name)
			}
			if bytes.Contains(data, []byte(input.key())) {
				t.Errorf("%s: cache key leaked into the object", name)
			}
			if i != 0 && !bytes.Equal(objects[name], data) {
				t.Errorf("%s: relocating TinyGo changed the trimmed object", name)
			}
			objects[name] = data
		}
	}
}

func TestTrimPathCFlagsMapTinyGoSource(t *testing.T) {
	for _, tc := range []struct {
		name      string
		overlap   bool
		untrimmed bool
	}{
		{name: "separate"},
		{name: "cache-prefix", overlap: true},
		{name: "untrimmed-separate", untrimmed: true},
		{name: "untrimmed-cache-prefix", overlap: true, untrimmed: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			home := t.TempDir()
			t.Setenv("HOME", home)
			t.Setenv("XDG_CACHE_HOME", home)
			t.Setenv("LocalAppData", home)
			cache := goenv.Get("GOCACHE")
			tinygoRoot := filepath.Join(filepath.Dir(cache), "source")
			if tc.overlap {
				tinygoRoot = cache + "-source"
			}
			for _, name := range []string{
				"src/runtime/internal/sys/zversion.go",
				"src/device/arm/arm.go",
			} {
				filename := filepath.Join(tinygoRoot, filepath.FromSlash(name))
				if err := os.MkdirAll(filepath.Dir(filename), 0o755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filename, nil, 0o644); err != nil {
					t.Fatal(err)
				}
			}
			t.Setenv("TINYGOROOT", tinygoRoot)

			source := filepath.Join(tinygoRoot, "probe.c")
			if err := os.WriteFile(source, []byte("int probe(void) { return 1; }\n"), 0o644); err != nil {
				t.Fatal(err)
			}
			config := &compileopts.Config{
				Options: &compileopts.Options{TrimPath: !tc.untrimmed, Opt: "0"},
				Target: &compileopts.TargetSpec{
					Triple: "x86_64-unknown-linux-gnu", GOOS: "linux", GOARCH: "amd64",
				},
			}
			objectPath := filepath.Join(home, "probe.o")
			flags := append([]string{"-c", "-o", objectPath, source}, config.CFlags(false)...)
			if err := runCCompiler(flags...); err != nil {
				t.Fatal(err)
			}
			unit := readELFCompileUnit(t, objectPath)
			got, ok := unit.Val(dwarf.AttrName).(string)
			if !ok {
				t.Fatalf("DWARF source path is not a string: %v", unit.Val(dwarf.AttrName))
			}
			want := config.CSourcePath("github.com/tinygo-org/tinygo/probe.c")
			if tc.untrimmed {
				got = filepath.ToSlash(got)
				want = filepath.ToSlash(source)
			}
			if got != want {
				t.Errorf("DWARF source path = %q, want %q", got, want)
			}
		})
	}
}

func readELFCompileUnit(t *testing.T, path string) *dwarf.Entry {
	t.Helper()
	object, err := elf.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer object.Close()
	debugInfo, err := object.DWARF()
	if err != nil {
		t.Fatal(err)
	}
	unit, err := debugInfo.Reader().Next()
	if err != nil {
		t.Fatal(err)
	}
	if unit == nil || unit.Tag != dwarf.TagCompileUnit {
		t.Fatalf("expected a DWARF compile unit, got %v", unit)
	}
	return unit
}
