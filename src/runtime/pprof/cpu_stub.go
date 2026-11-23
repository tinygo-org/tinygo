//go:build !darwin && !(linux && scheduler.threads && (amd64 || arm64))

package pprof

import "io"

func StartCPUProfile(w io.Writer) error {
	return nil
}

func StopCPUProfile() {
}
