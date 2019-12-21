package main

// This file tests the compiler by running Go files in testdata/*.go and
// comparing their output with the expected output in testdata/*.txt.

import (
	"bufio"
	"bytes"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"sync"
	"testing"

	"github.com/tinygo-org/tinygo/compileopts"
	"github.com/tinygo-org/tinygo/loader"
)

const TESTDATA = "testdata"

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

	if runtime.GOOS != "windows" {
		t.Run("Host", func(t *testing.T) {
			runPlatTests("", matches, t)
		})
	}

	if testing.Short() {
		return
	}

	t.Run("EmulatedCortexM3", func(t *testing.T) {
		runPlatTests("cortex-m-qemu", matches, t)
	})

	if runtime.GOOS == "linux" {
		t.Run("ARMLinux", func(t *testing.T) {
			runPlatTests("arm--linux-gnueabihf", matches, t)
		})
		t.Run("ARM64Linux", func(t *testing.T) {
			runPlatTests("aarch64--linux-gnu", matches, t)
		})
		t.Run("WebAssembly", func(t *testing.T) {
			runPlatTests("wasm", matches, t)
		})
	}
}

func runPlatTests(target string, matches []string, t *testing.T) {
	t.Parallel()

	for _, path := range matches {
		switch {
		case target == "wasm":
			// testdata/gc.go is known not to work on WebAssembly
			if path == filepath.Join("testdata", "gc.go") {
				continue
			}
		case target == "":
			// run all tests on host
		case target == "cortex-m-qemu":
			// all tests are supported
		default:
			// cross-compilation of cgo is not yet supported
			if path == filepath.Join("testdata", "cgo")+string(filepath.Separator) {
				continue
			}
		}

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

func runTest(path, target string, t *testing.T) {
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
		Debug:      false,
		PrintSizes: "",
		WasmAbi:    "js",
	}
	binary := filepath.Join(tmpdir, "test")
	err = runBuild("./"+path, binary, config)
	if err != nil {
		if errLoader, ok := err.(loader.Errors); ok {
			for _, err := range errLoader.Errs {
				t.Log("failed to build:", err)
			}
		} else {
			t.Log("failed to build:", err)
		}
		t.Fail()
		return
	}

	// Run the test.
	var cmd *exec.Cmd
	if target == "" {
		cmd = exec.Command(binary)
	} else {
		spec, err := compileopts.LoadTarget(target)
		if err != nil {
			t.Fatal("failed to load target spec:", err)
		}
		if len(spec.Emulator) == 0 {
			t.Fatal("no emulator available for target:", target)
		}
		args := append(spec.Emulator[1:], binary)
		cmd = exec.Command(spec.Emulator[0], args...)
	}
	stdout := &bytes.Buffer{}
	cmd.Stdout = stdout
	if target != "" {
		cmd.Stderr = os.Stderr
	}
	err = cmd.Run()
	if _, ok := err.(*exec.ExitError); ok && target != "" {
		err = nil // workaround for QEMU
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
