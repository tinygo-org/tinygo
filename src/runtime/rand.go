package runtime

import _ "unsafe"

// TODO: Use hardware when available
//
//go:linkname vgetrandom
func vgetrandom(p []byte, flags uint32) (ret int, supported bool) { return 0, false }

//go:linkname fatal crypto/internal/sysrand.fatal
func fatal(msg string) { runtimePanic(msg) }
