package builder

import (
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
