package main

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/tinygo-org/tinygo/builder"
)

// Unproven //go:nopanic checks fail the build at every optimizing level. At
// -opt=0 nothing is proven, so the checks stay and panic as usual.
func TestNoPanicOptLevels(t *testing.T) {
	t.Parallel()
	for _, opt := range []string{"0", "1", "2", "s", "z"} {
		t.Run(opt, func(t *testing.T) {
			t.Parallel()
			options := optionsFromTarget("wasip1", sema)
			options.Opt = opt
			config, err := builder.NewConfig(&options)
			if err != nil {
				t.Fatal(err)
			}
			tmpdir := t.TempDir()
			_, err = builder.Build("testdata/errors/nopanic-check.go", filepath.Join(tmpdir, "main.wasm"), tmpdir, config)
			if opt == "0" {
				if err != nil {
					t.Fatal(err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), "//go:nopanic check is not proven") {
				t.Fatalf("expected nopanic error, got %v", err)
			}
		})
	}
}
