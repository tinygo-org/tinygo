package main

// This is the same as examples/serial, but it also imports the machine package.
// It is used as a smoke test for the machine package (for boards that don't
// have an on-board LED and therefore can't use examples/blinky1).

import (
	_ "machine" // smoke test for the machine package
	"time"
)

func main() {
	for {
		println("hello world!")
		time.Sleep(time.Second)
	}
}
