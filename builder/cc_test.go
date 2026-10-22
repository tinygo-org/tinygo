package builder

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

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
