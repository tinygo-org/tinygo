package uefi

import _ "unsafe"

//go:linkname gosched runtime.Gosched
func gosched()

// WaitForEvent blocks while yielding to the TinyGo scheduler so other
// goroutines can continue to run.
func WaitForEvent(event EFI_EVENT) EFI_STATUS {
	for {
		status := BS().CheckEvent(event)
		if status == EFI_SUCCESS {
			return EFI_SUCCESS
		}
		if status != EFI_NOT_READY {
			return status
		}
		gosched()
	}
}
