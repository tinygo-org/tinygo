package transform_test

import (
	"errors"
	"os"
	"testing"

	"github.com/tinygo-org/tinygo/transform"
	"tinygo.org/x/go-llvm"
)

func TestMakeGCStackSlots(t *testing.T) {
	t.Parallel()
	testTransform(t, "testdata/gc-stackslots", func(mod llvm.Module) {
		transform.MakeGCStackSlots(mod)
	})
}

func TestMakeGCGlobalRootsAVR(t *testing.T) {
	t.Parallel()
	testTransform(t, "testdata/gc-globals-avr", func(mod llvm.Module) {
		transform.MakeGCStackSlots(mod)
	})
}

// TestGCStackSlotCounts checks slot counts without comparing unrelated IR to a golden file.
// See https://github.com/tinygo-org/tinygo/pull/5762#discussion_r4113188988.
func TestGCStackSlotCounts(t *testing.T) {
	t.Parallel()

	want := map[string]int{
		// The inputs of the acyclic merge feeding this loop do not need slots.
		"acyclicPhiIntoLoop": 3,
		// Every input here does, and must keep them.
		"nestedLoopPhis": 4,
	}

	ctx := llvm.NewContext()
	defer ctx.Dispose()
	buf, err := llvm.NewMemoryBufferFromFile("testdata/gc-slotcounts.ll")
	os.Stat("testdata/gc-slotcounts.ll") // make sure `go test` caching tracks this file
	if err != nil {
		t.Fatalf("could not read file: %v", err)
	}
	mod, err := ctx.ParseIR(buf)
	if err != nil {
		t.Fatalf("could not load module:\n%v", err)
	}
	defer mod.Dispose()

	transform.MakeGCStackSlots(mod)
	if err := llvm.VerifyModule(mod, llvm.PrintMessageAction); err != nil {
		t.Fatal("IR verification failed")
	}

	for name, want := range want {
		fn := mod.NamedFunction(name)
		if fn.IsNil() {
			t.Errorf("%s: not found in module", name)
			continue
		}
		got, err := stackSlotCount(fn)
		if err != nil {
			t.Errorf("%s: %v", name, err)
			continue
		}
		if got != want {
			t.Errorf("%s: got %d stack slots, want %d", name, got, want)
		}
	}
}

// stackSlotCount returns the number of pointer slots in the gc.stackobject of
// fn. The stack object is {parent, numSlots, slots...}, so the count is the
// number of struct fields beyond the first two.
func stackSlotCount(fn llvm.Value) (int, error) {
	entry := fn.EntryBasicBlock()
	for inst := entry.FirstInstruction(); !inst.IsNil(); inst = llvm.NextInstruction(inst) {
		if inst.IsAAllocaInst().IsNil() || inst.Name() != "gc.stackobject" {
			continue
		}
		return inst.AllocatedType().StructElementTypesCount() - 2, nil
	}
	return 0, errors.New("no gc.stackobject in entry block")
}
