package builder

// MultiError is a list of multiple errors (actually: diagnostics) returned
// during LLVM IR generation.
type MultiError struct {
	ImportPath string
	Errs       []error
}

func (e *MultiError) Error() string {
	// Return the first error, to conform to the error interface. Clients should
	// really do a type-assertion on *MultiError.
	return e.Errs[0].Error()
}

// newMultiError returns a *MultiError if there is more than one error, or
// returns that error directly when there is only one. Passing an empty slice
// will lead to a panic.
// The importPath may be passed if this error is for a single package.
func newMultiError(errs []error, importPath string) error {
	switch len(errs) {
	case 0:
		panic("attempted to create empty MultiError")
	case 1:
		return errs[0]
	default:
		return &MultiError{importPath, errs}
	}
}

// commandError is an error type to wrap os/exec.Command errors. This provides
// some more information regarding what went wrong while running a command.
type commandError struct {
	Msg  string
	File string
	Err  error
}

func (e *commandError) Error() string {
	return e.Msg + " " + e.File + ": " + e.Err.Error()
}
