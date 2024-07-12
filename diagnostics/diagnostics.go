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
	switch err := err.(type) {
	case *builder.MultiError:
		var diags ProgramDiagnostic
		for _, err := range err.Errs {
			diags = append(diags, createPackageDiagnostic(err))
		}
		return diags
	default:
		return ProgramDiagnostic{
			createPackageDiagnostic(err),
		}
	}
}

// Create diagnostics for a single package (though, in practice, it may also be
// used for whole-program diagnostics in some cases).
func createPackageDiagnostic(err error) PackageDiagnostic {
	var pkgDiag PackageDiagnostic
	switch err := err.(type) {
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
	// TODO: sort
	return pkgDiag
}

// Extract diagnostics from the given error message and return them as a slice
// of errors (which in many cases will just be a single diagnostic).
func createDiagnostics(err error) []Diagnostic {
	switch err := err.(type) {
	case types.Error:
		return []Diagnostic{
			{
				Pos: err.Fset.Position(err.Pos),
				Msg: err.Msg,
			},
		}
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
					// not the last pacakge
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
	pos := diag.Pos // make a copy
	if !strings.HasPrefix(pos.Filename, filepath.Join(goenv.Get("GOROOT"), "src")) && !strings.HasPrefix(pos.Filename, filepath.Join(goenv.Get("TINYGOROOT"), "src")) {
		// This file is not from the standard library (either the GOROOT or the
		// TINYGOROOT). Make the path relative, for easier reading.  Ignore any
		// errors in the process (falling back to the absolute path).
		pos.Filename = tryToMakePathRelative(pos.Filename, wd)
	}
	fmt.Fprintf(w, "%s: %s\n", pos, diag.Msg)
}

// try to make the path relative to the current working directory. If any error
// occurs, this error is ignored and the absolute path is returned instead.
func tryToMakePathRelative(dir, wd string) string {
	if wd == "" {
		return dir // working directory not found
	}
	relpath, err := filepath.Rel(wd, dir)
	if err != nil {
		return dir
	}
	return relpath
}
