package transform

// This file defines some helper functions for testing transforms.

import (
	"flag"
	"io/ioutil"
	"os"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"tinygo.org/x/go-llvm"
)

var update = flag.Bool("update", false, "update transform package tests")

// testTransform runs a transformation pass on an input file (pathPrefix+".ll")
// and checks whether it matches the expected output (pathPrefix+".out.ll"). The
// output is compared with a fuzzy match that ignores some irrelevant lines such
// as empty lines.
func testTransform(t *testing.T, pathPrefix string, transform func(mod llvm.Module)) {
	// Read the input IR.
	ctx := llvm.NewContext()
	buf, err := llvm.NewMemoryBufferFromFile(pathPrefix + ".ll")
	os.Stat(pathPrefix + ".ll") // make sure this file is tracked by `go test` caching
	if err != nil {
		t.Fatalf("could not read file %s: %v", pathPrefix+".ll", err)
	}
	mod, err := ctx.ParseIR(buf)
	if err != nil {
		t.Fatalf("could not load module:\n%v", err)
	}

	// Perform the transform.
	transform(mod)

	// Get the output from the test and filter some irrelevant lines.
	actual := mod.String()
	actual = actual[strings.Index(actual, "\ntarget datalayout = ")+1:]

	if *update {
		err := ioutil.WriteFile(pathPrefix+".out.ll", []byte(actual), 0666)
		if err != nil {
			t.Error("failed to write out new output:", err)
		}
	} else {
		// Read the expected output IR.
		out, err := ioutil.ReadFile(pathPrefix + ".out.ll")
		if err != nil {
			t.Fatalf("could not read output file %s: %v", pathPrefix+".out.ll", err)
		}

		// See whether the transform output matches with the expected output IR.
		expected := string(out)
		if !fuzzyEqualIR(expected, actual) {
			t.Logf("output does not match expected output:\n%s", actual)
			t.Fail()
		}
	}
}

var alignRegexp = regexp.MustCompile(", align [0-9]+$")

// fuzzyEqualIR returns true if the two LLVM IR strings passed in are roughly
// equal. That means, only relevant lines are compared (excluding comments
// etc.).
func fuzzyEqualIR(s1, s2 string) bool {
	lines1 := filterIrrelevantIRLines(strings.Split(s1, "\n"))
	lines2 := filterIrrelevantIRLines(strings.Split(s2, "\n"))
	if len(lines1) != len(lines2) {
		return false
	}
	for i, line1 := range lines1 {
		line2 := lines2[i]
		match1 := alignRegexp.MatchString(line1)
		match2 := alignRegexp.MatchString(line2)
		if match1 != match2 {
			// Only one of the lines has the align keyword. Remove it.
			// This is a change to make the test work in both LLVM 10 and LLVM
			// 11 (LLVM 11 appears to automatically add alignment everywhere).
			line1 = alignRegexp.ReplaceAllString(line1, "")
			line2 = alignRegexp.ReplaceAllString(line2, "")
		}
		if line1 != line2 {
			return false
		}
	}

	return true
}

// filterIrrelevantIRLines removes lines from the input slice of strings that
// are not relevant in comparing IR. For example, empty lines and comments are
// stripped out.
func filterIrrelevantIRLines(lines []string) []string {
	var out []string
	llvmVersion, err := strconv.Atoi(strings.Split(llvm.Version, ".")[0])
	if err != nil {
		// Note: this should never happen and if it does, it will always happen
		// for a particular build because llvm.Version is a constant.
		panic(err)
	}
	for _, line := range lines {
		line = strings.Split(line, ";")[0]    // strip out comments/info
		line = strings.TrimRight(line, "\r ") // drop '\r' on Windows and remove trailing spaces from comments
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "source_filename = ") {
			continue
		}
		if llvmVersion < 10 && strings.HasPrefix(line, "attributes ") {
			// Ignore attribute groups. These may change between LLVM versions.
			// Right now test outputs are for LLVM 10.
			continue
		}
		if llvmVersion < 10 && strings.HasPrefix(line, "define ") {
			// Remove parameter values such as %0 in function definitions. These
			// were added in LLVM 10 so to get the tests to pass on older
			// versions, ignore them there (there are other tests that verify
			// correct behavior).
			re := regexp.MustCompile(` %[0-9]+(\)|,)`)
			line = re.ReplaceAllString(line, "$1")
		}
		out = append(out, line)
	}
	return out
}
