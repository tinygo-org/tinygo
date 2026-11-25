package builder

import (
	"debug/dwarf"
	"debug/elf"
	"os"
	"path/filepath"
	"testing"

	"github.com/tinygo-org/tinygo/compileopts"
	"github.com/tinygo-org/tinygo/goenv"
)

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
