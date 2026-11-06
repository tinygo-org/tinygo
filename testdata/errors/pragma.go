package main

//go:noinline extra
func flagParams() {}

//go:noescape
func noescapeDef(p *int) {}

//go:inline
func inlineDecl()

//go:noheap
func noheapDecl()

//go:nobounds
func noboundsNoUnsafe() {}

//go:interrupt
func interruptNoUnsafe() {}

//go:section .text.foo
func sectionNoUnsafe() {}

//go:section
func sectionNoParam() {}

//go:noescape
func genericDef[T any]() {}

//go:align
var alignNoParam int

//go:align eight
var alignNotInt int

//go:section a b
var sectionParams int

//go:extern a b
var externParams int

//go:section .text.foo
func sectionDecl()

func main() {
	flagParams()
	noescapeDef(nil)
	inlineDecl()
	noheapDecl()
	noboundsNoUnsafe()
	interruptNoUnsafe()
	sectionNoUnsafe()
	sectionNoParam()
	genericDef[int]()
	genericDef[string]()
	alignNoParam++
	alignNotInt++
	sectionParams++
	externParams++
	sectionDecl()
}

// ERROR: # command-line-arguments
// ERROR: pragma.go:3:1: //go:noinline does not accept parameters
// ERROR: pragma.go:6:1: can only use //go:noescape on declarations
// ERROR: pragma.go:9:1: can only use //go:inline on definitions
// ERROR: pragma.go:12:1: can only use //go:noheap on definitions
// ERROR: pragma.go:15:1: can only use //go:nobounds in packages that import unsafe
// ERROR: pragma.go:18:1: can only use //go:interrupt in packages that import unsafe
// ERROR: pragma.go:21:1: can only use //go:section in packages that import unsafe
// ERROR: pragma.go:24:1: expected one parameter to //go:section, not 0
// ERROR: pragma.go:27:1: can only use //go:noescape on declarations
// ERROR: pragma.go:30:1: expected one parameter to //go:align, not 0
// ERROR: pragma.go:33:1: //go:align parameter must be an integer
// ERROR: pragma.go:36:1: expected one parameter to //go:section, not 2
// ERROR: pragma.go:39:1: expected at most one parameter to //go:extern, not 2
// ERROR: pragma.go:42:1: can only use //go:section on definitions
