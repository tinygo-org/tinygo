package runtime

// The Error interface identifies a run time error.
type Error interface {
	error

	// Method to indicate this is indeed a runtime error.
	RuntimeError()
}

type runtimeError struct {
	msg string
}

func (r runtimeError) Error() string {
	return r.msg
}

// Purely here to satisfy the Error interface.
func (r runtimeError) RuntimeError() {}

var (
	divideError   error = runtimeError{"runtime error: integer divide by zero"}
	lookupError   error = runtimeError{"runtime error: index out of range"}
	overflowError error = runtimeError{"runtime error: integer overflow"}
)
