//go:build stm32wlx

package machine

import (
	"device/stm32"
)

const flashPageSizeValue = 2048

func flashPageSize(address uintptr) uint32 {
	return flashPageSizeValue
}

func erasePage(page uint32) error {
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
