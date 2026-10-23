package compiler

import (
	"testing"

	"github.com/tinygo-org/tinygo/compileopts"
)

func TestFramePointers(t *testing.T) {
	for _, tc := range []struct {
		options *compileopts.Options
		want    string
	}{
		{&compileopts.Options{GOOS: "linux", GOARCH: "amd64"}, "all"},
		{&compileopts.Options{GOOS: "linux", GOARCH: "arm64"}, "all"},
		{&compileopts.Options{Target: "wasm"}, ""},
	} {
		name := tc.options.GOOS + "/" + tc.options.GOARCH + tc.options.Target
		t.Run(name, func(t *testing.T) {
			mod, errs := testCompilePackage(t, tc.options, "basic.go")
			for _, err := range errs {
				t.Fatal(err)
			}
			defer mod.Dispose()
			fn := mod.NamedFunction("main.addInt")
			if fn.IsNil() {
				t.Fatal("missing function main.addInt")
			}
			got := ""
			if attr := fn.GetStringAttributeAtIndex(-1, "frame-pointer"); !attr.IsNil() {
				got = attr.GetStringValue()
			}
			if got != tc.want {
				t.Errorf("frame-pointer = %q, want %q", got, tc.want)
			}
		})
	}
}
