//go:build !darwin

package pprof

import "io"

func StartCPUProfile(w io.Writer) error {
	return nil
}

func StopCPUProfile() {
}
