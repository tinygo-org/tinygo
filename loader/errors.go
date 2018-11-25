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

// ImportCycleErrors is returned when encountering an import cycle. The list of
// packages is a list from the root package to the leaf package that imports one
// of the packages in the list.
type ImportCycleError struct {
	Packages []string
}

func (e *ImportCycleError) Error() string {
	msg := "import cycle: " + e.Packages[0]
	for _, path := range e.Packages[1:] {
		msg += " → " + path
	}
	return msg
}
