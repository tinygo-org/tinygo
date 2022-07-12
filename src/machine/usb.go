//go:build sam || nrf52840 || rp2040
// +build sam nrf52840 rp2040

package machine

import (
	"errors"
	"runtime/volatile"
)

type USBDevice struct {
	initcomplete bool
}

var (
	USBDev = &USBDevice{}
	USBCDC Serialer
)

type Serialer interface {
	WriteByte(c byte) error
	Write(data []byte) (n int, err error)
	Configure(config UARTConfig) error
	Buffered() int
	ReadByte() (byte, error)
	DTR() bool
	RTS() bool
}

var usbDescriptor = descriptorCDC

const (
	usbDescriptorConfigCDC = 1 << iota
	usbDescriptorConfigHID
)

var usbDescriptorConfig uint8 = usbDescriptorConfigCDC

// strToUTF16LEDescriptor converts a utf8 string into a string descriptor
// note: the following code only converts ascii characters to UTF16LE. In order
// to do a "proper" conversion, we would need to pull in the 'unicode/utf16'
// package, which at the time this was written added 512 bytes to the compiled
// binary.
func strToUTF16LEDescriptor(in string, out []byte) {
	out[0] = byte(len(out))
	out[1] = usb_STRING_DESCRIPTOR_TYPE
	for i, rune := range in {
		out[(i<<1)+2] = byte(rune)
		out[(i<<1)+3] = 0
	}
	return
}

var (
	// TODO: allow setting these
	usb_STRING_LANGUAGE = [2]uint16{(3 << 8) | (2 + 2), 0x0409} // English
)

const cdcLineInfoSize = 7

var (
	ErrUSBReadTimeout = errors.New("USB read timeout")
	ErrUSBBytesRead   = errors.New("USB invalid number of bytes read")
)

var (
	usbEndpointDescriptors [numberOfEndpoints]usbDeviceDescriptor

	udd_ep_control_cache_buffer [256]uint8
	udd_ep_in_cache_buffer      [7][64]uint8
	udd_ep_out_cache_buffer     [7][64]uint8

	isEndpointHalt        = false
	isRemoteWakeUpEnabled = false

	usbConfiguration uint8
	usbSetInterface  uint8
)

const (
	usb_IMANUFACTURER = 1
	usb_IPRODUCT      = 2
	usb_ISERIAL       = 3

	usb_ENDPOINT_TYPE_DISABLE     = 0xFF
	usb_ENDPOINT_TYPE_CONTROL     = 0x00
	usb_ENDPOINT_TYPE_ISOCHRONOUS = 0x01
	usb_ENDPOINT_TYPE_BULK        = 0x02
	usb_ENDPOINT_TYPE_INTERRUPT   = 0x03

	usb_DEVICE_DESCRIPTOR_TYPE        = 1
	usb_CONFIGURATION_DESCRIPTOR_TYPE = 2
	usb_STRING_DESCRIPTOR_TYPE        = 3
	usb_INTERFACE_DESCRIPTOR_TYPE     = 4
	usb_ENDPOINT_DESCRIPTOR_TYPE      = 5
	usb_DEVICE_QUALIFIER              = 6
	usb_OTHER_SPEED_CONFIGURATION     = 7
	usb_SET_REPORT_TYPE               = 33
	usb_HID_REPORT_TYPE               = 34

	usbEndpointOut = 0x00
	usbEndpointIn  = 0x80

	numberOfEndpoints     = 8
	usbEndpointPacketSize = 64 // 64 for Full Speed, EPT size max is 1024

	// standard requests
	usb_GET_STATUS        = 0
	usb_CLEAR_FEATURE     = 1
	usb_SET_FEATURE       = 3
	usb_SET_ADDRESS       = 5
	usb_GET_DESCRIPTOR    = 6
	usb_SET_DESCRIPTOR    = 7
	usb_GET_CONFIGURATION = 8
	usb_SET_CONFIGURATION = 9
	usb_GET_INTERFACE     = 10
	usb_SET_INTERFACE     = 11

	// non standard requests
	usb_SET_IDLE = 10

	usb_DEVICE_CLASS_COMMUNICATIONS  = 0x02
	usb_DEVICE_CLASS_HUMAN_INTERFACE = 0x03
	usb_DEVICE_CLASS_STORAGE         = 0x08
	usb_DEVICE_CLASS_VENDOR_SPECIFIC = 0xFF

	usb_CONFIG_POWERED_MASK  = 0x40
	usb_CONFIG_BUS_POWERED   = 0x80
	usb_CONFIG_SELF_POWERED  = 0xC0
	usb_CONFIG_REMOTE_WAKEUP = 0x20

	// Interface
	numberOfInterfaces     = 3
	usb_CDC_ACM_INTERFACE  = 0 // CDC ACM
	usb_CDC_DATA_INTERFACE = 1 // CDC Data
	usb_CDC_FIRST_ENDPOINT = 1
	usb_HID_INTERFACE      = 2 // HID

	// Endpoint
	usb_CONTROL_ENDPOINT = 0
	usb_CDC_ENDPOINT_ACM = 1
	usb_CDC_ENDPOINT_OUT = 2
	usb_CDC_ENDPOINT_IN  = 3
	usb_HID_ENDPOINT_IN  = 4

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
)

var (
	usbTxHandler    [numberOfEndpoints]func()
	usbRxHandler    [numberOfEndpoints]func([]byte)
	usbSetupHandler [numberOfInterfaces]func(USBSetup) bool

	endPoints = []uint32{
		usb_CONTROL_ENDPOINT: usb_ENDPOINT_TYPE_CONTROL,
		usb_CDC_ENDPOINT_ACM: (usb_ENDPOINT_TYPE_INTERRUPT | usbEndpointIn),
		usb_CDC_ENDPOINT_OUT: (usb_ENDPOINT_TYPE_BULK | usbEndpointOut),
		usb_CDC_ENDPOINT_IN:  (usb_ENDPOINT_TYPE_BULK | usbEndpointIn),
		usb_HID_ENDPOINT_IN:  (usb_ENDPOINT_TYPE_DISABLE), // Interrupt In
	}
)

// usbDeviceDescBank is the USB device endpoint descriptor.
type usbDeviceDescBank struct {
	ADDR      volatile.Register32
	PCKSIZE   volatile.Register32
	EXTREG    volatile.Register16
	STATUS_BK volatile.Register8
	_reserved [5]volatile.Register8
}

type usbDeviceDescriptor struct {
	DeviceDescBank [2]usbDeviceDescBank
}

type USBSetup struct {
	BmRequestType uint8
	BRequest      uint8
	WValueL       uint8
	WValueH       uint8
	WIndex        uint16
	WLength       uint16
}

func newUSBSetup(data []byte) USBSetup {
	u := USBSetup{}
	u.BmRequestType = uint8(data[0])
	u.BRequest = uint8(data[1])
	u.WValueL = uint8(data[2])
	u.WValueH = uint8(data[3])
	u.WIndex = uint16(data[4]) | (uint16(data[5]) << 8)
	u.WLength = uint16(data[6]) | (uint16(data[7]) << 8)
	return u
}

// sendDescriptor creates and sends the various USB descriptor types that
// can be requested by the host.
func sendDescriptor(setup USBSetup) {
	switch setup.WValueH {
	case usb_CONFIGURATION_DESCRIPTOR_TYPE:
		sendUSBPacket(0, usbDescriptor.Configuration, setup.WLength)
		return
	case usb_DEVICE_DESCRIPTOR_TYPE:
		// composite descriptor
		if (usbDescriptorConfig & usbDescriptorConfigHID) > 0 {
			usbDescriptor = descriptorCDCHID
		} else {
			usbDescriptor = descriptorCDC
		}
		usbDescriptor.Configure(usb_VID, usb_PID)
		sendUSBPacket(0, usbDescriptor.Device, setup.WLength)
		return

	case usb_STRING_DESCRIPTOR_TYPE:
		switch setup.WValueL {
		case 0:
			b := []byte{0x04, 0x03, 0x09, 0x04}
			sendUSBPacket(0, b, setup.WLength)

		case usb_IPRODUCT:
			b := make([]byte, (len(usb_STRING_PRODUCT)<<1)+2)
			strToUTF16LEDescriptor(usb_STRING_PRODUCT, b)
			sendUSBPacket(0, b, setup.WLength)

		case usb_IMANUFACTURER:
			b := make([]byte, (len(usb_STRING_MANUFACTURER)<<1)+2)
			strToUTF16LEDescriptor(usb_STRING_MANUFACTURER, b)
			sendUSBPacket(0, b, setup.WLength)

		case usb_ISERIAL:
			// TODO: allow returning a product serial number
			SendZlp()
		}
		return
	case usb_HID_REPORT_TYPE:
		if h, ok := usbDescriptor.HID[setup.WIndex]; ok {
			sendUSBPacket(0, h, setup.WLength)
			return
		}
	case usb_DEVICE_QUALIFIER:
		// skip
	default:
	}

	// do not know how to handle this message, so return zero
	SendZlp()
	return
}

func handleStandardSetup(setup USBSetup) bool {
	switch setup.BRequest {
	case usb_GET_STATUS:
		buf := []byte{0, 0}

		if setup.BmRequestType != 0 { // endpoint
			if isEndpointHalt {
				buf[0] = 1
			}
		}

		sendUSBPacket(0, buf, setup.WLength)
		return true

	case usb_CLEAR_FEATURE:
		if setup.WValueL == 1 { // DEVICEREMOTEWAKEUP
			isRemoteWakeUpEnabled = false
		} else if setup.WValueL == 0 { // ENDPOINTHALT
			isEndpointHalt = false
		}
		SendZlp()
		return true

	case usb_SET_FEATURE:
		if setup.WValueL == 1 { // DEVICEREMOTEWAKEUP
			isRemoteWakeUpEnabled = true
		} else if setup.WValueL == 0 { // ENDPOINTHALT
			isEndpointHalt = true
		}
		SendZlp()
		return true

	case usb_SET_ADDRESS:
		return handleUSBSetAddress(setup)

	case usb_GET_DESCRIPTOR:
		sendDescriptor(setup)
		return true

	case usb_SET_DESCRIPTOR:
		return false

	case usb_GET_CONFIGURATION:
		buff := []byte{usbConfiguration}
		sendUSBPacket(0, buff, setup.WLength)
		return true

	case usb_SET_CONFIGURATION:
		if setup.BmRequestType&usb_REQUEST_RECIPIENT == usb_REQUEST_DEVICE {
			for i := 1; i < len(endPoints); i++ {
				initEndpoint(uint32(i), endPoints[i])
			}

			usbConfiguration = setup.WValueL

			SendZlp()
			return true
		} else {
			return false
		}

	case usb_GET_INTERFACE:
		buff := []byte{usbSetInterface}
		sendUSBPacket(0, buff, setup.WLength)
		return true

	case usb_SET_INTERFACE:
		usbSetInterface = setup.WValueL

		SendZlp()
		return true

	default:
		return true
	}
}

func EnableCDC(txHandler func(), rxHandler func([]byte), setupHandler func(USBSetup) bool) {
	usbDescriptorConfig |= usbDescriptorConfigCDC
	endPoints[usb_CDC_ENDPOINT_ACM] = (usb_ENDPOINT_TYPE_INTERRUPT | usbEndpointIn)
	endPoints[usb_CDC_ENDPOINT_OUT] = (usb_ENDPOINT_TYPE_BULK | usbEndpointOut)
	endPoints[usb_CDC_ENDPOINT_IN] = (usb_ENDPOINT_TYPE_BULK | usbEndpointIn)
	usbRxHandler[usb_CDC_ENDPOINT_OUT] = rxHandler
	usbTxHandler[usb_CDC_ENDPOINT_IN] = txHandler
	usbSetupHandler[usb_CDC_ACM_INTERFACE] = setupHandler // 0x02 (Communications and CDC Control)
	usbSetupHandler[usb_CDC_DATA_INTERFACE] = nil         // 0x0A (CDC-Data)
}

// EnableHID enables HID. This function must be executed from the init().
func EnableHID(txHandler func(), rxHandler func([]byte), setupHandler func(USBSetup) bool) {
	usbDescriptorConfig |= usbDescriptorConfigHID
	endPoints[usb_HID_ENDPOINT_IN] = (usb_ENDPOINT_TYPE_INTERRUPT | usbEndpointIn)
	usbTxHandler[usb_HID_ENDPOINT_IN] = txHandler
	usbSetupHandler[usb_HID_INTERFACE] = setupHandler // 0x03 (HID - Human Interface Device)
}
