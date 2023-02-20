//go:build stm32

package machine

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
