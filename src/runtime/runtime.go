
package runtime

const Compiler = "tgo"

// The bitness of the CPU (e.g. 8, 32, 64). Set by the compiler as a constant.
var TargetBits uint8

func Sleep(d Duration) {
	// This function is treated specially by the compiler: when goroutines are
	// used, it is transformed into a llvm.coro.suspend() call.
	// When goroutines are not used this function behaves as normal.
	sleep(d)
}

func _panic(message interface{}) {
	printstring("panic: ")
	printitf(message)
	printnl()
	abort()
}

func boundsCheck(outOfRange bool) {
	if outOfRange {
		// printstring() here is safe as this function is excluded from bounds
		// checking.
		printstring("panic: runtime error: index out of range\n")
		abort()
	}
}
