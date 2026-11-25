package transform_test

import (
	"testing"

	"github.com/tinygo-org/tinygo/transform"
	"tinygo.org/x/go-llvm"
)

func TestOptimizeStringToBytes(t *testing.T) {
	t.Parallel()
	testTransform(t, "testdata/stringtobytes", func(mod llvm.Module) {
		// Run optimization pass.
		transform.OptimizeStringToBytes(mod)
	})
}

func TestOptimizeStringEqual(t *testing.T) {
	t.Parallel()
	testTransform(t, "testdata/stringequal", func(mod llvm.Module) {
		// Run optimization pass.
		transform.OptimizeStringEqual(mod)
	})
}

func TestOptimizeStringFromBytes(t *testing.T) {
	t.Parallel()
	testTransform(t, "testdata/stringfrombytes", func(mod llvm.Module) {
		// Run optimization pass.
		transform.OptimizeStringFromBytes(mod)
	})
}

func TestOptimizeStringFromBytesSnapshot(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name string
		path string
	}{
		{"returnCapture", "testdata/stringfrombytes-capture.ll"},
		{"selectedSource", "testdata/stringfrombytes-alias.ll"},
		{"writableArgument", "testdata/stringfrombytes-alias.ll"},
		{"readOnlyControl", "testdata/stringfrombytes-alias.ll"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if tc.name == "returnCapture" && llvm.VersionMajor < 21 {
				t.Skip("requires LLVM 21 capture attributes")
			}
			ctx := llvm.NewContext()
			defer ctx.Dispose()
			buf, err := llvm.NewMemoryBufferFromFile(tc.path)
			if err != nil {
				t.Fatal(err)
			}
			mod, err := ctx.ParseIR(buf)
			if err != nil {
				t.Fatal(err)
			}
			defer mod.Dispose()
			if err := llvm.VerifyModule(mod, llvm.ReturnStatusAction); err != nil {
				t.Fatal(err)
			}

			transform.OptimizeStringFromBytes(mod)
			if err := llvm.VerifyModule(mod, llvm.ReturnStatusAction); err != nil {
				t.Fatal(err)
			}

			fn := mod.NamedFunction(tc.name)
			if fn.IsNil() {
				t.Fatal("function not found")
			}
			copies := 0
			for bb := fn.FirstBasicBlock(); !bb.IsNil(); bb = llvm.NextBasicBlock(bb) {
				for inst := bb.FirstInstruction(); !inst.IsNil(); inst = llvm.NextInstruction(inst) {
					if !inst.IsACallInst().IsNil() && inst.CalledValue() == mod.NamedFunction("runtime.stringFromBytes") {
						copies++
					}
				}
			}
			want := 1
			if tc.name == "readOnlyControl" {
				want = 0
			}
			if copies != want {
				t.Errorf("got %d string copies, want %d:\n%s", copies, want, fn.String())
			}
		})
	}
}
