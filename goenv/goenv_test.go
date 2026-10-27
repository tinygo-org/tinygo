package goenv

import (
	"slices"
	"testing"
)

func TestSplitFlags(t *testing.T) {
	for _, tc := range []struct {
		in   string
		want []string
		err  bool
	}{
		{in: "", want: nil},
		{in: "  \t\n ", want: nil},
		{in: "-a -b\t-c\n-d\r-e", want: []string{"-a", "-b", "-c", "-d", "-e"}},
		{in: `'a b' "c d"`, want: []string{"a b", "c d"}},
		{in: `"it's" 'say "hi"'`, want: []string{"it's", `say "hi"`}},
		{in: `a"b c"`, want: []string{`a"b`, `c"`}},
		{in: `"" ''`, want: []string{"", ""}},
		{in: `-IC:\Users\x\envinc -DC_ENV=1`, want: []string{`-IC:\Users\x\envinc`, "-DC_ENV=1"}},
		{in: `"-IC:\Program Files\x"`, want: []string{`-IC:\Program Files\x`}},
		{in: `a\ b`, want: []string{`a\`, "b"}},
		{in: `"abc`, err: true},
		{in: `x 'abc`, err: true},
	} {
		got, err := SplitFlags(tc.in)
		if tc.err {
			if err == nil {
				t.Errorf("SplitFlags(%q) = %q, want error", tc.in, got)
			}
			continue
		}
		if err != nil {
			t.Errorf("SplitFlags(%q) failed: %v", tc.in, err)
			continue
		}
		if !slices.Equal(got, tc.want) {
			t.Errorf("SplitFlags(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}
