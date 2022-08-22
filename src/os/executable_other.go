//go:build !linux || baremetal || unknown_wasm
// +build !linux baremetal unknown_wasm

package os

import "errors"

func Executable() (string, error) {
	return "", errors.New("Executable not implemented")
}
