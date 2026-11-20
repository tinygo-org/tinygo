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

func TestRP2BoardPackage(t *testing.T) {
	t.Parallel()

	mainSrc := `package main

import (
	"device/rp"
	"machine"
	"runtime/interrupt"
)

var testUART = machine.UART{
	Buffer: machine.NewRingBuffer(),
	Bus:    rp.UART0,
}

func main() {
	_ = machine.Watchdog
	_ = machine.Flash
	_, _ = machine.GetRNG()
	machine.InitADC()
	machine.GPIO0.Configure(machine.PinConfig{Mode: machine.PinInput})
	u := &testUART
	u.Interrupt = interrupt.New(rp.IRQ_UART0_IRQ, testUART.HandleInterrupt)

	// Verify explicit pin requirements and GPIO0 acceptance.
	if err := u.Configure(machine.UARTConfig{}); err == nil {
		println("expected uart error")
	}
	if err := u.Configure(machine.UARTConfig{TX: machine.NoPin, RX: machine.NoPin}); err == nil {
		println("expected uart nopin error")
	}
	if err := u.Configure(machine.UARTConfig{TX: machine.GPIO0, RX: machine.GPIO1}); err != nil {
		println(err.Error())
	}
	if err := u.Configure(machine.UARTConfig{TX: machine.GPIO0, RX: machine.NoPin}); err != nil {
		println(err.Error())
	}

	i2c := &machine.I2C{Bus: rp.I2C0}
	if err := i2c.Configure(machine.I2CConfig{}); err == nil {
		println("expected i2c error")
	}
	i2c1 := &machine.I2C{Bus: rp.I2C1}
	if err := i2c1.Configure(machine.I2CConfig{SCL: machine.NoPin, SDA: machine.GPIO2}); err == nil {
		println("expected i2c1 nopin scl error")
	}
	if err := i2c1.Configure(machine.I2CConfig{SCL: machine.GPIO3, SDA: machine.NoPin}); err == nil {
		println("expected i2c1 nopin sda error")
	}
	spi := &machine.SPI{Bus: rp.SPI0}
	if err := spi.Configure(machine.SPIConfig{}); err == nil {
		println("expected spi error")
	}

	if _, err := (machine.ADC{Pin: machine.GPIO26}).GetADCChannel(); err != nil {
		println("adc0 check:", err.Error())
	}

	var p *machine.PWM = machine.NewPWM(0)
	if p == nil {
		println("pwm0 nil")
	}
	if machine.NewPWM(7) == nil {
		println("pwm7 nil")
	}
	if machine.NewPWM(12) != nil {
		println("pwm12 non-nil")
	}

	println(machine.CPUFrequency())
}
`

	boardXTALSrc := `package board

import _ "unsafe"

//go:linkname boardXTAL machine.boardXTAL
func boardXTAL() uint32 {
	return 12_000_000
}
`

	boardUSBSrc := `//go:build serial.usb

package board

import _ "unsafe"

//go:linkname usbIdentity machine.boardUSBIdentity
func usbIdentity() (uint16, uint16, string, string) {
	return 0x1209, 0x0001, "TestMfg", "TestProd"
}
`

	writeModule := func(t *testing.T, baseTarget, src string) (dir, targetPath string) {
		t.Helper()
		dir = t.TempDir()
		ldflags := ""
		if baseTarget == "rp2040" {
			ldflags = "\n\t\"ldflags\": [\"--defsym=__flash_size=2M\"],"
		}
		targetJSON := `{
	"inherits": [` + `"` + baseTarget + `"` + `],` + ldflags + `
	"board-package": "example.com/rp2bp/board"
}
`
		files := map[string]string{
			"go.mod":            "module example.com/rp2bp\n\ngo 1.25.0\n",
			"go.work":           "go 1.25.0\n\nuse .\n",
			"target/board.json": targetJSON,
			"board/board.go":    boardXTALSrc,
			"board/usb.go":      boardUSBSrc,
			"app/main.go":       src,
		}
		for name, content := range files {
			full := filepath.Join(dir, filepath.FromSlash(name))
			if err := os.MkdirAll(filepath.Dir(full), 0o700); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(full, []byte(content), 0o600); err != nil {
				t.Fatal(err)
			}
		}
		return dir, filepath.Join(dir, "target", "board.json")
	}

	baseTargets := []struct {
		name        string
		inRangePWM  string
		outRangePWM string
		adc0Pin     string
		extraOmit   []string
	}{
		{name: "rp2040", inRangePWM: "7", outRangePWM: "8", adc0Pin: "GPIO26"},
		{name: "rp2350", inRangePWM: "11", outRangePWM: "12", adc0Pin: "GPIO26"},
		{name: "rp2350b", inRangePWM: "11", outRangePWM: "12", adc0Pin: "GPIO40", extraOmit: []string{"PWM8", "ADC7"}},
	}

	for _, bt := range baseTargets {
		t.Run(bt.name, func(t *testing.T) {
			t.Parallel()
			pwmSrc := strings.Replace(
				mainSrc,
				"machine.NewPWM(7)",
				"machine.NewPWM("+bt.inRangePWM+")",
				1,
			)
			pwmSrc = strings.Replace(
				pwmSrc,
				"machine.NewPWM(12)",
				"machine.NewPWM("+bt.outRangePWM+")",
				1,
			)
			pwmSrc = strings.Replace(
				pwmSrc,
				"machine.GPIO26",
				"machine."+bt.adc0Pin,
				1,
			)
			dir, targetPath := writeModule(t, bt.name, pwmSrc)

			for _, opt := range []string{"0", "1", "z"} {
				t.Run("pllTable/opt="+opt, func(t *testing.T) {
					t.Parallel()
					outLL := filepath.Join(t.TempDir(), "out.ll")
					options := &compileopts.Options{
						Target:                  targetPath,
						Opt:                     opt,
						Serial:                  "uart",
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
						"./app",
						outLL,
						t.TempDir(),
						config,
					); err != nil {
						t.Fatalf("Build(opt=%s) failed: %v", opt, err)
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

					checkRP2BoardPackageIR(t, mod, opt)
				})
			}

			for _, serial := range []string{"uart", "usb", "none"} {
				t.Run("serial="+serial, func(t *testing.T) {
					t.Parallel()
					outELF := filepath.Join(t.TempDir(), "out.elf")
					options := &compileopts.Options{
						Target:                  targetPath,
						Opt:                     "z",
						Serial:                  serial,
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
						"./app",
						outELF,
						t.TempDir(),
						config,
					)
					if err != nil {
						t.Fatalf("Build(serial=%s) failed: %v", serial, err)
					}
					info, err := os.Stat(result.Binary)
					if err != nil || info.Size() == 0 {
						t.Fatalf("invalid ELF %q: stat=%v err=%v", result.Binary, info, err)
					}
				})
			}

			t.Run("omitted globals undefined", func(t *testing.T) {
				t.Parallel()
				omitted := append([]string{
					"UART0",
					"UART1",
					"I2C0",
					"I2C1",
					"SPI0",
					"SPI1",
					"PWM0",
					"ADC0",
					"DefaultUART",
				}, bt.extraOmit...)
				for _, sym := range omitted {
					t.Run(sym, func(t *testing.T) {
						t.Parallel()
						badSrc := "package main\n\nimport \"machine\"\n\nfunc main() {\n\t_ = machine." + sym + "\n}\n"
						badDir, badTarget := writeModule(t, bt.name, badSrc)
						options := &compileopts.Options{
							Target:                  badTarget,
							Opt:                     "z",
							Serial:                  "uart",
							Directory:               badDir,
							Semaphore:               make(chan struct{}, 2),
							InterpTimeout:           60 * time.Second,
							InterpMaxLoopIterations: interp.DefaultMaxInterpBlockEntries,
						}
						config, err := NewConfig(options)
						if err != nil {
							t.Fatal(err)
						}
						_, err = Build(
							"./app",
							filepath.Join(t.TempDir(), "out.ll"),
							t.TempDir(),
							config,
						)
						if err == nil {
							t.Fatalf("expected error referencing omitted machine.%s under boardpackage", sym)
						}
						if !strings.Contains(err.Error(), "undefined: machine."+sym) {
							t.Errorf("unexpected error for machine.%s: %v", sym, err)
						}
					})
				}
			})
		})
	}
}

func checkRP2BoardPackageIR(t *testing.T, mod llvm.Module, opt string) {
	t.Helper()

	if fn := mod.NamedFunction("machine.boardUSBIdentity"); !fn.IsNil() {
		t.Errorf("expected unused machine.boardUSBIdentity to be stripped under -serial=uart")
	}

	g := mod.NamedGlobal("machine.pllTable")
	checkTargets := map[llvm.Value]string{}
	if g.IsNil() {
		if opt == "0" {
			t.Errorf("expected global machine.pllTable at -opt=0")
		}
	} else {
		if g.IsDeclaration() {
			t.Errorf("machine.pllTable is unexpectedly external")
		}
		initVal := g.Initializer()
		if initVal.IsNil() || initVal.IsNull() {
			t.Errorf("machine.pllTable has zero/nil initializer instead of interp-folded constants")
		}
		checkTargets[g] = "machine.pllTable"
	}

	for _, name := range []string{"machine.clocks", "machine.xosc", "machine.pllSys", "machine.pllUSB", "machine.timer", "machine.resets"} {
		mg := mod.NamedGlobal(name)
		if mg.IsNil() {
			if opt == "0" {
				t.Errorf("expected global %s at -opt=0", name)
			}
			continue
		}
		initVal := mg.Initializer()
		if initVal.IsNil() || initVal.IsNull() {
			t.Errorf("global %s has zero/nil initializer", name)
		}
		checkTargets[mg] = name
	}

	visited := map[llvm.Value]bool{}
	for _, fnName := range []string{"runtime.initAll", "machine.init"} {
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
