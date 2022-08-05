package cgo

import (
	"bytes"
	"flag"
	"fmt"
	"go/ast"
	"go/build"
	"go/format"
	"go/parser"
	"go/token"
	"go/types"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// Pass -update to go test to update the output of the test files.
var flagUpdate = flag.Bool("update", false, "Update images based on test output.")

// normalizeResult normalizes Go source code that comes out of tests across
// platforms and Go versions.
func normalizeResult(result string) string {
	actual := strings.ReplaceAll(result, "\r\n", "\n")
	return actual
}

func TestCGo(t *testing.T) {
	var cflags = []string{"--target=armv6m-unknown-unknown-eabi"}

	for _, name := range []string{
		"basic",
		"errors",
		"types",
		"symbols",
		"flags",
		"const",
	} {
		name := name // avoid a race condition
		t.Run(name, func(t *testing.T) {
			// Skip tests that require specific Go version.
			if name == "errors" {
				ok := false
				for _, version := range build.Default.ReleaseTags {
					if version == "go1.16" {
						ok = true
						break
					}
				}
				if !ok {
					t.Skip("Results for errors test are only valid for Go 1.16+")
				}
			}

			// Read the AST in memory.
			path := filepath.Join("testdata", name+".go")
			fset := token.NewFileSet()
			f, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
			if err != nil {
				t.Fatal("could not parse Go source file:", err)
			}

			// Process the AST with CGo.
			cgoAST, _, _, _, _, cgoErrors := Process([]*ast.File{f}, "testdata", fset, cflags, "")

			// Check the AST for type errors.
			var typecheckErrors []error
			config := types.Config{
				Error: func(err error) {
					typecheckErrors = append(typecheckErrors, err)
				},
				Importer: simpleImporter{},
				Sizes:    types.SizesFor("gccgo", "arm"),
			}
			_, err = config.Check("", fset, []*ast.File{f, cgoAST}, nil)
			if err != nil && len(typecheckErrors) == 0 {
				// Only report errors when no type errors are found (an
				// unexpected condition).
				t.Error(err)
			}

			// Store the (formatted) output in a buffer. Format it, so it
			// becomes easier to read (and will hopefully change less with CGo
			// changes).
			buf := &bytes.Buffer{}
			if len(cgoErrors) != 0 {
				buf.WriteString("// CGo errors:\n")
				for _, err := range cgoErrors {
					buf.WriteString(formatDiagnostic(err))
				}
				buf.WriteString("\n")
			}
			if len(typecheckErrors) != 0 {
				buf.WriteString("// Type checking errors after CGo processing:\n")
				for _, err := range typecheckErrors {
					buf.WriteString(formatDiagnostic(err))
				}
				buf.WriteString("\n")
			}
			err = format.Node(buf, fset, cgoAST)
			if err != nil {
				t.Errorf("could not write out CGo AST: %v", err)
			}
			actual := normalizeResult(buf.String())

			// Read the file with the expected output, to compare against.
			outfile := filepath.Join("testdata", name+".out.go")
			expectedBytes, err := os.ReadFile(outfile)
			if err != nil {
				t.Fatalf("could not read expected output: %v", err)
			}
			expected := strings.ReplaceAll(string(expectedBytes), "\r\n", "\n")

			// Check whether the output is as expected.
			if expected != actual {
				// It is not. Test failed.
				if *flagUpdate {
					// Update the file with the expected data.
					err := os.WriteFile(outfile, []byte(actual), 0666)
					if err != nil {
						t.Error("could not write updated output file:", err)
					}
					return
				}
				t.Errorf("output did not match:\n%s", string(actual))
			}
		})
	}
}

// simpleImporter implements the types.Importer interface, but only allows
// importing the unsafe package.
type simpleImporter struct {
}

// Import implements the Importer interface. For testing usage only: it only
// supports importing the unsafe package.
func (i simpleImporter) Import(path string) (*types.Package, error) {
	switch path {
	case "unsafe":
		return types.Unsafe, nil
	default:
		return nil, fmt.Errorf("importer not implemented for package %s", path)
	}
}

// formatDiagnostics formats the error message to be an indented comment. It
// also fixes Windows path name issues (backward slashes).
func formatDiagnostic(err error) string {
	msg := err.Error()
	if runtime.GOOS == "windows" {
		// Fix Windows path slashes.
		msg = strings.ReplaceAll(msg, "testdata\\", "testdata/")
	}
	return "//     " + msg + "\n"
}
