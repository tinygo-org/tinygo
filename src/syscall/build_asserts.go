//go:build syscall_asserts

package syscall

// enable assertions for checking Environ during runtime preinitialization.
// This is to catch people using os.Environ() during package inits with wizer.
const panicOnEnvironDuringInitAll = true
