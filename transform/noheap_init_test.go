package transform_test

import (
	"strings"
	"testing"

	"github.com/tinygo-org/tinygo/compileopts"
	"github.com/tinygo-org/tinygo/transform"
	"tinygo.org/x/go-llvm"
)

func TestCheckNoHeapInit(t *testing.T) {
	t.Parallel()

	t.Run("missing hwInit0", func(t *testing.T) {
		ctx := llvm.NewContext()
		defer ctx.Dispose()
		mod := ctx.NewModule("empty")
		defer mod.Dispose()

		if errs := transform.CheckNoHeapInit(mod); len(errs) != 0 {
			t.Fatalf("unexpected errors: %v", errs)
		}
	})

	t.Run("clean path with inline asm and intrinsic", func(t *testing.T) {
		ctx := llvm.NewContext()
		defer ctx.Dispose()
		mod := ctx.NewModule("clean")
		defer mod.Dispose()
		builder := ctx.NewBuilder()
		defer builder.Dispose()

		voidType := ctx.VoidType()
		fnType := llvm.FunctionType(voidType, nil, false)
		ptrType := llvm.PointerType(ctx.Int8Type(), 0)
		allocType := llvm.FunctionType(ptrType, []llvm.Type{ctx.Int32Type(), ptrType, ptrType}, false)

		alloc := llvm.AddFunction(mod, "runtime.alloc", allocType)
		trap := llvm.AddFunction(mod, "llvm.trap", fnType)
		helper := llvm.AddFunction(mod, "runtime.socInit0", fnType)
		builder.SetInsertPointAtEnd(ctx.AddBasicBlock(helper, "entry"))
		asm := llvm.InlineAsm(
			fnType,
			"nop",
			"",
			true,
			false,
			0,
			false,
		)
		builder.CreateCall(
			fnType,
			asm,
			nil,
			"",
		)
		builder.CreateCall(
			fnType,
			trap,
			nil,
			"",
		)
		builder.CreateRetVoid()

		hwInit0 := llvm.AddFunction(mod, "runtime.hwInit0", fnType)
		builder.SetInsertPointAtEnd(ctx.AddBasicBlock(hwInit0, "entry"))
		builder.CreateCall(
			fnType,
			helper,
			nil,
			"",
		)
		builder.CreateCall(
			fnType,
			helper,
			nil,
			"",
		)
		builder.CreateRetVoid()

		// Allocation outside hwInit0 reachability must not be flagged.
		other := llvm.AddFunction(mod, "main.main", fnType)
		builder.SetInsertPointAtEnd(ctx.AddBasicBlock(other, "entry"))
		builder.CreateCall(
			allocType,
			alloc,
			[]llvm.Value{
				llvm.ConstInt(ctx.Int32Type(), 4, false),
				llvm.ConstNull(ptrType),
				llvm.ConstNull(ptrType),
			},
			"",
		)
		builder.CreateRetVoid()

		if errs := transform.CheckNoHeapInit(mod); len(errs) != 0 {
			t.Fatalf("unexpected errors: %v", errs)
		}
	})

	t.Run("transitive alloc rejected", func(t *testing.T) {
		for _, allocName := range []string{"runtime.alloc", "runtime.alloc_noheap"} {
			t.Run(allocName, func(t *testing.T) {
				ctx := llvm.NewContext()
				defer ctx.Dispose()
				mod := ctx.NewModule("alloc")
				defer mod.Dispose()
				builder := ctx.NewBuilder()
				defer builder.Dispose()

				voidType := ctx.VoidType()
				fnType := llvm.FunctionType(voidType, nil, false)
				ptrType := llvm.PointerType(ctx.Int8Type(), 0)
				allocType := llvm.FunctionType(ptrType, []llvm.Type{ctx.Int32Type(), ptrType, ptrType}, false)

				alloc := llvm.AddFunction(mod, allocName, allocType)
				helper := llvm.AddFunction(mod, "board.helper", fnType)
				builder.SetInsertPointAtEnd(ctx.AddBasicBlock(helper, "entry"))
				builder.CreateCall(
					allocType,
					alloc,
					[]llvm.Value{
						llvm.ConstInt(ctx.Int32Type(), 8, false),
						llvm.ConstNull(ptrType),
						llvm.ConstNull(ptrType),
					},
					"",
				)
				builder.CreateRetVoid()

				boardInit0 := llvm.AddFunction(mod, "runtime.boardInit0", fnType)
				builder.SetInsertPointAtEnd(ctx.AddBasicBlock(boardInit0, "entry"))
				builder.CreateCall(
					fnType,
					helper,
					nil,
					"",
				)
				builder.CreateRetVoid()

				hwInit0 := llvm.AddFunction(mod, "runtime.hwInit0", fnType)
				builder.SetInsertPointAtEnd(ctx.AddBasicBlock(hwInit0, "entry"))
				builder.CreateCall(
					fnType,
					boardInit0,
					nil,
					"",
				)
				builder.CreateRetVoid()

				errs := transform.CheckNoHeapInit(mod)
				if len(errs) != 1 || !strings.Contains(errs[0].Error(), "heap allocation is not allowed in pre-heap initialization") {
					t.Fatalf("got errors %v, want 1 heap allocation error", errs)
				}
			})
		}
	})

	t.Run("indirect call rejected", func(t *testing.T) {
		ctx := llvm.NewContext()
		defer ctx.Dispose()
		mod := ctx.NewModule("indirect")
		defer mod.Dispose()
		builder := ctx.NewBuilder()
		defer builder.Dispose()

		voidType := ctx.VoidType()
		fnType := llvm.FunctionType(voidType, nil, false)
		ptrType := llvm.PointerType(ctx.Int8Type(), 0)
		indirectCallerType := llvm.FunctionType(voidType, []llvm.Type{ptrType}, false)

		caller := llvm.AddFunction(mod, "runtime.socInit0", indirectCallerType)
		builder.SetInsertPointAtEnd(ctx.AddBasicBlock(caller, "entry"))
		builder.CreateCall(
			fnType,
			caller.Param(0),
			nil,
			"",
		)
		builder.CreateRetVoid()

		hwInit0 := llvm.AddFunction(mod, "runtime.hwInit0", fnType)
		builder.SetInsertPointAtEnd(ctx.AddBasicBlock(hwInit0, "entry"))
		builder.CreateCall(
			indirectCallerType,
			caller,
			[]llvm.Value{llvm.ConstNull(ptrType)},
			"",
		)
		builder.CreateRetVoid()

		errs := transform.CheckNoHeapInit(mod)
		if len(errs) != 1 || !strings.Contains(errs[0].Error(), "indirect call is not allowed in pre-heap initialization") {
			t.Fatalf("got errors %v, want 1 indirect call error", errs)
		}
	})
}

func TestOptimizeNoHeapInit(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		opt      string
		escaping bool
		wantErr  bool
	}{
		{name: "opt1 non-escaping", opt: "1", escaping: false, wantErr: false},
		{name: "opt1 escaping", opt: "1", escaping: true, wantErr: true},
		{name: "opt0 non-escaping", opt: "0", escaping: false, wantErr: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx := llvm.NewContext()
			defer ctx.Dispose()
			buf, err := llvm.NewMemoryBufferFromFile("testdata/noheap-init.ll")
			if err != nil {
				t.Fatal(err)
			}
			mod, err := ctx.ParseIR(buf)
			if err != nil {
				t.Fatal(err)
			}
			defer mod.Dispose()

			if tc.escaping {
				orig := mod.NamedFunction("runtime.hwInit0")
				esc := mod.NamedFunction("runtime.hwInit0Escaping")
				orig.ReplaceAllUsesWith(esc)
				orig.EraseFromParentAsFunction()
				esc.SetName("runtime.hwInit0")
			}
			if err := llvm.VerifyModule(mod, llvm.ReturnStatusAction); err != nil {
				t.Fatal(err)
			}

			config := &compileopts.Config{
				Options: &compileopts.Options{Opt: tc.opt},
				Target:  &compileopts.TargetSpec{},
			}
			errs := transform.Optimize(mod, config)
			if tc.wantErr {
				if len(errs) == 0 || !strings.Contains(errs[0].Error(), "heap allocation is not allowed in pre-heap initialization") {
					t.Fatalf("got errors %v, want heap allocation error", errs)
				}
				return
			}
			if len(errs) != 0 {
				t.Fatalf("unexpected errors: %v", errs)
			}
			if err := llvm.VerifyModule(mod, llvm.ReturnStatusAction); err != nil {
				t.Fatal(err)
			}
		})
	}
}
