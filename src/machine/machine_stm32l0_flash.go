//go:build stm32l0

package machine

// The STM32L0 series of MCUs has a different type of flash than other STM32
// series chips. The programming interface is different, and the flash is erased
// to zero bits instead of one bits as on most flash. So this requires a
// different implementation.

import (
	"device/stm32"
	"runtime/interrupt"
	"runtime/volatile"
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
// Only word-sized (32 bits) length data can be programmed.
// If the length of p is not long enough it will be padded with zero bytes.
// This method assumes that the destination is already erased.
func (f flashBlockDevice) WriteAt(p []byte, off int64) (n int, err error) {
	if FlashDataStart()+uintptr(off)+uintptr(len(p)) > FlashDataEnd() {
		return 0, errFlashCannotWritePastEOF
	}
	if uintptr(off)%4 != 0 {
		// Offset must be aligned on a word boundary.
		return 0, errFlashCannotWriteData
	}

	unlockFlash()
	defer lockFlash()

	// Write words in this area.
	for i := 0; i < len(p); i += 4 {
		// Construct the word to write.
		word := uint32(p[i])
		if i+1 < len(p) {
			word |= uint32(p[i+1]) << 8
		}
		if i+2 < len(p) {
			word |= uint32(p[i+2]) << 16
		}
		if i+3 < len(p) {
			word |= uint32(p[i+3]) << 24
		}

		// Find the pointer address to write.
		address := FlashDataStart() + uintptr(off) + uintptr(i)

		// Write the word to flash.
		(*volatile.Register32)(unsafe.Pointer(address)).Set(word)

		// Check for any errors.
		if stm32.FLASH.SR.Get()&(stm32.Flash_SR_WRPERR|stm32.Flash_SR_NOTZEROERR|stm32.Flash_SR_SIZERR) != 0 {
			return i, errFlashCannotWriteData
		}
	}

	return len(p), nil
}

// Size returns the number of bytes in this block device.
func (f flashBlockDevice) Size() int64 {
	return int64(FlashDataEnd() - FlashDataStart())
}

// WriteBlockSize returns the block size in which data can be written to
// memory. It can be used by a client to optimize writes, non-aligned writes
// should always work correctly.
func (f flashBlockDevice) WriteBlockSize() int64 {
	return 4
}

func eraseBlockSize() int64 {
	return 128
}

// EraseBlockSize returns the smallest erasable area on this particular chip
// in bytes. This is used for the block size in EraseBlocks.
// It must be a power of two, and may be as small as 1. A typical size is 4096.
func (f flashBlockDevice) EraseBlockSize() int64 {
	return eraseBlockSize()
}

// EraseBlocks erases the given number of blocks. An implementation may
// transparently coalesce ranges of blocks into larger bundles if the chip
// supports this. The start and len parameters are in block numbers, use
// EraseBlockSize to map addresses to blocks.
// Note that block 0 should map to the address of FlashDataStart().
func (f flashBlockDevice) EraseBlocks(start, len int64) error {
	// Flash needs to be unlocked to be able to erase it.
	unlockFlash()
	defer lockFlash()

	// Set the flash programming mode to erase a page.
	// Note: lockFlash() will reset these flags to 0 so we don't need to
	// explicitly set them to 0.
	stm32.FLASH.PECR.Set(stm32.Flash_PECR_ERASE | stm32.Flash_PECR_PROG)

	// Erase all pages in this range.
	for i := uintptr(start); i < uintptr(start)+uintptr(len); i++ {
		// Find the pointer address somewhere in the page to erase.
		address := FlashDataStart() + i*uintptr(eraseBlockSize())

		// To erase, write any value to that address.
		(*volatile.Register32)(unsafe.Pointer(address)).Set(uint32(address))

		// Check for any errors.
		// The only error (that is not a programming error) that could happen is
		// if a row is in a protected sector.
		if stm32.FLASH.SR.Get()&(stm32.Flash_SR_WRPERR|stm32.Flash_SR_SIZERR) != 0 {
			return errFlashCannotErasePage
		}
	}

	return nil
}

func unlockFlash() {
	// Make sure the flash peripheral clock is enabled.
	stm32.RCC.AHBENR.SetBits(stm32.RCC_AHBENR_MIFEN)

	// Wait for the flash memory not to be busy.
	for stm32.FLASH.GetSR_BSY() != 0 {
	}

	// Disable interrupts while writing, since no memory operations may happen
	// while the unlock sequence is ongoing.
	mask := interrupt.Disable()

	// Remove PELOCK bit.
	stm32.FLASH.PEKEYR.Set(0x89ABCDEF)
	stm32.FLASH.PEKEYR.Set(0x02030405)

	// Remove PRGLOCK bit.
	stm32.FLASH.PRGKEYR.Set(0x8C9DAEBF)
	stm32.FLASH.PRGKEYR.Set(0x13141516)

	interrupt.Restore(mask)
}

func lockFlash() {
	// Set PELOCK to 1, which also automatically sets PRGLOCK to 1.
	stm32.FLASH.PECR.Set(stm32.Flash_PECR_PELOCK)
}
