//go:build !syscall_asserts

package syscall

// enable assertions for checking Environ during runtime preinitialization.
// This is to catch people using os.Environ() during package init with wizer.
const panicOnEnvironDuringInitAll = false
