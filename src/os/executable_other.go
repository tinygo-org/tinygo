//go:build !linux || baremetal || unknow_wasm
// +build !linux baremetal unknow_wasm

package os

import "errors"

func Executable() (string, error) {
	return "", errors.New("Executable not implemented")
}
