// Serial I/O Protocol (SIOP) – §13.4 UEFI 2.10
package uefi

import (
	"io"
	"unsafe"
)

//---------------------------------------------------------------------------
//  GUID                                                                     //
//---------------------------------------------------------------------------

// {BB25CF6F-F1A1-4F11-9E5A-AE8C109A771F}
var EFI_SERIAL_IO_PROTOCOL_GUID = EFI_GUID{
	0xBB25CF6F, 0xF1D4, 0x11D2,
	[8]uint8{0x9a, 0x0c, 0x00, 0x90, 0x27, 0x3f, 0xc1, 0xfd},
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
	EFI_SERIAL_DATA_TERMINAL_READY          = 0x0001
	EFI_SERIAL_REQUEST_TO_SEND              = 0x0002
	EFI_SERIAL_CLEAR_TO_SEND                = 0x0010
	EFI_SERIAL_DATA_SET_READY               = 0x0020
	EFI_SERIAL_RING_INDICATE                = 0x0040
	EFI_SERIAL_CARRIER_DETECT               = 0x0080
	EFI_SERIAL_INPUT_BUFFER_EMPTY           = 0x0100
	EFI_SERIAL_OUTPUT_BUFFER_EMPTY          = 0x0200
	EFI_SERIAL_HARDWARE_LOOPBACK_ENABLE     = 0x1000
	EFI_SERIAL_SOFTWARE_LOOPBACK_ENABLE     = 0x2000
	EFI_SERIAL_HARDWARE_FLOW_CONTROL_ENABLE = 0x4000
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

// TODO: make serial ports implement os.File
type SerialPort struct {
	*EFI_SERIAL_IO_PROTOCOL
}

// Init configures sp to use 115200 baud, 8N1.
// Read/Write timeout and Receive FIFO depth
// are left ot the serial driver's discretion.
func (sp *SerialPort) Init() error {
	// can't hurt, can it?
	status := sp.Reset(true)
	if status != EFI_SUCCESS {
		return StatusError(status)
	}
	status = sp.SetAttributes(
		115200,     // BaudRate
		0,          // ReceiveFifoDepth (0 = default)
		0,          // Timeout (0 = default, but maybe try 1?)
		ParityNone, // EFI_PARITY_TYPE (1 = none)
		8,          // DataBits
		StopBits1,  // StopBits (1)
	)
	if status != EFI_SUCCESS {
		return StatusError(status)
	}

	return nil
}

func (sp *SerialPort) Read(buf []byte) (n int, err error) {
	bufLen := UINTN(len(buf))
	for {
		status := sp.EFI_SERIAL_IO_PROTOCOL.Read(&bufLen, unsafe.Pointer(&buf[0]))
		switch status {
		case EFI_SUCCESS:
			return int(bufLen), nil
		case EFI_TIMEOUT, EFI_NO_RESPONSE:
			gosched() // let other stuff run
			continue
		default:
			return 0, StatusError(status)
		}
	}
}

func (sp *SerialPort) WriteTo(w io.Writer) (n int64, err error) {
	buf := make([]byte, 1024)
	for {
		bufLen := UINTN(len(buf))
		status := sp.EFI_SERIAL_IO_PROTOCOL.Read(&bufLen, unsafe.Pointer(&buf[0]))
		switch status {
		case EFI_SUCCESS:
			nw, err := w.Write(buf[:bufLen])
			n += int64(nw)
			if err != nil {
				return n, err
			}
		case EFI_TIMEOUT, EFI_NO_RESPONSE:
			gosched() // let other stuff run
			continue
		default:
			return 0, StatusError(status)
		}
	}
}

func (sp *SerialPort) Write(buf []byte) (n int, err error) {
	bufLen := UINTN(len(buf))
	status := sp.EFI_SERIAL_IO_PROTOCOL.Write(&bufLen, unsafe.Pointer(&buf[0]))
	if status != EFI_SUCCESS {
		return int(bufLen), StatusError(status)
	}
	return int(bufLen), nil
}

// EnumerateSerialPorts uses UEFI's handle walking API to discover
// serial ports.
func EnumerateSerialPorts() ([]*SerialPort, error) {
	var (
		handleCount  UINTN
		handleBuffer *EFI_HANDLE
	)
	status := BS().LocateHandleBuffer(ByProtocol, &EFI_SERIAL_IO_PROTOCOL_GUID, nil, &handleCount, &handleBuffer)
	if status != EFI_SUCCESS {
		return nil, StatusError(status)
	}
	// if none were found, we should have gotten EFI_NOT_FOUND

	//turn handleBuffer into a slice of EFI_HANDLEs
	handleSlice := unsafe.Slice((*EFI_HANDLE)(unsafe.Pointer(handleBuffer)), int(handleCount))
	ports := make([]*SerialPort, int(handleCount))

	for i := range int(handleCount) {
		var serial *EFI_SERIAL_IO_PROTOCOL
		status := BS().HandleProtocol(
			handleSlice[i],
			&EFI_SERIAL_IO_PROTOCOL_GUID,
			unsafe.Pointer(&serial),
		)
		if status != EFI_SUCCESS {
			return nil, StatusError(status)
		}
		ports[i] = &SerialPort{EFI_SERIAL_IO_PROTOCOL: serial}

		// arm each port with sane default
		ports[i].Init()
	}

	return ports, nil
}
