package compiler

import (
	"testing"

	"tinygo.org/x/go-llvm"
)

func TestParamNeedsSpill(t *testing.T) {
	t.Parallel()
	ctx := llvm.NewContext()
	defer ctx.Dispose()
	targetData := llvm.NewTargetData("e-m:e-p:32:32-p10:8:8-p20:8:8-i64:64-i128:128-n32:64-S128-ni:1:10:20")
	defer targetData.Dispose()
	c := &compilerContext{ctx: ctx, targetData: targetData}

	i32 := ctx.Int32Type()
	makeStruct := func(n int) llvm.Type {
		fields := make([]llvm.Type, n)
		for i := range fields {
			fields[i] = i32
		}
		return ctx.StructType(fields, false)
	}

	for _, tc := range []struct {
		name   string
		typ    llvm.Type
		leaves int
		spill  bool
	}{
		{"i32", i32, 1, false},
		{"empty", makeStruct(0), 0, false},
		{"flat16", makeStruct(16), 16, false},
		{"flat17", makeStruct(17), 17, true},
		{"nested16", ctx.StructType([]llvm.Type{makeStruct(8), makeStruct(8)}, false), 16, false},
		{"nested17", ctx.StructType([]llvm.Type{makeStruct(8), makeStruct(9)}, false), 17, true},
		{"zeroSizeField", ctx.StructType([]llvm.Type{ctx.StructType(nil, false), i32}, false), 1, false},
		{"array16", llvm.ArrayType(i32, 16), 16, false},
		{"array17", llvm.ArrayType(i32, 17), 17, true},
		{"structWithArray", ctx.StructType([]llvm.Type{llvm.ArrayType(i32, 16), i32}, false), 17, true},
		{"hugeArray", llvm.ArrayType(makeStruct(4), 1<<29), 1 << 30, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if leaves := c.countFlattenedLeaves(tc.typ); leaves != tc.leaves {
				t.Errorf("countFlattenedLeaves = %d, want %d", leaves, tc.leaves)
			}
			if spill := c.paramNeedsSpill(tc.typ); spill != tc.spill {
				t.Errorf("paramNeedsSpill = %v, want %v", spill, tc.spill)
			}
		})
	}
}
