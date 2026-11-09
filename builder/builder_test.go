package builder

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/tinygo-org/tinygo/compileopts"
	"tinygo.org/x/go-llvm"
)

func TestBuildRejectsUnsupportedFirmwareExtension(t *testing.T) {
	for _, targetName := range []string{"pico", "pico2"} {
		t.Run(targetName, func(t *testing.T) {
			options := &compileopts.Options{Target: targetName, Semaphore: sema}
			target, err := compileopts.LoadTarget(options)
			if err != nil {
				t.Fatal(err)
			}
			config := &compileopts.Config{
				Options: options,
				Target:  target,
			}
			_, err = Build("does/not/exist", "firmware.bin", t.TempDir(), config)
			want := "target " + targetName + " does not support firmware output extension .bin; use .elf, .hex, or .uf2"
			if err == nil || err.Error() != want {
				t.Fatalf("Build() error = %v, want %q", err, want)
			}
		})
	}
}

func TestBuildErrorReturnsNil(t *testing.T) {
	config := &compileopts.Config{
		Options: &compileopts.Options{Semaphore: sema},
		Target:  &compileopts.TargetSpec{Libc: "unknown"},
	}
	result, err := Build("", "", t.TempDir(), config)
	if err == nil || err.Error() != "unknown libc: unknown" {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != nil {
		t.Fatalf("expected nil result on error, got %+v", result)
	}
}

func TestNewConfigAcceptsSemaphore(t *testing.T) {
	for _, tc := range []struct {
		name      string
		semaphore chan struct{}
		capacity  int
	}{
		{name: "minimum", semaphore: make(chan struct{}, 1), capacity: 1},
		{name: "provided", semaphore: make(chan struct{}, 2), capacity: 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			options := &compileopts.Options{Target: "pico", Semaphore: tc.semaphore}
			config, err := NewConfig(options)
			if err != nil {
				t.Fatal(err)
			}
			semaphore := config.Options.Semaphore
			if cap(semaphore) != tc.capacity {
				t.Fatalf("semaphore capacity = %d, want %d", cap(semaphore), tc.capacity)
			}
			if semaphore != tc.semaphore {
				t.Fatal("NewConfig replaced the provided semaphore")
			}
			other, err := NewConfig(options)
			if err != nil {
				t.Fatal(err)
			}
			if other.Options.Semaphore != semaphore {
				t.Fatal("NewConfig replaced the semaphore when reusing options")
			}
		})
	}
}

func TestNewConfigRejectsZeroCapacitySemaphore(t *testing.T) {
	for _, tc := range []struct {
		name      string
		semaphore chan struct{}
	}{
		{name: "nil"},
		{name: "unbuffered", semaphore: make(chan struct{})},
	} {
		t.Run(tc.name, func(t *testing.T) {
			options := &compileopts.Options{Target: "pico", Semaphore: tc.semaphore}
			config, err := NewConfig(options)
			if err == nil || err.Error() != "cannot run 0 jobs at a time" {
				t.Fatalf("NewConfig() error = %v, want cannot run 0 jobs at a time", err)
			}
			if config != nil {
				t.Fatalf("expected nil config on error, got %+v", config)
			}
			if options.Semaphore != tc.semaphore {
				t.Fatal("NewConfig changed the provided semaphore")
			}
		})
	}
}

func TestBuildRejectsZeroCapacitySemaphore(t *testing.T) {
	for _, tc := range []struct {
		name      string
		semaphore chan struct{}
	}{
		{name: "nil"},
		{name: "unbuffered", semaphore: make(chan struct{})},
	} {
		t.Run(tc.name, func(t *testing.T) {
			config := &compileopts.Config{
				Options: &compileopts.Options{Semaphore: tc.semaphore},
				Target:  &compileopts.TargetSpec{Libc: "unknown"},
			}
			result, err := Build("", "", t.TempDir(), config)
			if err == nil || err.Error() != "cannot run 0 jobs at a time" {
				t.Fatalf("Build() error = %v, want cannot run 0 jobs at a time", err)
			}
			if result != nil {
				t.Fatalf("expected nil result on error, got %+v", result)
			}
			if config.Options.Semaphore != tc.semaphore {
				t.Fatal("Build changed the semaphore")
			}
		})
	}
}

func TestBuildSharesSemaphore(t *testing.T) {
	config, err := NewConfig(&compileopts.Options{Target: "pico", Semaphore: make(chan struct{}, 2)})
	if err != nil {
		t.Fatal(err)
	}
	semaphore := config.Options.Semaphore
	tmpdir := t.TempDir()
	start := make(chan struct{})
	results := make(chan error, 16)
	for range 16 {
		go func() {
			<-start
			_, err := Build("does/not/exist", "firmware.bin", tmpdir, config)
			results <- err
		}()
	}
	close(start)
	for range 16 {
		err := <-results
		if err == nil || !strings.Contains(err.Error(), "does not support firmware output extension") {
			t.Fatalf("Build() error = %v, want unsupported extension error", err)
		}
	}
	if config.Options.Semaphore != semaphore {
		t.Fatal("Build changed the semaphore")
	}
}

// Test whether the Clang generated "target-cpu" and "target-features"
// attributes match the CPU and Features property in TinyGo target files.
func TestClangAttributes(t *testing.T) {
	var targetNames = []string{
		// Please keep this list sorted!
		"atmega328p",
		"atmega1280",
		"atmega1284p",
		"atmega2560",
		"attiny85",
		"cortex-m0",
		"cortex-m0plus",
		"cortex-m3",
		"cortex-m33",
		"cortex-m4",
		"cortex-m7",
		"esp32c3",
		"esp32c6",
		"esp32h2",
		"esp32s3",
		"fe310",
		"gameboy-advance",
		"k210",
		"nintendoswitch",
		"riscv-qemu",
		"tkey",
		"uefi-amd64",
		"wasip1",
		"wasip2",
		"wasm",
		"wasm-unknown",
	}
	if hasBuiltinTools {
		// hasBuiltinTools is set when TinyGo is statically linked with LLVM,
		// which also implies it was built with Xtensa support.
		targetNames = append(targetNames, "esp32", "esp8266")
	}
	for _, targetName := range targetNames {
		t.Run(targetName, func(t *testing.T) {
			testClangAttributes(t, &compileopts.Options{Target: targetName})
		})
	}

	for _, options := range []*compileopts.Options{
		{GOOS: "linux", GOARCH: "386"},
		{GOOS: "linux", GOARCH: "amd64"},
		{GOOS: "linux", GOARCH: "arm", GOARM: "5,softfloat"},
		{GOOS: "linux", GOARCH: "arm", GOARM: "6,softfloat"},
		{GOOS: "linux", GOARCH: "arm", GOARM: "7,softfloat"},
		{GOOS: "linux", GOARCH: "arm", GOARM: "5,hardfloat"},
		{GOOS: "linux", GOARCH: "arm", GOARM: "6,hardfloat"},
		{GOOS: "linux", GOARCH: "arm", GOARM: "7,hardfloat"},
		{GOOS: "linux", GOARCH: "arm64"},
		{GOOS: "linux", GOARCH: "mips", GOMIPS: "hardfloat"},
		{GOOS: "linux", GOARCH: "mipsle", GOMIPS: "hardfloat"},
		{GOOS: "linux", GOARCH: "mips", GOMIPS: "softfloat"},
		{GOOS: "linux", GOARCH: "mipsle", GOMIPS: "softfloat"},
		{GOOS: "darwin", GOARCH: "amd64"},
		{GOOS: "darwin", GOARCH: "arm64"},
		{GOOS: "windows", GOARCH: "386"},
		{GOOS: "windows", GOARCH: "amd64"},
		{GOOS: "windows", GOARCH: "arm64"},
	} {
		name := "GOOS=" + options.GOOS + ",GOARCH=" + options.GOARCH
		if options.GOARCH == "arm" {
			name += ",GOARM=" + options.GOARM
		}
		if options.GOARCH == "mips" || options.GOARCH == "mipsle" {
			name += ",GOMIPS=" + options.GOMIPS
		}
		t.Run(name, func(t *testing.T) {
			testClangAttributes(t, options)
		})
	}
}

func testClangAttributes(t *testing.T, options *compileopts.Options) {
	testDir := t.TempDir()

	ctx := llvm.NewContext()
	defer ctx.Dispose()

	target, err := compileopts.LoadTarget(options)
	if err != nil {
		t.Fatalf("could not load target: %s", err)
	}
	config := compileopts.Config{
		Options: options,
		Target:  target,
	}

	// Create a very simple C input file.
	srcpath := filepath.Join(testDir, "test.c")
	err = os.WriteFile(srcpath, []byte("int add(int a, int b) { return a + b; }"), 0o666)
	if err != nil {
		t.Fatalf("could not write target file %s: %s", srcpath, err)
	}

	// Compile this file using Clang.
	outpath := filepath.Join(testDir, "test.bc")
	flags := append([]string{"-c", "-emit-llvm", "-o", outpath, srcpath}, config.CFlags(false)...)
	if config.GOOS() == "darwin" {
		// Silence some warnings that happen when testing GOOS=darwin on
		// something other than MacOS.
		flags = append(flags, "-Wno-missing-sysroot", "-Wno-incompatible-sysroot")
	}
	err = runCCompiler(flags...)
	if err != nil {
		t.Fatalf("failed to compile %s: %s", srcpath, err)
	}

	// Read the resulting LLVM bitcode.
	mod, err := ctx.ParseBitcodeFile(outpath)
	if err != nil {
		t.Fatalf("could not parse bitcode file %s: %s", outpath, err)
	}
	defer mod.Dispose()

	// Check whether the LLVM target matches.
	// Use ClangTriple since LLVM 22 normalizes wasm32-unknown-wasi to wasip1.
	expectedTriple := compileopts.ClangTriple(config.Triple())
	if mod.Target() != expectedTriple {
		t.Errorf("target has LLVM triple %#v but Clang makes it LLVM triple %#v", expectedTriple, mod.Target())
	}

	// Check the "target-cpu" and "target-features" string attribute of the add
	// function.
	add := mod.NamedFunction("add")
	var cpu, features string
	cpuAttr := add.GetStringAttributeAtIndex(-1, "target-cpu")
	featuresAttr := add.GetStringAttributeAtIndex(-1, "target-features")
	if !cpuAttr.IsNil() {
		cpu = cpuAttr.GetStringValue()
	}
	if !featuresAttr.IsNil() {
		features = featuresAttr.GetStringValue()
	}
	if cpu != config.CPU() {
		t.Errorf("target has CPU %#v but Clang makes it CPU %#v", config.CPU(), cpu)
	}
	if features != config.Features() {
		if hasBuiltinTools || runtime.GOOS != "linux" {
			// Skip this step when using an external Clang invocation on Linux.
			// The reason is that Debian has patched Clang in a way that
			// modifies the LLVM features string, changing lots of FPU/float
			// related flags. We want to test vanilla Clang, not Debian Clang.
			//
			// Rather than requiring an exact match (which breaks across LLVM
			// versions as new features are added), check that every feature
			// TinyGo specifies is consistent with what Clang produces:
			//  - A "+feature" in TinyGo must appear in Clang's output.
			//  - A "-feature" in TinyGo must not be "+feature" in Clang's output.
			checkFeatureFlags(t, config.Features(), features)
		}
	}
}

// checkFeatureFlags verifies that all features specified by TinyGo's target
// configuration are consistent with Clang's output. This allows Clang to add
// new features across LLVM versions without breaking the test.
func checkFeatureFlags(t *testing.T, targetFeatures, clangFeatures string) {
	t.Helper()

	// Build a set of Clang's features for fast lookup.
	clangSet := make(map[string]bool) // feature name -> enabled
	for f := range strings.SplitSeq(clangFeatures, ",") {
		f = strings.TrimSpace(f)
		if len(f) < 2 {
			continue
		}
		enabled := f[0] == '+'
		name := f[1:]
		clangSet[name] = enabled
	}

	// Check each feature that TinyGo specifies.
	var missing, conflicts []string
	for f := range strings.SplitSeq(targetFeatures, ",") {
		f = strings.TrimSpace(f)
		if len(f) < 2 {
			continue
		}
		wantEnabled := f[0] == '+'
		name := f[1:]

		clangEnabled, inClang := clangSet[name]
		if wantEnabled && (!inClang || !clangEnabled) {
			// TinyGo requires +feature but Clang doesn't enable it.
			missing = append(missing, f)
		} else if !wantEnabled && inClang && clangEnabled {
			// TinyGo requires -feature but Clang enables it.
			conflicts = append(conflicts, fmt.Sprintf("target has %q but Clang has %q", f, "+"+name))
		}
	}

	if len(missing) > 0 {
		t.Errorf("target specifies features not present in Clang output: %s\n\ttarget features: %s\n\tclang features:  %s",
			strings.Join(missing, ", "), targetFeatures, clangFeatures)
	}
	if len(conflicts) > 0 {
		t.Errorf("target disables features that Clang enables: %s",
			strings.Join(conflicts, "; "))
	}
}

// This TestMain is necessary because TinyGo may also be invoked to run certain
// LLVM tools in a separate process. Not capturing these invocations would lead
// to recursive tests.
func TestMain(m *testing.M) {
	if len(os.Args) >= 2 {
		switch os.Args[1] {
		case "clang", "ld.lld", "wasm-ld":
			// Invoke a specific tool.
			err := RunTool(os.Args[1], os.Args[2:]...)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
			os.Exit(0)
		}
	}

	// Run normal tests.
	os.Exit(m.Run())
}
