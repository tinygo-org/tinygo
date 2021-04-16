package main

import (
	"os"
)

func main() {
	// Check for environment variables (set by the test runner).
	println("ENV1:", os.Getenv("ENV1"))
	v, ok := os.LookupEnv("ENV2")
	if !ok {
		println("ENV2 not found")
	}
	println("ENV2:", v)

	// Check for command line arguments.
	// Argument 0 is skipped because it is the program name, which varies by
	// test run.
	println()
	for _, arg := range os.Args[1:] {
		println("arg:", arg)
	}
}
