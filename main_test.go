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
	"testing"

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

	targets := []string{""} // always test the native host

	if !testing.Short() {
		targets = append(targets, "qemu") // emulated cortex-m3
	}

	if !testing.Short() && runtime.GOOS == "linux" {
		targets = append(targets,
			"arm--linux-gnueabihf", // emulated linux/arm
			"aarch64--linux-gnu",   // "linux/arm"
			"wasm",                 // "WebAssembly"
		)
	}

	for _, path := range matches {
		for _, target := range targets {
			testName := filepath.Join(target, path)
			t.Run(testName, func(t *testing.T) {
				runTest(path, target, t)
			})
		}
	}

}

func runTest(path string, target string, t *testing.T) {
	t.Parallel()

	if target == "arm--linux-gnueabihf" && path == "testdata/cgo/" {
		t.Skip("TODO: improve CGo")
	}

	if target == "aarch64--linux-gnu" && path == "testdata/cgo/" {
		t.Skip("TODO: improve CGo")
	}

	if target == "wasm" && path == "testdata/gc.go" {
		t.Skip("known to fail")
	}

	tmpFile, err := ioutil.TempFile("", "tinygo-test-build")
	if err != nil {
		t.Fatal("could not create temporary file:", err)
	}
	defer os.Remove(tmpFile.Name())

	// Get the expected output for this test.
	txtpath := path[:len(path)-3] + ".txt"
	if path[len(path)-1] == os.PathSeparator {
		txtpath = path + "out.txt"
	}
	f, err := os.Open(txtpath)
	if err != nil {
		t.Fatal("could not open expected output file:", err)
	}
	expected, err := ioutil.ReadAll(f)
	if err != nil {
		t.Fatal("could not read expected output file:", err)
	}

	// Build the test binary.
	config := &BuildConfig{
		opt:        "z",
		printIR:    false,
		dumpSSA:    false,
		debug:      false,
		printSizes: "",
		wasmAbi:    "js",
	}

	binary := tmpFile.Name()
	err = Build("./"+path, binary, target, config)
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
		spec, err := LoadTarget(target)
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
