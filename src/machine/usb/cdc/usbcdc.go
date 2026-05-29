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
	tx ring512
	rx ring512

	// inflight is the number of bytes currently submitted to the USB IN endpoint.
	inflight atomic.Uint32

	// txActive serializes the USB CDC TX pump between Write and the TX
	// completion handler. This matters on multicore targets where they can run
	// concurrently.
	txActive atomic.Uint32

	rbuf [1]byte
	wbuf [1]byte
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

// Flush flushes buffered data.
func (usbcdc *USBCDC) Flush() {
	for usbcdc.tx.Used() > 0 || usbcdc.txActive.Load() != 0 {
		gosched()
	}
}

// Write data to the USBCDC.
func (usbcdc *USBCDC) Write(data []byte) (n int, err error) {
	n = len(data)
	if usbLineInfo.lineState <= 0 {
		return n, nil
	}
	for len(data) > 0 {
		tosend := min(len(data), int(usbcdc.tx.Free()))
		if tosend == 0 {
			gosched()
			continue
		}
		usbcdc.tx.Put(data[:tosend])
		data = data[tosend:]
		usbcdc.kickTx()
	}
	return n, nil
}

// kickTx starts the TX pump if it is idle.
func (usbcdc *USBCDC) kickTx() {
	if !usbcdc.txActive.CompareAndSwap(0, 1) {
		return
	}
	usbcdc.sendFromRing()
}

func (usbcdc *USBCDC) txhandler() {
	inflight := usbcdc.inflight.Load()
	if inflight == 0 {
		return
	}
	usbcdc.tx.Discard(inflight)
	usbcdc.inflight.Store(0)
	usbcdc.sendFromRing()
}

// sendFromRing sends one USB packet from the ring and sets inflight.
//
// The caller must own txActive. Ownership starts in kickTx and is kept across
// TX completion handling until the TX ring is empty.
func (usbcdc *USBCDC) sendFromRing() {
	// This is the main TX pump loop. While txActive is held, each iteration
	// peeks the TX ring and sends one packet if data is available.
	//
	// The empty-ring case below is the shutdown path: release txActive, then
	// re-check the ring to avoid missing a Write that appended data while
	// txActive was still set.
	for {
		d1, _ := usbcdc.tx.Peek()
		if len(d1) == 0 {
			usbcdc.txActive.Store(0)

			// Avoid a missed wakeup: Write may append data while txActive is
			// still set, causing kickTx to return without starting a transfer.
			switch {
			case usbcdc.tx.Used() == 0:
				return
			case !usbcdc.txActive.CompareAndSwap(0, 1):
				return
			}

			// New data appeared while txActive was set, and this TX pump owner
			// re-acquired ownership. Continue the pump and re-peek the ring.
			continue
		}

		chunk := d1[:min(usb.EndpointPacketSize, len(d1))]
		usbcdc.inflight.Store(uint32(len(chunk)))
		machine.SendUSBInPacket(cdcEndpointIn, chunk)
		return
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
