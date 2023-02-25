//go:build nrf || stm32f4 || stm32l4 || stm32wlx

package machine

import (
	"errors"
	"unsafe"
)

//go:extern __flash_data_start
var flashDataStart [0]byte

//go:extern __flash_data_end
var flashDataEnd [0]byte

// Return the start of the writable flash area, aligned on a page boundary. This
// is usually just after the program and static data.
func FlashDataStart() uintptr {
	pagesize := uintptr(FlashPageSize(uintptr(unsafe.Pointer(&flashDataStart))))
	return (uintptr(unsafe.Pointer(&flashDataStart)) + pagesize - 1) &^ (pagesize - 1)
}

// Return the end of the writable flash area. Usually this is the address one
// past the end of the on-chip flash.
func FlashDataEnd() uintptr {
	return uintptr(unsafe.Pointer(&flashDataEnd))
}

// FlashPageSize returns the page size for the address requested.
// Some processors have different page or sector sizes in different regions of flash.
func FlashPageSize(address uintptr) uint32 {
	return flashPageSize(address)
}

var (
	errFlashCannotErasePage     = errors.New("cannot erase flash page")
	errFlashInvalidWriteLength  = errors.New("write flash data must align to correct number of bits")
	errFlashNotAllowedWriteData = errors.New("not allowed to write flash data")
	errFlashCannotWriteData     = errors.New("cannot write flash data")
)

// Flasher interface is what a processor needs to implement for Flash API.
type Flasher interface {
	ErasePage(address uintptr) error
	WriteData(address uintptr, data []byte) error
	ReadData(address uintptr, data []byte) (n int, err error)
}

// FlashBuffer implements the ReadWriteCloser interface using the Flasher interface.
type FlashBuffer struct {
	f       Flasher
	start   uintptr
	current uintptr
}

// OpenFlashBuffer opens a FlashBuffer.
func OpenFlashBuffer(f Flasher, address uintptr) *FlashBuffer {
	return &FlashBuffer{f: f, start: address, current: address}
}

// Read data from a FlashBuffer.
func (fl *FlashBuffer) Read(p []byte) (n int, err error) {
	fl.f.ReadData(fl.current, p)

	fl.current += uintptr(len(p))

	return len(p), nil
}

// Write data to a FlashBuffer.
func (fl *FlashBuffer) Write(p []byte) (n int, err error) {
	// any new pages needed?
	// NOTE probably will not work as expected if you try to write over page boundry
	// of different sizes.
	pagesize := uintptr(FlashPageSize(fl.current))
	currentPageCount := (fl.current - fl.start + pagesize - 1) / pagesize
	totalPagesNeeded := (fl.current - fl.start + uintptr(len(p)) + pagesize - 1) / pagesize
	if currentPageCount == totalPagesNeeded {
		// just write the data
		err := fl.f.WriteData(fl.current, p)
		if err != nil {
			return 0, err
		}
		fl.current += uintptr(len(p))
		return len(p), nil
	}

	// erase enough pages to hold all data
	nextPageAddress := fl.start + (currentPageCount * pagesize)
	for i := 0; i < int(totalPagesNeeded-currentPageCount); i++ {
		if err := fl.f.ErasePage(nextPageAddress); err != nil {
			return 0, err
		}
		nextPageAddress += pagesize
	}

	// write the data
	for i := 0; i < len(p); i += int(pagesize) {
		var last int = i + int(pagesize)
		if i+int(pagesize) > len(p) {
			last = len(p)
		}

		err := fl.f.WriteData(fl.current, p[i:last])
		if err != nil {
			return 0, err
		}
		fl.current += uintptr(last - i)
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
	fl.current = fl.start + uintptr(offset)

	return offset, nil
}
