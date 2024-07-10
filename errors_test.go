package main

import (
	"bytes"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/tinygo-org/tinygo/compileopts"
	"github.com/tinygo-org/tinygo/diagnostics"
)

// Test the error messages of the TinyGo compiler.
func TestErrors(t *testing.T) {
	// TODO: nicely formatted error messages for:
	//   - duplicate symbols in ld.lld (currently only prints bitcode file)
	type errorTest struct {
		name   string
		target string
	}
	for _, tc := range []errorTest{
		{name: "cgo"},
		{name: "compiler"},
		{name: "interp"},
		{name: "linker-flashoverflow", target: "cortex-m-qemu"},
		{name: "linker-ramoverflow", target: "cortex-m-qemu"},
		{name: "linker-undefined", target: "darwin/arm64"},
		{name: "linker-undefined", target: "linux/amd64"},
		//{name: "linker-undefined", target: "windows/amd64"}, // TODO: no source location
		{name: "linker-undefined", target: "cortex-m-qemu"},
		//{name: "linker-undefined", target: "wasip1"}, // TODO: no source location
		{name: "loader-importcycle"},
		{name: "loader-invaliddep"},
		{name: "loader-invalidpackage"},
		{name: "loader-nopackage"},
		{name: "optimizer"},
		{name: "syntax"},
		{name: "types"},
	} {
		name := tc.name
		if tc.target != "" {
			name += "#" + tc.target
		}
		target := tc.target
		if target == "" {
			target = "wasip1"
		}
		t.Run(name, func(t *testing.T) {
			options := optionsFromTarget(target, sema)
			testErrorMessages(t, "./testdata/errors/"+tc.name+".go", &options)
		})
	}
}

func testErrorMessages(t *testing.T, filename string, options *compileopts.Options) {
	t.Parallel()

	// Parse expected error messages.
	expected := readErrorMessages(t, filename)

	// Try to build a binary (this should fail with an error).
	tmpdir := t.TempDir()
	err := Build(filename, tmpdir+"/out", options)
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
	diagnostics.CreateDiagnostics(err).WriteTo(&buf, wd)
	actual := strings.TrimRight(buf.String(), "\n")

	// Check whether the error is as expected.
	if !matchErrors(t, expected, actual) {
		t.Errorf("expected error:\n%s\ngot:\n%s", indentText(expected, "> "), indentText(actual, "> "))
	}
}

func matchErrors(t *testing.T, pattern, actual string) bool {
	patternLines := strings.Split(pattern, "\n")
	actualLines := strings.Split(actual, "\n")
	if len(patternLines) != len(actualLines) {
		return false
	}
	for i, patternLine := range patternLines {
		indices := regexp.MustCompile(`\{\{.*?\}\}`).FindAllStringIndex(patternLine, -1)
		patternParts := []string{"^"}
		lastStop := 0
		for _, startstop := range indices {
			start := startstop[0]
			stop := startstop[1]
			patternParts = append(patternParts,
				regexp.QuoteMeta(patternLine[lastStop:start]),
				patternLine[start+2:stop-2])
			lastStop = stop
		}
		patternParts = append(patternParts, regexp.QuoteMeta(patternLine[lastStop:]), "$")
		pattern := strings.Join(patternParts, "")
		re, err := regexp.Compile(pattern)
		if err != nil {
			t.Fatalf("could not compile regexp for %#v: %v", patternLine, err)
		}
		if !re.MatchString(actualLines[i]) {
			return false
		}
	}
	return true
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
