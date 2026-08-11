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
