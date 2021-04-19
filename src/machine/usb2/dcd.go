package usb2

import "unsafe"

type dcd interface {
	init() status
	enable(enable bool) status
	critical(enter bool) status
	interrupt()
	receive(endpoint uint8, transfer *dcdTransfer)
	transmit(endpoint uint8, transfer *dcdTransfer)
	control(setup dcdSetup)
}

const (
	dcdEndpointSize = 64 // bytes
	dcdTransferSize = 32 //
	dcdSetupSize    = 8  //
)

type dcdEndpoint struct {
	config   uint32       // 4
	current  *dcdTransfer // 4 *dcdTransfer
	transfer dcdTransfer  // 32
	setup    dcdSetup     // 8
	// Endpoints are 48-byte data structures. The remaining data extends this to
	// 64-byte, and also makes it simpler for implementations to align endpoints
	// allocated contiguously on 64-byte boundaries.
	first *dcdTransfer // 4 *dcdTransfer
	last  *dcdTransfer // 4 *dcdTransfer
	// After some discussion, perhaps the simplest change to support 64-bit (or
	// future TinyGo versions that don't use 8 bytes to refer to a function) is to
	// allocate a separate buffer of callbacks. The device controller would assign
	// callbacks to unused elements in that buffer, and only the index of that
	// callback would be stored here in the descriptor.
	callback dcdTransferCallback // 8
}

type dcdTransferCallback func(transfer *dcdTransfer)
type dcdTransfer struct {
	next    *dcdTransfer // 4 *dcdTransfer
	token   uint32       // 4
	pointer [5]uintptr   // 20
	param   uint32       // 4
}

type dcdSetup struct {
	bmRequestType uint8
	bRequest      uint8
	wValue        uint16
	wIndex        uint16
	wLength       uint16
}

func (s dcdSetup) pack() uint64 {
	return ((uint64(s.bmRequestType) & 0xFF) << 0) | // uint8
		((uint64(s.bRequest) & 0xFF) << 8) | // uint8
		((uint64(s.wValue) & 0xFFFF) << 16) | // uint16
		((uint64(s.wIndex) & 0xFFFF) << 32) | // uint16
		((uint64(s.wLength) & 0xFFFF) << 48) // uint16
}

var (
	// dcdPointerNil is a sentinel value used to indicate a pointer to an invalid
	// value. Do not attempt to dereference!
	dcdPointerNil = uintptr(0)
	// dcdTransferEOL is a sentinel value used to indicate the final node in a
	// linked list of transfer descriptors.
	dcdTransferEOL = (*dcdTransfer)(unsafe.Pointer(dcdTransferPointerEOL))
	// dcdTransferPointerEOL is the notional memory address of dcdTransferEOL.
	// The address does not refer to an actual dcdTransfer, so no attempt should
	// be made to dereference it.
	dcdTransferPointerEOL = uintptr(1)
)

// nextTransfer returns the next transfer descriptor pointed to by the receiver
// transfer descriptor, and whether or not that next descriptor is the final
// descriptor in the list.
func (t *dcdTransfer) nextTransfer() (*dcdTransfer, bool) {
	return t.next, uintptr(unsafe.Pointer(t.next)) == dcdTransferPointerEOL
}
