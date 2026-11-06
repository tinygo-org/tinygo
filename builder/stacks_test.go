package builder

import (
	"testing"
	"time"

	"github.com/tinygo-org/tinygo/compileopts"
	"github.com/tinygo-org/tinygo/interp"
)

// -print-stacks reads the program module after its bitcode is written, so the
// module must outlive the output job on a target without automatic stack sizes.
func TestPrintStacksWithoutAutomaticStackSize(t *testing.T) {
	options := compileopts.Options{
		Target:                  "wasip1",
		Opt:                     "z",
		Semaphore:               sema,
		InterpTimeout:           60 * time.Second,
		InterpMaxLoopIterations: interp.DefaultMaxInterpBlockEntries,
		PrintStacks:             true,
	}
	target, err := compileopts.LoadTarget(&options)
	if err != nil {
		t.Fatal(err)
	}
	config := &compileopts.Config{Options: &options, Target: target}
	if config.AutomaticStackSize() {
		t.Fatal("wasip1 sizes stacks automatically, so this test no longer covers the case")
	}
	// Stack sizes cannot be computed for wasm, so an error is expected. A freed
	// module would crash here instead.
	if _, err := Build("examples/serial", "", t.TempDir(), config); err == nil {
		t.Log("built with -print-stacks")
	}
}
