package transform

import (
	"strings"
	"testing"

	"github.com/tinygo-org/tinygo/compiler/llvmutil"
	"tinygo.org/x/go-llvm"
)

func TestNoRuntimeRoots(t *testing.T) {
	for _, permanent := range []bool{false, true} {
		t.Run(map[bool]string{false: "temporary-only", true: "permanent-and-temporary"}[permanent], func(t *testing.T) {
			ctx := llvm.NewContext()
			defer ctx.Dispose()
			mod := ctx.NewModule("test")
			defer mod.Dispose()
			builder := ctx.NewBuilder()
			defer builder.Dispose()
			fnType := llvm.FunctionType(ctx.VoidType(), nil, false)
			marked := llvm.AddFunction(mod, "marked", fnType)
			marked.SetLinkage(llvm.InternalLinkage)
			marked.AddFunctionAttr(ctx.CreateStringAttribute("tinygo-noruntime", ""))
			builder.SetInsertPointAtEnd(ctx.AddBasicBlock(marked, "entry"))
			builder.CreateRetVoid()
			llvmutil.AppendToGlobal(mod, "llvm.used", marked)
			llvmutil.AppendToGlobal(mod, "tinygo.indirect-abi", marked)
			if permanent {
				llvmutil.AppendToGlobal(mod, "llvm.used", marked)
			}

			// A live unmarked ABI root must remain protected after the check.
			live := llvm.AddFunction(mod, "live", fnType)
			live.SetLinkage(llvm.InternalLinkage)
			builder.SetInsertPointAtEnd(ctx.AddBasicBlock(live, "entry"))
			builder.CreateRetVoid()
			root := llvm.AddFunction(mod, "root", fnType)
			builder.SetInsertPointAtEnd(ctx.AddBasicBlock(root, "entry"))
			builder.CreateCall(fnType, live, nil, "")
			builder.CreateRetVoid()
			llvmutil.AppendToGlobal(mod, "llvm.used", live)
			llvmutil.AppendToGlobal(mod, "tinygo.indirect-abi", live)

			errs := CheckNoRuntime(mod)
			if permanent {
				if len(errs) != 1 || !strings.Contains(errs[0].Error(), "marked remains reachable") {
					t.Fatalf("expected one marked function error, got %v", errs)
				}
			} else {
				if len(errs) != 0 {
					t.Fatal(errs)
				}
				if !mod.NamedFunction("marked").IsNil() {
					t.Fatal("dead marked function remains")
				}
			}
			roots := mod.NamedGlobal("tinygo.indirect-abi")
			wantRoots := 1
			if permanent {
				wantRoots++
			}
			if roots.IsNil() || roots.Initializer().Type().ArrayLength() != wantRoots {
				t.Fatal("surviving ABI roots were not restored")
			}
			if err := llvm.VerifyModule(mod, llvm.ReturnStatusAction); err != nil {
				t.Fatal(err)
			}
		})
	}
}
