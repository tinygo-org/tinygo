package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/tinygo-org/tinygo/compileopts"
)

// Test the error messages of the TinyGo compiler.
func TestErrors(t *testing.T) {
	for _, name := range []string{
		"cgo",
		"syntax",
		"types",
	} {
		t.Run(name, func(t *testing.T) {
			testErrorMessages(t, "./testdata/errors/"+name+".go")
		})
	}
}

func testErrorMessages(t *testing.T, filename string) {
	// Parse expected error messages.
	expected := readErrorMessages(t, filename)

	// Try to build a binary (this should fail with an error).
	tmpdir := t.TempDir()
	err := Build(filename, tmpdir+"/out", &compileopts.Options{
		Target:        "wasip1",
		Semaphore:     sema,
		InterpTimeout: 180 * time.Second,
		Debug:         true,
		VerifyIR:      true,
		Opt:           "z",
	})
	if err == nil {
		t.Fatal("expected to get a compiler error")
	}

	// Get the full ./testdata/errors directory.
	wd, absErr := filepath.Abs("testdata/errors")
	if absErr != nil {
		t.Fatal(absErr)
	}

	// Write error message out as plain text.
	var buf bytes.Buffer
	printCompilerError(err, func(v ...interface{}) {
		fmt.Fprintln(&buf, v...)
	}, wd)
	actual := strings.TrimRight(buf.String(), "\n")

	// Check whether the error is as expected.
	if actual != expected {
		t.Errorf("expected error:\n%s\ngot:\n%s", indentText(expected, "> "), indentText(actual, "> "))
	}
}

// Indent the given text with a given indentation string.
func indentText(text, indent string) string {
	return indent + strings.ReplaceAll(text, "\n", "\n"+indent)
}

// Read "// ERROR:" prefixed messages from the given file.
func readErrorMessages(t *testing.T, file string) string {
	data, err := os.ReadFile(file)
	if err != nil {
		t.Fatal("could not read input file:", err)
	}

	var errors []string
	for _, line := range strings.Split(string(data), "\n") {
		if strings.HasPrefix(line, "// ERROR: ") {
			errors = append(errors, strings.TrimRight(line[len("// ERROR: "):], "\r\n"))
		}
	}
	return strings.Join(errors, "\n")
}
