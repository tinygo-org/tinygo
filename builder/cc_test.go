package builder

import (
	"bytes"
	"debug/macho"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestAssemblyObjectsSurviveCleanup(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_CACHE_HOME", home)
	t.Setenv("LocalAppData", home)
	for _, extension := range []string{".s", ".S"} {
		t.Run(extension, func(t *testing.T) {
			dir := t.TempDir()
			buildDir := filepath.Join(dir, "build")
			if err := os.Mkdir(buildDir, 0o755); err != nil {
				t.Fatal(err)
			}
			source := filepath.Join(dir, "main"+extension)
			if err := os.WriteFile(source, []byte(".text\n.globl _main\n_main:\nxorl %eax, %eax\nret\n"), 0o644); err != nil {
				t.Fatal(err)
			}
			flags := []string{"--target=x86_64-apple-darwin", "-gdwarf-4"}
			object, err := compileAndCacheCFile(source, buildDir, flags, nil, nil)
			if err != nil {
				t.Fatal(err)
			}
			before, err := os.Stat(object)
			if err != nil {
				t.Fatal(err)
			}
			rebuilt, err := compileAndCacheCFile(source, buildDir, flags, nil, nil)
			if err != nil {
				t.Fatal(err)
			}
			after, err := os.Stat(rebuilt)
			if err != nil {
				t.Fatal(err)
			}
			if rebuilt != object || !after.ModTime().Equal(before.ModTime()) {
				t.Fatal("unchanged assembly replaced an existing Darwin debug object")
			}
			binary := filepath.Join(dir, "main")
			if err := link("ld.lld", []string{
				"-flavor", "darwin", "-arch", "x86_64",
				"-platform_version", "macos", "11.0", "11.0",
				"-o", binary, object,
			}); err != nil {
				t.Fatal(err)
			}
			if err := os.RemoveAll(buildDir); err != nil {
				t.Fatal(err)
			}
			executable, err := macho.Open(binary)
			if err != nil {
				t.Fatal(err)
			}
			defer executable.Close()
			if executable.Symtab == nil {
				t.Fatal("Darwin executable has no debug map")
			}
			found := false
			for _, symbol := range executable.Symtab.Syms {
				const nOSO = 0x66
				if symbol.Type != nOSO {
					continue
				}
				if symbol.Name != object {
					t.Errorf("debug object = %q, want %q", symbol.Name, object)
				}
				if _, err := os.Stat(symbol.Name); err != nil {
					t.Errorf("Darwin debug object did not survive cleanup: %v", err)
				}
				found = true
			}
			if !found {
				t.Fatal("Darwin executable has no object debug reference")
			}
		})
	}
}

func TestAssemblyIncludes(t *testing.T) {
	for _, extension := range []string{".s", ".S"} {
		t.Run(extension, func(t *testing.T) {
			dir := t.TempDir()
			include1 := t.TempDir()
			include2 := t.TempDir()
			source := filepath.Join(dir, "test"+extension)
			if err := os.WriteFile(source, []byte(".text\n.globl value\nvalue:\n.include \"defs.inc\"\nret\n"), 0o666); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(include2, "defs.inc"), []byte(".include \"nested.inc\"\n"), 0o666); err != nil {
				t.Fatal(err)
			}
			nested := filepath.Join(include2, "nested.inc")
			if err := os.WriteFile(nested, []byte("movl $1, %eax\n"), 0o666); err != nil {
				t.Fatal(err)
			}
			flags := []string{"--target=x86_64-unknown-linux-gnu", "-I", include1, "-I", include2}
			compile := func() []byte {
				t.Helper()
				object, err := compileAndCacheCFile(source, dir, flags, nil, nil)
				if err != nil {
					t.Fatal(err)
				}
				data, err := os.ReadFile(object)
				if err != nil {
					t.Fatal(err)
				}
				return data
			}
			before := compile()
			if !bytes.Equal(before, compile()) {
				t.Fatal("unchanged assembly produced different machine code")
			}
			if err := os.WriteFile(nested, []byte("movl $2, %eax\n"), 0o666); err != nil {
				t.Fatal(err)
			}
			changed := compile()
			if bytes.Equal(before, changed) {
				t.Fatal("editing a nested assembler include reused stale machine code")
			}
			if err := os.WriteFile(filepath.Join(include1, "defs.inc"), []byte("movl $3, %eax\n"), 0o666); err != nil {
				t.Fatal(err)
			}
			if bytes.Equal(changed, compile()) {
				t.Fatal("shadowing an assembler include reused stale machine code")
			}
		})
	}
}

func TestCompileCXXFile(t *testing.T) {
	dir := t.TempDir()
	source := filepath.Join(dir, "flags.cpp")
	code := `#if defined(C_ONLY) || !defined(SHARED)
#error incorrect C++ flags
#endif
template <bool enabled> int constant() {
    if constexpr (enabled) {
        return SHARED + CXX_ONLY;
    } else {
        return 0;
    }
}
extern "C" int value(void) { return constant<true>(); }
`
	if err := os.WriteFile(source, []byte(code), 0o600); err != nil {
		t.Fatal(err)
	}
	flags := []string{"-std=c++17", "-DSHARED=23", "-DCXX_ONLY=19"}
	first, err := compileAndCacheCFile(source, dir, flags, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	cached, err := compileAndCacheCFile(source, dir, flags, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if cached != first {
		t.Errorf("cached output = %q, want %q", cached, first)
	}
	flags[2] = "-DCXX_ONLY=20"
	changed, err := compileAndCacheCFile(source, dir, flags, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if changed == first {
		t.Error("changed C++ flags reused the old cached output")
	}
}

func TestSplitDepFile(t *testing.T) {
	for i, tc := range []struct {
		in  string
		out []string
	}{
		{`deps: foo bar`, []string{"foo", "bar"}},
		{`deps: foo "bar"`, []string{"foo", "bar"}},
		{`deps: "foo" bar`, []string{"foo", "bar"}},
		{`deps: "foo bar"`, []string{"foo bar"}},
		{`deps: "foo bar" `, []string{"foo bar"}},
		{"deps: foo\nbar", []string{"foo"}},
		{"deps: foo \\\nbar", []string{"foo", "bar"}},
		{"deps: foo\\bar \\\nbaz", []string{"foo\\bar", "baz"}},
		{"deps: foo\\bar \\\r\n baz", []string{"foo\\bar", "baz"}}, // Windows uses CRLF line endings
	} {
		out, err := parseDepFile(tc.in)
		if err != nil {
			t.Errorf("test #%d failed: %v", i, err)
			continue
		}
		if !reflect.DeepEqual(out, tc.out) {
			t.Errorf("test #%d failed: expected %#v but got %#v", i, tc.out, out)
			continue
		}
	}
}

func TestCFileCacheIncludePathShadowing(t *testing.T) {
	t.Setenv("GOCACHEPROG", "")

	dir := t.TempDir()

	include1 := filepath.Join(dir, "include1")
	include2 := filepath.Join(dir, "include2")
	if err := os.Mkdir(include1, 0o777); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(include2, 0o777); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(include2, "value.h"), []byte("#define VALUE 1\n"), 0o666); err != nil {
		t.Fatal(err)
	}
	source := filepath.Join(dir, "test.c")
	if err := os.WriteFile(source, []byte("#include \"value.h\"\nint value(void) { return VALUE; }\n"), 0o666); err != nil {
		t.Fatal(err)
	}

	flags := []string{
		"-I", include1,
		"-I", include2,
		"--target=x86_64-unknown-linux-gnu",
	}
	first, err := compileAndCacheCFile(source, dir, flags, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	second, err := compileAndCacheCFile(source, dir, flags, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if first != second {
		t.Fatalf("unchanged compile did not hit cache: %s != %s", first, second)
	}

	if err := os.WriteFile(filepath.Join(include1, "value.h"), []byte("#define VALUE 2\n"), 0o666); err != nil {
		t.Fatal(err)
	}
	shadowed, err := compileAndCacheCFile(source, dir, flags, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if shadowed == first {
		t.Fatal("include path shadowing reused stale cached object")
	}
}
