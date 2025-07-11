// Serial I/O Protocol (SIOP) – §13.4 UEFI 2.10
package uefi

import (
	"unsafe"
)

//---------------------------------------------------------------------------
//  GUID                                                                     //
//---------------------------------------------------------------------------

// {BB25CF6F-F1A1-4F11-9E5A-AE8C109A771F}
var EFI_SERIAL_IO_PROTOCOL_GUID = EFI_GUID{
	0xbb25cf6f, 0xf1a1, 0x4f11,
	[8]uint8{0x9e, 0x5a, 0xae, 0x8c, 0x10, 0x9a, 0x77, 0x1f},
}

//---------------------------------------------------------------------------
//  Enums / bit-fields                                                       //
//---------------------------------------------------------------------------

// Parity – §13.4.1 Table 13-3
const (
	ParityDefault = iota
	ParityNone
	ParityEven
	ParityOdd
	ParityMark
	ParitySpace
)

// Stop bits – §13.4.1 Table 13-4
const (
	StopBitsDefault = iota
	StopBits1
	StopBits1_5
	StopBits2
)

// Control-bit masks – §13.4.2 Table 13-5
const (
	SERIAL_CLEAR_TO_SEND = 1 << iota
	SERIAL_DATA_SET_READY
	SERIAL_RING_INDICATE
	SERIAL_CARRIER_DETECT
	SERIAL_INPUT_BUFFER_EMPTY
	SERIAL_OUTPUT_BUFFER_EMPTY
	SERIAL_HARDWARE_LOOPBACK_ENABLE
	SERIAL_SOFTWARE_LOOPBACK_ENABLE
	SERIAL_HARDWARE_FLOW_CONTROL
	SERIAL_SOFTWARE_FLOW_CONTROL
	SERIAL_DEVICE_ENABLE
)

//---------------------------------------------------------------------------
//  Helper structs                                                           //
//---------------------------------------------------------------------------

// 13.4.1 EFI_SERIAL_IO_MODE
type EFI_SERIAL_IO_MODE struct {
	ControlMask, Timeout                         uint32
	BaudRate                                     uint64
	ReceiveFifoDepth, DataBits, Parity, StopBits uint32
}

//---------------------------------------------------------------------------
//  EFI_SERIAL_IO_PROTOCOL                                                   //
//---------------------------------------------------------------------------

// Function table order matches §13.4
type EFI_SERIAL_IO_PROTOCOL struct {
	Revision      uint32
	reset         uintptr // (*this, extendedVerification)
	setAttributes uintptr // (*this, baud, depth, timeout, parity, databits, stopbits)
	setControl    uintptr // (*this, control)
	getControl    uintptr // (*this, *control)
	write         uintptr // (*this, *bufSize, buf)
	read          uintptr // (*this, *bufSize, buf)
	Mode          *EFI_SERIAL_IO_MODE
}

// ----------------- method wrappers -----------------

// Reset the device.
func (p *EFI_SERIAL_IO_PROTOCOL) Reset(extendedVerification BOOLEAN) EFI_STATUS {
	return UefiCall2(
		p.reset,
		uintptr(unsafe.Pointer(p)),
		convertBoolean(extendedVerification),
	)
}

// SetAttributes – configure baud/format.
func (p *EFI_SERIAL_IO_PROTOCOL) SetAttributes(
	baudRate uint64,
	receiveFifoDepth uint64,
	timeout uint32,
	parity, dataBits, stopBits uint32,
) EFI_STATUS {
	return UefiCall7(
		p.setAttributes,
		uintptr(unsafe.Pointer(p)),
		uintptr(baudRate),
		uintptr(receiveFifoDepth),
		uintptr(timeout),
		uintptr(parity),
		uintptr(dataBits),
		uintptr(stopBits),
	)
}

// SetControl – raise/clear control bits.
func (p *EFI_SERIAL_IO_PROTOCOL) SetControl(control uint32) EFI_STATUS {
	return UefiCall2(
		p.setControl,
		uintptr(unsafe.Pointer(p)),
		uintptr(control),
	)
}

// GetControl – query control bits.
func (p *EFI_SERIAL_IO_PROTOCOL) GetControl(control *uint32) EFI_STATUS {
	return UefiCall2(
		p.getControl,
		uintptr(unsafe.Pointer(p)),
		uintptr(unsafe.Pointer(control)),
	)
}

// Write bytes to the UART. *bufSize* is in/out.
func (p *EFI_SERIAL_IO_PROTOCOL) Write(bufSize *UINTN, buffer unsafe.Pointer) EFI_STATUS {
	return UefiCall3(
		p.write,
		uintptr(unsafe.Pointer(p)),
		uintptr(unsafe.Pointer(bufSize)),
		uintptr(buffer),
	)
}

// Read bytes from the UART. *bufSize* is in/out.
func (p *EFI_SERIAL_IO_PROTOCOL) Read(bufSize *UINTN, buffer unsafe.Pointer) EFI_STATUS {
	return UefiCall3(
		p.read,
		uintptr(unsafe.Pointer(p)),
		uintptr(unsafe.Pointer(bufSize)),
		uintptr(buffer),
	)
}
