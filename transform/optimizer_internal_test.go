package transform

import (
	"testing"

	"github.com/tinygo-org/tinygo/compileopts"
	"tinygo.org/x/go-llvm"
)

func TestOptimizeGCGlobalRoots(t *testing.T) {
	t.Parallel()
	for _, used := range []bool{false, true} {
		name := "unused"
		if used {
			name = "used"
		}
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			ctx := llvm.NewContext()
			defer ctx.Dispose()
			buf, err := llvm.NewMemoryBufferFromFile("testdata/optimizer-gc-roots.ll")
			if err != nil {
				t.Fatal(err)
			}
			mod, err := ctx.ParseIR(buf)
			if err != nil {
				t.Fatal(err)
			}
			defer mod.Dispose()

			if used {
				start := mod.NamedFunction("start")
				store := llvm.NextInstruction(start.EntryBasicBlock().FirstInstruction())
				store.SetOperand(0, start.FirstParam())
			}
			config := &compileopts.Config{
				Options: &compileopts.Options{Opt: "z"},
				Target:  &compileopts.TargetSpec{},
			}
			if errs := Optimize(mod, config); len(errs) != 0 {
				t.Fatal(errs)
			}
			if err := llvm.VerifyModule(mod, llvm.ReturnStatusAction); err != nil {
				t.Fatal(err)
			}
			want := uint64(0)
			if used {
				want = 1
			}
			rootCount := mod.NamedFunction("runtime.gcGlobalRootCount")
			if rootCount.IsNil() || rootCount.FirstBasicBlock().IsNil() {
				t.Fatal("GC global roots were not generated")
			}
			count := rootCount.EntryBasicBlock().LastInstruction().Operand(0).ZExtValue()
			if count != want {
				t.Errorf("got %d global roots, want %d", count, want)
			}
			if !used && !mod.NamedGlobal("queue").IsNil() {
				t.Error("unused queue was retained")
			}
		})
	}
}

func TestBlockGlobalAllocPromotionUses(t *testing.T) {
	ctx := llvm.NewContext()
	defer ctx.Dispose()
	buf, err := llvm.NewMemoryBufferFromFile("testdata/optimizer-alloc-uses.ll")
	if err != nil {
		t.Fatal(err)
	}
	mod, err := ctx.ParseIR(buf)
	if err != nil {
		t.Fatal(err)
	}
	defer mod.Dispose()

	blockGlobalAllocPromotion(mod)
	if err := llvm.VerifyModule(mod, llvm.ReturnStatusAction); err != nil {
		t.Fatal(err)
	}
	marker := mod.NamedFunction("tinygo.gc.alloc.marker")
	if marker.IsNil() {
		t.Fatal("allocation marker was not created")
	}
	if uses := getAllUses(marker); len(uses) != 1 {
		t.Fatalf("got %d marker uses, want 1", len(uses))
	}
}

func TestPruneUnusedRuntimeStartup(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		path        string
		wantEnabled bool
	}{
		{
			name: "unused",
			path: "testdata/runtime-startup-unused.ll",
		},
		{
			name:        "used",
			path:        "testdata/runtime-startup-used.ll",
			wantEnabled: true,
		},
		{
			name:        "referenced",
			path:        "testdata/runtime-startup-referenced.ll",
			wantEnabled: true,
		},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
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

			if err := PrepareRuntimeStartup(mod); err != nil {
				t.Fatal(err)
			}
			for _, name := range []string{"runtime.gorootEnvEnabled", "runtime.godebugEnvEnabled"} {
				if !mod.NamedFunction(name).IsNil() {
					t.Errorf("%s was not removed", name)
				}
			}
			for _, name := range []string{"test.goroot", "test.godebug"} {
				ret := mod.NamedFunction(name).EntryBasicBlock().LastInstruction()
				got := ret.Operand(0).ZExtValue() != 0
				if got != tc.wantEnabled {
					t.Errorf("%s returned %v, want %v", name, got, tc.wantEnabled)
				}
			}
		})
	}
}
