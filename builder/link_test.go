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
			counts := map[string]linkSymbolCounts{}
			for i, pkg := range pkgs {
				ir := fmt.Sprintf("define %s i32 @helper() {\n  ret i32 %d\n}\n\n"+
					"define i32 @%s() {\n  %%r = call i32 @helper()\n  ret i32 %%r\n}\n", tc.helper, i+1, pkg)
				if err := linkPackageModule(
					linker,
					mod,
					parseTestIR(t, ctx, ir),
					counts,
				); err != nil {
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

func TestLinkPackageModuleWeak(t *testing.T) {
	t.Run("unoverridden weak", func(t *testing.T) {
		ctx := llvm.NewContext()
		defer ctx.Dispose()
		mod := ctx.NewModule("main")
		defer mod.Dispose()
		linker := llvm.NewLinker(mod)
		defer linker.Dispose()
		counts := map[string]linkSymbolCounts{}

		ir1 := "define weak hidden i32 @hook() {\n  ret i32 1\n}\n"
		ir2 := "declare i32 @hook()\ndefine i32 @caller() {\n  %r = call i32 @hook()\n  ret i32 %r\n}\n"
		for _, ir := range []string{ir1, ir2} {
			if err := linkPackageModule(
				linker,
				mod,
				parseTestIR(t, ctx, ir),
				counts,
			); err != nil {
				t.Fatal(err)
			}
		}
		if err := llvm.VerifyModule(mod, llvm.ReturnStatusAction); err != nil {
			t.Fatal(err)
		}
		if got := calledResult(t, mod, "caller"); got != 1 {
			t.Errorf("caller calls hook returning %d, want 1", got)
		}
		fn := mod.NamedFunction("hook")
		if fn.Linkage() != llvm.WeakAnyLinkage {
			t.Errorf("hook linkage = %v, want WeakAnyLinkage", fn.Linkage())
		}
		if fn.Visibility() != llvm.HiddenVisibility {
			t.Errorf("hook visibility = %v, want HiddenVisibility", fn.Visibility())
		}
	})

	orders := []struct {
		name string
		irs  []string
	}{
		{
			name: "weak then strong",
			irs: []string{
				"define weak hidden i32 @hook() {\n  ret i32 1\n}\ndefine i32 @callerA() {\n  %r = call i32 @hook()\n  ret i32 %r\n}\n",
				"define hidden i32 @hook() {\n  ret i32 2\n}\ndefine i32 @callerB() {\n  %r = call i32 @hook()\n  ret i32 %r\n}\n",
			},
		},
		{
			name: "strong then weak",
			irs: []string{
				"define hidden i32 @hook() {\n  ret i32 2\n}\ndefine i32 @callerB() {\n  %r = call i32 @hook()\n  ret i32 %r\n}\n",
				"define weak hidden i32 @hook() {\n  ret i32 1\n}\ndefine i32 @callerA() {\n  %r = call i32 @hook()\n  ret i32 %r\n}\n",
			},
		},
		{
			name: "decl then weak then strong",
			irs: []string{
				"declare i32 @hook()\ndefine i32 @callerA() {\n  %r = call i32 @hook()\n  ret i32 %r\n}\n",
				"define weak hidden i32 @hook() {\n  ret i32 1\n}\n",
				"define hidden i32 @hook() {\n  ret i32 2\n}\ndefine i32 @callerB() {\n  %r = call i32 @hook()\n  ret i32 %r\n}\n",
			},
		},
		{
			name: "decl then strong then weak",
			irs: []string{
				"declare i32 @hook()\ndefine i32 @callerA() {\n  %r = call i32 @hook()\n  ret i32 %r\n}\n",
				"define hidden i32 @hook() {\n  ret i32 2\n}\ndefine i32 @callerB() {\n  %r = call i32 @hook()\n  ret i32 %r\n}\n",
				"define weak hidden i32 @hook() {\n  ret i32 1\n}\n",
			},
		},
	}
	for _, tc := range orders {
		t.Run(tc.name, func(t *testing.T) {
			ctx := llvm.NewContext()
			defer ctx.Dispose()
			mod := ctx.NewModule("main")
			defer mod.Dispose()
			linker := llvm.NewLinker(mod)
			defer linker.Dispose()
			counts := map[string]linkSymbolCounts{}

			for _, ir := range tc.irs {
				if err := linkPackageModule(
					linker,
					mod,
					parseTestIR(t, ctx, ir),
					counts,
				); err != nil {
					t.Fatal(err)
				}
			}
			if err := llvm.VerifyModule(mod, llvm.ReturnStatusAction); err != nil {
				t.Fatal(err)
			}
			for _, caller := range []string{"callerA", "callerB"} {
				if got := calledResult(t, mod, caller); got != 2 {
					t.Errorf("%s calls hook returning %d, want 2", caller, got)
				}
			}
			if got := mod.NamedFunction("hook").Linkage(); got != llvm.ExternalLinkage {
				t.Errorf("hook linkage = %v, want ExternalLinkage", got)
			}
		})
	}

	weakIR := "define weak hidden i32 @hook() {\n  ret i32 1\n}\n"
	strong1IR := "define hidden i32 @hook() {\n  ret i32 2\n}\n"
	strong2IR := "define hidden i32 @hook() {\n  ret i32 3\n}\n"

	errorCases := []struct {
		name string
		irs  []string
	}{
		{name: "weak weak", irs: []string{weakIR, weakIR}},
		{name: "weak strong strong", irs: []string{weakIR, strong1IR, strong2IR}},
		{name: "strong weak strong", irs: []string{strong1IR, weakIR, strong2IR}},
		{name: "strong strong weak", irs: []string{strong1IR, strong2IR, weakIR}},
		{name: "weak strong weak", irs: []string{weakIR, strong1IR, weakIR}},
		{name: "strong weak weak", irs: []string{strong1IR, weakIR, weakIR}},
	}
	for _, tc := range errorCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := llvm.NewContext()
			defer ctx.Dispose()
			mod := ctx.NewModule("main")
			defer mod.Dispose()
			linker := llvm.NewLinker(mod)
			defer linker.Dispose()
			counts := map[string]linkSymbolCounts{}

			var linkErr error
			for _, ir := range tc.irs {
				if err := linkPackageModule(
					linker,
					mod,
					parseTestIR(t, ctx, ir),
					counts,
				); err != nil {
					linkErr = err
					break
				}
			}
			if linkErr == nil {
				t.Fatal("expected duplicate link error, got nil")
			}
		})
	}
}

func TestInternalizeModule(t *testing.T) {
	ctx := llvm.NewContext()
	defer ctx.Dispose()

	ir := `
define weak hidden void @weakHidden() {
  ret void
}
define weak void @weakDefault() {
  ret void
}
define hidden void @extHidden() {
  ret void
}
define linkonce_odr void @odrDefault() {
  ret void
}
define void @extDefault() {
  ret void
}
`
	mod := parseTestIR(t, ctx, ir)
	defer mod.Dispose()

	internalizeModule(mod)
	if err := llvm.VerifyModule(mod, llvm.ReturnStatusAction); err != nil {
		t.Fatal(err)
	}

	for _, tc := range []struct {
		name string
		want llvm.Linkage
	}{
		{name: "weakHidden", want: llvm.InternalLinkage},
		{name: "weakDefault", want: llvm.InternalLinkage},
		{name: "extHidden", want: llvm.InternalLinkage},
		{name: "odrDefault", want: llvm.InternalLinkage},
		{name: "extDefault", want: llvm.ExternalLinkage},
	} {
		fn := mod.NamedFunction(tc.name)
		if fn.Linkage() != tc.want {
			t.Errorf("%s linkage = %v, want %v", tc.name, fn.Linkage(), tc.want)
		}
		if fn.Visibility() != llvm.DefaultVisibility {
			t.Errorf("%s visibility = %v, want DefaultVisibility", tc.name, fn.Visibility())
		}
	}
}
