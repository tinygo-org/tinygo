package runtime

// NumCPU returns the number of logical CPUs usable by the current process.
//
// The set of available CPUs is checked by querying the operating system
// at process startup. Changes to operating system CPU allocation after
// process startup are not reflected.
func NumCPU() int {
	return 1
}

// Stub for NumCgoCall, does not return the real value
func NumCgoCall() int {
	return 0
}

// Stub for NumGoroutine, does not return the real value
func NumGoroutine() int {
	return 1
}
