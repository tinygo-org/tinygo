package usb

import "unsafe"

// deviceController provides hardware abstraction over a platform's physical
// USB device port controller (e.g., an EHCI-compliant peripheral).
//
// To connect the USB 2.0 device driver stack to a particular platform, the
// following method must be implemented:
//
//		func (d *device) initController() deviceController
//
// This method shall return an implementation of deviceController, or nil if a
// controller cannot be allocated for the given port.
type (
	deviceController interface {
		init() status                                                // Controller initialization
		deinit() status                                              // Controller de-initialization
		enable(enable bool) status                                   // Controller enable runtime
		interrupt()                                                  // Controller interrupt handler
		process() status                                             // Controller process interrupts
		send(address uint8, buffer []uint8, length uint32) status    // Controller send data
		receive(address uint8, buffer []uint8, length uint32) status // Controller receive data
		cancel(address uint8) status                                 // Controller cancel transfer
		control(command deviceControlID, param interface{}) status   // Controller device control
		critical(enter bool) status                                  // Controller critical section
	}

	// deviceControllerInterruptQueue represents a fixed-length, circular queue
	// (FIFO) of interrupts.
	//
	// The enqueue and dequeue operations (methods enq and deq, respectively) are
	// NOT interrupt-/thread-safe. The user must protect against race conditions
	// using appropriate resources for their system. For example, the interface
	// method (deviceController).critical(bool) implements the concept of critical
	// sections, which could be used to prevent concurrent accesses.
	// This restriction also applies to the higher-level methods fill and drain.
	deviceControllerInterruptQueue struct {
		queue [configInterruptQueueSize]uintptr // circular queue
		head  int
		count int
	}

	deviceControllerCapabilitiesBitmap uint32
	deviceControllerCapabilities       struct {
		reserved1     uint16 // 15 (bits)
		ios           uint8  // 1
		maxPacketSize uint16 // 11
		reserved2     uint8  // 2
		zlt           uint8  // 1
		mult          uint8  // 2 (= 32 bits)
	}

	deviceControllerDTDTokenBitmap uint32
	deviceControllerDTDToken       struct {
		status             uint8  // 8 (bits)
		reserved1          uint8  // 2
		multiplierOverride uint8  // 2
		reserved2          uint8  // 3
		ioc                uint8  // 1
		totalBytes         uint16 // 15
		reserved3          uint8  // 1 (= 32 bits)
	}

	deviceControllerEndpointStatusBitmap uint32
	deviceControllerEndpointStatus       struct {
		isOpened uint8  // 1 (bits)
		zlt      uint8  // 1
		_        uint32 // 30 (= 32 bits)
	}

	deviceControllerOriginalBufferBitmap uint32
	deviceControllerOriginalBuffer       struct {
		originalBufferOffset uint16 // 12 (bits)
		originalBufferLength uint32 // 19
		dtdInvalid           uint8  // 1 (= 32 bits)
	}

	deviceControllerQH struct {
		capabilities      deviceControllerCapabilitiesBitmap   // 4 (bytes)
		currentDTDPointer *deviceControllerDTD                 // 4
		nextDTDPointer    *deviceControllerDTD                 // 4
		dtdToken          deviceControllerDTDTokenBitmap       // 4
		bufferPointerPage [5]uint32                            // 20
		reserved1         uint32                               // 4
		setupBuffer       deviceSetupBuffer                    // 8
		setupBufferBack   deviceSetupBuffer                    // 8
		endpointStatus    deviceControllerEndpointStatusBitmap // 4
		reserved2         uint32                               // 4 (= 64 bytes)
	}

	deviceControllerDTDList [2 * configDeviceMaxEndpoints]*deviceControllerDTD
	deviceControllerDTD     struct {
		nextDTDPointer    *deviceControllerDTD                 // 4 (bytes)
		dtdToken          deviceControllerDTDTokenBitmap       // 4
		bufferPointerPage [5]uint32                            // 20
		originalBuffer    deviceControllerOriginalBufferBitmap // 4 (= 32 bytes)
	}
)

const (

	// device QH
	deviceControllerQHSize       = 64 // bytes
	deviceControllerQHBufferSize = (configDeviceCount-1)*configDeviceControllerQHAlign +
		2*configDeviceMaxEndpoints*2*deviceControllerQHSize

	deviceControllerQHPointerMsk       = 0xFFFFFFC0
	deviceControllerQHMultMsk          = 0xC0000000
	deviceControllerQHZLTMsk           = 0x20000000
	deviceControllerQHMaxPacketSizeMsk = 0x07FF0000
	deviceControllerQHMaxPacketSize    = 0x00000800
	deviceControllerQHIOSMsk           = 0x00008000

	// device DTD
	deviceControllerDTDSize       = 32 // bytes
	deviceControllerDTDBufferSize = (configDeviceCount-1)*configDeviceControllerDTDAlign +
		configDeviceControllerMaxDTD*deviceControllerDTDSize

	deviceControllerDTDPointerMsk             = 0xFFFFFFE0
	deviceControllerDTDTerminateMsk           = 0x00000001
	deviceControllerDTDPageMsk                = 0xFFFFF000
	deviceControllerDTDPageOffsetMsk          = 0x00000FFF
	deviceControllerDTDPageBlock              = 0x00001000
	deviceControllerDTDTotalBytesMsk          = 0x7FFF0000
	deviceControllerDTDTotalBytes             = 0x00004000
	deviceControllerDTDIOCMsk                 = 0x00008000
	deviceControllerDTDMultIOMsk              = 0x00000C00
	deviceControllerDTDStatusMsk              = 0x000000FF
	deviceControllerDTDStatusErrorMsk         = 0x00000068
	deviceControllerDTDStatusActive           = 0x00000080
	deviceControllerDTDStatusHalted           = 0x00000040
	deviceControllerDTDStatusDataBufferError  = 0x00000020
	deviceControllerDTDStatusTransactionError = 0x00000008
)

var (
	// special invalid pointer, indicating the end of a list of DTDs
	deviceControllerDTDTerminate = (*deviceControllerDTD)(unsafe.Pointer(uintptr(
		deviceControllerDTDTerminateMsk)))
)

// enq enqueues the given mask into tail position of the receiver iq's queue,
// increasing queue length by 1.
//
// When the queue fills to capacity, any subsequent enqueue will overwrite the
// current head with the given value, positioning its following element at the
// front of the queue, and leaves queue length unaffected.
func (iq *deviceControllerInterruptQueue) enq(mask uintptr) {
	if iq.count == configInterruptQueueSize {
		// queue is full; overwrite oldest element (queue head) and increment head
		iq.queue[iq.head] = mask
		iq.head++
		iq.head %= configInterruptQueueSize
	} else {
		// queue is not full; place element at queue tail and increment count
		iq.queue[(iq.head+iq.count)%configInterruptQueueSize] = mask
		iq.count++
	}
}

// deq dequeues the mask in tail position from the receiver iq's queue, reducing
// queue length by 1, and returns the dequeued value with true.
//
// When the queue is empty, queue length remains unaffected, and it returns 0
// with false.
func (iq *deviceControllerInterruptQueue) deq() (uintptr, bool) {
	if iq.count == 0 {
		// queue is empty; reset head and return an invalid value
		iq.head = 0
		return 0, false
	}
	// queue is not empty; decrement count, reset and return value at queue tail
	iq.count--
	n := (iq.head + iq.count) % configInterruptQueueSize
	m := iq.queue[n]
	iq.queue[n] = 0 // ensure the queue contains no spurious data
	return m, true
}

// fill enqueues each given mask (in order) into tail position of the receiver
// iq's queue, increasing queue length up to capacity, if possible.
//
// If the number of values given is greater than available queue positions, each
// element in head position - at the time the value is enqueued - will be
// overwritten, so that queue length never exceeds capacity, and the element in
// tail position is always the final element given.
func (iq *deviceControllerInterruptQueue) fill(mask ...uintptr) {
	for _, m := range mask {
		iq.enq(m)
	}
}

// drain dequeues all elements in the receiver iq's queue, reducing queue length
// to 0, and returns the dequeued values (in order) and the number of number of
// values dequeued.
func (iq *deviceControllerInterruptQueue) drain() (
	q [configInterruptQueueSize]uintptr, n int,
) {
	for {
		m, ok := iq.deq()
		if !ok {
			return
		}
		q[n] = m
		n++
	}
}

func getQHBuffer(port uint8, qh int, ep int) *deviceControllerQH {
	if port < configDeviceCount && qh < 2 && ep < 2*configDeviceMaxEndpoints {
		return (*deviceControllerQH)(unsafe.Pointer(
			&deviceControllerQHBuffer[int(port)*configDeviceControllerQHAlign+
				qh*2*configDeviceMaxEndpoints*deviceControllerQHSize+
				ep*deviceControllerQHSize]))
	}
	return nil
}

func getDTDBuffer(port uint8, dtd int) *deviceControllerDTD {
	if port < configDeviceCount && dtd < configDeviceControllerMaxDTD {
		return (*deviceControllerDTD)(unsafe.Pointer(
			&deviceControllerDTDBuffer[int(port)*configDeviceControllerDTDAlign+
				dtd*deviceControllerDTDSize]))
	}
	return nil
}

func (s deviceControllerCapabilities) pack() deviceControllerCapabilitiesBitmap {
	return deviceControllerCapabilitiesBitmap(
		((uint32(s.reserved1) & 0x7FFF) << 0) | // uint16 // 15 (bits)
			((uint32(s.ios) & 0x1) << 15) | // uint8  // 1
			((uint32(s.maxPacketSize) & 0x7FF) << 16) | // uint16 // 11
			((uint32(s.reserved2) & 0x3) << 27) | // uint8  // 2
			((uint32(s.zlt) & 0x1) << 29) | // uint8  // 1
			((uint32(s.mult) & 0x3) << 30)) // uint8  // 2 (= 32 bits)
}

func (s deviceControllerDTDToken) pack() deviceControllerDTDTokenBitmap {
	return deviceControllerDTDTokenBitmap(
		((uint32(s.status) & 0xFF) << 0) | // uint8  // 8 (bits)
			((uint32(s.reserved1) & 0x3) << 8) | // uint8  // 2
			((uint32(s.multiplierOverride) & 0x3) << 10) | // uint8  // 2
			((uint32(s.reserved2) & 0x7) << 12) | // uint8  // 3
			((uint32(s.ioc) & 0x1) << 15) | // uint8  // 1
			((uint32(s.totalBytes) & 0x7FFF) << 16) | // uint16 // 15
			((uint32(s.reserved3) & 0x1) << 31)) // uint8  // 1 (= 32 bits)
}

func (b deviceControllerDTDTokenBitmap) unpack() deviceControllerDTDToken {
	return deviceControllerDTDToken{
		status:             uint8(b>>0) & 0xFF,
		reserved1:          uint8(b>>8) & 0x3,
		multiplierOverride: uint8(b>>10) & 0x3,
		reserved2:          uint8(b>>12) & 0x7,
		ioc:                uint8(b>>15) & 0x1,
		totalBytes:         uint16(b>>16) & 0x7FFF,
		reserved3:          uint8(b>>31) & 0x1,
	}
}

func (s deviceControllerEndpointStatus) pack() deviceControllerEndpointStatusBitmap {
	return deviceControllerEndpointStatusBitmap(
		((uint32(s.isOpened) & 0x1) << 0) | // uint8  // 1 (bits)
			((uint32(s.zlt) & 0x1) << 1)) // uint8  // 1
}

func (s deviceControllerOriginalBuffer) pack() deviceControllerOriginalBufferBitmap {
	return deviceControllerOriginalBufferBitmap(
		((uint32(s.originalBufferOffset) & 0xFFF) << 0) | // uint16 // 12 (bits)
			((uint32(s.originalBufferLength) & 0x7FFFF) << 12) | // uint32 // 19
			((uint32(s.dtdInvalid) & 0x1) << 31)) // uint8  // 1 (= 32 bits)
}

func (b deviceControllerOriginalBufferBitmap) unpack() deviceControllerOriginalBuffer {
	return deviceControllerOriginalBuffer{
		originalBufferOffset: uint16(b>>0) & 0xFFF,
		originalBufferLength: uint32(b>>12) & 0x7FFFF,
		dtdInvalid:           uint8(b>>31) & 0x1,
	}
}

// cycles converts the given number of microseconds to CPU cycles for a CPU with
// given frequency.
//go:inline
func cycles(microsec, cpuFreqHz uint32) uint32 {
	return uint32((uint64(microsec) * uint64(cpuFreqHz)) / 1000000)
}
