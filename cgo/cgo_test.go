package cgo

import (
	"bytes"
	"flag"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"go/types"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"
)

// Pass -update to go test to update the output of the test files.
var flagUpdate = flag.Bool("update", false, "Update images based on test output.")

// normalizeResult normalizes Go source code that comes out of tests across
// platforms and Go versions.
func normalizeResult(t *testing.T, result string) string {
	result = strings.ReplaceAll(result, "\r\n", "\n")

	// This changed to 'undefined:', in Go 1.20.
	result = strings.ReplaceAll(result, ": undeclared name:", ": undefined:")
	// Go 1.20 added a bit more detail
	result = regexp.MustCompile(`(unknown field z in struct literal).*`).ReplaceAllString(result, "$1")

	return result
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
			// Read the AST in memory.
			path := filepath.Join("testdata", name+".go")
			fset := token.NewFileSet()
			f, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
			if err != nil {
				t.Fatal("could not parse Go source file:", err)
			}

			// Process the AST with CGo.
			cgoAST, _, _, _, _, cgoErrors := Process([]*ast.File{f}, "testdata", "main", fset, cflags, "")

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
			actual := normalizeResult(t, buf.String())

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

func Test_cgoPackage_isEquivalentAST(t *testing.T) {
	fieldA := &ast.Field{Type: &ast.BasicLit{Kind: token.STRING, Value: "a"}}
	fieldB := &ast.Field{Type: &ast.BasicLit{Kind: token.STRING, Value: "b"}}
	listOfFieldA := &ast.FieldList{List: []*ast.Field{fieldA}}
	listOfFieldB := &ast.FieldList{List: []*ast.Field{fieldB}}
	funcDeclA := &ast.FuncDecl{Name: &ast.Ident{Name: "a"}, Type: &ast.FuncType{Params: &ast.FieldList{}, Results: listOfFieldA}}
	funcDeclB := &ast.FuncDecl{Name: &ast.Ident{Name: "b"}, Type: &ast.FuncType{Params: &ast.FieldList{}, Results: listOfFieldB}}
	funcDeclNoResults := &ast.FuncDecl{Name: &ast.Ident{Name: "C"}, Type: &ast.FuncType{Params: &ast.FieldList{}}}

	testCases := []struct {
		name     string
		a, b     ast.Node
		expected bool
	}{
		{
			name:     "both nil",
			expected: true,
		},
		{
			name:     "not same type",
			a:        fieldA,
			b:        &ast.FuncDecl{},
			expected: false,
		},
		{
			name:     "Field same",
			a:        fieldA,
			b:        fieldA,
			expected: true,
		},
		{
			name:     "Field different",
			a:        fieldA,
			b:        fieldB,
			expected: false,
		},
		{
			name:     "FuncDecl Type Results nil",
			a:        funcDeclNoResults,
			b:        funcDeclNoResults,
			expected: true,
		},
		{
			name:     "FuncDecl Type Results same",
			a:        funcDeclA,
			b:        funcDeclA,
			expected: true,
		},
		{
			name:     "FuncDecl Type Results different",
			a:        funcDeclA,
			b:        funcDeclB,
			expected: false,
		},
		{
			name:     "FuncDecl Type Results a nil",
			a:        funcDeclNoResults,
			b:        funcDeclB,
			expected: false,
		},
		{
			name:     "FuncDecl Type Results b nil",
			a:        funcDeclA,
			b:        funcDeclNoResults,
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			p := &cgoPackage{}
			if got := p.isEquivalentAST(tc.a, tc.b); tc.expected != got {
				t.Errorf("expected %v, got %v", tc.expected, got)
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
