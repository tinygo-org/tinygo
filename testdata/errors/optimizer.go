package main

import "runtime/interrupt"

var num = 5

func main() {
	// Error coming from LowerInterrupts.
	interrupt.New(num, func(interrupt.Interrupt) {
	})

	// 2nd error
	interrupt.New(num, func(interrupt.Interrupt) {
	})
}

// ERROR: # command-line-arguments
// ERROR: optimizer.go:9:15: interrupt ID is not a constant
// ERROR: optimizer.go:13:15: interrupt ID is not a constant
