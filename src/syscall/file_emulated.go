//go:build baremetal || (wasm && !wasip1) || wasm_unknown

// This file emulates some file-related functions that are only available
// under a real operating system.

package syscall

func Getwd() (string, error) {
	return "", nil
}
