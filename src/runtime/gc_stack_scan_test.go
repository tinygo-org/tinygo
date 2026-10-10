//go:build gc.conservative || gc.precise

package runtime

import (
	"internal/gclayout"
	"unsafe"
)

// GCStackScanProbe checks that markCurrentGoroutineStack scans a stack object
// through its last body word before finishMark runs, and that it ignores a
// stack pointer that does not land in an allocation.
func GCStackScanProbe() string {
	const size = 64
	guard := allocManual(16)
	stack := alloc(size, gclayout.Conservative.AsPtr())
	target := alloc(16, gclayout.NoPtrs.AsPtr())
	blocks := (size + unsafe.Sizeof(objHeader{}) + bytesPerBlock - 1) / bytesPerBlock
	lastWord := unsafe.Add(stack, blocks*bytesPerBlock-unsafe.Sizeof(objHeader{})-unsafe.Sizeof(uintptr(0)))
	*(*unsafe.Pointer)(lastWord) = target

	gcLock.Lock()
	markCurrentGoroutineStack(uintptr(stack))
	marked := blockFromAddr(uintptr(target)).findHead().state() == blockStateMark
	finishMark()
	for block := gcBlock(0); block < endBlock; block++ {
		if block.state() == blockStateMark {
			header := (*objHeader)(unsafe.Add(block.pointer(), bytesPerBlock-unsafe.Sizeof(objHeader{})))
			if header.next != 1 {
				block.unmark()
			}
		}
	}
	manualMarked := blockFromAddr(uintptr(guard)).findHead().state() == blockStateMark
	gcLock.Unlock()

	*(*unsafe.Pointer)(lastWord) = nil

	// A freed block is not an allocation. findHead would walk off it and
	// fatal under gcAsserts, so this must return without scanning.
	free(guard)
	gcLock.Lock()
	markCurrentGoroutineStack(uintptr(guard))
	freedMarked := blockFromAddr(uintptr(guard)).state() != blockStateFree
	gcLock.Unlock()

	if freedMarked {
		return "stack scan marked a freed block"
	}
	if !marked {
		return "stack scan missed a root in the last body word"
	}
	if !manualMarked {
		return "probe cleanup unmarked a manual allocation"
	}
	return ""
}
