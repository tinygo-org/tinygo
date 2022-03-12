package runtime

// NumCPU returns the number of logical CPUs usable by the current process.
//
// The set of available CPUs is checked by querying the operating system
// at process startup. Changes to operating system CPU allocation after
// process startup are not reflected.
func NumCPU() int {
	return 1
}

// NumCgoCall returns the number of cgo calls made by the current process.
func NumCgoCall() int {
	return 0
}

// NumGoroutine returns the number of goroutines that currently exist.
func NumGoroutine() int {
	return 1
}

func Version() string {
	return "v0.1.0"
}
