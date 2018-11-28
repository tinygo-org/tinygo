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
	"testing"
)

const TESTDATA = "testdata"

func TestCompiler(t *testing.T) {
	matches, err := filepath.Glob(TESTDATA + "/*.go")
	if err != nil {
		t.Fatal("could not read test files:", err)
	}
	if len(matches) == 0 {
		t.Fatal("no test files found")
	}

	// Create a temporary directory for test output files.
	tmpdir, err := ioutil.TempDir("", "tinygo-test")
	if err != nil {
		t.Fatal("could not create temporary directory:", err)
	}
	defer os.RemoveAll(tmpdir)

	t.Log("running tests on the host...")
	for _, path := range matches {
		t.Run(path, func(t *testing.T) {
			runTest(path, tmpdir, "", t)
		})
	}

	t.Log("running tests on the qemu target...")
	for _, path := range matches {
		t.Run(path, func(t *testing.T) {
			runTest(path, tmpdir, "qemu", t)
		})
	}
}

func runTest(path, tmpdir string, target string, t *testing.T) {
	// Get the expected output for this test.
	txtpath := path[:len(path)-3] + ".txt"
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
		initInterp: true,
	}
	binary := filepath.Join(tmpdir, "test")
	err = Build(path, binary, target, config)
	if err != nil {
		t.Log("failed to build:", err)
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
