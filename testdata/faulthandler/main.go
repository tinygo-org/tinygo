package main

import "runtime"

func main() {
	println("registering fault handler...")
	runtime.SetFaultHandler(func(info runtime.FaultInfo) {
		println("trap instruction was caught")
		abort()
	})

	// Try running an invalid instruction.
	//
	println("running trap instruction...")
	callTrap()

	// This line should be unreachable.
	println("this line should not be printed")
}
