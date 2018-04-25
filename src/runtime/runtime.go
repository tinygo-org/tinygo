
package runtime

const Compiler = "tgo"

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
