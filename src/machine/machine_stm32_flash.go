//go:build stm32

package machine

import (
	"device/stm32"

	"errors"
	"unsafe"
)

const FlashPageSize = 2048

var (
	errFlashCannotErasePage    = errors.New("cannot erase flash page")
	errFlashInvalidWriteLength = errors.New("write flash data must align to 64 bits")
	errFlashCannotWriteData    = errors.New("cannot write flash data")
)

var Flash flash

type flash struct {
}

// ErasePage erases the page of flash data that starts at address.
func (f flash) ErasePage(address uintptr) error {
	unlockFlash()
	defer lockFlash()

	// calculate page number from address
	var page uint32 = uint32(address) / FlashPageSize

	// wait until other flash operations are done
	for stm32.FLASH.GetSR_BSY() != 0 {
	}

	// check if operation is allowed.
	if stm32.FLASH.GetSR_PESD() == 0 {
		return errFlashCannotErasePage
	}

	// TODO: clear any previous errors

	// page erase operation
	stm32.FLASH.SetCR_PER(1)

	// set the address to the page to be written
	stm32.FLASH.SetCR_PNB(page)

	// start the page erase
	stm32.FLASH.SetCR_STRT(1)

	// wait until page erase is done
	for stm32.FLASH.GetSR_BSY() != 0 {
	}

	return nil
}

// WriteData writes the flash that starts at address with data.
// Only double-word (64 bits) can be programmed. See rm0461 page 78.
func (f flash) WriteData(address uintptr, data []byte) error {
	if len(data)%4 != 0 {
		return errFlashInvalidWriteLength
	}

	unlockFlash()
	defer lockFlash()

	// wait until other flash operations are done
	for stm32.FLASH.GetSR_BSY() != 0 {
	}

	// check if operation is allowed
	if stm32.FLASH.GetSR_PESD() == 0 {
		return errFlashCannotWriteData
	}

	// TODO: clear any previous errors

	// start page write operation
	stm32.FLASH.SetCR_PG(1)

	// end page write when done
	defer stm32.FLASH.SetCR_PG(0)

	for i := 0; i < len(data); i += 4 {
		// write first word in an address aligned with double-word
		*(*uint32)(unsafe.Pointer(address)) = uint32(data[i] & data[i+1] << 16)

		// now write second word
		*(*uint32)(unsafe.Pointer(address + 2)) = uint32(data[i+2] & data[i+3] << 16)

		// wait until not busy
		for stm32.FLASH.GetSR_BSY() != 0 {
		}

		// verify write
		if stm32.FLASH.GetSR_EOP() == 0 {
			return errFlashCannotWriteData
		}

		// prepare for next operation
		stm32.FLASH.SetSR_EOP(0)
	}

	return nil
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
