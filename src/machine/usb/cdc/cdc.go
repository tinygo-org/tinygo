package cdc

import (
	"machine"
)

const (
	cdcEndpointACM = 1
	cdcEndpointOut = 2
	cdcEndpointIn  = 3
)

var CDC *cdc

type cdc struct {
	buf            *RingBuffer
	callbackFuncRx func([]byte)
}

// New returns hid-mouse.
func New() *USBCDC {
	USB = &USBCDC{
		Buffer:  NewRingBuffer(),
		Buffer2: NewRingBuffer2(),
	}
	return USB
}

func newCDC() *cdc {
	m := &cdc{
		buf: NewRingBuffer(),
	}
	//machine.EnableCDC(m.Callback, m.CallbackRx)
	return m
}

func (c *cdc) SetCallback(callbackRx func([]byte)) {
	c.callbackFuncRx = callbackRx
}

func (c *cdc) Write(b []byte) (n int, err error) {
	i := 0
	for i = 0; i < len(b); i++ {
		c.buf.Put(b[i])
	}
	return i, nil
}

func (c *cdc) sendUSBPacket(b []byte) {
	machine.SendUSBInPacket(cdcEndpointIn, b)
}

// from BulkIn
func (c *cdc) Callback() {
	if b, ok := c.buf.Get(); ok {
		c.sendUSBPacket([]byte{b})
	}
}

// from BulkOut
func (c *cdc) CallbackRx(b []byte) {
	if c.callbackFuncRx != nil {
		c.callbackFuncRx(b)
	}
}

const (
	// bmRequestType
	usb_REQUEST_HOSTTODEVICE = 0x00
	usb_REQUEST_DEVICETOHOST = 0x80
	usb_REQUEST_DIRECTION    = 0x80

	usb_REQUEST_STANDARD = 0x00
	usb_REQUEST_CLASS    = 0x20
	usb_REQUEST_VENDOR   = 0x40
	usb_REQUEST_TYPE     = 0x60

	usb_REQUEST_DEVICE    = 0x00
	usb_REQUEST_INTERFACE = 0x01
	usb_REQUEST_ENDPOINT  = 0x02
	usb_REQUEST_OTHER     = 0x03
	usb_REQUEST_RECIPIENT = 0x1F

	usb_REQUEST_DEVICETOHOST_CLASS_INTERFACE    = (usb_REQUEST_DEVICETOHOST | usb_REQUEST_CLASS | usb_REQUEST_INTERFACE)
	usb_REQUEST_HOSTTODEVICE_CLASS_INTERFACE    = (usb_REQUEST_HOSTTODEVICE | usb_REQUEST_CLASS | usb_REQUEST_INTERFACE)
	usb_REQUEST_DEVICETOHOST_STANDARD_INTERFACE = (usb_REQUEST_DEVICETOHOST | usb_REQUEST_STANDARD | usb_REQUEST_INTERFACE)

	// CDC Class requests
	usb_CDC_SET_LINE_CODING        = 0x20
	usb_CDC_GET_LINE_CODING        = 0x21
	usb_CDC_SET_CONTROL_LINE_STATE = 0x22
	usb_CDC_SEND_BREAK             = 0x23

	usb_CDC_V1_10                         = 0x0110
	usb_CDC_COMMUNICATION_INTERFACE_CLASS = 0x02

	usb_CDC_CALL_MANAGEMENT             = 0x01
	usb_CDC_ABSTRACT_CONTROL_MODEL      = 0x02
	usb_CDC_HEADER                      = 0x00
	usb_CDC_ABSTRACT_CONTROL_MANAGEMENT = 0x02
	usb_CDC_UNION                       = 0x06
	usb_CDC_CS_INTERFACE                = 0x24
	usb_CDC_CS_ENDPOINT                 = 0x25
	usb_CDC_DATA_INTERFACE_CLASS        = 0x0A

	usb_CDC_LINESTATE_DTR = 0x01
	usb_CDC_LINESTATE_RTS = 0x02
)

func (c *cdc) handleSetup(setup machine.USBSetup) bool {
	if setup.BmRequestType == usb_REQUEST_DEVICETOHOST_CLASS_INTERFACE {
		if setup.BRequest == usb_CDC_GET_LINE_CODING {
			var b [cdcLineInfoSize]byte
			b[0] = byte(usbLineInfo.dwDTERate)
			b[1] = byte(usbLineInfo.dwDTERate >> 8)
			b[2] = byte(usbLineInfo.dwDTERate >> 16)
			b[3] = byte(usbLineInfo.dwDTERate >> 24)
			b[4] = byte(usbLineInfo.bCharFormat)
			b[5] = byte(usbLineInfo.bParityType)
			b[6] = byte(usbLineInfo.bDataBits)

			//c.sendUSBPacket(0, b[:], setup.WLength)
			c.sendUSBPacket(b[:])
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
				machine.ResetProcessor()
			} else {
				// TODO: cancel any reset
			}
			sendZlp()
		}

		if setup.BRequest == usb_CDC_SEND_BREAK {
			// TODO: something with this value?
			// breakValue = ((uint16_t)setup.WValueH << 8) | setup.WValueL;
			// return false;
			sendZlp()
		}
		return true
	}
	return false
}
