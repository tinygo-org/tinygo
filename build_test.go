package main

import (
	"bytes"
	"os/exec"
	"regexp"
	"runtime"
	"strings"
	"testing"

	"github.com/tinygo-org/tinygo/builder"
	"github.com/tinygo-org/tinygo/compileopts"
	"github.com/tinygo-org/tinygo/diagnostics"
	"github.com/tinygo-org/tinygo/goenv"
)

func TestBuild(t *testing.T) {
	t.Parallel()

	tests := []string{
		"alias.go",
		"atomic.go",
		"binop.go",
		"buildinfo.go",
		"calls.go",
		"cgo/",
		"channel.go",
		"coreassign.go",
		"embed/",
		"finalizer.go",
		"finalizerbits.go",
		"finalizeridle.go",
		"finalizerinvariants.go",
		"finalizerlarge.go",
		"float.go",
		"gc.go",
		"generics.go",
		"go1.21.go",
		"go1.22/",
		"go1.23/",
		"goroutines.go",
		"init.go",
		"init_multi.go",
		"interface-blank.go",
		"interface.go",
		"json.go",
		"localtypes/",
		"localtypes.go",
		"map.go",
		"map_bigkey.go",
		"math.go",
		"oldgo/",
		"print.go",
		"reflect.go",
		"signal.go",
		"signalnotify.go",
		"slice.go",
		"sort.go",
		"stdlib.go",
		"string.go",
		"structs.go",
		"testing.go",
		"timers.go",
		"typealias.go",
		"zeroalloc.go",
	}

	if *testTarget != "" {
		// This makes it possible to run one specific test (instead of all),
		// which is especially useful to quickly check whether some changes
		// affect a particular target architecture.
		options := optionsFromTarget(*testTarget, sema)
		runPlatTests(options, tests, t)
		if *testTarget == "wasip1" {
			t.Run("cgo-realloc", func(t *testing.T) {
				runTest("cgo-realloc/", options, t, nil, nil)
			})
		}
		return
	}

	t.Run("Host", func(t *testing.T) {
		t.Parallel()
		hostOptions := optionsFromTarget("", sema)
		runPlatTests(hostOptions, tests, t)
		t.Run("testing.go-verbose", func(t *testing.T) {
			t.Parallel()
			runTest("testing-verbose.go", hostOptions, t, nil, nil)
		})

		// scheduler.threads needs threadID, which exists only on Linux and Darwin.
		// scheduler.none does not link on Windows.
		switch runtime.GOOS {
		case "darwin", "linux":
			for _, scheduler := range []string{"threads", "none"} {
				t.Run("finalizerinvariants.go-gc-conservative-scheduler-"+scheduler, func(t *testing.T) {
					t.Parallel()
					options := compileopts.Options(hostOptions)
					options.GC = "conservative"
					options.Scheduler = scheduler
					options.Tags = append(append([]string(nil), hostOptions.Tags...), "runtime_asserts")
					runTest("finalizerinvariants.go", options, t, nil, nil)
				})
			}
		}
	})

	t.Run("boehm-finalizers", func(t *testing.T) {
		t.Parallel()
		options := optionsFromTarget("wasm", sema)
		options.GC = "boehm"
		t.Run("finalizerlarge.go-graph", func(t *testing.T) {
			t.Parallel()
			testOptions := compileopts.Options(options)
			testOptions.GlobalValues = map[string]map[string]string{
				"main": {"finalizerLargeMode": "graph"},
			}
			runTest("finalizerlarge.go", testOptions, t, nil, nil)
		})
		for _, name := range []string{
			"finalizer.go",
			"finalizerbits.go",
			"finalizeridle.go",
			"finalizerlarge.go",
			"finalizerinvariants.go",
		} {
			name := name
			t.Run(name, func(t *testing.T) {
				t.Parallel()
				testOptions := compileopts.Options(options)
				if name == "finalizerlarge.go" {
					testOptions.GlobalValues = map[string]map[string]string{
						"main": {"finalizerLargeMode": "runner-payload"},
					}
				}
				runTest(name, testOptions, t, nil, nil)
			})
		}
	})

	// Test a few build options.
	t.Run("build-options", func(t *testing.T) {
		t.Parallel()

		// Test with few optimizations enabled (no inlining, etc).
		t.Run("opt=1", func(t *testing.T) {
			t.Parallel()
			opts := optionsFromTarget("", sema)
			opts.Opt = "1"
			runTestWithConfig("stdlib.go", t, opts, nil, nil)
		})

		// Test with only the bare minimum of optimizations enabled.
		// TODO: fix this for stdlib.go, which currently fails.
		t.Run("opt=0", func(t *testing.T) {
			t.Parallel()
			opts := optionsFromTarget("", sema)
			opts.Opt = "0"
			runTestWithConfig("print.go", t, opts, nil, nil)
		})

		t.Run("opt=0-gc=boehm", func(t *testing.T) {
			t.Parallel()
			opts := optionsFromTarget("", sema)
			opts.Opt = "0"
			opts.GC = "boehm"
			runTestWithConfig("gc-boehm-opt0.go", t, opts, nil, nil)
		})

		// Regression test: at -opt=0 the compiler does not always remove a
		// local escaping through a pointer cast, so printitf used to
		// allocate on the panic path. printitf is //go:noheap, so a
		// regression here makes the build fail with a linker error.
		t.Run("opt=0-printitf-cortex-m-qemu", func(t *testing.T) {
			t.Parallel()
			opts := optionsFromTarget("cortex-m-qemu", sema)
			opts.Opt = "0"
			config, err := builder.NewConfig(&opts)
			if err != nil {
				t.Fatal(err)
			}
			err = Build("testdata/panic-value.go", t.TempDir()+"/panic-value", config)
			if err != nil {
				w := &bytes.Buffer{}
				diagnostics.CreateDiagnostics(err).WriteTo(w, "")
				t.Fatal(w.String())
			}
		})

		// Regression test: at -opt=0 the compiler does not always remove a
		// local escaping through Queue.Push, so runGC used to allocate while
		// marking the runqueue. runGC is //go:noheap, so a regression here
		// makes the build fail with a linker error instead of silently
		// hanging like it used to.
		t.Run("opt=0-gc-cortex-m-qemu", func(t *testing.T) {
			t.Parallel()
			opts := optionsFromTarget("cortex-m-qemu", sema)
			opts.Opt = "0"
			emuCheck(t, opts)
			runTestWithConfig("gc.go", t, opts, nil, nil)
		})

		t.Run("gc=none-runtime-panic", func(t *testing.T) {
			t.Parallel()
			opts := optionsFromTarget("cortex-m-qemu", sema)
			opts.GC = "none"
			opts.Scheduler = "none"
			config, err := builder.NewConfig(&opts)
			if err != nil {
				t.Fatal(err)
			}
			err = Build("testdata/trivialpanic.go", t.TempDir()+"/trivialpanic", config)
			if err != nil {
				w := &bytes.Buffer{}
				diagnostics.CreateDiagnostics(err).WriteTo(w, "")
				t.Fatal(w.String())
			}
		})

		t.Run("scheduler=none-uefi", func(t *testing.T) {
			t.Parallel()
			opts := optionsFromTarget("uefi-amd64", sema)
			opts.Scheduler = "none"
			config, err := builder.NewConfig(&opts)
			if err != nil {
				t.Fatal(err)
			}
			err = Build("testdata/trivialpanic.go", t.TempDir()+"/trivialpanic", config)
			if err != nil {
				w := &bytes.Buffer{}
				diagnostics.CreateDiagnostics(err).WriteTo(w, "")
				t.Fatal(w.String())
			}
		})

		t.Run("ldflags", func(t *testing.T) {
			t.Parallel()
			opts := optionsFromTarget("", sema)
			opts.GlobalValues = map[string]map[string]string{
				"main": {
					"someGlobal": "foobar",
				},
			}
			runTestWithConfig("ldflags.go", t, opts, nil, nil)
		})

		// Same as ldflags, but the global has a default value in source.
		// -ldflags -X must override it, matching standard Go behaviour.
		t.Run("ldflags-initialized", func(t *testing.T) {
			t.Parallel()
			opts := optionsFromTarget("", sema)
			opts.GlobalValues = map[string]map[string]string{
				"main": {
					"someGlobal": "foobar",
				},
			}
			runTestWithConfig("ldflags-initialized.go", t, opts, nil, nil)
		})
	})

	// Short mode skips the targets that every host can test. Linux CI tests
	// them without -short.
	if !testing.Short() {
		t.Run("EmulatedCortexM3", func(t *testing.T) {
			t.Parallel()
			runPlatTests(optionsFromTarget("cortex-m-qemu", sema), tests, t)
		})

		t.Run("EmulatedRISCV", func(t *testing.T) {
			t.Parallel()
			runPlatTests(optionsFromTarget("riscv-qemu", sema), tests, t)
		})

		t.Run("AVR", func(t *testing.T) {
			t.Parallel()
			runPlatTests(optionsFromTarget("simavr", sema), tests, t)
		})

		t.Run("WebAssembly", func(t *testing.T) {
			t.Parallel()

			runPlatTests(optionsFromTarget("wasm", sema), tests, t)
			runGCLivenessTest(optionsFromTarget("wasm", sema), t)
			// Test with -gc=boehm.
			t.Run("gc.go-boehm", func(t *testing.T) {
				t.Parallel()
				optionsBoehm := optionsFromTarget("wasm", sema)
				optionsBoehm.GC = "boehm"
				runTest("gc.go", optionsBoehm, t, nil, nil)
				runTest("gc-boehm.go", optionsBoehm, t, nil, nil)
			})
		})
		t.Run("WASIp1", func(t *testing.T) {
			t.Parallel()
			options := optionsFromTarget("wasip1", sema)
			runPlatTests(options, tests, t)
			runGCLivenessTest(options, t)
			t.Run("cgo-realloc", func(t *testing.T) {
				runTest("cgo-realloc/", options, t, nil, nil)
			})

			// Test with -gc=boehm.
			t.Run("gc.go-boehm", func(t *testing.T) {
				t.Parallel()
				optionsBoehm := optionsFromTarget("wasip1", sema)
				optionsBoehm.GC = "boehm"
				runTest("gc.go", optionsBoehm, t, nil, nil)
				runTest("gc-boehm.go", optionsBoehm, t, nil, nil)
			})
		})
		t.Run("WASIp2", func(t *testing.T) {
			t.Parallel()
			runPlatTests(optionsFromTarget("wasip2", sema), tests, t)
			runGCLivenessTest(optionsFromTarget("wasip2", sema), t)
		})
	}

	// Linux CI also tests the other Linux architectures without -short, so
	// short mode skips them. Windows CI runs only in short mode, so
	// Windows386 continues to run.
	if runtime.GOOS == "linux" && !testing.Short() {
		for name, osArch := range supportedLinuxArches {
			options := optionsFromOSARCH(osArch, sema)
			if options.GOARCH != runtime.GOARCH { // Native architecture already run above.
				t.Run(name, func(t *testing.T) {
					runPlatTests(options, tests, t)
				})
			}
		}
		t.Run("MIPS little-endian", func(t *testing.T) {
			// Run a single test for GOARCH=mipsle to see whether it works at
			// all. It is already mostly tested because GOARCH=mips and
			// GOARCH=mipsle are so similar, but it's good to have an extra test
			// to be sure.
			t.Parallel()
			options := optionsFromOSARCH("linux/mipsle/softfloat", sema)
			emuCheck(t, options)
			runTest("cgo/", options, t, nil, nil)
		})
	} else if runtime.GOOS == "windows" {
		if runtime.GOARCH != "386" {
			t.Run("Windows386", func(t *testing.T) {
				t.Parallel()
				runPlatTests(optionsFromOSARCH("windows/386", sema), tests, t)
			})
		}
	}
}

func TestESP32QEMU(t *testing.T) {
	t.Parallel()

	options := optionsFromTarget("esp32-qemu", sema)
	emuCheck(t, options)
	machines, err := exec.Command("qemu-system-xtensa", "-machine", "help").Output()
	if err != nil {
		t.Fatal("failed to list qemu-system-xtensa machines:", err)
	}
	if !regexp.MustCompile(`(?m)^esp32\s`).Match(machines) {
		t.Skip("qemu-system-xtensa does not support the ESP32 machine")
	}
	for _, name := range []string{"print.go", "recover.go"} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			runTest(name, options, t, nil, nil)
		})
	}
}

func runPlatTests(options compileopts.Options, tests []string, t *testing.T) {
	emuCheck(t, options)

	spec, err := compileopts.LoadTarget(&options)
	if err != nil {
		t.Fatal("failed to load target spec:", err)
	}

	// FIXME: this should really be:
	// isWebAssembly := strings.HasPrefix(spec.Triple, "wasm")
	isWASI := strings.HasPrefix(options.Target, "wasi")
	isWebAssembly := isWASI || strings.HasPrefix(options.Target, "wasm") || (options.Target == "" && strings.HasPrefix(options.GOARCH, "wasm"))
	isBaremetal := options.Target == "simavr" || options.Target == "cortex-m-qemu" || options.Target == "riscv-qemu"
	_, goMinor, err := goenv.GetGorootVersion()
	if err != nil {
		t.Fatal("could not get version:", goMinor)
	}

	for _, name := range tests {
		if name == "goroutines.go" && (spec.Scheduler == "threads" || spec.Scheduler == "cores") {
			// This test intentionally checks concurrent scheduling by comparing
			// output order, so only run it with non-threaded schedulers.
			continue
		}
		if name == "coreassign.go" && spec.Scheduler != "cores" {
			// Core assignment needs the cores scheduler.
			continue
		}
		if options.GOOS == "linux" && (options.GOARCH == "arm" || options.GOARCH == "386") {
			switch name {
			case "timers.go":
				// Timer tests do not work because syscall.seek is implemented
				// as Assembly in mainline Go and causes linker failure
				continue
			}
		}
		if options.GOOS == "linux" && (options.GOARCH == "mips" || options.GOARCH == "mipsle") {
			if name == "atomic.go" || name == "timers.go" {
				// 64-bit atomic operations aren't currently supported on MIPS.
				continue
			}
		}
		if options.GOOS == "linux" && options.GOARCH == "mips" {
			if name == "cgo/" {
				// CGo isn't supported yet on big-endian systems (needs updates
				// to bitfield access methods).
				continue
			}
		}
		if options.Target == "cortex-m-qemu" && goMinor >= 27 && name == "json.go" {
			// Go 1.27 jsonv2 exceeds the LM3S6965's 256KiB flash. json.go
			// is still covered by larger targets such as riscv-qemu.
			continue
		}
		if options.Target == "simavr" {
			// Not all tests are currently supported on AVR.
			// Skip the ones that aren't.
			switch name {
			case "reflect.go":
				// Reflect tests do not run correctly, probably because of the
				// limited amount of memory.
				continue

			case "map_bigkey.go":
				// Compiler generates many large stack temporaries for [256]byte
				// map keys, overflowing the goroutine stack (384 bytes).
				continue

			case "gc.go":
				// Does not pass due to high mark false positive rate.
				continue

			case "buildinfo.go", "json.go", "stdlib.go", "testing.go":
				// Too big for AVR. Doesn't fit in flash/RAM.
				continue

			case "cgo/":
				// CGo function pointers don't work on AVR (needs LLVM 16 and
				// some compiler changes).
				continue

			case "timers.go":
				// Crashes starting with Go 1.23.
				// Bug: https://github.com/llvm/llvm-project/issues/104032
				continue

			default:
			}
		}
		if options.Target == "wasip2" {
			switch name {
			case "cgo/":
				// waisp2 use our own libc; cgo tests fail
				continue
			}
		}
		if options.GOOS == "windows" {
			switch name {
			case "signal.go", "signalnotify.go":
				// os/signal does not link on Windows.
				continue
			}
		}
		if isWebAssembly || isBaremetal {
			switch name {
			case "signal.go":
				// Signals only work on POSIX-like systems.
				continue
			}
		}
		if options.Target != "wasm" {
			switch name {
			case "finalizer.go", "finalizerbits.go", "finalizeridle.go", "finalizerlarge.go":
				// These tests require deterministic finalization on target wasm.
				// finalizerinvariants.go covers other block GC targets.
				continue
			}
		}
		if options.Target == "simavr" {
			switch name {
			case "finalizerinvariants.go":
				// Skip because runtime.GC does not return. See the gc.go exclusion above.
				continue
			}
		}
		name := name // redefine to avoid race condition
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			testOptions := compileopts.Options(options)
			switch name {
			case "finalizerinvariants.go", "finalizerlarge.go", "interface.go", "interface-blank.go":
				testOptions.Tags = append(append([]string(nil), options.Tags...), "runtime_asserts")
			}
			if testOptions.Target == "simavr" && (name == "math.go" || name == "string.go") {
				// These tests exceed simavr's default 384-byte goroutine stack.
				testOptions.StackSize = 512
			}
			runTest(name, testOptions, t, nil, nil)
		})
	}
	if !strings.HasPrefix(spec.Emulator, "simavr ") {
		t.Run("env.go", func(t *testing.T) {
			t.Parallel()
			runTest("env.go", options, t, []string{"first", "second"}, []string{"ENV1=VALUE1", "ENV2=VALUE2"})
		})
	}
	if isWebAssembly {
		t.Run("alias.go-scheduler-none", func(t *testing.T) {
			t.Parallel()
			options := compileopts.Options(options)
			options.Scheduler = "none"
			runTest("alias.go", options, t, nil, nil)
		})
	}
	if options.Target == "" || isWASI {
		t.Run("filesystem.go", func(t *testing.T) {
			t.Parallel()
			runTest("filesystem.go", options, t, nil, nil)
		})
	}
	if options.Target == "wasm" || isBaremetal {
		t.Run("devnull.go", func(t *testing.T) {
			t.Parallel()
			runTest("devnull.go", options, t, nil, nil)
		})
	}
	if options.Target == "" || options.Target == "wasm" || isWASI {
		t.Run("rand.go", func(t *testing.T) {
			t.Parallel()
			runTest("rand.go", options, t, nil, nil)
		})
	}
	t.Run("recover.go", func(t *testing.T) {
		t.Parallel()
		runTest("recover.go", options, t, nil, nil)
	})
	if isWebAssembly {
		t.Run("recover-explicit.go", func(t *testing.T) {
			t.Parallel()
			options := compileopts.Options(options)
			options.Scheduler = "none"
			options.PanicUnwind = "explicit"
			runTest("recover-explicit.go", options, t, nil, nil)
		})
		t.Run("calls.go-unwind-none", func(t *testing.T) {
			t.Parallel()
			options := compileopts.Options(options)
			options.PanicUnwind = "none"
			runTest("calls.go", options, t, nil, nil)
		})
	}
}

// runGCLivenessTest runs testdata/gc-liveness-repro.go. Only targets with
// stack objects can lose a GC root the way it reproduces, so it is registered
// per wasm target instead of being part of the shared test list.
func runGCLivenessTest(options compileopts.Options, t *testing.T) {
	t.Run("gc-liveness-repro.go", func(t *testing.T) {
		t.Parallel()
		runTest("gc-liveness-repro.go", options, t, nil, nil)
	})
}
