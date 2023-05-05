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

// FlashBuffer implements the ReadWriteCloser interface using the BlockDevice interface.
type FlashBuffer struct {
	b BlockDevice

	// start is actual address
	start uintptr

	// offset is relative to start
	offset uintptr
}

// OpenFlashBuffer opens a FlashBuffer.
func OpenFlashBuffer(b BlockDevice, address uintptr) *FlashBuffer {
	return &FlashBuffer{b: b, start: address}
}

// Read data from a FlashBuffer.
func (fl *FlashBuffer) Read(p []byte) (n int, err error) {
	n, err = fl.b.ReadAt(p, int64(fl.offset))
	fl.offset += uintptr(n)

	return
}

// Write data to a FlashBuffer.
func (fl *FlashBuffer) Write(p []byte) (n int, err error) {
	// any new pages needed?
	// NOTE probably will not work as expected if you try to write over page boundary
	// of pages with different sizes.
	pagesize := uintptr(fl.b.EraseBlockSize())

	// calculate currentPageBlock relative to fl.start, meaning that
	// block 0 -> fl.start
	// block 1 -> fl.start + pagesize
	// block 2 -> fl.start + pagesize*2
	// ...
	currentPageBlock := ((fl.start + fl.offset - FlashDataStart()) + (pagesize - 1)) / pagesize
	lastPageBlockNeeded := ((fl.start + fl.offset + uintptr(len(p)) - FlashDataStart()) + (pagesize - 1)) / pagesize

	// erase enough blocks to hold the data
	if err := fl.b.EraseBlocks(int64(currentPageBlock), int64(lastPageBlockNeeded-currentPageBlock)); err != nil {
		return 0, err
	}

	// write the data
	for i := 0; i < len(p); i += int(pagesize) {
		var last int = i + int(pagesize)
		if i+int(pagesize) > len(p) {
			last = len(p)
		}

		_, err := fl.b.WriteAt(p[i:last], int64(fl.offset))
		if err != nil {
			return 0, err
		}
		fl.offset += uintptr(len(p[i:last]))
	}

	return len(p), nil
}

// Close the FlashBuffer.
func (fl *FlashBuffer) Close() error {
	return nil
}

// Seek implements io.Seeker interface, but with limitations.
// You can only seek relative to the start.
// Also, you cannot use seek before write operations, only read.
func (fl *FlashBuffer) Seek(offset int64, whence int) (int64, error) {
	if whence != io.SeekStart {
		panic("you can only Seek relative to Start")
	}

	fl.offset = uintptr(offset)

	return offset, nil
}
