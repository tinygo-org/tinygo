package main

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/aykevl/go-wasm"
	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
	"github.com/tetratelabs/wazero/sys"
	"github.com/tinygo-org/tinygo/builder"
)

// Test WebAssembly files for certain properties.
func TestWebAssembly(t *testing.T) {
	t.Parallel()
	type testCase struct {
		name          string
		target        string
		panicStrategy string
		imports       []string
	}
	for _, tc := range []testCase{
		// Test whether there really are no imports when using -panic=trap. This
		// tests the bugfix for https://github.com/tinygo-org/tinygo/issues/4161.
		{name: "panic-default", target: "wasip1", imports: []string{"wasi_snapshot_preview1.fd_write", "wasi_snapshot_preview1.proc_exit", "wasi_snapshot_preview1.random_get"}},
		{name: "panic-trap", target: "wasm-unknown", panicStrategy: "trap", imports: []string{}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			tmpdir := t.TempDir()
			options := optionsFromTarget(tc.target, sema)
			options.PanicStrategy = tc.panicStrategy
			config, err := builder.NewConfig(&options)
			if err != nil {
				t.Fatal(err)
			}

			result, err := builder.Build("testdata/trivialpanic.go", ".wasm", tmpdir, config)
			if err != nil {
				t.Fatal("failed to build binary:", err)
			}
			f, err := os.Open(result.Binary)
			if err != nil {
				t.Fatal("could not open output binary:", err)
			}
			defer f.Close()
			module, err := wasm.Parse(f)
			if err != nil {
				t.Fatal("could not parse output binary:", err)
			}

			// Test the list of imports.
			if tc.imports != nil {
				var imports []string
				for _, section := range module.Sections {
					switch section := section.(type) {
					case *wasm.SectionImport:
						for _, symbol := range section.Entries {
							imports = append(imports, symbol.Module+"."+symbol.Field)
						}
					}
				}
				if !slices.Equal(imports, tc.imports) {
					t.Errorf("import list not as expected!\nexpected: %v\nactual:   %v", tc.imports, imports)
				}
			}
		})
	}
}

func TestWasmExport(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name      string
		target    string
		buildMode string
		scheduler string
		file      string
		noOutput  bool
		command   bool // call _start (command mode) instead of _initialize
	}

	tests := []testCase{
		// "command mode" WASI
		{
			name:    "WASIp1-command",
			target:  "wasip1",
			command: true,
		},
		// "reactor mode" WASI (with -buildmode=c-shared)
		{
			name:      "WASIp1-reactor",
			target:    "wasip1",
			buildMode: "c-shared",
		},
		// Make sure reactor mode also works without a scheduler.
		{
			name:      "WASIp1-reactor-noscheduler",
			target:    "wasip1",
			buildMode: "c-shared",
			scheduler: "none",
			file:      "wasmexport-noscheduler.go",
		},
		// Test -target=wasm-unknown with the default build mode (which is
		// c-shared).
		{
			name:     "wasm-unknown-reactor",
			target:   "wasm-unknown",
			file:     "wasmexport-noscheduler.go",
			noOutput: true, // wasm-unknown cannot produce output
		},
		// Test -target=wasm-unknown with -buildmode=default, which makes it run
		// in command mode.
		{
			name:      "wasm-unknown-command",
			target:    "wasm-unknown",
			buildMode: "default",
			file:      "wasmexport-noscheduler.go",
			noOutput:  true, // wasm-unknown cannot produce output
			command:   true,
		},
		// Test buildmode=wasi-legacy with WASI.
		{
			name:      "WASIp1-legacy",
			target:    "wasip1",
			buildMode: "wasi-legacy",
			scheduler: "none",
			file:      "wasmexport-noscheduler.go",
			command:   true,
		},
		// Test buildmode=wasi-legacy with the default scheduler.
		{
			name:      "WASIp1-legacy-scheduler",
			target:    "wasip1",
			buildMode: "wasi-legacy",
			command:   true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Build the wasm binary.
			tmpdir := t.TempDir()
			options := optionsFromTarget(tc.target, sema)
			options.BuildMode = tc.buildMode
			options.Scheduler = tc.scheduler
			buildConfig, err := builder.NewConfig(&options)
			if err != nil {
				t.Fatal(err)
			}
			filename := "wasmexport.go"
			if tc.file != "" {
				filename = tc.file
			}
			result, err := builder.Build("testdata/"+filename, ".wasm", tmpdir, buildConfig)
			if err != nil {
				t.Fatal("failed to build binary:", err)
			}

			// Read the wasm binary back into memory.
			data, err := os.ReadFile(result.Binary)
			if err != nil {
				t.Fatal("could not read wasm binary: ", err)
			}

			// Set up the wazero runtime.
			output := &bytes.Buffer{}
			ctx := context.Background()
			r := wazero.NewRuntimeWithConfig(ctx, wazero.NewRuntimeConfigInterpreter())
			defer r.Close(ctx)
			config := wazero.NewModuleConfig().
				WithStdout(output).WithStderr(output).
				WithStartFunctions()

			// Prepare for testing.
			var mod api.Module
			mustCall := func(results []uint64, err error) []uint64 {
				if err != nil {
					t.Error("failed to run function:", err)
				}
				return results
			}
			checkResult := func(name string, results []uint64, expected []uint64) {
				if len(results) != len(expected) {
					t.Errorf("%s: expected %v but got %v", name, expected, results)
				}
				for i, result := range results {
					if result != expected[i] {
						t.Errorf("%s: expected %v but got %v", name, expected, results)
						break
					}
				}
			}
			runTests := func() {
				// Test an exported function without params or return value.
				checkResult("hello()", mustCall(mod.ExportedFunction("hello").Call(ctx)), nil)

				// Test that we can call an exported function more than once.
				checkResult("add(3, 5)", mustCall(mod.ExportedFunction("add").Call(ctx, 3, 5)), []uint64{8})
				checkResult("add(7, 9)", mustCall(mod.ExportedFunction("add").Call(ctx, 7, 9)), []uint64{16})
				checkResult("add(6, 1)", mustCall(mod.ExportedFunction("add").Call(ctx, 6, 1)), []uint64{7})

				// Test that imported functions can call exported functions
				// again.
				checkResult("reentrantCall(2, 3)", mustCall(mod.ExportedFunction("reentrantCall").Call(ctx, 2, 3)), []uint64{5})
				checkResult("reentrantCall(1, 8)", mustCall(mod.ExportedFunction("reentrantCall").Call(ctx, 1, 8)), []uint64{9})

				// Check that goroutines started inside //go:wasmexport don't
				// block the called function from returning.
				checkResult("goroutineExit()", mustCall(mod.ExportedFunction("goroutineExit").Call(ctx)), nil)
			}

			// Add wasip1 module.
			wasi_snapshot_preview1.MustInstantiate(ctx, r)

			// Add custom "tester" module.
			callOutside := func(a, b int32) int32 {
				results, err := mod.ExportedFunction("add").Call(ctx, uint64(a), uint64(b))
				if err != nil {
					t.Error("could not call exported add function:", err)
				}
				return int32(results[0])
			}
			callTestMain := func() {
				runTests()
			}
			builder := r.NewHostModuleBuilder("tester")
			builder.NewFunctionBuilder().WithFunc(callOutside).Export("callOutside")
			builder.NewFunctionBuilder().WithFunc(callTestMain).Export("callTestMain")
			_, err = builder.Instantiate(ctx)
			if err != nil {
				t.Fatal(err)
			}

			// Parse and instantiate the wasm.
			mod, err = r.InstantiateWithConfig(ctx, data, config)
			if err != nil {
				t.Fatal("could not instantiate wasm module:", err)
			}

			// Initialize the module and run the tests.
			if tc.command {
				// Call _start (the entry point), which calls
				// tester.callTestMain, which then runs all the tests.
				_, err := mod.ExportedFunction("_start").Call(ctx)
				if err != nil {
					if exitErr, ok := err.(*sys.ExitError); ok && exitErr.ExitCode() == 0 {
						// Exited with code 0. Nothing to worry about.
					} else {
						t.Error("failed to run _start:", err)
					}
				}
				if tc.buildMode == "wasi-legacy" && mod.IsClosed() {
					// A legacy command returns from _start without exiting.
					t.Error("module exited in _start")
				}
			} else {
				// Run the _initialize call, because this is reactor mode wasm.
				mustCall(mod.ExportedFunction("_initialize").Call(ctx))
				runTests()
			}

			// Check that the output matches the expected output.
			// (Skip this for wasm-unknown because it can't produce output).
			if !tc.noOutput {
				checkOutput(t, "testdata/wasmexport.txt", output.Bytes())
			}
		})
	}
}

// Test that with -buildmode=wasi-legacy and the default scheduler, _start
// returns and the host can call an //export function afterwards.
func TestWasmLegacyExport(t *testing.T) {
	t.Parallel()

	// Build the wasm binary.
	options := optionsFromTarget("wasip1", sema)
	options.BuildMode = "wasi-legacy"
	buildConfig, err := builder.NewConfig(&options)
	if err != nil {
		t.Fatal(err)
	}
	result, err := builder.Build("testdata/wasi-legacy.go", ".wasm", t.TempDir(), buildConfig)
	if err != nil {
		t.Fatal("failed to build binary:", err)
	}
	data, err := os.ReadFile(result.Binary)
	if err != nil {
		t.Fatal("could not read wasm binary:", err)
	}

	// Run _start, which waits for a goroutine to set the result.
	output := &bytes.Buffer{}
	ctx := context.Background()
	r := wazero.NewRuntimeWithConfig(ctx, wazero.NewRuntimeConfigInterpreter())
	defer r.Close(ctx)
	wasi_snapshot_preview1.MustInstantiate(ctx, r)
	config := wazero.NewModuleConfig().WithStdout(output).WithStderr(output)
	mod, err := r.InstantiateWithConfig(ctx, data, config)
	if err != nil {
		t.Fatal("failed to run _start:", err)
	}

	// Call the //export function after _start has returned.
	results, err := mod.ExportedFunction("getResult").Call(ctx)
	if err != nil {
		t.Fatal("failed to call getResult after _start:", err)
	}
	if len(results) != 1 || results[0] != 42 {
		t.Errorf("getResult() = %v, want [42]", results)
	}
	checkOutput(t, "testdata/wasi-legacy.txt", output.Bytes())

	// An //export function runs outside any goroutine, so it cannot block.
	output.Reset()
	if _, err := mod.ExportedFunction("sleep").Call(ctx); err == nil {
		t.Error("sleep() after _start did not fail")
	}
	if !strings.Contains(output.String(), "blocking operation in exported function") {
		t.Errorf("sleep() after _start: unexpected output %q", output.String())
	}
}

// Test js.FuncOf (for syscall/js).
// This test might be extended in the future to cover more cases in syscall/js.
func TestWasmFuncOf(t *testing.T) {
	// Build the wasm binary.
	tmpdir := t.TempDir()
	options := optionsFromTarget("wasm", sema)
	buildConfig, err := builder.NewConfig(&options)
	if err != nil {
		t.Fatal(err)
	}
	result, err := builder.Build("testdata/wasmfunc.go", ".wasm", tmpdir, buildConfig)
	if err != nil {
		t.Fatal("failed to build binary:", err)
	}

	// Test the resulting binary using NodeJS.
	output := &bytes.Buffer{}
	cmd := exec.Command("node", "testdata/wasmfunc.js", result.Binary, buildConfig.BuildMode())
	cmd.Stdout = output
	cmd.Stderr = output
	err = cmd.Run()
	if err != nil {
		t.Error("failed to run node:", err)
	}
	checkOutput(t, "testdata/wasmfunc.txt", output.Bytes())
}

// Test //go:wasmexport in JavaScript (using NodeJS).
func TestWasmExportJS(t *testing.T) {
	t.Parallel()
	type testCase struct {
		name      string
		buildMode string
	}

	tests := []testCase{
		{name: "default"},
		{name: "c-shared", buildMode: "c-shared"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			// Build the wasm binary.
			tmpdir := t.TempDir()
			options := optionsFromTarget("wasm", sema)
			options.BuildMode = tc.buildMode
			buildConfig, err := builder.NewConfig(&options)
			if err != nil {
				t.Fatal(err)
			}
			result, err := builder.Build("testdata/wasmexport-noscheduler.go", ".wasm", tmpdir, buildConfig)
			if err != nil {
				t.Fatal("failed to build binary:", err)
			}

			// Test the resulting binary using NodeJS.
			output := &bytes.Buffer{}
			cmd := exec.Command("node", "testdata/wasmexport.js", result.Binary, buildConfig.BuildMode())
			cmd.Stdout = output
			cmd.Stderr = output
			err = cmd.Run()
			if err != nil {
				t.Error("failed to run node:", err)
			}
			checkOutput(t, "testdata/wasmexport.txt", output.Bytes())
		})
	}
}

func TestWasmExportFinalizersJS(t *testing.T) {
	t.Parallel()

	tmpdir := t.TempDir()
	options := optionsFromTarget("wasm", sema)
	options.BuildMode = "c-shared"
	buildConfig, err := builder.NewConfig(&options)
	if err != nil {
		t.Fatal(err)
	}
	result, err := builder.Build("testdata/wasmexport-finalizer.go", ".wasm", tmpdir, buildConfig)
	if err != nil {
		t.Fatal("failed to build binary:", err)
	}

	output := &bytes.Buffer{}
	cmd := exec.Command("node", "testdata/wasmexport-finalizer.js", result.Binary)
	cmd.Stdout = output
	cmd.Stderr = output
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to run node: %v\n%s", err, output)
	}
}

// Test whether Go.run() (in wasm_exec.js) normally returns and returns the
// right exit code.
func TestWasmExit(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name   string
		output string
	}

	tests := []testCase{
		{name: "normal", output: "exit code: 0\n"},
		{name: "exit-0", output: "exit code: 0\n"},
		{name: "exit-0-sleep", output: "slept\nexit code: 0\n"},
		{name: "exit-1", output: "exit code: 1\n"},
		{name: "exit-1-sleep", output: "slept\nexit code: 1\n"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			options := optionsFromTarget("wasm", sema)
			buildConfig, err := builder.NewConfig(&options)
			if err != nil {
				t.Fatal(err)
			}
			buildConfig.Target.Emulator = "node testdata/wasmexit.js {}"
			output := &bytes.Buffer{}
			_, err = buildAndRun("testdata/wasmexit.go", buildConfig, output, []string{tc.name}, nil, time.Minute, func(cmd *exec.Cmd, result builder.BuildResult) error {
				return cmd.Run()
			})
			if err != nil {
				t.Error(err)
			}
			expected := "wasmexit test: " + tc.name + "\n" + tc.output
			checkOutputData(t, []byte(expected), output.Bytes())
		})
	}
}

func TestWASIPanicTraceback(t *testing.T) {
	t.Parallel()

	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	options := optionsFromTarget("wasip1", sema)
	config, err := builder.NewConfig(&options)
	if err != nil {
		t.Fatal(err)
	}

	result, err := builder.Build("testdata/panic-traceback.go", ".wasm", t.TempDir(), config)
	if err != nil {
		t.Fatal("failed to build binary:", err)
	}
	data, err := os.ReadFile(result.Binary)
	if err != nil {
		t.Fatal("failed to read binary:", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	r := wazero.NewRuntimeWithConfig(ctx, wazero.NewRuntimeConfigInterpreter())
	defer r.Close(ctx)
	wasi_snapshot_preview1.MustInstantiate(ctx, r)
	_, err = r.InstantiateWithConfig(ctx, data, wazero.NewModuleConfig())
	if err == nil {
		t.Fatal("program unexpectedly exited successfully")
	}
	if !strings.Contains(err.Error(), "main.panicHere") {
		t.Fatalf("panic traceback does not contain main.panicHere:\n%s", err)
	}
}
