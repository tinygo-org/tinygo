//go:build scheduler.asyncify && !(gc.conservative || gc.precise)

package task

import (
	"internal/gclayout"
	"unsafe"
)

func stackLayout() unsafe.Pointer { return gclayout.Conservative.AsPtr() }

// ScanStacks does nothing, since the GC scans each stack as a whole object.
func ScanStacks() {}
