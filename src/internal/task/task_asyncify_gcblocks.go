//go:build scheduler.asyncify && (gc.conservative || gc.precise)

package task

import (
	"internal/gclayout"
	"unsafe"
)

// The GC does not scan stacks as objects. ScanStacks marks their live parts.
func stackLayout() unsafe.Pointer { return gclayout.NoPtrs.AsPtr() }

//go:linkname markRoots runtime.markRoots
func markRoots(start, end uintptr)

// ScanStacks marks the live part of every goroutine stack.
func ScanStacks() { liveStackRanges(markRoots) }
