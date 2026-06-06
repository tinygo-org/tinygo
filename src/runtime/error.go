package runtime

// The Error interface identifies a run time error.
type Error interface {
	error

	RuntimeError()
}

// plainError is a runtime.Error implementation for plain string messages.
type plainError string

func (e plainError) Error() string { return string(e) }
func (e plainError) RuntimeError() {}
