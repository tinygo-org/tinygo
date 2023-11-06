//go:build sam || nrf52840 || rp2040

package machine

import (
	"machine/usb"
	"machine/usb/descriptor"

	"errors"
)

type USBDevice struct {
	initcomplete         bool
	InitEndpointComplete bool
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

var usbDescriptor descriptor.Descriptor

func usbVendorID() uint16 {
	if usb.VendorID != 0 {
		return usb.VendorID
	}

	return usb_VID
}

func usbProductID() uint16 {
	if usb.ProductID != 0 {
		return usb.ProductID
	}

	return usb_PID
}

func usbManufacturer() string {
	if usb.Manufacturer != "" {
		return usb.Manufacturer
	}

	return usb_STRING_MANUFACTURER
}

func usbProduct() string {
	if usb.Product != "" {
		return usb.Product
	}

	return usb_STRING_PRODUCT
}

func usbSerial() string {
	if usb.Serial != "" {
		return usb.Serial
	}
	return ""
}

// strToUTF16LEDescriptor converts a utf8 string into a string descriptor
// note: the following code only converts ascii characters to UTF16LE. In order
// to do a "proper" conversion, we would need to pull in the 'unicode/utf16'
// package, which at the time this was written added 512 bytes to the compiled
// binary.
func strToUTF16LEDescriptor(in string, out []byte) {
	out[0] = byte(len(out))
	out[1] = descriptor.TypeString
	for i, rune := range in {
		out[(i<<1)+2] = byte(rune)
		out[(i<<1)+3] = 0
	}
	return
}

const cdcLineInfoSize = 7

var (
	ErrUSBReadTimeout = errors.New("USB read timeout")
	ErrUSBBytesRead   = errors.New("USB invalid number of bytes read")
)

var (
	usbEndpointDescriptors [usb.NumberOfEndpoints]descriptor.Device

	isEndpointHalt        = false
	isRemoteWakeUpEnabled = false

	usbConfiguration uint8
	usbSetInterface  uint8
)

//go:align 4
var udd_ep_control_cache_buffer [256]uint8

//go:align 4
var udd_ep_in_cache_buffer [usb.NumberOfEndpoints][64]uint8

//go:align 4
var udd_ep_out_cache_buffer [usb.NumberOfEndpoints][64]uint8

// usb_trans_buffer max size is 255 since that is max size
// for a descriptor (bLength is 1 byte), and the biggest use
// for this buffer is to transmit string descriptors.  If
// this buffer is used for new purposes in future the length
// must be revisited.
var usb_trans_buffer [255]uint8

var (
	usbTxHandler    [usb.NumberOfEndpoints]func()
	usbRxHandler    [usb.NumberOfEndpoints]func([]byte)
	usbSetupHandler [usb.NumberOfInterfaces]func(usb.Setup) bool

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

// sendDescriptor creates and sends the various USB descriptor types that
// can be requested by the host.
func sendDescriptor(setup usb.Setup) {
	switch setup.WValueH {
	case descriptor.TypeConfiguration:
		sendUSBPacket(0, usbDescriptor.Configuration, setup.WLength)
		return
	case descriptor.TypeDevice:
		usbDescriptor.Configure(usbVendorID(), usbProductID())
		sendUSBPacket(0, usbDescriptor.Device, setup.WLength)
		return

	case descriptor.TypeString:
		switch setup.WValueL {
		case 0:
			usb_trans_buffer[0] = 0x04
			usb_trans_buffer[1] = 0x03
			usb_trans_buffer[2] = 0x09
			usb_trans_buffer[3] = 0x04
			sendUSBPacket(0, usb_trans_buffer[:4], setup.WLength)

		case usb.IPRODUCT:
			b := usb_trans_buffer[:(len(usbProduct())<<1)+2]
			strToUTF16LEDescriptor(usbProduct(), b)
			sendUSBPacket(0, b, setup.WLength)

		case usb.IMANUFACTURER:
			b := usb_trans_buffer[:(len(usbManufacturer())<<1)+2]
			strToUTF16LEDescriptor(usbManufacturer(), b)
			sendUSBPacket(0, b, setup.WLength)

		case usb.ISERIAL:
			sz := len(usbSerial())
			if sz == 0 {
				SendZlp()
			} else {
				b := usb_trans_buffer[:(sz<<1)+2]
				strToUTF16LEDescriptor(usbSerial(), b)
				sendUSBPacket(0, b, setup.WLength)
			}
		}
		return
	case descriptor.TypeHIDReport:
		if h, ok := usbDescriptor.HID[setup.WIndex]; ok {
			sendUSBPacket(0, h, setup.WLength)
			return
		}
	case descriptor.TypeDeviceQualifier:
		// skip
	default:
	}

	// do not know how to handle this message, so return zero
	SendZlp()
	return
}

func handleStandardSetup(setup usb.Setup) bool {
	switch setup.BRequest {
	case usb.GET_STATUS:
		usb_trans_buffer[0] = 0
		usb_trans_buffer[1] = 0

		if setup.BmRequestType != 0 { // endpoint
			if isEndpointHalt {
				usb_trans_buffer[0] = 1
			}
		}

		sendUSBPacket(0, usb_trans_buffer[:2], setup.WLength)
		return true

	case usb.CLEAR_FEATURE:
		if setup.WValueL == 1 { // DEVICEREMOTEWAKEUP
			isRemoteWakeUpEnabled = false
		} else if setup.WValueL == 0 { // ENDPOINTHALT
			isEndpointHalt = false
		}
		SendZlp()
		return true

	case usb.SET_FEATURE:
		if setup.WValueL == 1 { // DEVICEREMOTEWAKEUP
			isRemoteWakeUpEnabled = true
		} else if setup.WValueL == 0 { // ENDPOINTHALT
			isEndpointHalt = true
		}
		SendZlp()
		return true

	case usb.SET_ADDRESS:
		return handleUSBSetAddress(setup)

	case usb.GET_DESCRIPTOR:
		sendDescriptor(setup)
		return true

	case usb.SET_DESCRIPTOR:
		return false

	case usb.GET_CONFIGURATION:
		usb_trans_buffer[0] = usbConfiguration
		sendUSBPacket(0, usb_trans_buffer[:1], setup.WLength)
		return true

	case usb.SET_CONFIGURATION:
		if setup.BmRequestType&usb.REQUEST_RECIPIENT == usb.REQUEST_DEVICE {
			for i := 1; i < len(endPoints); i++ {
				initEndpoint(uint32(i), endPoints[i])
			}

			usbConfiguration = setup.WValueL
			USBDev.InitEndpointComplete = true

			SendZlp()
			return true
		} else {
			return false
		}

	case usb.GET_INTERFACE:
		usb_trans_buffer[0] = usbSetInterface
		sendUSBPacket(0, usb_trans_buffer[:1], setup.WLength)
		return true

	case usb.SET_INTERFACE:
		usbSetInterface = setup.WValueL

		SendZlp()
		return true

	default:
		return true
	}
}

func EnableCDC(txHandler func(), rxHandler func([]byte), setupHandler func(usb.Setup) bool) {
	if len(usbDescriptor.Device) == 0 {
		usbDescriptor = descriptor.CDC
	}
	// Initialization of endpoints is required even for non-CDC
	ConfigureUSBEndpoint(usbDescriptor,
		[]usb.EndpointConfig{
			{
				Index: usb.CDC_ENDPOINT_ACM,
				IsIn:  true,
				Type:  usb.ENDPOINT_TYPE_INTERRUPT,
			},
			{
				Index:     usb.CDC_ENDPOINT_OUT,
				IsIn:      false,
				Type:      usb.ENDPOINT_TYPE_BULK,
				RxHandler: rxHandler,
			},
			{
				Index:     usb.CDC_ENDPOINT_IN,
				IsIn:      true,
				Type:      usb.ENDPOINT_TYPE_BULK,
				TxHandler: txHandler,
			},
		},
		[]usb.SetupConfig{
			{
				Index:   usb.CDC_ACM_INTERFACE,
				Handler: setupHandler,
			},
		})
}

func ConfigureUSBEndpoint(desc descriptor.Descriptor, epSettings []usb.EndpointConfig, setup []usb.SetupConfig) {
	usbDescriptor = desc

	for _, ep := range epSettings {
		if ep.IsIn {
			endPoints[ep.Index] = uint32(ep.Type | usb.EndpointIn)
			if ep.TxHandler != nil {
				usbTxHandler[ep.Index] = ep.TxHandler
			}
		} else {
			endPoints[ep.Index] = uint32(ep.Type | usb.EndpointOut)
			if ep.RxHandler != nil {
				usbRxHandler[ep.Index] = ep.RxHandler
			}
		}
	}

	for _, s := range setup {
		usbSetupHandler[s.Index] = s.Handler
	}
}
