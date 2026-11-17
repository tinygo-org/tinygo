package transform_test

import (
	"go/types"
	"path/filepath"
	"testing"

	"github.com/tinygo-org/tinygo/compileopts"
	"github.com/tinygo-org/tinygo/compiler"
	"github.com/tinygo-org/tinygo/loader"
	"tinygo.org/x/go-llvm"
)

// refBlockInCycle is the algorithm that BasicBlock.InCycle replaced.
func refBlockInCycle(bb llvm.BasicBlock) bool {
	seen := map[llvm.BasicBlock]struct{}{}
	worklist := []llvm.BasicBlock{bb}
	for len(worklist) != 0 {
		cur := worklist[len(worklist)-1]
		worklist = worklist[:len(worklist)-1]
		term := cur.LastInstruction()
		if term.IsNil() {
			continue
		}
		for i := 0; i < term.SuccessorsCount(); i++ {
			succ := term.Successor(i)
			if succ == bb {
				return true
			}
			if _, ok := seen[succ]; ok {
				continue
			}
			seen[succ] = struct{}{}
			worklist = append(worklist, succ)
		}
	}
	return false
}

// compilePackagesForTesting compiles every package of the program into its own
// module, including the standard library.
func compilePackagesForTesting(t *testing.T, filename string) (mods []llvm.Module, machine llvm.TargetMachine) {
	t.Helper()
	target, err := compileopts.LoadTarget(&compileopts.Options{GOOS: "linux", GOARCH: "386"})
	if err != nil {
		t.Fatal("failed to load target:", err)
	}
	config := &compileopts.Config{
		Options: &compileopts.Options{},
		Target:  target,
	}
	compilerConfig := &compiler.Config{
		Triple:             config.Triple(),
		GOOS:               config.GOOS(),
		GOARCH:             config.GOARCH(),
		CodeModel:          config.CodeModel(),
		RelocationModel:    config.RelocationModel(),
		Scheduler:          config.Scheduler(),
		AutomaticStackSize: config.AutomaticStackSize(),
		PanicStrategy:      config.PanicStrategy(),
		PanicUnwind:        config.PanicUnwind(),
	}
	machine, err = compiler.NewTargetMachine(compilerConfig)
	if err != nil {
		t.Fatal("failed to create target machine:", err)
	}
	lprogram, err := loader.Load(config, filename, types.Config{Sizes: compiler.Sizes(machine)})
	if err != nil {
		t.Fatal("failed to load program:", err)
	}
	if err := lprogram.Parse(); err != nil {
		t.Fatal("could not parse:", err)
	}
	program := lprogram.LoadSSA()
	program.Build()
	for _, pkg := range lprogram.Sorted() {
		mod, errs := compiler.CompilePackage(pkg.ImportPath, pkg, program.Package(pkg.Pkg), machine, compilerConfig, false)
		if errs != nil {
			for _, err := range errs {
				t.Error(err)
			}
			t.FailNow()
		}
		mods = append(mods, mod)
	}
	return mods, machine
}

// compareHelpers compares the LLVM based helpers with the old algorithms on
// the module. It checks every stride-th block.
func compareHelpers(t *testing.T, mod llvm.Module, stride int, blocks, cyclic *int) {
	t.Helper()
	n := 0
	for fn := mod.FirstFunction(); !fn.IsNil(); fn = llvm.NextFunction(fn) {
		for bb := fn.FirstBasicBlock(); !bb.IsNil(); bb = llvm.NextBasicBlock(bb) {
			n++
			if n%stride != 0 {
				continue
			}
			*blocks++
			want := refBlockInCycle(bb)
			if want {
				*cyclic++
			}
			if got := bb.InCycle(); got != want {
				t.Errorf("%s block %s: InCycle() = %v, want %v", fn.Name(), bb.AsValue().Name(), got, want)
			}
		}
	}
}

// TestInCycleMatchesReference checks BasicBlock.InCycle on real programs, before
// and after LLVM's optimizations. The old algorithm is quadratic, so a short
// run samples the blocks.
func TestInCycleMatchesReference(t *testing.T) {
	t.Parallel()
	programs := []string{"generics.go", "reflect.go"}
	stride := 1
	if testing.Short() {
		stride = 25
	} else {
		programs = append(programs, "json.go", "stdlib.go", "goroutines.go")
	}
	for _, name := range programs {
		for _, optimize := range []bool{false, true} {
			desc := name
			if optimize {
				desc += "/optimized"
			}
			t.Run(desc, func(t *testing.T) {
				mods, machine := compilePackagesForTesting(t, filepath.Join("..", "testdata", name))
				defer machine.Dispose()
				var blocks, cyclic int
				for _, mod := range mods {
					if optimize {
						po := llvm.NewPassBuilderOptions()
						err := mod.RunPasses("default<O2>", machine, po)
						po.Dispose()
						if err != nil {
							t.Fatal(err)
						}
					}
					compareHelpers(t, mod, stride, &blocks, &cyclic)
					ctx := mod.Context()
					mod.Dispose()
					ctx.Dispose()
				}
				if cyclic == 0 || cyclic == blocks {
					t.Errorf("the blocks checked do not cover both answers, %d of %d are in a cycle", cyclic, blocks)
				}
			})
		}
	}
}
