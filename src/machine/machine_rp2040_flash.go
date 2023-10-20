//go:build rp2040

package machine

import (
	"bytes"
	"unsafe"
)

// EnterBootloader should perform a system reset in preparation
// to switch to the bootloader to flash new firmware.
func EnterBootloader() {
	enterBootloader()
}

// 13 = 1 + FLASH_RUID_DUMMY_BYTES(4) + FLASH_RUID_DATA_BYTES(8)
var deviceIDBuf [13]byte

// DeviceID returns an identifier that is unique within
// a particular chipset.
//
// The identity is one burnt into the MCU itself, or the
// flash chip at time of manufacture.
//
// It's possible that two different vendors may allocate
// the same DeviceID, so callers should take this into
// account if needing to generate a globally unique id.
//
// The length of the hardware ID is vendor-specific, but
// 8 bytes (64 bits) is common.
func DeviceID() []byte {
	deviceIDBuf[0] = 0x4b // FLASH_RUID_CMD

	err := doFlashCommand(deviceIDBuf[:], deviceIDBuf[:])
	if err != nil {
		panic(err)
	}

	return deviceIDBuf[5:13]
}

// compile-time check for ensuring we fulfill BlockDevice interface
var _ BlockDevice = flashBlockDevice{}

var Flash flashBlockDevice

type flashBlockDevice struct {
}

// ReadAt reads the given number of bytes from the block device.
func (f flashBlockDevice) ReadAt(p []byte, off int64) (n int, err error) {
	if readAddress(off) > FlashDataEnd() {
		return 0, errFlashCannotReadPastEOF
	}

	data := unsafe.Slice((*byte)(unsafe.Pointer(readAddress(off))), len(p))
	copy(p, data)

	return len(p), nil
}

// WriteAt writes the given number of bytes to the block device.
// Only word (32 bits) length data can be programmed.
// If the length of p is not long enough it will be padded with 0xFF bytes.
// This method assumes that the destination is already erased.
func (f flashBlockDevice) WriteAt(p []byte, off int64) (n int, err error) {
	return f.writeAt(p, off)
}

// Size returns the number of bytes in this block device.
func (f flashBlockDevice) Size() int64 {
	return int64(FlashDataEnd() - FlashDataStart())
}

const writeBlockSize = 1 << 8

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
func (f flashBlockDevice) EraseBlocks(start, length int64) error {
	return f.eraseBlocks(start, length)
}

// pad data if needed so it is long enough for correct byte alignment on writes.
func (f flashBlockDevice) pad(p []byte) []byte {
	overflow := int64(len(p)) % f.WriteBlockSize()
	if overflow == 0 {
		return p
	}

	padding := bytes.Repeat([]byte{0xff}, int(f.WriteBlockSize()-overflow))
	return append(p, padding...)
}

// return the correct address to be used for write
func writeAddress(off int64) uintptr {
	return readAddress(off) - uintptr(memoryStart)
}

// return the correct address to be used for reads
func readAddress(off int64) uintptr {
	return FlashDataStart() + uintptr(off)
}
