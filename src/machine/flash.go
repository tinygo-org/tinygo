//go:build nrf

package machine

import "unsafe"

//go:extern __flash_data_start
var flashDataStart [0]byte

//go:extern __flash_data_end
var flashDataEnd [0]byte

// Return the start of the writable flash area, aligned on a page boundary. This
// is usually just after the program and static data.
func FlashDataStart() uintptr {
	return (uintptr(unsafe.Pointer(&flashDataStart)) + flashPageSize - 1) &^ (flashPageSize - 1)
}

// Return the end of the writable flash area. Usually this is the address one
// past the end of the on-chip flash.
func FlashDataEnd() uintptr {
	return uintptr(unsafe.Pointer(&flashDataEnd))
}
