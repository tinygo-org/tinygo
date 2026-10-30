package compileopts

import (
	"slices"
	"testing"
)

func TestExtraFilesBoehm(t *testing.T) {
	target := &TargetSpec{
		GC: "precise",
		ExtraFiles: []string{
			"src/runtime/asm_tinygowasm.S",
		},
	}
	config := &Config{
		Options: &Options{},
		Target:  target,
	}

	got := config.ExtraFiles()
	want := []string{"src/runtime/asm_tinygowasm.S"}
	if !slices.Equal(got, want) {
		t.Fatalf("unexpected precise GC files: got %v, want %v", got, want)
	}

	config.Options.GC = "boehm"
	got = config.ExtraFiles()
	want = []string{
		"src/runtime/asm_tinygowasm.S",
		"src/runtime/gc_boehm.c",
	}
	if !slices.Equal(got, want) {
		t.Fatalf("unexpected Boehm GC files: got %v, want %v", got, want)
	}
}

func TestLDFlagsWasmStackSize(t *testing.T) {
	config := &Config{
		Options: &Options{
			Scheduler: "none",
			StackSize: 16 << 20,
		},
		Target: &TargetSpec{
			Linker:    "wasm-ld",
			Scheduler: "asyncify",
		},
	}
	if got := config.LDFlags(); !slices.Contains(got, "stack-size=16777216") {
		t.Fatalf("expected stack-size=16777216 in LDFlags for scheduler=none, got %v", got)
	}

	config.Options.Scheduler = "asyncify"
	if got := config.LDFlags(); slices.Contains(got, "stack-size=16777216") {
		t.Fatalf("unexpected stack-size in LDFlags for scheduler=asyncify, got %v", got)
	}
}
