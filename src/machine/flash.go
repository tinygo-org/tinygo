//go:build nrf || nrf51 || nrf52 || nrf528xx || stm32f4 || stm32l4 || stm32wlx || atsamd21 || atsamd51 || atsame5x || rp2040

package machine

import (
	"errors"
	"io"
	"unsafe"
)

//go:extern __flash_data_start
var flashDataStart [0]byte

//go:extern __flash_data_end
var flashDataEnd [0]byte

// Return the start of the writable flash area, aligned on a page boundary. This
// is usually just after the program and static data.
func FlashDataStart() uintptr {
	pagesize := uintptr(eraseBlockSize())
	return (uintptr(unsafe.Pointer(&flashDataStart)) + pagesize - 1) &^ (pagesize - 1)
}

// Return the end of the writable flash area. Usually this is the address one
// past the end of the on-chip flash.
func FlashDataEnd() uintptr {
	return uintptr(unsafe.Pointer(&flashDataEnd))
}

var (
	errFlashCannotErasePage     = errors.New("cannot erase flash page")
	errFlashInvalidWriteLength  = errors.New("write flash data must align to correct number of bits")
	errFlashNotAllowedWriteData = errors.New("not allowed to write flash data")
	errFlashCannotWriteData     = errors.New("cannot write flash data")
	errFlashCannotReadPastEOF   = errors.New("cannot read beyond end of flash data")
	errFlashCannotWritePastEOF  = errors.New("cannot write beyond end of flash data")
	errFlashCannotErasePastEOF  = errors.New("cannot erase beyond end of flash data")
)

// BlockDevice is the raw device that is meant to store flash data.
type BlockDevice interface {
	// ReadAt reads the given number of bytes from the block device.
	io.ReaderAt

	// WriteAt writes the given number of bytes to the block device.
	io.WriterAt

	// Size returns the number of bytes in this block device.
	Size() int64

	// WriteBlockSize returns the block size in which data can be written to
	// memory. It can be used by a client to optimize writes, non-aligned writes
	// should always work correctly.
	WriteBlockSize() int64

	// EraseBlockSize returns the smallest erasable area on this particular chip
	// in bytes. This is used for the block size in EraseBlocks.
	// It must be a power of two, and may be as small as 1. A typical size is 4096.
	EraseBlockSize() int64

	// EraseBlocks erases the given number of blocks. An implementation may
	// transparently coalesce ranges of blocks into larger bundles if the chip
	// supports this. The start and len parameters are in block numbers, use
	// EraseBlockSize to map addresses to blocks.
	EraseBlocks(start, len int64) error
}
