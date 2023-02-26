//go:build stm32f4 || stm32l4 || stm32wlx

package machine

import (
	"device/stm32"

	"unsafe"
)

// compile-time check for ensuring we fulfill BlockDevice interface
var _ BlockDevice = flashBlockDevice{}

var Flash flashBlockDevice

type flashBlockDevice struct {
}

// ReadAt reads the given number of bytes from the block device.
func (f flashBlockDevice) ReadAt(p []byte, off int64) (n int, err error) {
	if FlashDataStart()+uintptr(off)+uintptr(len(p)) > FlashDataEnd() {
		return 0, errFlashCannotReadPastEOF
	}

	data := unsafe.Slice((*byte)(unsafe.Pointer(FlashDataStart()+uintptr(off))), len(p))
	copy(p, data)

	return len(p), nil
}

// WriteAt writes the given number of bytes to the block device.
// Only double-word (64 bits) length data can be programmed. See rm0461 page 78.
func (f flashBlockDevice) WriteAt(p []byte, off int64) (n int, err error) {
	if FlashDataStart()+uintptr(off)+uintptr(len(p)) > FlashDataEnd() {
		return 0, errFlashCannotWritePastEOF
	}

	unlockFlash()
	defer lockFlash()

	return writeFlashData(FlashDataStart()+uintptr(off), p)
}

// Size returns the number of bytes in this block device.
func (f flashBlockDevice) Size() int64 {
	return int64(FlashDataEnd() - FlashDataStart())
}

// WriteBlockSize returns the block size in which data can be written to
// memory. It can be used by a client to optimize writes, non-aligned writes
// should always work correctly.
func (f flashBlockDevice) WriteBlockSize() int64 {
	return writeBlockSize
}

func eraseBlockSize() int64 {
	return eraseBlockSizeValue
}

// EraseBlockSize returns the smallest erasable area on this particular chip
// in bytes. This is used for the block size in EraseBlocks.
// It must be a power of two, and may be as small as 1. A typical size is 4096.
// TODO: correctly handle processors that have differently sized blocks
// in different areas of memory like the STM32F40x and STM32F1x.
func (f flashBlockDevice) EraseBlockSize() int64 {
	return eraseBlockSize()
}

// EraseBlocks erases the given number of blocks. An implementation may
// transparently coalesce ranges of blocks into larger bundles if the chip
// supports this. The start and len parameters are in block numbers, use
// EraseBlockSize to map addresses to blocks.
func (f flashBlockDevice) EraseBlocks(start, len int64) error {
	unlockFlash()
	defer lockFlash()

	for i := start; i < start+len; i++ {
		if err := eraseBlock(uint32(i)); err != nil {
			return err
		}
	}

	return nil
}

const memoryStart = 0x08000000

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
