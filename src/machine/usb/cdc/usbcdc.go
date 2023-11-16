package cdc

import (
	"errors"
	"machine"
	"machine/usb"
	"runtime/interrupt"
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

// Read from the RX buffer.
func (usbcdc *USBCDC) Read(data []byte) (n int, err error) {
	// If RX buffer is empty, block until data is available.
	size := usbcdc.Buffered()
	if size == 0 {
		<-usbcdc.RXChan
		size = usbcdc.Buffered()
	}

	// Make sure we do not read more from buffer than the data slice can hold.
	if len(data) < size {
		size = len(data)
	}

	// only read number of bytes used from buffer
	for i := 0; i < size; i++ {
		v, _ := usbcdc.ReadByte()
		data[i] = v
	}

	return size, nil
}

// ReadByte reads a single byte from the RX buffer.
// If there is no data in the buffer, returns an error.
func (usbcdc *USBCDC) ReadByte() (byte, error) {
	// check if RX buffer is empty
	buf, ok := usbcdc.rxBuffer.Get()
	if !ok {
		return 0, ErrBufferEmpty
	}
	return buf, nil
}

// Buffered returns the number of bytes currently stored in the RX buffer.
func (usbcdc *USBCDC) Buffered() int {
	return int(usbcdc.rxBuffer.Used())
}

// Receive handles adding data to the UART's data buffer.
// Usually called by the IRQ handler for a machine.
func (usbcdc *USBCDC) Receive(data byte) {
	usbcdc.rxBuffer.Put(data)

	select {
	case USB.RXChan <- true:
	default:
	}
}

// USBCDC is the USB CDC aka serial over USB interface.
type USBCDC struct {
	rxBuffer *rxRingBuffer
	RXChan   chan bool
	txBuffer *txRingBuffer
	waitTxc  bool
}

var (
	// USB is a USB CDC interface.
	USB *USBCDC

	usbLineInfo = cdcLineInfo{115200, 0x00, 0x00, 0x08, 0x00}
)

// Configure the USB CDC interface. The config is here for compatibility with the UART interface.
func (usbcdc *USBCDC) Configure(config machine.UARTConfig) error {
	return nil
}

// Flush flushes buffered data.
func (usbcdc *USBCDC) Flush() {
	mask := interrupt.Disable()
	if b, ok := usbcdc.txBuffer.Get(); ok {
		machine.SendUSBInPacket(cdcEndpointIn, b)
	} else {
		usbcdc.waitTxc = false
	}
	interrupt.Restore(mask)
}

// Write data to the USBCDC.
func (usbcdc *USBCDC) Write(data []byte) (n int, err error) {
	if usbLineInfo.lineState > 0 {
		mask := interrupt.Disable()
		usbcdc.txBuffer.Put(data)
		if !usbcdc.waitTxc {
			usbcdc.waitTxc = true
			usbcdc.Flush()
		}
		interrupt.Restore(mask)
	}
	return len(data), nil
}

// WriteByte writes a byte of data to the USB CDC interface.
func (usbcdc *USBCDC) WriteByte(c byte) error {
	usbcdc.Write([]byte{c})
	return nil
}

func (usbcdc *USBCDC) DTR() bool {
	return (usbLineInfo.lineState & usb_CDC_LINESTATE_DTR) > 0
}

func (usbcdc *USBCDC) RTS() bool {
	return (usbLineInfo.lineState & usb_CDC_LINESTATE_RTS) > 0
}

func cdcCallbackRx(b []byte) {
	for i := range b {
		USB.Receive(b[i])
	}
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
	machine.EnableCDC(USB.Flush, cdcCallbackRx, cdcSetup)
}
