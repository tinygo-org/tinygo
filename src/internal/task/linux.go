//go:build linux && !baremetal

package task

import "unsafe"

// Musl uses a pointer (or unsigned long for C++) so unsafe.Pointer should be
// fine.
type threadID unsafe.Pointer
