//go:build esp32c3 || esp32s3

package machine

import (
	"unsafe"
)

/*
#include <stdint.h>
extern int esp_rom_spiflash_read(uint32_t src_addr, uint32_t *data, uint32_t len);
extern int esp_rom_spiflash_write(uint32_t dest_addr, const uint32_t *data, uint32_t len);
extern int esp_rom_spiflash_erase_sector(uint32_t sector_num);
extern int esp_rom_spiflash_unlock(void);
extern void Cache_Invalidate_Addr(uint32_t addr, uint32_t size);
*/
import "C"

// compile-time check for ensuring we fulfill BlockDevice interface
var _ BlockDevice = flashBlockDevice{}

var Flash flashBlockDevice

type flashBlockDevice struct {
}

// ReadAt reads the given number of bytes from the block device.
func (f flashBlockDevice) ReadAt(p []byte, off int64) (n int, err error) {
	if readAddress(off)+uintptr(len(p)) > FlashDataEnd() {
		return 0, errFlashCannotReadPastEOF
	}

	data := unsafe.Slice((*byte)(unsafe.Add(unsafe.Pointer(FlashDataStart()), off)), len(p))
	copy(p, data)

	return len(p), nil
}

// WriteAt writes the given number of bytes to the block device.
// Only word (32 bits) length data can be programmed.
// It must not be called from an interrupt or with interrupts disabled.
// If the length of p is not long enough it will be padded with 0xFF bytes.
// This method assumes that the destination is already erased.
func (f flashBlockDevice) WriteAt(p []byte, off int64) (n int, err error) {
	return f.writeAt(p, off)
}

// Size returns the number of bytes in this block device.
func (f flashBlockDevice) Size() int64 {
	return int64(FlashDataEnd() - FlashDataStart())
}

const writeBlockSize = 4

// WriteBlockSize returns the block size in which data can be written to
// memory. It can be used by a client to optimize writes, non-aligned writes
// should always work correctly.
func (f flashBlockDevice) WriteBlockSize() int64 {
	return writeBlockSize
}

const eraseBlockSizeValue = 1 << 12

func eraseBlockSize() int64 {
	return eraseBlockSizeValue
}

// EraseBlockSize returns the smallest erasable area on this particular chip
// in bytes. This is used for the block size in EraseBlocks.
func (f flashBlockDevice) EraseBlockSize() int64 {
	return eraseBlockSize()
}

// EraseBlocks erases the given number of blocks. An implementation may
// transparently coalesce ranges of blocks into larger bundles if the chip
// supports this. The start and len parameters are in block numbers, use
// EraseBlockSize to map addresses to blocks.
// It must not be called from an interrupt or with interrupts disabled.
func (f flashBlockDevice) EraseBlocks(start, length int64) error {
	return f.eraseBlocks(start, length)
}

// return the correct address to be used for reads
func readAddress(off int64) uintptr {
	return FlashDataStart() + uintptr(off)
}

// Physical flash offset of FlashDataStart, set by the linker script.
//
//go:extern __flash_data_phys
var flashDataPhys [0]byte

// return the correct physical address to be used for write/erase
func writeAddress(off int64) uint32 {
	return uint32(uintptr(unsafe.Pointer(&flashDataPhys))) + uint32(off)
}

// romSpiflashChip matches esp_rom_spiflash_chip_t in esp-idf
// components/esp_rom/include/esp_rom_spiflash.h.
type romSpiflashChip struct {
	deviceID   uint32
	chipSize   uint32
	blockSize  uint32
	sectorSize uint32
	pageSize   uint32
	statusMask uint32
}

//go:extern rom_spiflash_legacy_data
var romSpiflashLegacyData *romSpiflashChip

var flashSizeSet bool

// setFlashSize gives the ROM flash routines the chip size from the image
// header, as update_flash_config in the esp-idf bootloader does.
func setFlashSize() {
	if flashSizeSet {
		return
	}
	flashSizeSet = true
	var header uint32
	if C.esp_rom_spiflash_read(0, (*C.uint32_t)(unsafe.Pointer(&header)), 4) != 0 {
		return
	}
	// The upper nibble of byte 3 is the flash size, 0 is 1 MB.
	size := header >> 28
	if byte(header) == 0xE9 && size <= 7 {
		romSpiflashLegacyData.chipSize = 1 << 20 << size
	}
}

func (f flashBlockDevice) writeAt(p []byte, off int64) (n int, err error) {
	if readAddress(off)+uintptr(len(p)) > FlashDataEnd() {
		return 0, errFlashCannotWritePastEOF
	}

	address := writeAddress(off)
	padded := flashPad(p, int(f.WriteBlockSize()))

	state, multicore := flashSafeEnter()
	defer flashSafeExit(state, multicore)

	setFlashSize()
	C.esp_rom_spiflash_unlock()
	res := C.esp_rom_spiflash_write(C.uint32_t(address), (*C.uint32_t)(unsafe.Pointer(&padded[0])), C.uint32_t(len(padded)))
	C.Cache_Invalidate_Addr(C.uint32_t(readAddress(off)), C.uint32_t(len(padded)))
	if res != 0 {
		return 0, errFlashCannotWriteData
	}

	return len(padded), nil
}

func (f flashBlockDevice) eraseBlocks(start, length int64) error {
	address := writeAddress(start * f.EraseBlockSize())
	if readAddress((start+length)*f.EraseBlockSize()) > FlashDataEnd() {
		return errFlashCannotErasePastEOF
	}

	state, multicore := flashSafeEnter()
	defer flashSafeExit(state, multicore)

	setFlashSize()
	C.esp_rom_spiflash_unlock()
	sector := address / uint32(f.EraseBlockSize())

	for i := int64(0); i < length; i++ {
		res := C.esp_rom_spiflash_erase_sector(C.uint32_t(sector + uint32(i)))
		C.Cache_Invalidate_Addr(C.uint32_t(readAddress((start+i)*f.EraseBlockSize())), C.uint32_t(f.EraseBlockSize()))
		if res != 0 {
			return errFlashCannotErasePage
		}
	}

	return nil
}
