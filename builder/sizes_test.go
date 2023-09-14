package builder

import (
	"runtime"
	"testing"
	"time"

	"github.com/tinygo-org/tinygo/compileopts"
)

var sema = make(chan struct{}, runtime.NumCPU())

type sizeTest struct {
	target     string
	path       string
	codeSize   uint64
	rodataSize uint64
	dataSize   uint64
	bssSize    uint64
}

// Test whether code and data size is as expected for the given targets.
// This tests both the logic of loadProgramSize and checks that code size
// doesn't change unintentionally.
//
// If you find that code or data size is reduced, then great! You can reduce the
// number in this test.
// If you find that the code or data size is increased, take a look as to why
// this is. It could be due to an update (LLVM version, Go version, etc) which
// is fine, but it could also mean that a recent change introduced this size
// increase. If so, please consider whether this new feature is indeed worth the
// size increase for all users.
func TestBinarySize(t *testing.T) {
	if runtime.GOOS == "linux" && !hasBuiltinTools {
		// Debian LLVM packages are modified a bit and tend to produce
		// different machine code. Ideally we'd fix this (with some attributes
		// or something?), but for now skip it.
		t.Skip("Skip: using external LLVM version so binary size might differ")
	}

	// This is a small number of very diverse targets that we want to test.
	tests := []sizeTest{
		// microcontrollers
		{"hifive1b", "examples/echo", 4612, 280, 0, 2252},
		{"microbit", "examples/serial", 2724, 388, 8, 2256},
		{"wioterminal", "examples/pininterrupt", 6063, 1485, 116, 6816},

		// TODO: also check wasm. Right now this is difficult, because
		// wasm binaries are run through wasm-opt and therefore the
		// output varies by binaryen version.
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.target+"/"+tc.path, func(t *testing.T) {
			t.Parallel()

			// Build the binary.
			options := compileopts.Options{
				Target:        tc.target,
				Opt:           "z",
				Semaphore:     sema,
				InterpTimeout: 60 * time.Second,
				Debug:         true,
				VerifyIR:      true,
			}
			target, err := compileopts.LoadTarget(&options)
			if err != nil {
				t.Fatal("could not load target:", err)
			}
			config := &compileopts.Config{
				Options: &options,
				Target:  target,
			}
			result, err := Build(tc.path, "", t.TempDir(), config)
			if err != nil {
				t.Fatal("could not build:", err)
			}

			// Check whether the size of the binary matches the expected size.
			sizes, err := loadProgramSize(result.Executable, nil)
			if err != nil {
				t.Fatal("could not read program size:", err)
			}
			if sizes.Code != tc.codeSize || sizes.ROData != tc.rodataSize || sizes.Data != tc.dataSize || sizes.BSS != tc.bssSize {
				t.Errorf("Unexpected code size when compiling: -target=%s %s", tc.target, tc.path)
				t.Errorf("            code rodata   data    bss")
				t.Errorf("expected: %6d %6d %6d %6d", tc.codeSize, tc.rodataSize, tc.dataSize, tc.bssSize)
				t.Errorf("actual:   %6d %6d %6d %6d", sizes.Code, sizes.ROData, sizes.Data, sizes.BSS)
			}
		})
	}
}
