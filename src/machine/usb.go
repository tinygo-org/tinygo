//go:build sam || nrf52840 || rp2040 || rp2350

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

type usbEndpointEntry struct {
	Endpoint uint32
	Config   uint32
}

var (
	USBDev = &USBDevice{}
	USBCDC Serialer

	endPoints = []usbEndpointEntry{
		{
			Endpoint: usb.CONTROL_ENDPOINT,
			Config:   usb.ENDPOINT_TYPE_CONTROL,
		},
	}
)

func initUSB() {
	enableUSBCDC()
	USBDev.Configure(UARTConfig{})
}

// Using go:linkname here because there's a circular dependency between the
// machine package and the machine/usb/cdc package.
//
//go:linkname enableUSBCDC machine/usb/cdc.EnableUSBCDC
func enableUSBCDC()

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

const cdcLineInfoSize = 7

var (
	ErrUSBReadTimeout  = errors.New("USB read timeout")
	ErrUSBBytesRead    = errors.New("USB invalid number of bytes read")
	ErrUSBBytesWritten = errors.New("USB invalid number of bytes written")
)

var (
	usbEndpointDescriptors [NumberOfUSBEndpoints]descriptor.Device

	isEndpointHalt        = false
	isRemoteWakeUpEnabled = false

	usbConfiguration uint8
	usbSetInterface  uint8
)

//go:align 4
var udd_ep_control_cache_buffer [256]uint8

//go:align 4
var udd_ep_in_cache_buffer [NumberOfUSBEndpoints][64]uint8

//go:align 4
var udd_ep_out_cache_buffer [NumberOfUSBEndpoints][64]uint8

// usb_trans_buffer max size is 255 since that is max size
// for a descriptor (bLength is 1 byte), and the biggest use
// for this buffer is to transmit string descriptors.  If
// this buffer is used for new purposes in future the length
// must be revisited.
var usb_trans_buffer [255]uint8

var (
	usbTxHandler    [NumberOfUSBEndpoints]func()
	usbRxHandler    [NumberOfUSBEndpoints]func([]byte) bool
	usbSetupHandler [usb.NumberOfInterfaces]func(usb.Setup) bool
	usbStallHandler [NumberOfUSBEndpoints]func(usb.Setup) bool
)

var usbLangInfo = [4]byte{
	// length = 4 bytes
	0x04,
	// descriptor type = string
	0x03,
	// language codes
	// 0x0409 = English (United States)
	0x09, 0x04,
}

// sendDescriptor creates and sends the various USB descriptor types that
// can be requested by the host.
func sendDescriptor(setup usb.Setup) {
	switch setup.WValueH {
	case descriptor.TypeConfiguration:
		sendDescriptorData(usbDescriptor.Configuration, setup.WLength)
		return

	case descriptor.TypeDevice:
		usbDescriptor.Configure(usbVendorID(), usbProductID())
		sendDescriptorData(usbDescriptor.Device, setup.WLength)
		return

	case descriptor.TypeString:
		switch setup.WValueL {
		case 0:
			sendDescriptorData(usbLangInfo[:], setup.WLength)

		case usb.IPRODUCT:
			sendDescriptorString(usbProduct(), setup.WLength)

		case usb.IMANUFACTURER:
			sendDescriptorString(usbManufacturer(), setup.WLength)

		case usb.ISERIAL:
			serial := usbSerial()
			if len(serial) == 0 {
				SendZlp()
			} else {
				sendDescriptorString(serial, setup.WLength)
			}
		}
		// TODO: why do we do this when WValueL is unknown?
		return
	case descriptor.TypeHIDReport:
		if h, ok := usbDescriptor.HID[setup.WIndex]; ok {
			sendDescriptorData(h, setup.WLength)
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

// sendDescriptorString sends a string descriptor, truncating it to fit maxLen or the buffer size.
// note: the following code only converts ascii characters to UTF16LE. In order
// to do a "proper" conversion, we would need to pull in the 'unicode/utf16'
// package, which at the time this was written added 512 bytes to the compiled
// binary.
// TODO: old comment, re-evaluate
func sendDescriptorString(data string, maxLen uint16) {
	if maxLen < 2 {
		// Something has gone horribly wrong.
		SendZlp()
		return
	}

	// Clamp the length.
	maxEncBytes := min(len(usb_trans_buffer), len(udd_ep_control_cache_buffer), int(maxLen))
	// Write the header.
	buf := usb_trans_buffer[:min(2*len(data)+2, maxEncBytes)]
	hdr, body := buf[:2], buf[2:]

	// hdr[0] (bLength) should convey the "original total string length" before being limited by the host's maxLen.
	hdr[0] = byte(2*len(data) + 2)
	hdr[1] = descriptor.TypeString

	// Convert the string to UTF16.
	// NOTE: Using range here would cause the length to disagree when multibyte codepoints are present.
	limit := min(len(data), len(body)/2)
	for i := 0; i < limit; i++ {
		body[2*i] = byte(data[i])
		body[2*i+1] = 0
	}

	sendUSBPacket(0, buf)
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

		sendDescriptorData(usb_trans_buffer[:2], setup.WLength)
		return true

	case usb.CLEAR_FEATURE:
		if setup.WValueL == 1 { // DEVICEREMOTEWAKEUP
			isRemoteWakeUpEnabled = false
		} else if setup.WValueL == 0 { // ENDPOINTHALT
			if idx := setup.WIndex & 0x7F; idx < NumberOfUSBEndpoints && usbStallHandler[idx] != nil {
				// Host has requested to clear an endpoint stall. If the request is addressed to
				// an endpoint with a configured StallHandler, forward the message on.
				// The 0x7F mask is used to clear the direction bit from the endpoint number
				return usbStallHandler[idx](setup)
			}
			isEndpointHalt = false
		}
		SendZlp()
		return true

	case usb.SET_FEATURE:
		if setup.WValueL == 1 { // DEVICEREMOTEWAKEUP
			isRemoteWakeUpEnabled = true
		} else if setup.WValueL == 0 { // ENDPOINTHALT
			if idx := setup.WIndex & 0x7F; idx < NumberOfUSBEndpoints && usbStallHandler[idx] != nil {
				// Host has requested to stall an endpoint. If the request is addressed to
				// an endpoint with a configured StallHandler, forward the message on.
				// The 0x7F mask is used to clear the direction bit from the endpoint number
				return usbStallHandler[idx](setup)
			}
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
		sendDescriptorData(usb_trans_buffer[:1], setup.WLength)
		return true

	case usb.SET_CONFIGURATION:
		if setup.BmRequestType&usb.REQUEST_RECIPIENT == usb.REQUEST_DEVICE {
			for _, entry := range endPoints {
				if entry.Endpoint == usb.CONTROL_ENDPOINT {
					continue
				}
				initEndpoint(uint32(entry.Endpoint), entry.Config)
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
		sendDescriptorData(usb_trans_buffer[:1], setup.WLength)
		return true

	case usb.SET_INTERFACE:
		usbSetInterface = setup.WValueL

		SendZlp()
		return true

	default:
		return true
	}
}

// sendDescriptorData sends a descriptor, truncating it to fit maxLen or the buffer size.
func sendDescriptorData(data []byte, maxLen uint16) {
	data = lenToCap(data)
	data = data[:min(len(data), len(udd_ep_control_cache_buffer), int(maxLen))]
	sendUSBPacket(0, data)
}

// Set the cap of the slice to the length.
// This is safe, but cannot be proven by the compiler.
//
//go:nobounds
func lenToCap(b []byte) []byte {
	return b[:len(b):len(b)]
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
			endPoints = append(endPoints, usbEndpointEntry{
				Endpoint: uint32(ep.Index),
				Config:   uint32(ep.Type | usb.EndpointIn),
			})
			if ep.TxHandler != nil {
				usbTxHandler[ep.Index] = ep.TxHandler
			}
		} else {
			endPoints = append(endPoints, usbEndpointEntry{
				Endpoint: uint32(ep.Index),
				Config:   uint32(ep.Type | usb.EndpointOut),
			})
			if ep.RxHandler != nil {
				usbRxHandler[ep.Index] = func(b []byte) bool {
					ep.RxHandler(b)
					return true
				}
			} else if ep.DelayRxHandler != nil {
				usbRxHandler[ep.Index] = ep.DelayRxHandler
			}
		}
		if ep.StallHandler != nil {
			usbStallHandler[ep.Index] = ep.StallHandler
		}
	}

	for _, s := range setup {
		usbSetupHandler[s.Index] = s.Handler
	}
}
