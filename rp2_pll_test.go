package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tinygo-org/tinygo/builder"
	"github.com/tinygo-org/tinygo/interp"
)

func TestRP2PLLPreEvaluation(t *testing.T) {
	for _, target := range []string{"pico", "pico2"} {
		t.Run(target, func(t *testing.T) {
			tmpdir := t.TempDir()
			options := optionsFromTarget(target, sema)
			options.Opt = "1"
			options.InterpMaxLoopIterations = interp.DefaultMaxInterpBlockEntries
			config, err := builder.NewConfig(&options)
			if err != nil {
				t.Fatal(err)
			}
			path := filepath.Join(tmpdir, "blinky.ll")
			if _, err := builder.Build("examples/blinky1", path, tmpdir, config); err != nil {
				t.Fatal(err)
			}
			ir, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			for _, name := range []string{
				"machine.pdTable", "machine.pllsysFB", "machine.pllsysPD1", "machine.pllsysPD2",
				"machine.genTable", "machine.pllFreqOutPostdiv",
			} {
				if strings.Contains(string(ir), "@"+name) {
					t.Errorf("PLL calculation was not pre-evaluated: %s remains", name)
				}
			}
		})
	}
}
