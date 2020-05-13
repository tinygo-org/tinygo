package runtime

// The Error interface identifies a run time error.
type Error interface {
	error

	RuntimeError()
}
