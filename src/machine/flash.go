//go:build stm32 || nrf

package machine

import "unsafe"

//go:extern __flash_data_start
var flashDataStart [0]byte

//go:extern __flash_data_end
var flashDataEnd [0]byte

// Return the start of the writable flash area, aligned on a page boundary. This
// is usually just after the program and static data.
func FlashDataStart() uintptr {
	return (uintptr(unsafe.Pointer(&flashDataStart)) + flashPageSize - 1) &^ (flashPageSize - 1)
}

// Return the end of the writable flash area. Usually this is the address one
// past the end of the on-chip flash.
func FlashDataEnd() uintptr {
	return uintptr(unsafe.Pointer(&flashDataEnd))
}

type Flasher interface {
	ErasePage(address uintptr) error
	WriteData(address uintptr, data []byte) error
	ReadData(address uintptr, data []byte) (n int, err error)
}

type FlashBuffer struct {
	f       Flasher
	start   uintptr
	current uintptr
}

func OpenFlashBuffer(f Flasher, address uintptr) *FlashBuffer {
	return &FlashBuffer{f: f, start: address, current: address}
}

func (fl *FlashBuffer) Read(p []byte) (n int, err error) {
	fl.f.ReadData(fl.current, p)

	fl.current += uintptr(len(p))

	return len(p), nil
}

func (fl *FlashBuffer) Write(p []byte) (n int, err error) {
	// any new pages needed?
	currentPageCount := (fl.current - fl.start + FlashPageSize - 1) / FlashPageSize
	totalPagesNeeded := (fl.current - fl.start + uintptr(len(p)) + FlashPageSize - 1) / FlashPageSize
	pageAddress := fl.start + (currentPageCount * FlashPageSize)
	for i := 0; i < int(totalPagesNeeded-currentPageCount); i++ {
		fl.f.ErasePage(pageAddress)
		pageAddress += FlashPageSize
	}

	// TODO: write the data
	return 0, nil
}

func (fl *FlashBuffer) Close() error {
	return nil
}
