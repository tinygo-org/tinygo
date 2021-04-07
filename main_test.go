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
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/tinygo-org/tinygo/builder"
	"github.com/tinygo-org/tinygo/compileopts"
	"github.com/tinygo-org/tinygo/goenv"
)

const TESTDATA = "testdata"

var testTarget = flag.String("target", "", "override test target")

func TestCompiler(t *testing.T) {
	matches, err := filepath.Glob(filepath.Join(TESTDATA, "*.go"))
	if err != nil {
		t.Fatal("could not read test files:", err)
	}

	dirMatches, err := filepath.Glob(filepath.Join(TESTDATA, "*", "main.go"))
	if err != nil {
		t.Fatal("could not read test packages:", err)
	}
	if len(matches) == 0 || len(dirMatches) == 0 {
		t.Fatal("no test files found")
	}
	for _, m := range dirMatches {
		matches = append(matches, filepath.Dir(m)+string(filepath.Separator))
	}

	sort.Strings(matches)

	if *testTarget != "" {
		// This makes it possible to run one specific test (instead of all),
		// which is especially useful to quickly check whether some changes
		// affect a particular target architecture.
		runPlatTests(*testTarget, matches, t)
		return
	}

	if runtime.GOOS != "windows" {
		t.Run("Host", func(t *testing.T) {
			runPlatTests("", matches, t)
			if runtime.GOOS == "darwin" {
				runTest("testdata/libc/filesystem.go", "", t,
					nil, nil)
				runTest("testdata/libc/env.go", "", t,
					[]string{"ENV1=VALUE1", "ENV2=VALUE2"}, nil)
			}
		})
	}

	if testing.Short() {
		return
	}

	t.Run("EmulatedCortexM3", func(t *testing.T) {
		runPlatTests("cortex-m-qemu", matches, t)
	})

	if runtime.GOOS == "windows" || runtime.GOOS == "darwin" {
		// Note: running only on Windows and macOS because Linux (as of 2020)
		// usually has an outdated QEMU version that doesn't support RISC-V yet.
		t.Run("EmulatedRISCV", func(t *testing.T) {
			runPlatTests("riscv-qemu", matches, t)
		})
	}

	if runtime.GOOS == "linux" {
		t.Run("X86Linux", func(t *testing.T) {
			runPlatTests("i386--linux-gnu", matches, t)
		})
		t.Run("ARMLinux", func(t *testing.T) {
			runPlatTests("arm--linux-gnueabihf", matches, t)
		})
		t.Run("ARM64Linux", func(t *testing.T) {
			runPlatTests("aarch64--linux-gnu", matches, t)
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
				runPlatTests("wasm", matches, t)
			})
		}

		t.Run("WASI", func(t *testing.T) {
			runPlatTests("wasi", matches, t)
			runTest("testdata/libc/env.go", "wasi", t,
				[]string{"--env", "ENV1=VALUE1", "--env", "ENV2=VALUE2"}, nil)
			runTest("testdata/libc/filesystem.go", "wasi", t, nil, []string{"--dir=."})
		})
	}
}

func runPlatTests(target string, matches []string, t *testing.T) {
	t.Parallel()

	for _, path := range matches {
		path := path // redefine to avoid race condition
		t.Run(filepath.Base(path), func(t *testing.T) {
			t.Parallel()
			runTest(path, target, t, nil, nil)
		})
	}
}

func runTest(path, target string, t *testing.T, environmentVars []string, additionalArgs []string) {
	// Get the expected output for this test.
	txtpath := path[:len(path)-3] + ".txt"
	if path[len(path)-1] == os.PathSeparator {
		txtpath = path + "out.txt"
	}
	expected, err := ioutil.ReadFile(txtpath)
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
	err = Build("./"+path, binary, config)
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
		cmd.Env = append(cmd.Env, environmentVars...)
	} else {
		spec, err := compileopts.LoadTarget(target)
		if err != nil {
			t.Fatal("failed to load target spec:", err)
		}
		if len(spec.Emulator) == 0 {
			cmd = exec.Command(binary)
		} else {
			args := append(spec.Emulator[1:], binary)
			cmd = exec.Command(spec.Emulator[0], append(args, additionalArgs...)...)
		}

		if len(spec.Emulator) != 0 && spec.Emulator[0] == "wasmtime" {
			for _, v := range environmentVars {
				cmd.Args = append(cmd.Args, "--env", v)
			}
		} else {
			cmd.Env = append(cmd.Env, environmentVars...)
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
