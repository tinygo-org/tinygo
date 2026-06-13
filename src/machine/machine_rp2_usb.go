//go:build rp2040 || rp2350

package machine

import (
	"machine/usb"
	"runtime/volatile"
	"unsafe"
)

const NumberOfUSBEndpoints = 8

var (
	sendOnEP0DATADONE struct {
		offset int
		data   []byte
		pid    uint32
	}
)

func initEndpoint(ep, config uint32) {
	val := uint32(usbEpControlEnable) | uint32(usbEpControlInterruptPerBuff)

	// Each endpoint has 128 bytes of DPRAM buffer space allocated (2 * usbBufferLen).
	// To support bidirectional configurations using the same endpoint number,
	// we allocate the first 64 bytes (Buffer0) to OUT transfers, and the remaining
	// 64 bytes (Buffer1) to IN transfers by shifting the IN offset by usbBufferLen.
	offset := ep*2*usbBufferLen + 0x100

	// Bulk and interrupt endpoints must have their Packet ID reset to DATA0 when un-stalled.
	// Since both directions share the same ep, we reset their PID flags independently.
	if (config & usb.EndpointIn) != 0 {
		epXPIDResetIn[ep] = false
	} else {
		epXPIDResetOut[ep] = false
	}

	switch config {
	case usb.ENDPOINT_TYPE_INTERRUPT | usb.EndpointIn:
		epXPIDResetIn[ep] = true
		epXdata0In[ep] = false
		val |= offset + usbBufferLen
		val |= usbEpControlEndpointTypeInterrupt
		_usbDPSRAM.EPxControl[ep].In.Set(val)

	case usb.ENDPOINT_TYPE_BULK | usb.EndpointOut:
		epXPIDResetOut[ep] = true
		epXdata0Out[ep] = false
		val |= offset
		val |= usbEpControlEndpointTypeBulk
		_usbDPSRAM.EPxControl[ep].Out.Set(val)
		_usbDPSRAM.EPxBufferControl[ep].Out.Set(usbBufferLen & usbBuf0CtrlLenMask)
		_usbDPSRAM.EPxBufferControl[ep].Out.SetBits(usbBuf0CtrlAvail)

	case usb.ENDPOINT_TYPE_INTERRUPT | usb.EndpointOut:
		epXPIDResetOut[ep] = true
		epXdata0Out[ep] = false
		val |= offset
		val |= usbEpControlEndpointTypeInterrupt
		_usbDPSRAM.EPxControl[ep].Out.Set(val)
		_usbDPSRAM.EPxBufferControl[ep].Out.Set(usbBufferLen & usbBuf0CtrlLenMask)
		_usbDPSRAM.EPxBufferControl[ep].Out.SetBits(usbBuf0CtrlAvail)

	case usb.ENDPOINT_TYPE_BULK | usb.EndpointIn:
		epXPIDResetIn[ep] = true
		epXdata0In[ep] = false
		val |= offset + usbBufferLen
		val |= usbEpControlEndpointTypeBulk
		_usbDPSRAM.EPxControl[ep].In.Set(val)

	case usb.ENDPOINT_TYPE_CONTROL:
		val |= offset
		val |= usbEpControlEndpointTypeControl
		_usbDPSRAM.EPxBufferControl[ep].Out.Set(usbBuf0CtrlData1Pid)
		_usbDPSRAM.EPxBufferControl[ep].Out.SetBits(usbBuf0CtrlAvail)

	}
}

// SendUSBInPacket sends a packet for USB (interrupt in / bulk in).
func SendUSBInPacket(ep uint32, data []byte) bool {
	sendUSBPacket(ep, data)
	return true
}

// Prevent file size increases: https://github.com/tinygo-org/tinygo/pull/998
//
//go:noinline
func sendUSBPacket(ep uint32, data []byte) {
	count := len(data)
	if ep == 0 {
		if count > usb.EndpointPacketSize {
			count = usb.EndpointPacketSize

			sendOnEP0DATADONE.offset = count
			sendOnEP0DATADONE.data = data
		} else {
			sendOnEP0DATADONE.offset = 0
		}
		epXdata0In[ep] = true
	}

	sendViaEPIn(ep, data, count)
}

func ReceiveUSBControlPacket() ([cdcLineInfoSize]byte, error) {
	var b [cdcLineInfoSize]byte
	ep := 0

	for !_usbDPSRAM.EPxBufferControl[ep].Out.HasBits(usbBuf0CtrlFull) {
		// TODO: timeout
	}

	ctrl := _usbDPSRAM.EPxBufferControl[ep].Out.Get()
	_usbDPSRAM.EPxBufferControl[ep].Out.Set(usbBufferLen & usbBuf0CtrlLenMask)
	sz := ctrl & usbBuf0CtrlLenMask

	copy(b[:], _usbDPSRAM.EPxBuffer[ep].Buffer0[:sz])

	_usbDPSRAM.EPxBufferControl[ep].Out.SetBits(usbBuf0CtrlData1Pid)
	_usbDPSRAM.EPxBufferControl[ep].Out.SetBits(usbBuf0CtrlAvail)

	return b, nil
}

func handleEndpointRx(ep uint32) []byte {
	ctrl := _usbDPSRAM.EPxBufferControl[ep].Out.Get()
	_usbDPSRAM.EPxBufferControl[ep].Out.Set(usbBufferLen & usbBuf0CtrlLenMask)
	sz := ctrl & usbBuf0CtrlLenMask

	return _usbDPSRAM.EPxBuffer[ep].Buffer0[:sz]
}

// AckUsbOutTransfer is called to acknowledge the completion of a USB OUT transfer.
func AckUsbOutTransfer(ep uint32) {
	ep = ep & 0x7F
	setEPDataPIDOut(ep, !epXdata0Out[ep])
}

// Set the USB endpoint Packet ID to DATA0 or DATA1 for OUT direction.
func setEPDataPIDOut(ep uint32, dataOne bool) {
	epXdata0Out[ep] = dataOne
	if epXdata0Out[ep] || ep == 0 {
		_usbDPSRAM.EPxBufferControl[ep].Out.SetBits(usbBuf0CtrlData1Pid)
	}

	_usbDPSRAM.EPxBufferControl[ep].Out.SetBits(usbBuf0CtrlAvail)
}

// Set the USB endpoint Packet ID to DATA0 or DATA1 for IN direction.
func setEPDataPIDIn(ep uint32, dataOne bool) {
	epXdata0In[ep] = dataOne
	if epXdata0In[ep] || ep == 0 {
		_usbDPSRAM.EPxBufferControl[ep].In.SetBits(usbBuf0CtrlData1Pid)
	}

	_usbDPSRAM.EPxBufferControl[ep].In.SetBits(usbBuf0CtrlAvail)
}

func SendZlp() {
	sendUSBPacket(0, nil)
}

func sendViaEPIn(ep uint32, data []byte, count int) {
	// Prepare buffer control register value
	val := uint32(count) | usbBuf0CtrlAvail

	// DATA0 or DATA1
	epXdata0In[ep&0x7F] = !epXdata0In[ep&0x7F]
	if !epXdata0In[ep&0x7F] {
		val |= usbBuf0CtrlData1Pid
	}

	// Mark as full
	val |= usbBuf0CtrlFull

	if (ep & 0x7F) == 0 {
		copy(_usbDPSRAM.EPxBuffer[0].Buffer0[:], data[:count])
	} else {
		copy(_usbDPSRAM.EPxBuffer[ep&0x7F].Buffer1[:], data[:count])
	}
	_usbDPSRAM.EPxBufferControl[ep&0x7F].In.Set(val)
}

// Set ENDPOINT_HALT/stall status on a USB IN endpoint.
func (dev *USBDevice) SetStallEPIn(ep uint32) {
	ep = ep & 0x7F
	// Prepare buffer control register value
	if ep == 0 {
		armEPZeroStall()
	}
	val := uint32(usbBuf0CtrlFull)
	_usbDPSRAM.EPxBufferControl[ep].In.Set(val)
	val |= uint32(usbBuf0CtrlStall)
	_usbDPSRAM.EPxBufferControl[ep].In.Set(val)
}

// Set ENDPOINT_HALT/stall status on a USB OUT endpoint.
func (dev *USBDevice) SetStallEPOut(ep uint32) {
	ep = ep & 0x7F
	if ep == 0 {
		panic("SetStallEPOut: EP0 OUT not valid")
	}
	val := uint32(usbBuf0CtrlStall)
	_usbDPSRAM.EPxBufferControl[ep].Out.Set(val)
}

// Clear the ENDPOINT_HALT/stall on a USB IN endpoint.
func (dev *USBDevice) ClearStallEPIn(ep uint32) {
	ep = ep & 0x7F
	val := uint32(usbBuf0CtrlStall)
	_usbDPSRAM.EPxBufferControl[ep].In.ClearBits(val)
	if epXPIDResetIn[ep] {
		// Reset the PID to DATA0
		setEPDataPIDIn(ep, false)
	}
}

// Clear the ENDPOINT_HALT/stall on a USB OUT endpoint.
func (dev *USBDevice) ClearStallEPOut(ep uint32) {
	ep = ep & 0x7F
	val := uint32(usbBuf0CtrlStall)
	_usbDPSRAM.EPxBufferControl[ep].Out.ClearBits(val)
	if epXPIDResetOut[ep] {
		// Reset the PID to DATA0
		setEPDataPIDOut(ep, false)
	}
}

type usbDPSRAM struct {
	// Note that EPxControl[0] is not EP0Control but 8-byte setup data.
	EPxControl [16]usbEndpointControlRegister

	EPxBufferControl [16]usbBufferControlRegister

	EPxBuffer [16]usbBuffer
}

type usbEndpointControlRegister struct {
	In  volatile.Register32
	Out volatile.Register32
}
type usbBufferControlRegister struct {
	In  volatile.Register32
	Out volatile.Register32
}

type usbBuffer struct {
	Buffer0 [usbBufferLen]byte
	Buffer1 [usbBufferLen]byte
}

var (
	_usbDPSRAM     = (*usbDPSRAM)(unsafe.Pointer(uintptr(0x50100000)))
	epXdata0In     [16]bool
	epXdata0Out    [16]bool
	epXPIDResetIn  [16]bool
	epXPIDResetOut [16]bool
	setupBytes     [8]byte
)

func (d *usbDPSRAM) setupBytes() []byte {

	data := d.EPxControl[usb.CONTROL_ENDPOINT].In.Get()
	setupBytes[0] = byte(data)
	setupBytes[1] = byte(data >> 8)
	setupBytes[2] = byte(data >> 16)
	setupBytes[3] = byte(data >> 24)

	data = d.EPxControl[usb.CONTROL_ENDPOINT].Out.Get()
	setupBytes[4] = byte(data)
	setupBytes[5] = byte(data >> 8)
	setupBytes[6] = byte(data >> 16)
	setupBytes[7] = byte(data >> 24)

	return setupBytes[:]
}

func (d *usbDPSRAM) clear() {
	for i := 0; i < len(d.EPxControl); i++ {
		d.EPxControl[i].In.Set(0)
		d.EPxControl[i].Out.Set(0)
		d.EPxBufferControl[i].In.Set(0)
		d.EPxBufferControl[i].Out.Set(0)
	}
}

const (
	// DPRAM : Endpoint control register
	usbEpControlEnable                 = 0x80000000
	usbEpControlDoubleBuffered         = 0x40000000
	usbEpControlInterruptPerBuff       = 0x20000000
	usbEpControlInterruptPerDoubleBuff = 0x10000000
	usbEpControlEndpointType           = 0x0c000000
	usbEpControlInterruptOnStall       = 0x00020000
	usbEpControlInterruptOnNak         = 0x00010000
	usbEpControlBufferAddress          = 0x0000ffff

	usbEpControlEndpointTypeControl   = 0x00000000
	usbEpControlEndpointTypeISO       = 0x04000000
	usbEpControlEndpointTypeBulk      = 0x08000000
	usbEpControlEndpointTypeInterrupt = 0x0c000000

	// Endpoint buffer control bits
	usbBuf1CtrlFull     = 0x80000000
	usbBuf1CtrlLast     = 0x40000000
	usbBuf1CtrlData0Pid = 0x20000000
	usbBuf1CtrlData1Pid = 0x00000000
	usbBuf1CtrlSel      = 0x10000000
	usbBuf1CtrlStall    = 0x08000000
	usbBuf1CtrlAvail    = 0x04000000
	usbBuf1CtrlLenMask  = 0x03FF0000
	usbBuf0CtrlFull     = 0x00008000
	usbBuf0CtrlLast     = 0x00004000
	usbBuf0CtrlData0Pid = 0x00000000
	usbBuf0CtrlData1Pid = 0x00002000
	usbBuf0CtrlSel      = 0x00001000
	usbBuf0CtrlStall    = 0x00000800
	usbBuf0CtrlAvail    = 0x00000400
	usbBuf0CtrlLenMask  = 0x000003FF

	usbBufferLen = 64
)
