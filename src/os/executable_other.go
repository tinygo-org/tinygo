//go:build !linux || baremetal || wasm_freestanding
// +build !linux baremetal wasm_freestanding

package os

import "errors"

func Executable() (string, error) {
	return "", errors.New("Executable not implemented")
}
