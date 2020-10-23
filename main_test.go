package main

// This file tests the compiler by running Go files in testdata/*.go and
// comparing their output with the expected output in testdata/*.txt.

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/tinygo-org/tinygo/builder"
	"github.com/tinygo-org/tinygo/compileopts"
	"github.com/tinygo-org/tinygo/goenv"
)

var testTarget = flag.String("target", "", "override test target")

func TestCompiler(t *testing.T) {
	tests := []string{
		"alias.go",
		"atomic.go",
		"binop.go",
		"calls.go",
		"cgo/",
		"channel.go",
		"coroutines.go",
		"float.go",
		"gc.go",
		"init.go",
		"init_multi.go",
		"interface.go",
		"map.go",
		"math.go",
		"print.go",
		"reflect.go",
		"slice.go",
		"stdlib.go",
		"string.go",
		"structs.go",
		"zeroalloc.go",
	}

	if *testTarget != "" {
		// This makes it possible to run one specific test (instead of all),
		// which is especially useful to quickly check whether some changes
		// affect a particular target architecture.
		runPlatTests(*testTarget, tests, t)
		return
	}

	if runtime.GOOS != "windows" {
		t.Run("Host", func(t *testing.T) {
			runPlatTests("", tests, t)
		})
	}

	if testing.Short() {
		return
	}

	t.Run("EmulatedCortexM3", func(t *testing.T) {
		runPlatTests("cortex-m-qemu", tests, t)
	})

	if runtime.GOOS == "windows" || runtime.GOOS == "darwin" {
		// Note: running only on Windows and macOS because Linux (as of 2020)
		// usually has an outdated QEMU version that doesn't support RISC-V yet.
		t.Run("EmulatedRISCV", func(t *testing.T) {
			runPlatTests("riscv-qemu", tests, t)
		})
	}

	if runtime.GOOS == "linux" {
		t.Run("X86Linux", func(t *testing.T) {
			runPlatTests("i386--linux-gnu", tests, t)
		})
		t.Run("ARMLinux", func(t *testing.T) {
			runPlatTests("arm--linux-gnueabihf", tests, t)
		})
		t.Run("ARM64Linux", func(t *testing.T) {
			runPlatTests("aarch64--linux-gnu", tests, t)
		})
		goVersion, err := goenv.GorootVersionString(goenv.Get("GOROOT"))
		if err != nil {
			t.Error("could not get Go version:", err)
			return
		}
		minorVersion := strings.Split(goVersion, ".")[1]
		if minorVersion != "13" {
			// WebAssembly tests fail on Go 1.13, so skip them there. Versions
			// below that are also not supported but still seem to pass, so
			// include them in the tests for now.
			t.Run("WebAssembly", func(t *testing.T) {
				runPlatTests("wasm", tests, t)
			})
		}

		t.Run("WASI", func(t *testing.T) {
			runPlatTests("wasi", tests, t)
		})
	}
}

func runPlatTests(target string, tests []string, t *testing.T) {
	t.Parallel()

	for _, path := range tests {
		path := path // redefine to avoid race condition

		t.Run(filepath.Base(path), func(t *testing.T) {
			t.Parallel()
			runTest(path, target, t)
		})
	}
}

// Due to some problems with LLD, we cannot run links in parallel, or in parallel with compiles.
// Therefore, we put a lock around builds and run everything else in parallel.
var buildLock sync.Mutex

// runBuild is a thread-safe wrapper around Build.
func runBuild(src, out string, opts *compileopts.Options) error {
	buildLock.Lock()
	defer buildLock.Unlock()

	return Build(src, out, opts)
}

func runTest(name, target string, t *testing.T) {
	// Get the expected output for this test.
	txtfile := name[:len(name)-3] + ".txt"
	if name[len(name)-1] == '/' {
		txtfile = name + "out.txt"
	}
	expected, err := ioutil.ReadFile(filepath.Join("testdata", txtfile))
	if err != nil {
		t.Fatal("could not read expected output file:", err)
	}

	// Create a temporary directory for test output files.
	tmpdir, err := ioutil.TempDir("", "tinygo-test")
	if err != nil {
		t.Fatal("could not create temporary directory:", err)
	}
	defer func() {
		rerr := os.RemoveAll(tmpdir)
		if rerr != nil {
			t.Errorf("failed to remove temporary directory %q: %s", tmpdir, rerr.Error())
		}
	}()

	// Build the test binary.
	config := &compileopts.Options{
		Target:     target,
		Opt:        "z",
		PrintIR:    false,
		DumpSSA:    false,
		VerifyIR:   true,
		Debug:      true,
		PrintSizes: "",
		WasmAbi:    "",
	}

	binary := filepath.Join(tmpdir, "test")
	err = runBuild("./testdata/"+name, binary, config)
	if err != nil {
		printCompilerError(t.Log, err)
		t.Fail()
		return
	}

	// Run the test.
	runComplete := make(chan struct{})
	var cmd *exec.Cmd
	ranTooLong := false
	if target == "" {
		cmd = exec.Command(binary)
	} else {
		spec, err := compileopts.LoadTarget(target)
		if err != nil {
			t.Fatal("failed to load target spec:", err)
		}
		if len(spec.Emulator) == 0 {
			cmd = exec.Command(binary)
		} else {
			args := append(spec.Emulator[1:], binary)
			cmd = exec.Command(spec.Emulator[0], args...)
		}
	}
	stdout := &bytes.Buffer{}
	cmd.Stdout = stdout
	cmd.Stderr = os.Stderr
	err = cmd.Start()
	if err != nil {
		t.Fatal("failed to start:", err)
	}
	go func() {
		// Terminate the process if it runs too long.
		timer := time.NewTimer(10 * time.Second)
		select {
		case <-runComplete:
			timer.Stop()
		case <-timer.C:
			ranTooLong = true
			if runtime.GOOS == "windows" {
				cmd.Process.Signal(os.Kill) // Windows doesn't support SIGINT.
			} else {
				cmd.Process.Signal(os.Interrupt)
			}
		}
	}()
	err = cmd.Wait()
	if _, ok := err.(*exec.ExitError); ok && target != "" {
		err = nil // workaround for QEMU
	}
	close(runComplete)

	if ranTooLong {
		stdout.WriteString("--- test ran too long, terminating...\n")
	}

	// putchar() prints CRLF, convert it to LF.
	actual := bytes.Replace(stdout.Bytes(), []byte{'\r', '\n'}, []byte{'\n'}, -1)
	expected = bytes.Replace(expected, []byte{'\r', '\n'}, []byte{'\n'}, -1) // for Windows

	// Check whether the command ran successfully.
	fail := false
	if err != nil {
		t.Log("failed to run:", err)
		fail = true
	} else if !bytes.Equal(expected, actual) {
		t.Log("output did not match")
		fail = true
	}

	if fail {
		r := bufio.NewReader(bytes.NewReader(actual))
		for {
			line, err := r.ReadString('\n')
			if err != nil {
				break
			}
			t.Log("stdout:", line[:len(line)-1])
		}
		t.Fail()
	}
}

// TestHostEnvironment tests command line arguments. In the future it may also
// test other things, such as environment variables.
func TestHostEnvironment(t *testing.T) {
	// Create a temporary directory for test output files.
	tmpdir, err := ioutil.TempDir("", "tinygo-test")
	if err != nil {
		t.Fatal("could not create temporary directory:", err)
	}
	defer func() {
		rerr := os.RemoveAll(tmpdir)
		if rerr != nil {
			t.Errorf("failed to remove temporary directory %q: %s", tmpdir, rerr.Error())
		}
	}()

	// Build the test binary.
	config := &compileopts.Options{
		Target:     "",
		Opt:        "z",
		PrintIR:    false,
		DumpSSA:    false,
		VerifyIR:   true,
		Debug:      true,
		PrintSizes: "",
		WasmAbi:    "",
	}
	binary := filepath.Join(tmpdir, "test")
	err = runBuild("./testdata/environment.go", binary, config)
	if err != nil {
		printCompilerError(t.Log, err)
		t.Fail()
		return
	}

	// Run the test.
	cmd := exec.Command(binary, "arg1", "\targ2 \n")
	stdout := &bytes.Buffer{}
	cmd.Stdout = stdout
	cmd.Stderr = stdout
	err = cmd.Start()
	if err != nil {
		t.Fatal("failed to start:", err)
	}
	err = cmd.Wait()

	expected := fmt.Sprintf("args: 3\narg: %s\narg: arg1\narg: \targ2 \n", binary)

	// Munge the output a bit to make it easier to work with: putchar() prints
	// CRLF, convert it to LF.
	actual := strings.Replace(string(stdout.Bytes()), "\r\n", "\n", -1)
	actual = actual[:len(actual)-1] // remove trailing '\n' char

	// Check whether the command ran successfully, and print the actual output
	// if it differs.
	fail := false
	if err != nil {
		t.Log("failed to run:", err)
		fail = true
	} else if expected != actual {
		t.Log("output did not match")
		fail = true
	}
	if fail {
		r := bufio.NewReader(strings.NewReader(actual))
		for {
			line, err := r.ReadString('\n')
			if err != nil {
				break
			}
			t.Log("stdout:", line[:len(line)-1])
		}
		t.Fail()
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
			err := builder.RunTool(os.Args[1], os.Args[2:]...)
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
