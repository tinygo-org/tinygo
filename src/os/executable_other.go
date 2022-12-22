//go:build !linux || baremetal

package os

import "errors"

func Executable() (string, error) {
	return "", errors.New("Executable not implemented")
}
