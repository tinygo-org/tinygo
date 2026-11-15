//go:build !linux || baremetal || nintendoswitch || tinygo.wasi || wasm_unknown

package os

import "time"

// Chtimes is a stub on non-Linux targets.
func Chtimes(name string, atime time.Time, mtime time.Time) error {
	return ErrNotImplemented
}
