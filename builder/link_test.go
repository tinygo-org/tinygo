package builder

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"tinygo.org/x/go-llvm"
)

func parseTestIR(t *testing.T, ctx llvm.Context, ir string) llvm.Module {
	t.Helper()
	path := filepath.Join(t.TempDir(), "mod.ll")
	if err := os.WriteFile(path, []byte(ir), 0o600); err != nil {
		t.Fatal(err)
	}
	buf, err := llvm.NewMemoryBufferFromFile(path)
	if err != nil {
		t.Fatal(err)
	}
	mod, err := ctx.ParseIR(buf)
	if err != nil {
		t.Fatal(err)
	}
	return mod
}

// calledResult returns the constant returned by the function that caller calls.
func calledResult(t *testing.T, mod llvm.Module, caller string) int64 {
	t.Helper()
	fn := mod.NamedFunction(caller)
	if fn.IsNil() {
		t.Fatalf("function %s not found", caller)
	}
	call := fn.FirstBasicBlock().FirstInstruction()
	callee := call.CalledValue()
	ret := callee.FirstBasicBlock().FirstInstruction()
	return ret.Operand(0).SExtValue()
}

func TestLinkPackageModule(t *testing.T) {
	tests := []struct {
		name   string
		helper string
		want   []int64
	}{
		{name: "internal", helper: "internal", want: []int64{1, 2, 3}},
		{name: "private", helper: "private", want: []int64{1, 2, 3}},
		// An external duplicate keeps the first definition.
		{name: "external", helper: "", want: []int64{1, 1, 1}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctx := llvm.NewContext()
			defer ctx.Dispose()
			mod := ctx.NewModule("main")
			defer mod.Dispose()
			linker := llvm.NewLinker(mod)
			defer linker.Dispose()

			pkgs := []string{"a", "b", "c"}
			for i, pkg := range pkgs {
				ir := fmt.Sprintf("define %s i32 @helper() {\n  ret i32 %d\n}\n\n"+
					"define i32 @%s() {\n  %%r = call i32 @helper()\n  ret i32 %%r\n}\n", tc.helper, i+1, pkg)
				if err := linkPackageModule(linker, mod, parseTestIR(t, ctx, ir)); err != nil {
					t.Fatal(err)
				}
			}
			if err := llvm.VerifyModule(mod, llvm.ReturnStatusAction); err != nil {
				t.Fatal(err)
			}
			for i, pkg := range pkgs {
				if got := calledResult(t, mod, pkg); got != tc.want[i] {
					t.Errorf("%s calls helper returning %d, want %d", pkg, got, tc.want[i])
				}
			}
		})
	}
}
