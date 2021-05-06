package loader

import "go/scanner"

// Errors contains a list of parser errors or a list of typechecker errors for
// the given package.
type Errors struct {
	Pkg  *Package
	Errs []error
}

func (e Errors) Error() string {
	return "could not compile: " + e.Errs[0].Error()
}

// Error is a regular error but with an added import stack. This is especially
// useful for debugging import cycle errors.
type Error struct {
	ImportStack []string
	Err         scanner.Error
}

func (e Error) Error() string {
	return e.Err.Error()
}

// Error returned when loading a *Program for a test binary but no test files
// are present.
type NoTestFilesError struct {
	ImportPath string
}

func (e NoTestFilesError) Error() string {
	return "no test files"
}
