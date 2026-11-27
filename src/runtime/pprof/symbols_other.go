//go:build !darwin && !(linux && scheduler.threads && (amd64 || arm64))

package pprof

// Code addresses are not symbolized on this OS yet.
func fileMappings() []fileMapping {
	return nil
}

func lookupSystem(pc uintptr) (lib string, base uintptr, sym string, ok bool) {
	return "", 0, "", false
}
