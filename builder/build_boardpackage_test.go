package builder

import (
	"bytes"
	"debug/elf"
	"fmt"
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

func TestBoardPackage(t *testing.T) {
	t.Parallel()

	fixtureDir := copyBoardPkgFixture(t)

	boards := []struct {
		name       string
		targetPath string
		usbProduct string
		usbPID     int64
	}{
		{
			name:       "custom2040",
			targetPath: filepath.Join(fixtureDir, "boards", "custom2040", "target.json"),
			usbProduct: "Custom2040",
			usbPID:     0x4242,
		},
		{
			name:       "custom2350",
			targetPath: filepath.Join(fixtureDir, "boards", "custom2350", "target.json"),
			usbProduct: "Custom2350",
			usbPID:     0x4243,
		},
	}
	apps := []string{"./cmd/minimal", "./cmd/full"}
	serials := []string{"uart", "usb"}
	opts := []string{"0", "1", "z"}

	for _, board := range boards {
		for _, app := range apps {
			for _, serial := range serials {
				for _, opt := range opts {
					sub := strings.Join([]string{board.name, filepath.Base(app), "serial=" + serial, "opt=" + opt}, "/")
					t.Run(sub, func(t *testing.T) {
						t.Parallel()

						outELF := filepath.Join(t.TempDir(), "firmware.elf")
						result := buildBoardPkgTarget(t, boardPkgBuildSpec{
							dir:     fixtureDir,
							pkg:     app,
							outPath: outELF,
							target:  board.targetPath,
							serial:  serial,
							opt:     opt,
						})
						info, err := os.Stat(result.Binary)
						if err != nil || info.Size() == 0 {
							t.Fatalf("missing or empty ELF %q: stat=%v err=%v", result.Binary, info, err)
						}

						if app != "./cmd/minimal" {
							return
						}
						if opt == "0" {
							verifyBoardPkgMinimalIR(
								t,
								fixtureDir,
								board.targetPath,
								serial,
							)
						}
						if opt == "1" || opt == "z" {
							verifyBoardPkgMinimalELF(t, result.Binary)
						}
						if serial == "usb" && opt == "z" {
							verifyBoardPkgUSBIdentity(t, usbCheckSpec{
								fixtureDir:  fixtureDir,
								targetPath:  board.targetPath,
								elfPath:     result.Binary,
								wantProduct: board.usbProduct,
								wantPID:     board.usbPID,
							})
						}
					})
				}
			}
		}
	}

	t.Run("nousb2040", func(t *testing.T) {
		t.Parallel()
		noUSBTarget := filepath.Join(fixtureDir, "boards", "nousb2040", "target.json")

		t.Run("serial=uart succeeds", func(t *testing.T) {
			t.Parallel()
			outELF := filepath.Join(t.TempDir(), "nousb.elf")
			result := buildBoardPkgTarget(t, boardPkgBuildSpec{
				dir:     fixtureDir,
				pkg:     "./cmd/minimal",
				outPath: outELF,
				target:  noUSBTarget,
				serial:  "uart",
				opt:     "z",
			})
			if info, err := os.Stat(result.Binary); err != nil || info.Size() == 0 {
				t.Fatalf("expected valid ELF for nousb2040 under serial=uart: stat=%v err=%v", info, err)
			}
		})

		t.Run("serial=usb fails link", func(t *testing.T) {
			t.Parallel()
			config := newBoardPkgConfig(t, boardPkgBuildSpec{
				dir:    fixtureDir,
				target: noUSBTarget,
				serial: "usb",
				opt:    "z",
			})
			_, err := Build(
				"./cmd/minimal",
				filepath.Join(t.TempDir(), "nousb.elf"),
				t.TempDir(),
				config,
			)
			if err == nil {
				t.Fatal("expected link error when boardUSBIdentity is missing under -serial=usb")
			}
			if !strings.Contains(err.Error(), "boardUSBIdentity") {
				t.Errorf("unexpected link error: %v", err)
			}
		})
	})

	t.Run("negative", func(t *testing.T) {
		t.Parallel()

		t.Run("missing boardXTAL fails link", func(t *testing.T) {
			t.Parallel()
			dir, targetPath := writeTempBoardModule(t, map[string]string{
				"board/board.go": "package board\n\nvar Touch int\n",
				"app/main.go":    "package main\n\nimport \"machine\"\n\nfunc main() {\n\t_ = machine.CPUFrequency()\n}\n",
			})
			config := newBoardPkgConfig(t, boardPkgBuildSpec{
				dir:    dir,
				target: targetPath,
				serial: "uart",
				opt:    "z",
			})
			_, err := Build(
				"./app",
				filepath.Join(t.TempDir(), "out.elf"),
				t.TempDir(),
				config,
			)
			if err == nil {
				t.Fatal("expected link failure when machine.boardXTAL is missing")
			}
			if !strings.Contains(err.Error(), "boardXTAL") {
				t.Errorf("unexpected link error: %v", err)
			}
		})

		t.Run("signature mismatch boardXTAL", func(t *testing.T) {
			t.Parallel()
			dir, targetPath := writeTempBoardModule(t, map[string]string{
				"board/board.go": `package board

import _ "unsafe"

//go:linkname badXTAL machine.boardXTAL
func badXTAL() int {
	return 12_000_000
}
`,
				"app/main.go": "package main\n\nimport \"machine\"\n\nfunc main() {\n\t_ = machine.CPUFrequency()\n}\n",
			})
			config := newBoardPkgConfig(t, boardPkgBuildSpec{
				dir:    dir,
				target: targetPath,
				serial: "uart",
				opt:    "z",
			})
			_, err := Build(
				"./app",
				filepath.Join(t.TempDir(), "out.ll"),
				t.TempDir(),
				config,
			)
			if err == nil {
				t.Fatal("expected signature mismatch error on machine.boardXTAL")
			}
			if !strings.Contains(err.Error(), "machine.boardXTAL") || !strings.Contains(err.Error(), "does not match") {
				t.Errorf("unexpected error: %v", err)
			}
		})

		t.Run("signature mismatch boardInit0", func(t *testing.T) {
			t.Parallel()
			dir, targetPath := writeTempBoardModule(t, map[string]string{
				"board/board.go": `package board

import _ "unsafe"

//go:linkname boardXTAL machine.boardXTAL
func boardXTAL() uint32 {
	return 12_000_000
}

//go:linkname badInit0 runtime.boardInit0
func badInit0(arg int) {
	_ = arg
}
`,
				"app/main.go": "package main\n\nimport \"machine\"\n\nfunc main() {\n\t_ = machine.CPUFrequency()\n}\n",
			})
			config := newBoardPkgConfig(t, boardPkgBuildSpec{
				dir:    dir,
				target: targetPath,
				serial: "uart",
				opt:    "z",
			})
			_, err := Build(
				"./app",
				filepath.Join(t.TempDir(), "out.ll"),
				t.TempDir(),
				config,
			)
			if err == nil {
				t.Fatal("expected signature mismatch error on runtime.boardInit0")
			}
			if !strings.Contains(err.Error(), "runtime.boardInit0") || !strings.Contains(err.Error(), "does not match") {
				t.Errorf("unexpected error: %v", err)
			}
		})

		t.Run("duplicate board packages fail", func(t *testing.T) {
			t.Parallel()
			dir, targetPath := writeTempBoardModule(t, map[string]string{
				"board/board.go": `package board

import _ "unsafe"

//go:linkname boardXTAL machine.boardXTAL
func boardXTAL() uint32 {
	return 12_000_000
}

//go:linkname boardInit1 runtime.boardInit1
func boardInit1() {}
`,
				"board2/board.go": `package board2

import _ "unsafe"

//go:linkname boardXTAL machine.boardXTAL
func boardXTAL() uint32 {
	return 12_000_000
}

//go:linkname boardInit1 runtime.boardInit1
func boardInit1() {}
`,
				"app/main.go": `package main

import (
	"machine"

	_ "example.com/tempboard/board2"
)

func main() {
	_ = machine.CPUFrequency()
}
`,
			})
			config := newBoardPkgConfig(t, boardPkgBuildSpec{
				dir:    dir,
				target: targetPath,
				serial: "uart",
				opt:    "z",
			})
			_, err := Build(
				"./app",
				filepath.Join(t.TempDir(), "out.ll"),
				t.TempDir(),
				config,
			)
			if err == nil {
				t.Fatal("expected build failure when two board packages define conflicting hooks")
			}
			if !strings.Contains(err.Error(), "duplicate strong override") {
				t.Errorf("unexpected duplicate board package error: %v", err)
			}
		})

		t.Run("noheap rejects escaping alloc in boardInit0", func(t *testing.T) {
			t.Parallel()
			for _, opt := range []string{"0", "z"} {
				t.Run("opt="+opt, func(t *testing.T) {
					t.Parallel()
					dir, targetPath := writeTempBoardModule(t, map[string]string{
						"board/board.go": `package board

import _ "unsafe"

var Leaked []*byte
var AllocCount = 4

//go:linkname boardXTAL machine.boardXTAL
func boardXTAL() uint32 {
	return 12_000_000
}

//go:linkname badBoardInit0 runtime.boardInit0
func badBoardInit0() {
	Leaked = make([]*byte, AllocCount)
}
`,
						"app/main.go": `package main

import (
	"machine"

	board "example.com/tempboard/board"
)

func main() {
	_ = machine.CPUFrequency()
	if len(board.Leaked) > 0 {
		println(len(board.Leaked), board.Leaked[0])
	}
}
`,
					})
					config := newBoardPkgConfig(t, boardPkgBuildSpec{
						dir:    dir,
						target: targetPath,
						serial: "uart",
						opt:    opt,
					})
					_, err := Build(
						"./app",
						filepath.Join(t.TempDir(), "out.ll"),
						t.TempDir(),
						config,
					)
					if err == nil {
						t.Fatalf("expected noheap rejection in boardInit0 at -opt=%s, got nil", opt)
					}
					if !strings.Contains(err.Error(), "heap allocation is not allowed in pre-heap initialization") {
						t.Errorf("unexpected error at -opt=%s: %v", opt, err)
					}
				})
			}
		})
	})
}

type boardPkgBuildSpec struct {
	dir     string
	pkg     string
	outPath string
	target  string
	serial  string
	opt     string
}

func newBoardPkgConfig(t *testing.T, spec boardPkgBuildSpec) *compileopts.Config {
	t.Helper()
	options := &compileopts.Options{
		Target:                  spec.target,
		Opt:                     spec.opt,
		Serial:                  spec.serial,
		Debug:                   true,
		Directory:               spec.dir,
		Semaphore:               make(chan struct{}, 2),
		InterpTimeout:           60 * time.Second,
		InterpMaxLoopIterations: interp.DefaultMaxInterpBlockEntries,
	}
	config, err := NewConfig(options)
	if err != nil {
		t.Fatal(err)
	}
	return config
}

func buildBoardPkgTarget(t *testing.T, spec boardPkgBuildSpec) *BuildResult {
	t.Helper()
	config := newBoardPkgConfig(t, spec)
	result, err := Build(
		spec.pkg,
		spec.outPath,
		t.TempDir(),
		config,
	)
	if err != nil {
		t.Fatalf("Build(pkg=%q, target=%q, serial=%q, opt=%q) failed: %v", spec.pkg, spec.target, spec.serial, spec.opt, err)
	}
	return result
}

func copyBoardPkgFixture(t *testing.T) string {
	t.Helper()
	fixtureDir := t.TempDir()
	srcDir := filepath.Join("..", "testdata", "boardpackage")
	if err := os.CopyFS(fixtureDir, os.DirFS(srcDir)); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(fixtureDir, "go.work"), []byte("go 1.25.0\n\nuse .\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	return fixtureDir
}

func writeTempBoardModule(t *testing.T, extraFiles map[string]string) (dir, targetPath string) {
	t.Helper()
	dir = t.TempDir()
	files := map[string]string{
		"go.mod":  "module example.com/tempboard\n\ngo 1.25.0\n",
		"go.work": "go 1.25.0\n\nuse .\n",
		"target/board.json": `{
    "inherits": ["rp2040"],
    "ldflags": ["--defsym=__flash_size=2M"],
    "board-package": "example.com/tempboard/board"
}
`,
	}
	for name, content := range extraFiles {
		files[name] = content
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

func verifyBoardPkgMinimalIR(t *testing.T, fixtureDir, targetPath, serial string) {
	t.Helper()
	outLL := filepath.Join(t.TempDir(), "minimal.ll")
	buildBoardPkgTarget(t, boardPkgBuildSpec{
		dir:     fixtureDir,
		pkg:     "./cmd/minimal",
		outPath: outLL,
		target:  targetPath,
		serial:  serial,
		opt:     "0",
	})

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

	runCalls := directCallsInNamedFunc(t, mod, "runtime.run")
	assertCallOrder(
		t,
		"runtime.run",
		runCalls,
		[]string{"runtime.hwInit0", "runtime.initRand", "runtime.initHeap"},
	)

	hw0Calls := directCallsInNamedFunc(t, mod, "runtime.hwInit0")
	assertCallOrder(
		t,
		"runtime.hwInit0",
		hw0Calls,
		[]string{"runtime.socInit0", "runtime.boardInit0"},
	)

	_, goroutineCalls := findRunGoroutineCalls(t, mod)
	assertCallOrder(
		t,
		"runtime.run$1",
		goroutineCalls,
		[]string{"runtime.hwInit1", "runtime.initAll", "main.main"},
	)

	hw1Calls := directCallsInNamedFunc(t, mod, "runtime.hwInit1")
	if !slices.Contains(hw1Calls, "runtime.boardInit1") {
		t.Errorf("runtime.hwInit1 calls %v, missing runtime.boardInit1", hw1Calls)
	}

	// interp must keep runtime.putchar loading machine.Serial at runtime instead of folding nil.
	putchar := mod.NamedFunction("runtime.putchar")
	if putchar.IsNil() || putchar.IsDeclaration() {
		t.Fatal("missing defined runtime.putchar in IR")
	}
	if !strings.Contains(putchar.String(), "@machine.Serial") {
		t.Errorf("runtime.putchar no longer references @machine.Serial (serial=%s):\n%s", serial, putchar.String())
	}
}

func verifyBoardPkgMinimalELF(t *testing.T, elfPath string) {
	t.Helper()
	f, err := elf.Open(elfPath)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	syms, err := f.Symbols()
	if err != nil {
		t.Fatal(err)
	}
	foundFlashEnd := false
	for _, sym := range syms {
		if sym.Name == "__flash_data_end" {
			foundFlashEnd = true
			const wantFlashEnd = 0x10400000
			if sym.Value != wantFlashEnd {
				t.Errorf("__flash_data_end = 0x%x, want 0x%x in %s", sym.Value, wantFlashEnd, elfPath)
			}
		}
		for _, forbidden := range []string{
			"machine.(*I2C).Configure",
			"machine.(*SPI).Configure",
			"machine.UART1",
			"machine._UART1",
			"machine._I2C0",
			"machine._SPI0",
		} {
			if strings.Contains(sym.Name, forbidden) {
				t.Errorf("unused symbol %q unexpectedly retained in %s", sym.Name, elfPath)
			}
		}
	}
	if !foundFlashEnd {
		t.Errorf("missing __flash_data_end symbol in %s", elfPath)
	}
}

type usbCheckSpec struct {
	fixtureDir  string
	targetPath  string
	elfPath     string
	wantProduct string
	wantPID     int64
}

func verifyBoardPkgUSBIdentity(t *testing.T, spec usbCheckSpec) {
	t.Helper()
	f, err := elf.Open(spec.elfPath)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	blob := []byte{}
	for _, sec := range f.Sections {
		if sec.Type != elf.SHT_PROGBITS {
			continue
		}
		data, err := sec.Data()
		if err != nil {
			t.Fatal(err)
		}
		blob = append(blob, data...)
	}
	for _, wantStr := range []string{"CustomHW", spec.wantProduct} {
		if !bytes.Contains(blob, []byte(wantStr)) {
			t.Errorf("ELF %s missing custom USB string %q", spec.elfPath, wantStr)
		}
	}

	// Verify VID (0x1209) and PID in optimized IR passed to usbDescriptor.Configure.
	outLL := filepath.Join(t.TempDir(), "usb.ll")
	buildBoardPkgTarget(t, boardPkgBuildSpec{
		dir:     spec.fixtureDir,
		pkg:     "./cmd/minimal",
		outPath: outLL,
		target:  spec.targetPath,
		serial:  "usb",
		opt:     "z",
	})
	irData, err := os.ReadFile(outLL)
	if err != nil {
		t.Fatal(err)
	}
	ir := string(irData)
	wantVIDArg := fmt.Sprintf("i16 %d", 0x1209)
	wantPIDArg := fmt.Sprintf("i16 %d", spec.wantPID)
	if !strings.Contains(ir, wantVIDArg) || !strings.Contains(ir, wantPIDArg) {
		t.Errorf("optimized USB IR missing custom VID (%s) or PID (%s)", wantVIDArg, wantPIDArg)
	}
}
