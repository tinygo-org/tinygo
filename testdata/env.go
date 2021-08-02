package main

import (
	"crypto/rand"
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

	// Check for crypto/rand support.
	checkRand()
}

func checkRand() {
	buf := make([]byte, 500)
	n, err := rand.Read(buf)
	if n != len(buf) || err != nil {
		println("could not read random numbers:", err)
	}

	// Very simple test that random numbers are at least somewhat random.
	sum := 0
	for _, b := range buf {
		sum += int(b)
	}
	if sum < 95*len(buf) || sum > 159*len(buf) {
		println("random numbers don't seem that random, the average byte is", sum/len(buf))
	} else {
		println("random number check was successful")
	}
}
