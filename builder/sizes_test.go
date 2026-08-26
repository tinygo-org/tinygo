package builder

import (
	"flag"
	"fmt"
	"os"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/tinygo-org/tinygo/compileopts"
)

var sema = make(chan struct{}, runtime.NumCPU())

var flagUpdate = flag.Bool("update", false, "update builder package tests")

type sizeTest struct {
	target string
	path   string
}

// Test whether code and data size is as expected for the given targets.
// This tests both the logic of loadProgramSize and checks that code size
// doesn't change unintentionally.
//
// If you find that code or data size is reduced, then great! You can update the
// golden file by passing -update to the test.
// If you find that the code or data size is increased, take a look as to why
// this is. It could be due to an update (LLVM version, Go version, etc) which
// is fine, but it could also mean that a recent change introduced this size
// increase. If so, please consider whether this new feature is indeed worth the
// size increase for all users.
func TestBinarySize(t *testing.T) {
	if runtime.GOOS == "linux" && !hasBuiltinTools {
		// Debian LLVM packages are modified a bit and tend to produce
		// different machine code. Ideally we'd fix this (with some attributes
		// or something?), but for now skip it.
		t.Skip("Skip: using external LLVM version so binary size might differ")
	}

	// This is a small number of very diverse targets that we want to test.
	tests := []sizeTest{
		// microcontrollers
		{"hifive1b", "examples/echo"},
		{"microbit", "examples/serial"},
		{"wioterminal", "examples/pininterrupt"},

		// TODO: also check wasm. Right now this is difficult, because
		// wasm binaries are run through wasm-opt and therefore the
		// output varies by binaryen version.
	}
	sizes := measureBinarySizes(t, tests)
	output := formatSizeTable(tests, sizes)
	checkGolden(t, "testdata/binary-size.txt", output)
}

func checkGolden(t *testing.T, path, actual string) {
	t.Helper()
	if *flagUpdate {
		if err := os.WriteFile(path, []byte(actual), 0o666); err != nil {
			t.Fatal("failed to write updated golden file:", err)
		}
		return
	}
	expected, err := os.ReadFile(path)
	if err != nil {
		t.Fatal("failed to read golden file:", err)
	}
	if actual != string(expected) {
		t.Errorf("%s does not match expected output (re-run with -update to regenerate):\nexpected:\n%sactual:\n%s", path, expected, actual)
	}
}

func measureBinarySizes(t *testing.T, tests []sizeTest) []*programSize {
	t.Helper()
	type result struct {
		index int
		size  *programSize
		err   error
	}

	results := make(chan result, len(tests))
	for i, tc := range tests {
		tmpdir := t.TempDir()
		go func() {
			size, err := measureBinarySize(tc, tmpdir)
			results <- result{i, size, err}
		}()
	}

	sizes := make([]*programSize, len(tests))
	failed := false
	for range tests {
		result := <-results
		if result.err != nil {
			tc := tests[result.index]
			t.Errorf("%s/%s: %v", tc.target, tc.path, result.err)
			failed = true
		}
		sizes[result.index] = result.size
	}
	if failed {
		t.FailNow()
	}
	return sizes
}

func measureBinarySize(tc sizeTest, tmpdir string) (*programSize, error) {
	result, err := buildBinaryInDir(tc.target, tc.path, tmpdir)
	if err != nil {
		return nil, err
	}
	size, err := loadProgramSize(result.Executable, nil)
	if err != nil {
		return nil, fmt.Errorf("could not read program size: %w", err)
	}
	return size, nil
}

func formatSizeTable(tests []sizeTest, sizes []*programSize) string {
	targetWidth := len("target")
	packageWidth := len("package")
	codeWidth := len("code")
	rodataWidth := len("rodata")
	dataWidth := len("data")
	bssWidth := len("bss")
	for i, tc := range tests {
		targetWidth = max(targetWidth, len(tc.target))
		packageWidth = max(packageWidth, len(tc.path))
		codeWidth = max(codeWidth, len(strconv.FormatUint(sizes[i].Code, 10)))
		rodataWidth = max(rodataWidth, len(strconv.FormatUint(sizes[i].ROData, 10)))
		dataWidth = max(dataWidth, len(strconv.FormatUint(sizes[i].Data, 10)))
		bssWidth = max(bssWidth, len(strconv.FormatUint(sizes[i].BSS, 10)))
	}

	var output strings.Builder
	fmt.Fprintf(&output, "%-*s %-*s %*s %*s %*s %*s\n",
		targetWidth, "target", packageWidth, "package",
		codeWidth, "code", rodataWidth, "rodata", dataWidth, "data", bssWidth, "bss")
	for i, tc := range tests {
		size := sizes[i]
		fmt.Fprintf(&output, "%-*s %-*s %*d %*d %*d %*d\n",
			targetWidth, tc.target, packageWidth, tc.path,
			codeWidth, size.Code, rodataWidth, size.ROData,
			dataWidth, size.Data, bssWidth, size.BSS)
	}
	return output.String()
}

// Check that the -size=full flag attributes binary size to the correct package
// without filesystem paths and things like that.
func TestSizeFull(t *testing.T) {
	tests := []string{
		"microbit",
		"wasip1",
	}

	libMatch := regexp.MustCompile(`^C [a-z -]+$`) // example: "C interrupt vector"
	pkgMatch := regexp.MustCompile(`^[a-z/]+$`)    // example: "internal/task"

	for _, target := range tests {
		t.Run(target, func(t *testing.T) {
			t.Parallel()

			// Build the binary.
			result := buildBinary(t, target, "examples/serial")

			// Check whether the binary doesn't contain any unexpected package
			// names.
			sizes, err := loadProgramSize(result.Executable, result.PackagePathMap)
			if err != nil {
				t.Fatal("could not read program size:", err)
			}
			for _, pkg := range sizes.sortedPackageNames() {
				if pkg == "(padding)" || pkg == "(unknown)" || pkg == "Go types" {
					// TODO: correctly attribute all unknown binary size.
					continue
				}
				if libMatch.MatchString(pkg) {
					continue
				}
				if pkgMatch.MatchString(pkg) {
					continue
				}
				t.Error("unexpected package name in size output:", pkg)
			}
		})
	}
}

func buildBinary(t *testing.T, targetString, pkgName string) BuildResult {
	t.Helper()
	result, err := buildBinaryInDir(targetString, pkgName, t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	return result
}

func buildBinaryInDir(targetString, pkgName, tmpdir string) (BuildResult, error) {
	options := compileopts.Options{
		Target:        targetString,
		Opt:           "z",
		Semaphore:     sema,
		InterpTimeout: 60 * time.Second,
		Debug:         true,
		VerifyIR:      true,
	}
	target, err := compileopts.LoadTarget(&options)
	if err != nil {
		return BuildResult{}, fmt.Errorf("could not load target: %w", err)
	}
	config := &compileopts.Config{
		Options: &options,
		Target:  target,
	}
	result, err := Build(pkgName, "", tmpdir, config)
	if err != nil {
		return BuildResult{}, fmt.Errorf("could not build: %w", err)
	}
	return result, nil
}
