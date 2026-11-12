//go:build !linux || baremetal || nintendoswitch || wasip1 || wasm_unknown || wasip2

package os

import "time"

// Chtimes is a stub on non-Linux targets.
func Chtimes(name string, atime time.Time, mtime time.Time) error {
	return ErrNotImplemented
}
