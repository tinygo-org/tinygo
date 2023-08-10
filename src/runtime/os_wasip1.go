//go:build wasip1

package runtime

// The actual GOOS=wasip1, as newly added in the Go 1.21 toolchain.
// Previously we supported -target=wasi, but that was essentially faked by using
// linux/arm instead because that was the closest thing that was already
// supported in the Go standard library.

const GOOS = "wasip1"
