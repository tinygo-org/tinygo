package loader

import (
	"go/token"
	"strings"
)

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
	Packages        []string
	ImportPositions []token.Position
}

func (e *ImportCycleError) Error() string {
	var msg strings.Builder
	msg.WriteString("import cycle:\n\t")
	msg.WriteString(strings.Join(e.Packages, "\n\t"))
	msg.WriteString("\n at ")
	for i, pos := range e.ImportPositions {
		if i > 0 {
			msg.WriteString(", ")
		}
		msg.WriteString(pos.String())
	}
	return msg.String()
}
