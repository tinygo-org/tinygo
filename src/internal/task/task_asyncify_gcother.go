//go:build scheduler.asyncify && !(gc.conservative || gc.precise)

package task

import (
	"internal/gclayout"
	"unsafe"
)

// The GC scans each stack as a whole object.
func stackLayout() unsafe.Pointer { return gclayout.Conservative.AsPtr() }

func stackBodySize(size uintptr) uintptr { return size }

// StackReached is never called, since the GC scans each stack as a whole object.
func StackReached(header uintptr) {}
