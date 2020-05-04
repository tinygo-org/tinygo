package loader

// Errors contains a list of parser errors or a list of typechecker errors for
// the given package.
type Errors struct {
	Pkg  *Package
	Errs []error
}

func (e Errors) Error() string {
	return "could not compile: " + e.Errs[0].Error()
}
