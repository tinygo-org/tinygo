package builder

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/tinygo-org/tinygo/compileopts"
	"github.com/tinygo-org/tinygo/interp"
	"tinygo.org/x/go-llvm"
)

type bootHooksBuildSpec struct {
	dir       string
	outPath   string
	target    string
	goos      string
	goarch    string
	scheduler string
	opt       string
	tags      []string
}

func TestBootHooks(t *testing.T) {
	t.Parallel()

	hooksMainSrc := `package main

import _ "unsafe"

var stageLog [8]uint32
var stageCount uint32

//go:noinline
func recordStage(stage uint32) {
	switch stageCount {
	case 0:
		stageLog[0] = stage
	case 1:
		stageLog[1] = stage
	case 2:
		stageLog[2] = stage
	case 3:
		stageLog[3] = stage
	case 4:
		stageLog[4] = stage
	case 5:
		stageLog[5] = stage
	default:
		return
	}
	stageCount++
}

//go:linkname testSocInit0 runtime.socInit0
func testSocInit0() {
	recordStage(1)
}

//go:linkname testBoardInit0 runtime.boardInit0
func testBoardInit0() {
	recordStage(2)
}

//go:linkname testSocInit1 runtime.socInit1
func testSocInit1() {
	recordStage(3)
}

//go:linkname testBoardInit1 runtime.boardInit1
func testBoardInit1() {
	recordStage(4)
}

func init() {
	recordStage(5)
}

func main() {
	recordStage(6)
	for i := uint32(0); i < stageCount; i++ {
		if i > 0 {
			print(",")
		}
		print(stageLog[i])
	}
	println()
}
`

	plainMainSrc := `package main

func main() {}
`

	t.Run("scheduler ordering with socinit", func(t *testing.T) {
		t.Parallel()

		cases := []struct {
			name          string
			target        string
			goos          string
			goarch        string
			scheduler     string
			goroutineInit bool
			checkCores    bool
		}{
			{
				name:          "tasks",
				target:        "cortex-m-qemu",
				scheduler:     "tasks",
				goroutineInit: true,
			},
			{
				name:      "none",
				target:    "cortex-m-qemu",
				scheduler: "none",
			},
			{
				name:          "cores",
				target:        "pico",
				scheduler:     "cores",
				goroutineInit: true,
				checkCores:    true,
			},
			{
				name:          "asyncify",
				target:        "wasm",
				scheduler:     "asyncify",
				goroutineInit: true,
			},
			{
				name:      "threads",
				goos:      "linux",
				goarch:    "amd64",
				scheduler: "threads",
			},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()
				dir := writeBootHooksTestModule(t, hooksMainSrc)
				outLL := filepath.Join(t.TempDir(), "out.ll")
				buildBootHooksTarget(t, bootHooksBuildSpec{
					dir:       dir,
					outPath:   outLL,
					target:    tc.target,
					goos:      tc.goos,
					goarch:    tc.goarch,
					scheduler: tc.scheduler,
					opt:       "0",
					tags:      []string{"socinit"},
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

				hw0Calls := directCallsInNamedFunc(t, mod, "runtime.hwInit0")
				assertCallOrder(
					t,
					"runtime.hwInit0",
					hw0Calls,
					[]string{"runtime.socInit0", "runtime.boardInit0"},
				)

				hw1Calls := directCallsInNamedFunc(t, mod, "runtime.hwInit1")
				assertCallOrder(
					t,
					"runtime.hwInit1",
					hw1Calls,
					[]string{"runtime.socInit1", "runtime.boardInit1"},
				)

				runCalls := directCallsInNamedFunc(t, mod, "runtime.run")
				assertCallOrder(
					t,
					"runtime.run",
					runCalls,
					[]string{"runtime.hwInit0", "runtime.initRand", "runtime.initHeap"},
				)

				if tc.goroutineInit {
					initFnName, goroutineCalls := findRunGoroutineCalls(t, mod)
					assertCallOrder(
						t,
						initFnName,
						goroutineCalls,
						[]string{"runtime.hwInit1", "runtime.initAll", "main.main"},
					)
					if tc.checkCores {
						assertCallOrder(
							t,
							initFnName,
							goroutineCalls,
							[]string{"runtime.initAll", "runtime.startSecondaryCores", "main.main"},
						)
					}
					return
				}

				assertCallOrder(
					t,
					"runtime.run",
					runCalls,
					[]string{"runtime.initHeap", "runtime.hwInit1", "runtime.initAll", "main.main"},
				)
			})
		}
	})

	t.Run("threads runtime execution order", func(t *testing.T) {
		t.Parallel()
		if runtime.GOOS != "linux" && runtime.GOOS != "darwin" {
			t.Skipf("scheduler=threads is unsupported on host %s/%s", runtime.GOOS, runtime.GOARCH)
		}
		dir := writeBootHooksTestModule(t, hooksMainSrc)
		result := buildBootHooksTarget(t, bootHooksBuildSpec{
			dir:       dir,
			outPath:   filepath.Join(t.TempDir(), "hooks-test"),
			goos:      runtime.GOOS,
			goarch:    runtime.GOARCH,
			scheduler: "threads",
			opt:       "1",
			tags:      []string{"socinit"},
		})
		out, err := exec.Command(result.Binary).CombinedOutput()
		if err != nil {
			t.Fatalf("running compiled binary failed: %v\n%s", err, out)
		}
		if got := strings.TrimSpace(string(out)); got != "1,2,3,4,5,6" {
			t.Fatalf("runtime stage sequence = %q, want %q", got, "1,2,3,4,5,6")
		}
	})

	t.Run("noheap rejects escaping alloc in boardInit0", func(t *testing.T) {
		t.Parallel()
		for _, tc := range []struct {
			name     string
			src      string
			wantFunc string
		}{
			{
				name: "direct",
				src: `package main

import _ "unsafe"

var leaked []*byte
var allocCount = 4

//go:linkname badBoardInit0 runtime.boardInit0
func badBoardInit0() {
	leaked = make([]*byte, allocCount)
}

func main() {
	if len(leaked) > 0 {
		println(len(leaked), leaked[0])
	}
}
`,
				wantFunc: "runtime.boardInit0",
			},
			{
				name: "transitive",
				src: `package main

import _ "unsafe"

var leaked []*byte
var allocCount = 4

//go:noinline
func allocHelper() []*byte {
	return make([]*byte, allocCount)
}

//go:linkname badBoardInit0 runtime.boardInit0
func badBoardInit0() {
	leaked = allocHelper()
}

func main() {
	if len(leaked) > 0 {
		println(len(leaked), leaked[0])
	}
}
`,
				wantFunc: "main.allocHelper",
			},
		} {
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()
				dir := writeBootHooksTestModule(t, tc.src)
				outLL := filepath.Join(t.TempDir(), "bad.ll")
				config := newBootHooksConfig(t, bootHooksBuildSpec{
					dir:       dir,
					target:    "cortex-m-qemu",
					scheduler: "tasks",
					opt:       "1",
					tags:      []string{"socinit"},
				})
				_, err := Build(
					".",
					outLL,
					t.TempDir(),
					config,
				)
				if err == nil {
					t.Fatal("expected Build to reject heap allocation in boardInit0, got nil")
				}
				msg := err.Error()
				wantMsg := "heap allocation is not allowed in pre-heap initialization"
				if !strings.Contains(msg, wantMsg) || !strings.Contains(msg, tc.wantFunc) {
					t.Errorf("unexpected error message: %v", err)
				}
			})
		}
	})

	t.Run("socinit without overrides strips weak no-ops at opt=1", func(t *testing.T) {
		t.Parallel()
		dir := writeBootHooksTestModule(t, plainMainSrc)
		outLL := filepath.Join(t.TempDir(), "stripped.ll")
		buildBootHooksTarget(t, bootHooksBuildSpec{
			dir:       dir,
			outPath:   outLL,
			target:    "cortex-m-qemu",
			scheduler: "tasks",
			opt:       "1",
			tags:      []string{"socinit"},
		})
		assertNoBootHookSymbols(t, outLL)
	})

	t.Run("without socinit strips boot hook symbols", func(t *testing.T) {
		t.Parallel()
		for _, opt := range []string{"1", "z"} {
			t.Run("opt="+opt, func(t *testing.T) {
				t.Parallel()
				dir := writeBootHooksTestModule(t, plainMainSrc)
				outLL := filepath.Join(t.TempDir(), "nosocinit.ll")
				buildBootHooksTarget(t, bootHooksBuildSpec{
					dir:       dir,
					outPath:   outLL,
					target:    "cortex-m-qemu",
					scheduler: "tasks",
					opt:       opt,
				})
				assertNoBootHookSymbols(t, outLL)
			})
		}
	})
}

func writeBootHooksTestModule(t *testing.T, mainSrc string) string {
	t.Helper()
	dir := t.TempDir()
	files := map[string]string{
		"go.mod":  "module example.com/boothooks\n\ngo 1.25.0\n",
		"go.work": "go 1.25.0\n\nuse .\n",
		"main.go": mainSrc,
	}
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

func newBootHooksConfig(t *testing.T, spec bootHooksBuildSpec) *compileopts.Config {
	t.Helper()
	options := &compileopts.Options{
		Target:                  spec.target,
		Scheduler:               spec.scheduler,
		Opt:                     spec.opt,
		Tags:                    spec.tags,
		Directory:               spec.dir,
		Semaphore:               make(chan struct{}, 2),
		InterpTimeout:           60 * time.Second,
		InterpMaxLoopIterations: interp.DefaultMaxInterpBlockEntries,
	}
	if spec.target == "" {
		options.GOOS = spec.goos
		options.GOARCH = spec.goarch
	}
	config, err := NewConfig(options)
	if err != nil {
		t.Fatal(err)
	}
	return config
}

func buildBootHooksTarget(t *testing.T, spec bootHooksBuildSpec) *BuildResult {
	t.Helper()
	config := newBootHooksConfig(t, spec)
	result, err := Build(
		".",
		spec.outPath,
		t.TempDir(),
		config,
	)
	if err != nil {
		t.Fatalf("Build(target=%q, scheduler=%q, opt=%q) failed: %v", spec.target, spec.scheduler, spec.opt, err)
	}
	return result
}

func directCallsInFunc(fn llvm.Value) []string {
	calls := []string{}
	for bb := fn.FirstBasicBlock(); !bb.IsNil(); bb = llvm.NextBasicBlock(bb) {
		for inst := bb.FirstInstruction(); !inst.IsNil(); inst = llvm.NextInstruction(inst) {
			op := inst.InstructionOpcode()
			if op != llvm.Call && op != llvm.Invoke {
				continue
			}
			callee := inst.CalledValue()
			if callee.IsNil() || callee.IsAFunction().IsNil() {
				continue
			}
			calls = append(calls, callee.Name())
		}
	}
	return calls
}

func directCallsInNamedFunc(t *testing.T, mod llvm.Module, name string) []string {
	t.Helper()
	fn := mod.NamedFunction(name)
	if fn.IsNil() || fn.IsDeclaration() {
		t.Fatalf("expected defined function %q in IR", name)
	}
	return directCallsInFunc(fn)
}

func findRunGoroutineCalls(t *testing.T, mod llvm.Module) (string, []string) {
	t.Helper()
	for fn := mod.FirstFunction(); !fn.IsNil(); fn = llvm.NextFunction(fn) {
		if fn.IsDeclaration() {
			continue
		}
		name := fn.Name()
		if !strings.HasPrefix(name, "runtime.run$") {
			continue
		}
		calls := directCallsInFunc(fn)
		if slices.Contains(calls, "runtime.initAll") {
			return name, calls
		}
	}
	t.Fatal("could not find runtime.run goroutine function calling runtime.initAll")
	return "", nil
}

func assertCallOrder(t *testing.T, fnName string, calls, expected []string) {
	t.Helper()
	pos := -1
	for _, target := range expected {
		found := -1
		for i := pos + 1; i < len(calls); i++ {
			if calls[i] == target {
				found = i
				break
			}
		}
		if found < 0 {
			t.Fatalf("%s calls %v missing ordered sequence %v (failed at %q)", fnName, calls, expected, target)
		}
		pos = found
	}
}

func assertNoBootHookSymbols(t *testing.T, irPath string) {
	t.Helper()
	data, err := os.ReadFile(irPath)
	if err != nil {
		t.Fatal(err)
	}
	ir := string(data)
	for _, sym := range []string{
		"runtime.hwInit0",
		"runtime.hwInit1",
		"runtime.socInit0",
		"runtime.socInit1",
		"runtime.boardInit0",
		"runtime.boardInit1",
	} {
		if strings.Contains(ir, sym) {
			t.Errorf("IR %s unexpectedly contains symbol %q", irPath, sym)
		}
	}
}
