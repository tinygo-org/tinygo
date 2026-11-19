package main

// ERROR: can only use //go:weak in runtime or machine
//
//go:weak
func weakInMain() {}

// ERROR: can only use //go:weak on definitions
//
//go:weak
func weakDeclaration()

// ERROR: //go:weak does not accept parameters
//
//go:weak arg
func weakWithArgs() {}

func main() {
	weakInMain()
	weakDeclaration()
	weakWithArgs()
}
