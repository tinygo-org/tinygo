//go:build baremetal || wasm || unknown_wasm
// +build baremetal wasm unknown_wasm

// This file emulates some file-related functions that are only available
// under a real operating system.

package syscall

func Getwd() (string, error) {
	return "", nil
}
