package builder

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/tinygo-org/tinygo/compileopts"
	"github.com/tinygo-org/tinygo/goenv"
	"tinygo.org/x/go-llvm"
)

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
		//"cortex-m33", // TODO: broken in LLVM 11, fixed in https://reviews.llvm.org/D90305
		"cortex-m4",
		"cortex-m7",
		"esp32c3",
		"fe310",
		"gameboy-advance",
		"k210",
		"nintendoswitch",
		"riscv-qemu",
		"wasi",
		"wasm",
	}
	if hasBuiltinTools {
		// hasBuiltinTools is set when TinyGo is statically linked with LLVM,
		// which also implies it was built with Xtensa support.
		targetNames = append(targetNames, "esp32", "esp8266")
	}
	for _, targetName := range targetNames {
		targetName := targetName
		t.Run(targetName, func(t *testing.T) {
			testClangAttributes(t, &compileopts.Options{Target: targetName})
		})
	}

	for _, options := range []*compileopts.Options{
		{GOOS: "linux", GOARCH: "386"},
		{GOOS: "linux", GOARCH: "amd64"},
		{GOOS: "linux", GOARCH: "arm", GOARM: "5"},
		{GOOS: "linux", GOARCH: "arm", GOARM: "6"},
		{GOOS: "linux", GOARCH: "arm", GOARM: "7"},
		{GOOS: "linux", GOARCH: "arm64"},
		{GOOS: "darwin", GOARCH: "amd64"},
		{GOOS: "darwin", GOARCH: "arm64"},
	} {
		t.Run("GOOS="+options.GOOS+",GOARCH="+options.GOARCH, func(t *testing.T) {
			testClangAttributes(t, options)
		})
	}
}

func testClangAttributes(t *testing.T, options *compileopts.Options) {
	testDir := t.TempDir()
	clangHeaderPath := getClangHeaderPath(goenv.Get("TINYGOROOT"))

	ctx := llvm.NewContext()
	defer ctx.Dispose()

	target, err := compileopts.LoadTarget(options)
	if err != nil {
		t.Fatalf("could not load target: %s", err)
	}
	config := compileopts.Config{
		Options:      options,
		Target:       target,
		ClangHeaders: clangHeaderPath,
	}

	// Create a very simple C input file.
	srcpath := filepath.Join(testDir, "test.c")
	err = ioutil.WriteFile(srcpath, []byte("int add(int a, int b) { return a + b; }"), 0o666)
	if err != nil {
		t.Fatalf("could not write target file %s: %s", srcpath, err)
	}

	// Compile this file using Clang.
	outpath := filepath.Join(testDir, "test.bc")
	flags := append([]string{"-c", "-emit-llvm", "-o", outpath, srcpath}, config.CFlags()...)
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
	if mod.Target() != config.Triple() {
		t.Errorf("target has LLVM triple %#v but Clang makes it LLVM triple %#v", config.Triple(), mod.Target())
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
		if llvm.Version != "11.0.0" {
			// This needs to be removed once we switch to LLVM 12.
			// LLVM 11.0.0 uses a different "target-features" string than LLVM
			// 11.1.0 for Thumb targets. The Xtensa fork is still based on LLVM
			// 11.0.0, so we need to skip this check on that version.
			t.Errorf("target has LLVM features %#v but Clang makes it %#v", config.Features(), features)
		}
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
