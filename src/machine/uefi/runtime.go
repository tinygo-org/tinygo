package uefi

//go:linkname gosched runtime.Gosched
func gosched()

// WaitForEvent blocks execution while yielding to other goroutines which differs
// from BS().WaitForEvent, which is a hard-block; the CPU is entirely stalled.
func WaitForEvent(event EFI_EVENT) {
	for {
		status := BS().CheckEvent(event)
		if status == EFI_SUCCESS {
			return
		}
		gosched()
	}
}
