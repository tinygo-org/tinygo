//go:build nrf52840
// +build nrf52840

package machine

import (
	"errors"
	"runtime/volatile"
)

const deviceDescriptorSize = 18

var (
	errUSBCDCBufferEmpty      = errors.New("USB-CDC buffer empty")
	errUSBCDCWriteByteTimeout = errors.New("USB-CDC write byte timeout")
	errUSBCDCReadTimeout      = errors.New("USB-CDC read timeout")
	errUSBCDCBytesRead        = errors.New("USB-CDC invalid number of bytes read")
)

// DeviceDescriptor implements the USB standard device descriptor.
//
// Table 9-8. Standard Device Descriptor
// bLength, bDescriptorType, bcdUSB, bDeviceClass, bDeviceSubClass, bDeviceProtocol, bMaxPacketSize0,
//    idVendor, idProduct, bcdDevice, iManufacturer, iProduct, iSerialNumber, bNumConfigurations */
//
type DeviceDescriptor struct {
	bLength            uint8  // 18
	bDescriptorType    uint8  // 1 USB_DEVICE_DESCRIPTOR_TYPE
	bcdUSB             uint16 // 0x200
	bDeviceClass       uint8
	bDeviceSubClass    uint8
	bDeviceProtocol    uint8
	bMaxPacketSize0    uint8 // Packet 0
	idVendor           uint16
	idProduct          uint16
	bcdDevice          uint16 // 0x100
	iManufacturer      uint8
	iProduct           uint8
	iSerialNumber      uint8
	bNumConfigurations uint8
}

// NewDeviceDescriptor returns a USB DeviceDescriptor.
func NewDeviceDescriptor(class, subClass, proto, packetSize0 uint8, vid, pid, version uint16, im, ip, is, configs uint8) DeviceDescriptor {
	return DeviceDescriptor{deviceDescriptorSize, 1, 0x200, class, subClass, proto, packetSize0, vid, pid, version, im, ip, is, configs}
}

// Bytes returns DeviceDescriptor data
func (d DeviceDescriptor) Bytes() [deviceDescriptorSize]byte {
	var b [deviceDescriptorSize]byte
	b[0] = byte(d.bLength)
	b[1] = byte(d.bDescriptorType)
	b[2] = byte(d.bcdUSB)
	b[3] = byte(d.bcdUSB >> 8)
	b[4] = byte(d.bDeviceClass)
	b[5] = byte(d.bDeviceSubClass)
	b[6] = byte(d.bDeviceProtocol)
	b[7] = byte(d.bMaxPacketSize0)
	b[8] = byte(d.idVendor)
	b[9] = byte(d.idVendor >> 8)
	b[10] = byte(d.idProduct)
	b[11] = byte(d.idProduct >> 8)
	b[12] = byte(d.bcdDevice)
	b[13] = byte(d.bcdDevice >> 8)
	b[14] = byte(d.iManufacturer)
	b[15] = byte(d.iProduct)
	b[16] = byte(d.iSerialNumber)
	b[17] = byte(d.bNumConfigurations)
	return b
}

const configDescriptorSize = 9

// ConfigDescriptor implements the standard USB configuration descriptor.
//
// Table 9-10. Standard Configuration Descriptor
// bLength, bDescriptorType, wTotalLength, bNumInterfaces, bConfigurationValue, iConfiguration
// bmAttributes, bMaxPower
//
type ConfigDescriptor struct {
	bLength             uint8  // 9
	bDescriptorType     uint8  // 2
	wTotalLength        uint16 // total length
	bNumInterfaces      uint8
	bConfigurationValue uint8
	iConfiguration      uint8
	bmAttributes        uint8
	bMaxPower           uint8
}

// NewConfigDescriptor returns a new USB ConfigDescriptor.
func NewConfigDescriptor(totalLength uint16, interfaces uint8) ConfigDescriptor {
	return ConfigDescriptor{configDescriptorSize, 2, totalLength, interfaces, 1, 0, usb_CONFIG_BUS_POWERED | usb_CONFIG_REMOTE_WAKEUP, 50}
}

// Bytes returns ConfigDescriptor data.
func (d ConfigDescriptor) Bytes() [configDescriptorSize]byte {
	var b [configDescriptorSize]byte
	b[0] = byte(d.bLength)
	b[1] = byte(d.bDescriptorType)
	b[2] = byte(d.wTotalLength)
	b[3] = byte(d.wTotalLength >> 8)
	b[4] = byte(d.bNumInterfaces)
	b[5] = byte(d.bConfigurationValue)
	b[6] = byte(d.iConfiguration)
	b[7] = byte(d.bmAttributes)
	b[8] = byte(d.bMaxPower)
	return b
}

const interfaceDescriptorSize = 9

// InterfaceDescriptor implements the standard USB interface descriptor.
//
// Table 9-12. Standard Interface Descriptor
// bLength, bDescriptorType, bInterfaceNumber, bAlternateSetting, bNumEndpoints, bInterfaceClass,
// bInterfaceSubClass, bInterfaceProtocol, iInterface
//
type InterfaceDescriptor struct {
	bLength            uint8 // 9
	bDescriptorType    uint8 // 4
	bInterfaceNumber   uint8
	bAlternateSetting  uint8
	bNumEndpoints      uint8
	bInterfaceClass    uint8
	bInterfaceSubClass uint8
	bInterfaceProtocol uint8
	iInterface         uint8
}

// NewInterfaceDescriptor returns a new USB InterfaceDescriptor.
func NewInterfaceDescriptor(n, numEndpoints, class, subClass, protocol uint8) InterfaceDescriptor {
	return InterfaceDescriptor{interfaceDescriptorSize, 4, n, 0, numEndpoints, class, subClass, protocol, 0}
}

// Bytes returns InterfaceDescriptor data.
func (d InterfaceDescriptor) Bytes() [interfaceDescriptorSize]byte {
	var b [interfaceDescriptorSize]byte
	b[0] = byte(d.bLength)
	b[1] = byte(d.bDescriptorType)
	b[2] = byte(d.bInterfaceNumber)
	b[3] = byte(d.bAlternateSetting)
	b[4] = byte(d.bNumEndpoints)
	b[5] = byte(d.bInterfaceClass)
	b[6] = byte(d.bInterfaceSubClass)
	b[7] = byte(d.bInterfaceProtocol)
	b[8] = byte(d.iInterface)
	return b
}

const endpointDescriptorSize = 7

// EndpointDescriptor implements the standard USB endpoint descriptor.
//
// Table 9-13. Standard Endpoint Descriptor
// bLength, bDescriptorType, bEndpointAddress, bmAttributes, wMaxPacketSize, bInterval
//
type EndpointDescriptor struct {
	bLength          uint8 // 7
	bDescriptorType  uint8 // 5
	bEndpointAddress uint8
	bmAttributes     uint8
	wMaxPacketSize   uint16
	bInterval        uint8
}

// NewEndpointDescriptor returns a new USB EndpointDescriptor.
func NewEndpointDescriptor(addr, attr uint8, packetSize uint16, interval uint8) EndpointDescriptor {
	return EndpointDescriptor{endpointDescriptorSize, 5, addr, attr, packetSize, interval}
}

// Bytes returns EndpointDescriptor data.
func (d EndpointDescriptor) Bytes() [endpointDescriptorSize]byte {
	var b [endpointDescriptorSize]byte
	b[0] = byte(d.bLength)
	b[1] = byte(d.bDescriptorType)
	b[2] = byte(d.bEndpointAddress)
	b[3] = byte(d.bmAttributes)
	b[4] = byte(d.wMaxPacketSize)
	b[5] = byte(d.wMaxPacketSize >> 8)
	b[6] = byte(d.bInterval)
	return b
}

const iadDescriptorSize = 8

// IADDescriptor is an Interface Association Descriptor, which is used
// to bind 2 interfaces together in CDC composite device.
//
// Standard Interface Association Descriptor:
// bLength, bDescriptorType, bFirstInterface, bInterfaceCount, bFunctionClass, bFunctionSubClass,
// bFunctionProtocol, iFunction
//
type IADDescriptor struct {
	bLength           uint8 // 8
	bDescriptorType   uint8 // 11
	bFirstInterface   uint8
	bInterfaceCount   uint8
	bFunctionClass    uint8
	bFunctionSubClass uint8
	bFunctionProtocol uint8
	iFunction         uint8
}

// NewIADDescriptor returns a new USB IADDescriptor.
func NewIADDescriptor(firstInterface, count, class, subClass, protocol uint8) IADDescriptor {
	return IADDescriptor{iadDescriptorSize, 11, firstInterface, count, class, subClass, protocol, 0}
}

// Bytes returns IADDescriptor data.
func (d IADDescriptor) Bytes() [iadDescriptorSize]byte {
	var b [iadDescriptorSize]byte
	b[0] = byte(d.bLength)
	b[1] = byte(d.bDescriptorType)
	b[2] = byte(d.bFirstInterface)
	b[3] = byte(d.bInterfaceCount)
	b[4] = byte(d.bFunctionClass)
	b[5] = byte(d.bFunctionSubClass)
	b[6] = byte(d.bFunctionProtocol)
	b[7] = byte(d.iFunction)
	return b
}

const cdcCSInterfaceDescriptorSize = 5

// CDCCSInterfaceDescriptor is a CDC CS interface descriptor.
type CDCCSInterfaceDescriptor struct {
	len     uint8 // 5
	dtype   uint8 // 0x24
	subtype uint8
	d0      uint8
	d1      uint8
}

// NewCDCCSInterfaceDescriptor returns a new USB CDCCSInterfaceDescriptor.
func NewCDCCSInterfaceDescriptor(subtype, d0, d1 uint8) CDCCSInterfaceDescriptor {
	return CDCCSInterfaceDescriptor{cdcCSInterfaceDescriptorSize, 0x24, subtype, d0, d1}
}

// Bytes returns CDCCSInterfaceDescriptor data.
func (d CDCCSInterfaceDescriptor) Bytes() [cdcCSInterfaceDescriptorSize]byte {
	var b [cdcCSInterfaceDescriptorSize]byte
	b[0] = byte(d.len)
	b[1] = byte(d.dtype)
	b[2] = byte(d.subtype)
	b[3] = byte(d.d0)
	b[4] = byte(d.d1)
	return b
}

const cmFunctionalDescriptorSize = 5

// CMFunctionalDescriptor is the functional descriptor general format.
type CMFunctionalDescriptor struct {
	bFunctionLength    uint8
	bDescriptorType    uint8 // 0x24
	bDescriptorSubtype uint8 // 1
	bmCapabilities     uint8
	bDataInterface     uint8
}

// NewCMFunctionalDescriptor returns a new USB CMFunctionalDescriptor.
func NewCMFunctionalDescriptor(subtype, d0, d1 uint8) CMFunctionalDescriptor {
	return CMFunctionalDescriptor{5, 0x24, subtype, d0, d1}
}

// Bytes returns the CMFunctionalDescriptor data.
func (d CMFunctionalDescriptor) Bytes() [cmFunctionalDescriptorSize]byte {
	var b [cmFunctionalDescriptorSize]byte
	b[0] = byte(d.bFunctionLength)
	b[1] = byte(d.bDescriptorType)
	b[2] = byte(d.bDescriptorSubtype)
	b[3] = byte(d.bmCapabilities)
	b[4] = byte(d.bDescriptorSubtype)
	return b
}

const acmFunctionalDescriptorSize = 4

// ACMFunctionalDescriptor is a Abstract Control Model (ACM) USB descriptor.
type ACMFunctionalDescriptor struct {
	len            uint8
	dtype          uint8 // 0x24
	subtype        uint8 // 1
	bmCapabilities uint8
}

// NewACMFunctionalDescriptor returns a new USB ACMFunctionalDescriptor.
func NewACMFunctionalDescriptor(subtype, d0 uint8) ACMFunctionalDescriptor {
	return ACMFunctionalDescriptor{4, 0x24, subtype, d0}
}

// Bytes returns the ACMFunctionalDescriptor data.
func (d ACMFunctionalDescriptor) Bytes() [acmFunctionalDescriptorSize]byte {
	var b [acmFunctionalDescriptorSize]byte
	b[0] = byte(d.len)
	b[1] = byte(d.dtype)
	b[2] = byte(d.subtype)
	b[3] = byte(d.bmCapabilities)
	return b
}

// CDCDescriptor is the Communication Device Class (CDC) descriptor.
type CDCDescriptor struct {
	//	IAD
	iad IADDescriptor // Only needed on compound device

	//	Control
	cif    InterfaceDescriptor
	header CDCCSInterfaceDescriptor

	// CDC control
	controlManagement    ACMFunctionalDescriptor  // ACM
	functionalDescriptor CDCCSInterfaceDescriptor // CDC_UNION
	callManagement       CMFunctionalDescriptor   // Call Management
	cifin                EndpointDescriptor

	//	CDC Data
	dif InterfaceDescriptor
	in  EndpointDescriptor
	out EndpointDescriptor
}

func NewCDCDescriptor(i IADDescriptor, c InterfaceDescriptor,
	h CDCCSInterfaceDescriptor,
	cm ACMFunctionalDescriptor,
	fd CDCCSInterfaceDescriptor,
	callm CMFunctionalDescriptor,
	ci EndpointDescriptor,
	di InterfaceDescriptor,
	outp EndpointDescriptor,
	inp EndpointDescriptor) CDCDescriptor {
	return CDCDescriptor{iad: i,
		cif:                  c,
		header:               h,
		controlManagement:    cm,
		functionalDescriptor: fd,
		callManagement:       callm,
		cifin:                ci,
		dif:                  di,
		in:                   inp,
		out:                  outp}
}

const cdcSize = iadDescriptorSize +
	interfaceDescriptorSize +
	cdcCSInterfaceDescriptorSize +
	acmFunctionalDescriptorSize +
	cdcCSInterfaceDescriptorSize +
	cmFunctionalDescriptorSize +
	endpointDescriptorSize +
	interfaceDescriptorSize +
	endpointDescriptorSize +
	endpointDescriptorSize

// Bytes returns CDCDescriptor data.
func (d CDCDescriptor) Bytes() [cdcSize]byte {
	var b [cdcSize]byte
	offset := 0

	iad := d.iad.Bytes()
	copy(b[offset:], iad[:])
	offset += len(iad)

	cif := d.cif.Bytes()
	copy(b[offset:], cif[:])
	offset += len(cif)

	header := d.header.Bytes()
	copy(b[offset:], header[:])
	offset += len(header)

	controlManagement := d.controlManagement.Bytes()
	copy(b[offset:], controlManagement[:])
	offset += len(controlManagement)

	functionalDescriptor := d.functionalDescriptor.Bytes()
	copy(b[offset:], functionalDescriptor[:])
	offset += len(functionalDescriptor)

	callManagement := d.callManagement.Bytes()
	copy(b[offset:], callManagement[:])
	offset += len(callManagement)

	cifin := d.cifin.Bytes()
	copy(b[offset:], cifin[:])
	offset += len(cifin)

	dif := d.dif.Bytes()
	copy(b[offset:], dif[:])
	offset += len(dif)

	out := d.out.Bytes()
	copy(b[offset:], out[:])
	offset += len(out)

	in := d.in.Bytes()
	copy(b[offset:], in[:])
	offset += len(in)

	return b
}

// MSCDescriptor is not used yet.
type MSCDescriptor struct {
	msc InterfaceDescriptor
	in  EndpointDescriptor
	out EndpointDescriptor
}

const cdcLineInfoSize = 7

type cdcLineInfo struct {
	dwDTERate   uint32
	bCharFormat uint8
	bParityType uint8
	bDataBits   uint8
	lineState   uint8
}

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

const (
	usb_IMANUFACTURER = 1
	usb_IPRODUCT      = 2
	usb_ISERIAL       = 3

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

	usbEndpointOut = 0x00
	usbEndpointIn  = 0x80

	usbEndpointPacketSize = 64 // 64 for Full Speed, EPT size max is 1024
	usb_EPT_NUM           = 7

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

	usb_DEVICE_CLASS_COMMUNICATIONS  = 0x02
	usb_DEVICE_CLASS_HUMAN_INTERFACE = 0x03
	usb_DEVICE_CLASS_STORAGE         = 0x08
	usb_DEVICE_CLASS_VENDOR_SPECIFIC = 0xFF

	usb_CONFIG_POWERED_MASK  = 0x40
	usb_CONFIG_BUS_POWERED   = 0x80
	usb_CONFIG_SELF_POWERED  = 0xC0
	usb_CONFIG_REMOTE_WAKEUP = 0x20

	// CDC
	usb_CDC_ACM_INTERFACE  = 0 // CDC ACM
	usb_CDC_DATA_INTERFACE = 1 // CDC Data
	usb_CDC_FIRST_ENDPOINT = 1
	usb_CDC_ENDPOINT_ACM   = 1
	usb_CDC_ENDPOINT_OUT   = 2
	usb_CDC_ENDPOINT_IN    = 3

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

// usbDeviceDescBank is the USB device endpoint descriptor.
// typedef struct {
// 	__IO USB_DEVICE_ADDR_Type      ADDR;        /**< \brief Offset: 0x000 (R/W 32) DEVICE_DESC_BANK Endpoint Bank, Adress of Data Buffer */
// 	__IO USB_DEVICE_PCKSIZE_Type   PCKSIZE;     /**< \brief Offset: 0x004 (R/W 32) DEVICE_DESC_BANK Endpoint Bank, Packet Size */
// 	__IO USB_DEVICE_EXTREG_Type    EXTREG;      /**< \brief Offset: 0x008 (R/W 16) DEVICE_DESC_BANK Endpoint Bank, Extended */
// 	__IO USB_DEVICE_STATUS_BK_Type STATUS_BK;   /**< \brief Offset: 0x00A (R/W  8) DEVICE_DESC_BANK Enpoint Bank, Status of Bank */
// 		 RoReg8                    Reserved1[0x5];
//   } UsbDeviceDescBank;
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

// typedef struct {
// 	union {
// 		uint8_t bmRequestType;
// 		struct {
// 			uint8_t direction : 5;
// 			uint8_t type : 2;
// 			uint8_t transferDirection : 1;
// 		};
// 	};
// 	uint8_t bRequest;
// 	uint8_t wValueL;
// 	uint8_t wValueH;
// 	uint16_t wIndex;
// 	uint16_t wLength;
// } USBSetup;
type usbSetup struct {
	bmRequestType uint8
	bRequest      uint8
	wValueL       uint8
	wValueH       uint8
	wIndex        uint16
	wLength       uint16
}

func newUSBSetup(data []byte) usbSetup {
	u := usbSetup{}
	u.bmRequestType = uint8(data[0])
	u.bRequest = uint8(data[1])
	u.wValueL = uint8(data[2])
	u.wValueH = uint8(data[3])
	u.wIndex = uint16(data[4]) | (uint16(data[5]) << 8)
	u.wLength = uint16(data[6]) | (uint16(data[7]) << 8)
	return u
}

// USBCDC is the serial interface that works over the USB port.
// To implement the USBCDC interface for a board, you must declare a concrete type as follows:
//
// 		type USBCDC struct {
// 			Buffer *RingBuffer
// 		}
//
// You can also add additional members to this struct depending on your implementation,
// but the *RingBuffer is required.
// When you are declaring the USBCDC for your board, make sure that you also declare the
// RingBuffer using the NewRingBuffer() function:
//
//		USBCDC{Buffer: NewRingBuffer()}
//

// Read from the RX buffer.
func (usbcdc *USBCDC) Read(data []byte) (n int, err error) {
	// check if RX buffer is empty
	size := usbcdc.Buffered()
	if size == 0 {
		return 0, nil
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

// Write data to the USBCDC.
func (usbcdc *USBCDC) Write(data []byte) (n int, err error) {
	for _, v := range data {
		usbcdc.WriteByte(v)
	}
	return len(data), nil
}

// ReadByte reads a single byte from the RX buffer.
// If there is no data in the buffer, returns an error.
func (usbcdc *USBCDC) ReadByte() (byte, error) {
	// check if RX buffer is empty
	buf, ok := usbcdc.Buffer.Get()
	if !ok {
		return 0, errUSBCDCBufferEmpty
	}
	return buf, nil
}

// Buffered returns the number of bytes currently stored in the RX buffer.
func (usbcdc *USBCDC) Buffered() int {
	return int(usbcdc.Buffer.Used())
}

// Receive handles adding data to the UART's data buffer.
// Usually called by the IRQ handler for a machine.
func (usbcdc *USBCDC) Receive(data byte) {
	usbcdc.Buffer.Put(data)
}

// sendDescriptor creates and sends the various USB descriptor types that
// can be requested by the host.
func sendDescriptor(setup usbSetup) {
	switch setup.wValueH {
	case usb_CONFIGURATION_DESCRIPTOR_TYPE:
		sendConfiguration(setup)
		return
	case usb_DEVICE_DESCRIPTOR_TYPE:
		// composite descriptor
		dd := NewDeviceDescriptor(0xef, 0x02, 0x01, 64, usb_VID, usb_PID, 0x100, usb_IMANUFACTURER, usb_IPRODUCT, usb_ISERIAL, 1)
		l := deviceDescriptorSize
		if setup.wLength < deviceDescriptorSize {
			l = int(setup.wLength)
		}
		buf := dd.Bytes()
		sendUSBPacket(0, buf[:l])
		return

	case usb_STRING_DESCRIPTOR_TYPE:
		switch setup.wValueL {
		case 0:
			b := []byte{0x04, 0x03, 0x09, 0x04}
			sendUSBPacket(0, b)

		case usb_IPRODUCT:
			b := make([]byte, (len(usb_STRING_PRODUCT)<<1)+2)
			strToUTF16LEDescriptor(usb_STRING_PRODUCT, b)
			sendUSBPacket(0, b)

		case usb_IMANUFACTURER:
			b := make([]byte, (len(usb_STRING_MANUFACTURER)<<1)+2)
			strToUTF16LEDescriptor(usb_STRING_MANUFACTURER, b)
			sendUSBPacket(0, b)

		case usb_ISERIAL:
			// TODO: allow returning a product serial number
			sendZlp()
		}
		return
	}

	// do not know how to handle this message, so return zero
	sendZlp()
	return
}

// sendConfiguration creates and sends the configuration packet to the host.
func sendConfiguration(setup usbSetup) {
	if setup.wLength == 9 {
		sz := uint16(configDescriptorSize + cdcSize)
		config := NewConfigDescriptor(sz, 2)
		configBuf := config.Bytes()
		sendUSBPacket(0, configBuf[:])
	} else {
		iad := NewIADDescriptor(0, 2, usb_CDC_COMMUNICATION_INTERFACE_CLASS, usb_CDC_ABSTRACT_CONTROL_MODEL, 0)

		cif := NewInterfaceDescriptor(usb_CDC_ACM_INTERFACE, 1, usb_CDC_COMMUNICATION_INTERFACE_CLASS, usb_CDC_ABSTRACT_CONTROL_MODEL, 0)

		header := NewCDCCSInterfaceDescriptor(usb_CDC_HEADER, usb_CDC_V1_10&0xFF, (usb_CDC_V1_10>>8)&0x0FF)

		controlManagement := NewACMFunctionalDescriptor(usb_CDC_ABSTRACT_CONTROL_MANAGEMENT, 6)

		functionalDescriptor := NewCDCCSInterfaceDescriptor(usb_CDC_UNION, usb_CDC_ACM_INTERFACE, usb_CDC_DATA_INTERFACE)

		callManagement := NewCMFunctionalDescriptor(usb_CDC_CALL_MANAGEMENT, 1, 1)

		cifin := NewEndpointDescriptor((usb_CDC_ENDPOINT_ACM | usbEndpointIn), usb_ENDPOINT_TYPE_INTERRUPT, 0x10, 0x10)

		dif := NewInterfaceDescriptor(usb_CDC_DATA_INTERFACE, 2, usb_CDC_DATA_INTERFACE_CLASS, 0, 0)

		out := NewEndpointDescriptor((usb_CDC_ENDPOINT_OUT | usbEndpointOut), usb_ENDPOINT_TYPE_BULK, usbEndpointPacketSize, 0)

		in := NewEndpointDescriptor((usb_CDC_ENDPOINT_IN | usbEndpointIn), usb_ENDPOINT_TYPE_BULK, usbEndpointPacketSize, 0)

		cdc := NewCDCDescriptor(iad,
			cif,
			header,
			controlManagement,
			functionalDescriptor,
			callManagement,
			cifin,
			dif,
			out,
			in)

		sz := uint16(configDescriptorSize + cdcSize)
		config := NewConfigDescriptor(sz, 2)

		configBuf := config.Bytes()
		cdcBuf := cdc.Bytes()
		var buf [configDescriptorSize + cdcSize]byte
		copy(buf[0:], configBuf[:])
		copy(buf[configDescriptorSize:], cdcBuf[:])

		sendUSBPacket(0, buf[:])
	}
}
