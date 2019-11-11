package builder

// MultiError is a list of multiple errors (actually: diagnostics) returned
// during LLVM IR generation.
type MultiError struct {
	Errs []error
}

func (e *MultiError) Error() string {
	// Return the first error, to conform to the error interface. Clients should
	// really do a type-assertion on *MultiError.
	return e.Errs[0].Error()
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
