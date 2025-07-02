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

	endPoints = []uint32{
		usb.CONTROL_ENDPOINT:  usb.ENDPOINT_TYPE_CONTROL,
		usb.CDC_ENDPOINT_ACM:  (usb.ENDPOINT_TYPE_INTERRUPT | usb.EndpointIn),
		usb.CDC_ENDPOINT_OUT:  (usb.ENDPOINT_TYPE_BULK | usb.EndpointOut),
		usb.CDC_ENDPOINT_IN:   (usb.ENDPOINT_TYPE_BULK | usb.EndpointIn),
		usb.HID_ENDPOINT_IN:   (usb.ENDPOINT_TYPE_DISABLE), // Interrupt In
		usb.HID_ENDPOINT_OUT:  (usb.ENDPOINT_TYPE_DISABLE), // Interrupt Out
		usb.MIDI_ENDPOINT_IN:  (usb.ENDPOINT_TYPE_DISABLE), // Bulk In
		usb.MIDI_ENDPOINT_OUT: (usb.ENDPOINT_TYPE_DISABLE), // Bulk Out
	}
)

func initEndpoint(ep, config uint32) {
	val := uint32(usbEpControlEnable) | uint32(usbEpControlInterruptPerBuff)
	offset := ep*2*usbBufferLen + 0x100
	val |= offset

	// Bulk and interrupt endpoints must have their Packet ID reset to DATA0 when un-stalled.
	epXPIDReset[ep] = false // Default to false in case an endpoint is re-initialized.

	switch config {
	case usb.ENDPOINT_TYPE_INTERRUPT | usb.EndpointIn:
		val |= usbEpControlEndpointTypeInterrupt
		_usbDPSRAM.EPxControl[ep].In.Set(val)
		epXPIDReset[ep] = true

	case usb.ENDPOINT_TYPE_BULK | usb.EndpointOut:
		val |= usbEpControlEndpointTypeBulk
		_usbDPSRAM.EPxControl[ep].Out.Set(val)
		_usbDPSRAM.EPxBufferControl[ep].Out.Set(usbBufferLen & usbBuf0CtrlLenMask)
		_usbDPSRAM.EPxBufferControl[ep].Out.SetBits(usbBuf0CtrlAvail)
		epXPIDReset[ep] = true

	case usb.ENDPOINT_TYPE_INTERRUPT | usb.EndpointOut:
		val |= usbEpControlEndpointTypeInterrupt
		_usbDPSRAM.EPxControl[ep].Out.Set(val)
		_usbDPSRAM.EPxBufferControl[ep].Out.Set(usbBufferLen & usbBuf0CtrlLenMask)
		_usbDPSRAM.EPxBufferControl[ep].Out.SetBits(usbBuf0CtrlAvail)
		epXPIDReset[ep] = true

	case usb.ENDPOINT_TYPE_BULK | usb.EndpointIn:
		val |= usbEpControlEndpointTypeBulk
		_usbDPSRAM.EPxControl[ep].In.Set(val)
		epXPIDReset[ep] = true

	case usb.ENDPOINT_TYPE_CONTROL:
		val |= usbEpControlEndpointTypeControl
		_usbDPSRAM.EPxBufferControl[ep].Out.Set(usbBuf0CtrlData1Pid)
		_usbDPSRAM.EPxBufferControl[ep].Out.SetBits(usbBuf0CtrlAvail)

	}
}

// SendUSBInPacket sends a packet for USB (interrupt in / bulk in).
func SendUSBInPacket(ep uint32, data []byte) bool {
	sendUSBPacket(ep, data, 0)
	return true
}

// Prevent file size increases: https://github.com/tinygo-org/tinygo/pull/998
//
//go:noinline
func sendUSBPacket(ep uint32, data []byte, maxsize uint16) {
	count := len(data)
	if 0 < int(maxsize) && int(maxsize) < count {
		count = int(maxsize)
	}

	if ep == 0 {
		if count > usb.EndpointPacketSize {
			count = usb.EndpointPacketSize

			sendOnEP0DATADONE.offset = count
			sendOnEP0DATADONE.data = data
		} else {
			sendOnEP0DATADONE.offset = 0
		}
		epXdata0[ep] = true
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
	setEPDataPID(ep, !epXdata0[ep])
}

// Set the USB endpoint Packet ID to DATA0 or DATA1.
func setEPDataPID(ep uint32, dataOne bool) {
	epXdata0[ep] = dataOne
	if epXdata0[ep] || ep == 0 {
		_usbDPSRAM.EPxBufferControl[ep].Out.SetBits(usbBuf0CtrlData1Pid)
	}

	_usbDPSRAM.EPxBufferControl[ep].Out.SetBits(usbBuf0CtrlAvail)
}

func SendZlp() {
	sendUSBPacket(0, []byte{}, 0)
}

func sendViaEPIn(ep uint32, data []byte, count int) {
	// Prepare buffer control register value
	val := uint32(count) | usbBuf0CtrlAvail

	// DATA0 or DATA1
	epXdata0[ep&0x7F] = !epXdata0[ep&0x7F]
	if !epXdata0[ep&0x7F] {
		val |= usbBuf0CtrlData1Pid
	}

	// Mark as full
	val |= usbBuf0CtrlFull

	copy(_usbDPSRAM.EPxBuffer[ep&0x7F].Buffer0[:], data[:count])
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
	if epXPIDReset[ep] {
		// Reset the PID to DATA0
		setEPDataPID(ep, false)
	}
}

// Clear the ENDPOINT_HALT/stall on a USB OUT endpoint.
func (dev *USBDevice) ClearStallEPOut(ep uint32) {
	ep = ep & 0x7F
	val := uint32(usbBuf0CtrlStall)
	_usbDPSRAM.EPxBufferControl[ep].Out.ClearBits(val)
	if epXPIDReset[ep] {
		// Reset the PID to DATA0
		setEPDataPID(ep, false)
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
	_usbDPSRAM  = (*usbDPSRAM)(unsafe.Pointer(uintptr(0x50100000)))
	epXdata0    [16]bool
	epXPIDReset [16]bool
	setupBytes  [8]byte
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
