package builder

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

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
	first, err := compileAndCacheCFile(source, dir, flags, nil)
	if err != nil {
		t.Fatal(err)
	}
	second, err := compileAndCacheCFile(source, dir, flags, nil)
	if err != nil {
		t.Fatal(err)
	}
	if first != second {
		t.Fatalf("unchanged compile did not hit cache: %s != %s", first, second)
	}

	if err := os.WriteFile(filepath.Join(include1, "value.h"), []byte("#define VALUE 2\n"), 0o666); err != nil {
		t.Fatal(err)
	}
	shadowed, err := compileAndCacheCFile(source, dir, flags, nil)
	if err != nil {
		t.Fatal(err)
	}
	if shadowed == first {
		t.Fatal("include path shadowing reused stale cached object")
	}
}
