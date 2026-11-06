package main

//go:noruntime
func missing()

//go:noruntime unexpected
func invalid() {}

func main() {
	missing()
	invalid()
}

// ERROR: # command-line-arguments
// ERROR: noruntime.go:3:1: can only use //go:noruntime on definitions
// ERROR: noruntime.go:6:1: //go:noruntime does not accept parameters
