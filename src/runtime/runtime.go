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

func GOMAXPROCS(n int) int {
	// Note: setting GOMAXPROCS is ignored.
	return 1
}

func stringequal(x, y string) bool {
	if len(x) != len(y) {
		return false
	}
	for i := 0; i < len(x); i++ {
		if x[i] != y[i] {
			return false
		}
	}
	return true
}

func _panic(message interface{}) {
	printstring("panic: ")
	printitf(message)
	printnl()
	abort()
}

// Check for bounds in *ssa.IndexAddr and *ssa.Lookup.
func lookupBoundsCheck(length, index int) {
	if index < 0 || index >= length {
		// printstring() here is safe as this function is excluded from bounds
		// checking.
		printstring("panic: runtime error: index out of range\n")
		abort()
	}
}

// Check for bounds in *ssa.Slice
func sliceBoundsCheck(length, low, high uint) {
	if !(0 <= low && low <= high && high <= length) {
		printstring("panic: runtime error: slice out of range\n")
		abort()
	}
}
