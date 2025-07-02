// Package diagnostics formats compiler errors and prints them in a consistent
// way.
package diagnostics

import (
	"bytes"
	"fmt"
	"go/scanner"
	"go/token"
	"go/types"
	"io"
	"path/filepath"
	"reflect"
	"sort"
	"strings"

	"github.com/tinygo-org/tinygo/builder"
	"github.com/tinygo-org/tinygo/goenv"
	"github.com/tinygo-org/tinygo/interp"
	"github.com/tinygo-org/tinygo/loader"
)

// A single diagnostic.
type Diagnostic struct {
	Pos token.Position
	Msg string

	// Start and end position, if available. For many errors these positions are
	// not available, but for some they are.
	StartPos token.Position
	EndPos   token.Position
}

// One or multiple errors of a particular package.
// It can also represent whole-program errors (like linker errors) that can't
// easily be connected to a single package.
type PackageDiagnostic struct {
	ImportPath  string // the same ImportPath as in `go list -json`
	Diagnostics []Diagnostic
}

// Diagnostics of a whole program. This can include errors belonging to multiple
// packages, or just a single package.
type ProgramDiagnostic []PackageDiagnostic

// CreateDiagnostics reads the underlying errors in the error object and creates
// a set of diagnostics that's sorted and can be readily printed.
func CreateDiagnostics(err error) ProgramDiagnostic {
	if err == nil {
		return nil
	}
	// Right now, the compiler will only show errors for the first package that
	// fails to build. This is likely to change in the future.
	return ProgramDiagnostic{
		createPackageDiagnostic(err),
	}
}

// Create diagnostics for a single package (though, in practice, it may also be
// used for whole-program diagnostics in some cases).
func createPackageDiagnostic(err error) PackageDiagnostic {
	// Extract diagnostics for this package.
	var pkgDiag PackageDiagnostic
	switch err := err.(type) {
	case *builder.MultiError:
		if err.ImportPath != "" {
			pkgDiag.ImportPath = err.ImportPath
		}
		for _, err := range err.Errs {
			diags := createDiagnostics(err)
			pkgDiag.Diagnostics = append(pkgDiag.Diagnostics, diags...)
		}
	case loader.Errors:
		if err.Pkg != nil {
			pkgDiag.ImportPath = err.Pkg.ImportPath
		}
		for _, err := range err.Errs {
			diags := createDiagnostics(err)
			pkgDiag.Diagnostics = append(pkgDiag.Diagnostics, diags...)
		}
	case *interp.Error:
		pkgDiag.ImportPath = err.ImportPath
		w := &bytes.Buffer{}
		fmt.Fprintln(w, err.Error())
		if len(err.Inst) != 0 {
			fmt.Fprintln(w, err.Inst)
		}
		if len(err.Traceback) > 0 {
			fmt.Fprintln(w, "\ntraceback:")
			for _, line := range err.Traceback {
				fmt.Fprintln(w, line.Pos.String()+":")
				fmt.Fprintln(w, line.Inst)
			}
		}
		pkgDiag.Diagnostics = append(pkgDiag.Diagnostics, Diagnostic{
			Msg: w.String(),
		})
	default:
		pkgDiag.Diagnostics = createDiagnostics(err)
	}

	// Sort these diagnostics by file/line/column.
	sort.SliceStable(pkgDiag.Diagnostics, func(i, j int) bool {
		posI := pkgDiag.Diagnostics[i].Pos
		posJ := pkgDiag.Diagnostics[j].Pos
		if posI.Filename != posJ.Filename {
			return posI.Filename < posJ.Filename
		}
		if posI.Line != posJ.Line {
			return posI.Line < posJ.Line
		}
		return posI.Column < posJ.Column
	})

	return pkgDiag
}

// Extract diagnostics from the given error message and return them as a slice
// of errors (which in many cases will just be a single diagnostic).
func createDiagnostics(err error) []Diagnostic {
	switch err := err.(type) {
	case types.Error:
		diag := Diagnostic{
			Pos: err.Fset.Position(err.Pos),
			Msg: err.Msg,
		}
		// There is a special unexported API since Go 1.16 that provides the
		// range (start and end position) where the type error exists.
		// There is no promise of backwards compatibility in future Go versions
		// so we have to be extra careful here to be resilient.
		v := reflect.ValueOf(err)
		start := v.FieldByName("go116start")
		end := v.FieldByName("go116end")
		if start.IsValid() && end.IsValid() && start.Int() != end.Int() {
			diag.StartPos = err.Fset.Position(token.Pos(start.Int()))
			diag.EndPos = err.Fset.Position(token.Pos(end.Int()))
		}
		return []Diagnostic{diag}
	case scanner.Error:
		return []Diagnostic{
			{
				Pos: err.Pos,
				Msg: err.Msg,
			},
		}
	case scanner.ErrorList:
		var diags []Diagnostic
		for _, err := range err {
			diags = append(diags, createDiagnostics(*err)...)
		}
		return diags
	case loader.Error:
		if err.Err.Pos.Filename != "" {
			// Probably a syntax error in a dependency.
			return createDiagnostics(err.Err)
		} else {
			// Probably an "import cycle not allowed" error.
			buf := &bytes.Buffer{}
			fmt.Fprintln(buf, "package", err.ImportStack[0])
			for i := 1; i < len(err.ImportStack); i++ {
				pkgPath := err.ImportStack[i]
				if i == len(err.ImportStack)-1 {
					// last package
					fmt.Fprintln(buf, "\timports", pkgPath+": "+err.Err.Error())
				} else {
					// not the last package
					fmt.Fprintln(buf, "\timports", pkgPath)
				}
			}
			return []Diagnostic{
				{Msg: buf.String()},
			}
		}
	default:
		return []Diagnostic{
			{Msg: err.Error()},
		}
	}
}

// Write program diagnostics to the given writer with 'wd' as the relative
// working directory.
func (progDiag ProgramDiagnostic) WriteTo(w io.Writer, wd string) {
	for _, pkgDiag := range progDiag {
		pkgDiag.WriteTo(w, wd)
	}
}

// Write package diagnostics to the given writer with 'wd' as the relative
// working directory.
func (pkgDiag PackageDiagnostic) WriteTo(w io.Writer, wd string) {
	if pkgDiag.ImportPath != "" {
		fmt.Fprintln(w, "#", pkgDiag.ImportPath)
	}
	for _, diag := range pkgDiag.Diagnostics {
		diag.WriteTo(w, wd)
	}
}

// Write this diagnostic to the given writer with 'wd' as the relative working
// directory.
func (diag Diagnostic) WriteTo(w io.Writer, wd string) {
	if diag.Pos == (token.Position{}) {
		fmt.Fprintln(w, diag.Msg)
		return
	}
	pos := RelativePosition(diag.Pos, wd)
	fmt.Fprintf(w, "%s: %s\n", pos, diag.Msg)
}

// Convert the position in pos (assumed to have an absolute path) into a
// relative path if possible. Paths inside GOROOT/TINYGOROOT will remain
// absolute.
func RelativePosition(pos token.Position, wd string) token.Position {
	// Check whether we even have a working directory.
	if wd == "" {
		return pos
	}

	// Paths inside GOROOT should be printed in full.
	if strings.HasPrefix(pos.Filename, filepath.Join(goenv.Get("GOROOT"), "src")) || strings.HasPrefix(pos.Filename, filepath.Join(goenv.Get("TINYGOROOT"), "src")) {
		return pos
	}

	// Make the path relative, for easier reading. Ignore any errors in the
	// process (falling back to the absolute path).
	relpath, err := filepath.Rel(wd, pos.Filename)
	if err == nil {
		pos.Filename = relpath
	}
	return pos
}
