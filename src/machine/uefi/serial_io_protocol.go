// Serial I/O Protocol (SIOP) – §12.8 UEFI 2.10
package uefi

import (
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

// Control-bit masks – §13.4.2
const (
	EFI_SERIAL_DATA_TERMINAL_READY          uint32 = 0x0001
	EFI_SERIAL_REQUEST_TO_SEND              uint32 = 0x0002
	EFI_SERIAL_CLEAR_TO_SEND                uint32 = 0x0010
	EFI_SERIAL_DATA_SET_READY               uint32 = 0x0020
	EFI_SERIAL_RING_INDICATE                uint32 = 0x0040
	EFI_SERIAL_CARRIER_DETECT               uint32 = 0x0080
	EFI_SERIAL_INPUT_BUFFER_EMPTY           uint32 = 0x0100
	EFI_SERIAL_OUTPUT_BUFFER_EMPTY          uint32 = 0x0200
	EFI_SERIAL_HARDWARE_LOOPBACK_ENABLE     uint32 = 0x1000
	EFI_SERIAL_SOFTWARE_LOOPBACK_ENABLE     uint32 = 0x2000
	EFI_SERIAL_HARDWARE_FLOW_CONTROL_ENABLE uint32 = 0x4000
)

// EFI_SERIAL_IO_MODE is the output-only struct for checking the status of
// a serial port.  §12.8.1
type EFI_SERIAL_IO_MODE struct {
	ControlMask, Timeout                         uint32
	BaudRate                                     uint64
	ReceiveFifoDepth, DataBits, Parity, StopBits uint32
}

// EFI_SERIAL_IO_PROTOCOL Function table order matches §12.8.1
type EFI_SERIAL_IO_PROTOCOL struct {
	Revision       uint32
	reset          uintptr // (*this)
	setAttributes  uintptr // (*this, baud, depth, timeout, parity, databits, stopbits)
	setControl     uintptr // (*this, control)
	getControl     uintptr // (*this, *control)
	write          uintptr // (*this, *bufSize, buf)
	read           uintptr // (*this, *bufSize, buf)
	Mode           *EFI_SERIAL_IO_MODE
	deviceTypeGuid uintptr
}

// Reset the device.  §12.8.3.1
func (p *EFI_SERIAL_IO_PROTOCOL) Reset() EFI_STATUS {
	return UefiCall1(
		p.reset,
		uintptr(unsafe.Pointer(p)),
	)
}

// SetAttributes configures baud/format.
// Setting baudRate, receiveFifoDepth, or timeout to 0 *SHOULD* tell
// the port driver to use sane default values. YMMV.
// §12.8.3.2
func (p *EFI_SERIAL_IO_PROTOCOL) SetAttributes(
	baudRate uint64,
	receiveFifoDepth uint32,
	timeout uint32,
	parity uint32,
	dataBits int8,
	stopBits uint32,
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

// SetControl sets or clears control bits.
// §12.8.3.3
func (p *EFI_SERIAL_IO_PROTOCOL) SetControl(control uint32) EFI_STATUS {
	return UefiCall2(
		p.setControl,
		uintptr(unsafe.Pointer(p)),
		uintptr(control),
	)
}

// GetControl queries control bits.
// §12.8.3.4
func (p *EFI_SERIAL_IO_PROTOCOL) GetControl(control *uint32) EFI_STATUS {
	return UefiCall2(
		p.getControl,
		uintptr(unsafe.Pointer(p)),
		uintptr(unsafe.Pointer(control)),
	)
}

// Write bytes to the UART. When calling, bufSize is the size of buffer. After Write has
// returned, bufSize will be set to the number of bytes actually written.
// §12.8.3.5
func (p *EFI_SERIAL_IO_PROTOCOL) Write(bufSize *UINTN, buffer unsafe.Pointer) EFI_STATUS {
	return UefiCall3(
		p.write,
		uintptr(unsafe.Pointer(p)),
		uintptr(unsafe.Pointer(bufSize)),
		uintptr(buffer),
	)
}

// Read bytes from the UART. When calling, bufSize is size of buffer.
// After Read has returned, bufSize is set to the number of bytes actually read.
// §12.8.3.6
func (p *EFI_SERIAL_IO_PROTOCOL) Read(bufSize *UINTN, buffer unsafe.Pointer) EFI_STATUS {
	return UefiCall3(
		p.read,
		uintptr(unsafe.Pointer(p)),
		uintptr(unsafe.Pointer(bufSize)),
		uintptr(buffer),
	)
}

// SerialPort wraps an EFI_SERIAL_IO_PROTOCOL and provides an idiomatic-go API
// TODO: make serial ports implement os.File
type SerialPort struct {
	*EFI_SERIAL_IO_PROTOCOL
	flowControl bool
}

// EnumerateSerialPorts uses UEFI's handle walking API to discover
// serial ports. SerialPorts may be in use for something else, so
// EnumerateSerialPorts doesn't Init() any of the ports it returns.
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
	}

	return ports, nil
}

// Init configures sp to use 115200 baud, 8N1.
// Read/Write timeout and Receive FIFO depth
// are left ot the serial driver's discretion.
func (sp *SerialPort) Init() error {
	// can't hurt, can it?
	status := sp.Reset()
	if status != EFI_SUCCESS {
		return StatusError(status)
	}
	// attempt a write to kick lazy FW/drivers into action
	sp.Write([]byte{0})

	status = sp.SetAttributes(
		115200,     // BaudRate
		0,          // ReceiveFifoDepth (0 = default)
		0,          // Timeout (0 = default)
		ParityNone, // EFI_PARITY_TYPE (1 = none)
		8,          // DataBits
		StopBits1,  // StopBits (1)
	)
	if status != EFI_SUCCESS {
		println("SetAttributes failed\r")
		return StatusError(status)
	}

	return nil
}

// Read implements io.Reader. Reads are blocking operations, however if the read
// operation times out, this method yields to other goroutines. If using hardware
// flow control, RTS is set and cleared accordingly.
func (sp *SerialPort) Read(buf []byte) (n int, err error) {
	var controlBits uint32
	sp.GetControl(&controlBits)
	for {
		bufLen := UINTN(len(buf))
		// assert RTS and DTR; should be no harm to set/clear these even if hardware flow control
		// is not in use.
		sp.SetControl(controlBits | EFI_SERIAL_REQUEST_TO_SEND | EFI_SERIAL_DATA_TERMINAL_READY)
		status := sp.EFI_SERIAL_IO_PROTOCOL.Read(&bufLen, unsafe.Pointer(&buf[0]))
		if bufLen == 0 {
			status = EFI_TIMEOUT
		}
		switch status {
		case EFI_SUCCESS:
			return int(bufLen), nil
		case EFI_TIMEOUT, EFI_NO_RESPONSE:
			// deassert RTS
			sp.SetControl(controlBits & ^(EFI_SERIAL_REQUEST_TO_SEND))
			gosched() // let other stuff run
			continue
		default:
			println(StatusError(status).Error())
			return 0, StatusError(status)
		}
	}
}

/*
func (sp *SerialPort) WriteTo(w io.Writer) (n int64, err error) {
	var nr, nw int
	// buf := make([]byte, sp.Mode.ReceiveFifoDepth)
	buf := make([]byte, 24)
	for err == nil {
		nr, err = sp.Read(buf)
		if err != nil {
			return
		}
		nw, err = sp.Write(buf[:nr])
		n += int64(nw)
	}
	return
}
*/

// Write implements io.Writer. If hardware flow control is enabled and CTS
// is cleared, Write will return (0, nil) indicating no bytes were written,
// but this is not an error. Some consumers of the io.Writer interface may
// not like this.
func (sp *SerialPort) Write(buf []byte) (n int, err error) {
	var controlBits uint32
	sp.GetControl(&controlBits)
	if sp.flowControl && controlBits&EFI_SERIAL_CLEAR_TO_SEND == 0 {
		return 0, nil
	}
	bufLen := UINTN(len(buf))
	status := sp.EFI_SERIAL_IO_PROTOCOL.Write(&bufLen, unsafe.Pointer(&buf[0]))
	if status != EFI_SUCCESS {
		return int(bufLen), StatusError(status)
	}
	return int(bufLen), nil
}

// UseFlowControl enables hardware flow control when fc is true.
func (sp *SerialPort) UseFlowControl(fc bool) {
	sp.flowControl = fc
}
