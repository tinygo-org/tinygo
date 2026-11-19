package builder

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/tinygo-org/tinygo/compileopts"
	"github.com/tinygo-org/tinygo/interp"
	"tinygo.org/x/go-llvm"
)

func TestRP2(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	files := map[string]string{
		"go.mod":  "module example.com/rp2test\n\ngo 1.25.0\n",
		"go.work": "go 1.25.0\n\nuse .\n",
		"main.go": "package main\n\nimport \"machine\"\n\nfunc main() {\n\tprintln(machine.CPUFrequency())\n}\n",
	}
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}

	targets := []string{"pico", "pico2", "pga2350"}

	opts := []string{"0", "1", "z"}
	serials := []string{"usb", "uart"}

	for _, target := range targets {
		for _, opt := range opts {
			for _, serial := range serials {
				name := target + "/opt=" + opt + "/serial=" + serial
				t.Run(name, func(t *testing.T) {
					t.Parallel()

					outLL := filepath.Join(t.TempDir(), "out.ll")
					options := &compileopts.Options{
						Target:                  target,
						Opt:                     opt,
						Serial:                  serial,
						Directory:               dir,
						Semaphore:               make(chan struct{}, 2),
						InterpTimeout:           60 * time.Second,
						InterpMaxLoopIterations: interp.DefaultMaxInterpBlockEntries,
					}
					config, err := NewConfig(options)
					if err != nil {
						t.Fatal(err)
					}
					if _, err := Build(
						".",
						outLL,
						t.TempDir(),
						config,
					); err != nil {
						t.Fatalf("Build(%s) failed: %v", name, err)
					}

					ctx := llvm.NewContext()
					defer ctx.Dispose()
					buf, err := llvm.NewMemoryBufferFromFile(outLL)
					if err != nil {
						t.Fatal(err)
					}
					mod, err := ctx.ParseIR(buf)
					if err != nil {
						buf.Dispose()
						t.Fatal(err)
					}
					defer mod.Dispose()

					verifyRP2BootHooksAndInitFolding(
						t,
						mod,
						opt,
					)
				})
			}
		}
	}

	t.Run("link elf", func(t *testing.T) {
		t.Parallel()
		for _, target := range targets {
			t.Run(target, func(t *testing.T) {
				t.Parallel()
				outELF := filepath.Join(t.TempDir(), "out.elf")
				options := &compileopts.Options{
					Target:                  target,
					Opt:                     "z",
					Debug:                   true,
					Directory:               dir,
					Semaphore:               make(chan struct{}, 2),
					InterpTimeout:           60 * time.Second,
					InterpMaxLoopIterations: interp.DefaultMaxInterpBlockEntries,
				}
				config, err := NewConfig(options)
				if err != nil {
					t.Fatal(err)
				}
				result, err := Build(
					".",
					outELF,
					t.TempDir(),
					config,
				)
				if err != nil {
					t.Fatalf("Build(%s) ELF failed: %v", target, err)
				}
				info, err := os.Stat(result.Binary)
				if err != nil || info.Size() == 0 {
					t.Fatalf("invalid ELF output %q: stat=%v err=%v", result.Binary, info, err)
				}
			})
		}
	})
}

func verifyRP2BootHooksAndInitFolding(t *testing.T, mod llvm.Module, opt string) {
	t.Helper()

	if opt == "0" {
		runCalls := directCallsInNamedFunc(t, mod, "runtime.run")
		assertCallOrder(
			t,
			"runtime.run",
			runCalls,
			[]string{"runtime.hwInit0", "runtime.initRand", "runtime.initHeap"},
		)
		hw0Calls := directCallsInNamedFunc(t, mod, "runtime.hwInit0")
		if !slices.Contains(hw0Calls, "runtime.socInit0") {
			t.Fatalf("runtime.hwInit0 calls %v, missing runtime.socInit0", hw0Calls)
		}
		soc0Calls := directCallsInNamedFunc(t, mod, "runtime.socInit0")
		if !slices.Contains(soc0Calls, "runtime.machineInit") {
			t.Fatalf("runtime.socInit0 calls %v, missing runtime.machineInit", soc0Calls)
		}
	}

	initFuncs := []string{"runtime.initAll", "runtime.init", "machine.init", "device/rp.init"}
	for _, fnName := range initFuncs {
		fn := mod.NamedFunction(fnName)
		if fn.IsNil() || fn.IsDeclaration() {
			continue
		}
		calls := directCallsInFunc(fn)
		if slices.Contains(calls, "runtime.machineInit") {
			t.Errorf("%s unexpectedly calls runtime.machineInit: %v", fnName, calls)
		}
	}

	requiredMachineGlobals := []string{
		"machine.clocks",
		"machine.xosc",
		"machine.pllSys",
		"machine.pllUSB",
		"machine.timer",
		"machine.resets",
		"machine.pllsysFB",
		"machine.pllsysPD1",
		"machine.pllsysPD2",
	}

	checkTargets := map[llvm.Value]string{}
	for _, globalName := range requiredMachineGlobals {
		g := mod.NamedGlobal(globalName)
		if g.IsNil() {
			if opt == "0" {
				t.Errorf("expected global %s at -opt=0", globalName)
			}
			continue
		}
		checkTargets[g] = globalName
		if g.IsDeclaration() {
			t.Errorf("global %s is unexpectedly external declaration", globalName)
			continue
		}
		initVal := g.Initializer()
		if initVal.IsNil() || initVal.IsNull() {
			t.Errorf("global %s has zero/nil initializer instead of folded constant", globalName)
		}
	}

	for g := mod.FirstGlobal(); !g.IsNil(); g = llvm.NextGlobal(g) {
		name := g.Name()
		if !strings.HasPrefix(name, "device/rp.") {
			continue
		}
		if g.IsDeclaration() || g.IsGlobalConstant() {
			continue
		}
		checkTargets[g] = name
		initVal := g.Initializer()
		if initVal.IsNil() || initVal.IsNull() {
			t.Errorf("global %s has zero/nil initializer instead of folded constant", name)
		}
	}

	visited := map[llvm.Value]bool{}
	for _, fnName := range []string{"runtime.initAll", "machine.init", "device/rp.init"} {
		fn := mod.NamedFunction(fnName)
		if fn.IsNil() || fn.IsDeclaration() {
			continue
		}
		assertNoStoresToGlobals(
			t,
			fn,
			checkTargets,
			visited,
		)
	}
}

func assertNoStoresToGlobals(t *testing.T, fn llvm.Value, globals map[llvm.Value]string, visited map[llvm.Value]bool) {
	t.Helper()
	if len(globals) == 0 || fn.IsNil() || fn.IsDeclaration() {
		return
	}
	if visited[fn] {
		return
	}
	visited[fn] = true
	for bb := fn.FirstBasicBlock(); !bb.IsNil(); bb = llvm.NextBasicBlock(bb) {
		for inst := bb.FirstInstruction(); !inst.IsNil(); inst = llvm.NextInstruction(inst) {
			op := inst.InstructionOpcode()
			if op == llvm.Call || op == llvm.Invoke {
				callee := inst.CalledValue()
				if !callee.IsNil() && !callee.IsAFunction().IsNil() {
					assertNoStoresToGlobals(
						t,
						callee,
						globals,
						visited,
					)
				}
				continue
			}
			if op != llvm.Store {
				continue
			}
			dst := stripPointerCasts(inst.Operand(1))
			if name, ok := globals[dst]; ok {
				t.Errorf("%s has residual store to folded global %s: %s", fn.Name(), name, strings.TrimSpace(inst.String()))
			}
		}
	}
}

func stripPointerCasts(val llvm.Value) llvm.Value {
	for {
		if val.IsNil() {
			return val
		}
		if !val.IsAInstruction().IsNil() {
			switch val.InstructionOpcode() {
			case llvm.BitCast, llvm.GetElementPtr:
				val = val.Operand(0)
				continue
			}
			return val
		}
		if !val.IsAConstantExpr().IsNil() {
			switch val.Opcode() {
			case llvm.BitCast, llvm.GetElementPtr:
				val = val.Operand(0)
				continue
			}
		}
		return val
	}
}
