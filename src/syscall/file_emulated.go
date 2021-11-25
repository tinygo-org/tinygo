// +build baremetal wasm

// This file emulates some file-related functions that are only available
// under a real operating system.

package syscall

func Getwd() (string, error) {
	return "", nil
}
