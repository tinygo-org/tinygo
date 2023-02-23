//go:build stm32f4 || stm32l4 || stm32wlx

package machine

import (
	"device/stm32"

	"errors"
	"unsafe"
)

var (
	errFlashCannotErasePage     = errors.New("cannot erase flash page")
	errFlashInvalidWriteLength  = errors.New("write flash data must align to 64 bits")
	errFlashNotAllowedWriteData = errors.New("not allowed to write flash data")
	errFlashCannotWriteData     = errors.New("cannot write flash data")
)

var Flash flash

type flash struct {
}

// ErasePage erases the page of flash data that starts at address.
func (f flash) ErasePage(address uintptr) error {
	unlockFlash()
	defer lockFlash()

	return eraseFlashPage(address)
}

// WriteData writes the flash that starts at address with data.
// Only double-word (64 bits) can be programmed. See rm0461 page 78.
func (f flash) WriteData(address uintptr, data []byte) error {
	unlockFlash()
	defer lockFlash()

	return writeFlashData(address, data)
}

// ReadData reads the data starting at address.
func (f flash) ReadData(address uintptr, data []byte) (n int, err error) {
	p := unsafe.Slice((*byte)(unsafe.Pointer(address)), len(data))
	copy(data, p)

	return len(data), nil
}

func unlockFlash() {
	// keys as described rm0461 page 76
	var fkey1 uint32 = 0x45670123
	var fkey2 uint32 = 0xCDEF89AB

	// Wait for the flash memory not to be busy
	for stm32.FLASH.GetSR_BSY() != 0 {
	}

	// Check if the controller is unlocked already
	if stm32.FLASH.GetCR_LOCK() != 0 {
		// Write the first key
		stm32.FLASH.SetKEYR(fkey1)
		// Write the second key
		stm32.FLASH.SetKEYR(fkey2)
	}
}

func lockFlash() {
	stm32.FLASH.SetCR_LOCK(1)
}
