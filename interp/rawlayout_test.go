package interp

import (
	"slices"
	"testing"
	"time"

	"tinygo.org/x/go-llvm"
)

// Test that rawLLVMValue keeps each pointer at its byte offset, and that it
// gives an error for part of a pointer.
func TestRawLLVMValue(t *testing.T) {
	ctx := llvm.NewContext()
	defer ctx.Dispose()
	mod := ctx.NewModule("rawlayout")
	defer mod.Dispose()
	mod.SetTarget("x86_64--linux")
	mod.SetDataLayout("e-m:e-i64:64-f80:128-n8:16:32:64-S128")
	x := llvm.AddGlobal(mod, ctx.Int64Type(), "x")
	r := newRunner(mod, time.Minute, DefaultMaxInterpBlockEntries, false)
	defer r.dispose()
	ptr, err := r.getValue(x).asPointer(r)
	if err != nil {
		t.Fatal(err)
	}
	p := slices.Repeat([]uint64{ptr.pointer}, 8)
	mem := memoryView{r: r}

	for _, tc := range []struct {
		buf  []uint64
		want string // empty if an error is expected
	}{
		{slices.Concat([]uint64{1, 2, 3, 4}, p, []uint64{5, 6, 7, 8}),
			`<{ [4 x i8], ptr, [4 x i8] }> <{ [4 x i8] c"\01\02\03\04", ptr @x, [4 x i8] c"\05\06\07\08" }>`},
		{slices.Concat([]uint64{9}, p, p),
			`<{ i8, ptr, ptr }> <{ i8 9, ptr @x, ptr @x }>`},
		{slices.Concat([]uint64{0, 0, 0, 0}, p, p[:4]), ""},
		{slices.Concat([]uint64{0}, p[:7]), ""},
	} {
		v, err := rawValue{tc.buf}.rawLLVMValue(&mem)
		if tc.want == "" {
			if err == nil {
				t.Errorf("expected an error for %v, got %s", tc.buf, v.String())
			}
			continue
		}
		if err != nil {
			t.Errorf("unexpected error for %v: %v", tc.buf, err)
			continue
		}
		if s := v.String(); s != tc.want {
			t.Errorf("unexpected value for %v:\nwant: %s\ngot:  %s", tc.buf, tc.want, s)
		}
		if size := r.targetData.TypeAllocSize(v.Type()); size != uint64(len(tc.buf)) {
			t.Errorf("unexpected size for %v: %d", tc.buf, size)
		}
	}
}
