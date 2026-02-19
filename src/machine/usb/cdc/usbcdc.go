//go:build baremetal

package cdc

import (
	"errors"
	"machine"
	"machine/usb"
	"sync/atomic"
	_ "unsafe"
)

var (
	ErrBufferEmpty = errors.New("USB-CDC buffer empty")
)

const cdcLineInfoSize = 7

type cdcLineInfo struct {
	dwDTERate   uint32
	bCharFormat uint8
	bParityType uint8
	bDataBits   uint8
	lineState   uint8
}

// USBCDC is the USB CDC aka serial over USB interface.
type USBCDC struct {
	tx       ring512
	rx       ring512
	inflight atomic.Uint32
	rbuf     [1]byte
	wbuf     [1]byte
}

var (
	// USB is a USB CDC interface.
	USB *USBCDC

	usbLineInfo = cdcLineInfo{115200, 0x00, 0x00, 0x08, 0x00}
)

// Read from the RX buffer.
func (usbcdc *USBCDC) Read(data []byte) (n int, err error) {
	data1, data2 := usbcdc.rx.Peek()
	n += copy(data, data1)
	n += copy(data[n:], data2)
	usbcdc.rx.Discard(uint32(n))
	return n, nil
}

// ReadByte reads a single byte from the RX buffer.
// If there is no data in the buffer, returns an error.
func (usbcdc *USBCDC) ReadByte() (byte, error) {
	// check if RX buffer is empty
	b, _ := usbcdc.rx.Peek()
	if len(b) > 0 {
		c := b[0]
		usbcdc.rx.Discard(1)
		return c, nil
	}
	return 0, ErrBufferEmpty
}

// Buffered returns the number of bytes currently stored in the RX buffer.
func (usbcdc *USBCDC) Buffered() int {
	return int(usbcdc.rx.Used())
}

// Receive handles adding data to the UART's data buffer.
// Usually called by the IRQ handler for a machine.
func (usbcdc *USBCDC) Receive(data byte) {
	usbcdc.rbuf[0] = data
	usbcdc.rx.Put(usbcdc.rbuf[:])
}

// Configure the USB CDC interface. The config is here for compatibility with the UART interface.
func (usbcdc *USBCDC) Configure(config machine.UARTConfig) error {
	return nil
}

func (usbcdc *USBCDC) txhandler() {
	// Mark data as sent.
	inflight := usbcdc.inflight.Load()
	usbcdc.tx.Discard(inflight)
	// Check if tx needs to send more data.
	used := usbcdc.tx.Used()
	usbcdc.inflight.Store(used)
	if used > 0 {
		data1, data2 := usbcdc.tx.Peek()
		usbcdc.send(data1)
		usbcdc.send(data2)
	}
}

// Flush flushes buffered data.
func (usbcdc *USBCDC) Flush() {
	for usbcdc.tx.Used() > 0 {
		gosched()
	}
}

// Write data to the USBCDC.
func (usbcdc *USBCDC) Write(data []byte) (n int, err error) {
	n = len(data)
	if usbLineInfo.lineState > 0 {
		if usbcdc.inflight.Load() == 0 && usbcdc.tx.Used() == 0 {
			// If no data inflight/inring, send directly to USB device.
			sz := min(len(data), 512)
			usbcdc.send(data[:sz])
			data = data[sz:]
		}
		for len(data) > 0 {
			tosend := min(len(data), int(usbcdc.tx.Free()))
			usbcdc.tx.Put(data[:tosend])
			data = data[tosend:]
			if len(data) > 0 {
				usbcdc.Flush()
			}
		}
	}
	return n, nil
}

func (usbcdc *USBCDC) send(data []byte) {
	for len(data) > 0 {
		off := min(usb.EndpointPacketSize, len(data))
		chunk := data[:off]
		data = data[off:]
		usbcdc.inflight.Add(uint32(len(data)))
		machine.SendUSBInPacket(cdcEndpointIn, chunk)
	}
}

// WriteByte writes a byte of data to the USB CDC interface.
func (usbcdc *USBCDC) WriteByte(c byte) error {
	usbcdc.wbuf[0] = c
	usbcdc.Write(usbcdc.wbuf[:])
	return nil
}

func (usbcdc *USBCDC) DTR() bool {
	return (usbLineInfo.lineState & usb_CDC_LINESTATE_DTR) > 0
}

func (usbcdc *USBCDC) RTS() bool {
	return (usbLineInfo.lineState & usb_CDC_LINESTATE_RTS) > 0
}

func cdcCallbackRx(b []byte) {
	free := USB.rx.Free()
	USB.rx.Put(b[:min(len(b), int(free))])
}

var cdcSetupBuff [cdcLineInfoSize]byte

func cdcSetup(setup usb.Setup) bool {
	if setup.BmRequestType == usb_REQUEST_DEVICETOHOST_CLASS_INTERFACE {
		if setup.BRequest == usb_CDC_GET_LINE_CODING {
			cdcSetupBuff[0] = byte(usbLineInfo.dwDTERate)
			cdcSetupBuff[1] = byte(usbLineInfo.dwDTERate >> 8)
			cdcSetupBuff[2] = byte(usbLineInfo.dwDTERate >> 16)
			cdcSetupBuff[3] = byte(usbLineInfo.dwDTERate >> 24)
			cdcSetupBuff[4] = byte(usbLineInfo.bCharFormat)
			cdcSetupBuff[5] = byte(usbLineInfo.bParityType)
			cdcSetupBuff[6] = byte(usbLineInfo.bDataBits)

			machine.SendUSBInPacket(0, cdcSetupBuff[:])
			return true
		}
	}

	if setup.BmRequestType == usb_REQUEST_HOSTTODEVICE_CLASS_INTERFACE {
		if setup.BRequest == usb_CDC_SET_LINE_CODING {
			b, err := machine.ReceiveUSBControlPacket()
			if err != nil {
				return false
			}

			usbLineInfo.dwDTERate = uint32(b[0]) | uint32(b[1])<<8 | uint32(b[2])<<16 | uint32(b[3])<<24
			usbLineInfo.bCharFormat = b[4]
			usbLineInfo.bParityType = b[5]
			usbLineInfo.bDataBits = b[6]
		}

		if setup.BRequest == usb_CDC_SET_CONTROL_LINE_STATE {
			usbLineInfo.lineState = setup.WValueL
		}

		if setup.BRequest == usb_CDC_SET_LINE_CODING || setup.BRequest == usb_CDC_SET_CONTROL_LINE_STATE {
			// auto-reset into the bootloader
			if usbLineInfo.dwDTERate == 1200 && usbLineInfo.lineState&usb_CDC_LINESTATE_DTR == 0 {
				machine.EnterBootloader()
			} else {
				// TODO: cancel any reset
			}
			machine.SendZlp()
		}

		if setup.BRequest == usb_CDC_SEND_BREAK {
			// TODO: something with this value?
			// breakValue = ((uint16_t)setup.wValueH << 8) | setup.wValueL;
			// return false;
			machine.SendZlp()
		}
		return true
	}
	return false
}

func EnableUSBCDC() {
	machine.USBCDC = New()
	machine.EnableCDC(USB.txhandler, cdcCallbackRx, cdcSetup)
}

//go:linkname gosched runtime.Gosched
func gosched()
