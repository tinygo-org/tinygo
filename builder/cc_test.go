package builder

import (
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

func TestMakeCCompilerPathsAbsolute(t *testing.T) {
	workingDir := t.TempDir()
	flags := []string{
		"-include", "config.h",
		"-Iinclude",
		"-isystem", "system",
		"--sysroot=sdk",
		`-DCONFIG_PATH="/work"`,
	}
	want := []string{
		"-include", filepath.Join(workingDir, "config.h"),
		"-I" + filepath.Join(workingDir, "include"),
		"-isystem", filepath.Join(workingDir, "system"),
		"--sysroot=" + filepath.Join(workingDir, "sdk"),
		`-DCONFIG_PATH="/work"`,
	}
	if got := makeCCompilerPathsAbsolute(flags, workingDir); !reflect.DeepEqual(got, want) {
		t.Fatalf("makeCCompilerPathsAbsolute() = %q, want %q", got, want)
	}
}
