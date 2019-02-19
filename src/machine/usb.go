// +build sam

package machine

import (
	"bytes"
	"device/sam"
	"encoding/binary"
)

const deviceDescriptorSize = 18

//	DeviceDescriptor
type DeviceDescriptor struct {
	len                uint8  // 18
	dtype              uint8  // 1 USB_DEVICE_DESCRIPTOR_TYPE
	usbVersion         uint16 // 0x200
	deviceClass        uint8
	deviceSubClass     uint8
	deviceProtocol     uint8
	packetSize0        uint8 // Packet 0
	idVendor           uint16
	idProduct          uint16
	deviceVersion      uint16 // 0x100
	iManufacturer      uint8
	iProduct           uint8
	iSerialNumber      uint8
	bNumConfigurations uint8
}

// usb_D_DEVICE(_class,_subClass,_proto,_packetSize0,_vid,_pid,_version,_im,_ip,_is,_configs) \
// 	{ 18, 1, 0x200, _class,_subClass,_proto,_packetSize0,_vid,_pid,_version,_im,_ip,_is,_configs }
/* Table 9-8. Standard Device Descriptor
 * bLength, bDescriptorType, bcdUSB, bDeviceClass, bDeviceSubClass, bDeviceProtocol, bMaxPacketSize0,
 *    idVendor, idProduct, bcdDevice, iManufacturer, iProduct, iSerialNumber, bNumConfigurations */
func NewDeviceDescriptor(class, subClass, proto, packetSize0 uint8, vid, pid, version uint16, im, ip, is, configs uint8) DeviceDescriptor {
	return DeviceDescriptor{deviceDescriptorSize, 1, 0x200, class, subClass, proto, packetSize0, vid, pid, version, im, ip, is, configs}
}

// Bytes returns DeviceDescriptor data
func (d DeviceDescriptor) Bytes() []byte {
	buf := bytes.NewBuffer(make([]byte, 0, deviceDescriptorSize))
	binary.Write(buf, binary.LittleEndian, d.len)
	binary.Write(buf, binary.LittleEndian, d.dtype)
	binary.Write(buf, binary.LittleEndian, d.usbVersion)
	binary.Write(buf, binary.LittleEndian, d.deviceClass)
	binary.Write(buf, binary.LittleEndian, d.deviceSubClass)
	binary.Write(buf, binary.LittleEndian, d.deviceProtocol)
	binary.Write(buf, binary.LittleEndian, d.packetSize0)
	binary.Write(buf, binary.LittleEndian, d.idVendor)
	binary.Write(buf, binary.LittleEndian, d.idProduct)
	binary.Write(buf, binary.LittleEndian, d.deviceVersion)
	binary.Write(buf, binary.LittleEndian, d.iManufacturer)
	binary.Write(buf, binary.LittleEndian, d.iProduct)
	binary.Write(buf, binary.LittleEndian, d.iSerialNumber)
	binary.Write(buf, binary.LittleEndian, d.bNumConfigurations)
	return buf.Bytes()
}

const configDescriptorSize = 9

//	ConfigDescriptor
type ConfigDescriptor struct {
	len           uint8  // 9
	dtype         uint8  // 2
	clen          uint16 // total length
	numInterfaces uint8
	config        uint8
	iconfig       uint8
	attributes    uint8
	maxPower      uint8
}

// usb_D_CONFIG(_totalLength,_interfaces) \
// 	{ 9, 2, _totalLength,_interfaces, 1, 0, USB_CONFIG_BUS_POWERED | USB_CONFIG_REMOTE_WAKEUP, USB_CONFIG_POWER_MA(USB_CONFIG_POWER) }
/* Table 9-10. Standard Configuration Descriptor
 * bLength, bDescriptorType, wTotalLength, bNumInterfaces, bConfigurationValue, iConfiguration
 * bmAttributes, bMaxPower */
func NewConfigDescriptor(totalLength uint16, interfaces uint8) ConfigDescriptor {
	return ConfigDescriptor{configDescriptorSize, 2, totalLength, interfaces, 1, 0, usb_CONFIG_BUS_POWERED | usb_CONFIG_REMOTE_WAKEUP, 50}
}

// Bytes returns ConfigDescriptor data
func (d ConfigDescriptor) Bytes() []byte {
	buf := bytes.NewBuffer(make([]byte, 0, configDescriptorSize))
	binary.Write(buf, binary.LittleEndian, d.len)
	binary.Write(buf, binary.LittleEndian, d.dtype)
	binary.Write(buf, binary.LittleEndian, d.clen)
	binary.Write(buf, binary.LittleEndian, d.numInterfaces)
	binary.Write(buf, binary.LittleEndian, d.config)
	binary.Write(buf, binary.LittleEndian, d.iconfig)
	binary.Write(buf, binary.LittleEndian, d.attributes)
	binary.Write(buf, binary.LittleEndian, d.maxPower)
	return buf.Bytes()
}

//	String

const interfaceDescriptorSize = 9

//	InterfaceDescriptor
type InterfaceDescriptor struct {
	len               uint8 // 9
	dtype             uint8 // 4
	number            uint8
	alternate         uint8
	numEndpoints      uint8
	interfaceClass    uint8
	interfaceSubClass uint8
	protocol          uint8
	iInterface        uint8
}

// usb_D_INTERFACE(_n,_numEndpoints,_class,_subClass,_protocol) \
// 	{ 9, 4, _n, 0, _numEndpoints, _class,_subClass, _protocol, 0 }
/* Table 9-12. Standard Interface Descriptor
 * bLength, bDescriptorType, bInterfaceNumber, bAlternateSetting, bNumEndpoints, bInterfaceClass,
 * bInterfaceSubClass, bInterfaceProtocol, iInterface */
func NewInterfaceDescriptor(n, numEndpoints, class, subClass, protocol uint8) InterfaceDescriptor {
	return InterfaceDescriptor{interfaceDescriptorSize, 4, n, 0, numEndpoints, class, subClass, protocol, 0}
}

// Bytes returns InterfaceDescriptor data
func (d InterfaceDescriptor) Bytes() []byte {
	buf := bytes.NewBuffer(make([]byte, 0, interfaceDescriptorSize))
	binary.Write(buf, binary.LittleEndian, d.len)
	binary.Write(buf, binary.LittleEndian, d.dtype)
	binary.Write(buf, binary.LittleEndian, d.number)
	binary.Write(buf, binary.LittleEndian, d.alternate)
	binary.Write(buf, binary.LittleEndian, d.numEndpoints)
	binary.Write(buf, binary.LittleEndian, d.interfaceClass)
	binary.Write(buf, binary.LittleEndian, d.interfaceSubClass)
	binary.Write(buf, binary.LittleEndian, d.protocol)
	binary.Write(buf, binary.LittleEndian, d.iInterface)
	return buf.Bytes()
}

const endpointDescriptorSize = 7

//	EndpointDescriptor
type EndpointDescriptor struct {
	len        uint8 // 7
	dtype      uint8 // 5
	addr       uint8
	attr       uint8
	packetSize uint16
	interval   uint8
}

// usb_D_ENDPOINT(_addr,_attr,_packetSize, _interval) \
// 	{ 7, 5, _addr,_attr,_packetSize, _interval }
/* Table 9-13. Standard Endpoint Descriptor
 * bLength, bDescriptorType, bEndpointAddress, bmAttributes, wMaxPacketSize, bInterval */
func NewEndpointDescriptor(addr, attr uint8, packetSize uint16, interval uint8) EndpointDescriptor {
	return EndpointDescriptor{endpointDescriptorSize, 5, addr, attr, packetSize, interval}
}

// Bytes returns EndpointDescriptor data
func (d EndpointDescriptor) Bytes() []byte {
	buf := bytes.NewBuffer(make([]byte, 0, endpointDescriptorSize))
	binary.Write(buf, binary.LittleEndian, d.len)
	binary.Write(buf, binary.LittleEndian, d.dtype)
	binary.Write(buf, binary.LittleEndian, d.addr)
	binary.Write(buf, binary.LittleEndian, d.attr)
	binary.Write(buf, binary.LittleEndian, d.packetSize)
	binary.Write(buf, binary.LittleEndian, d.interval)
	return buf.Bytes()
}

const iadDescriptorSize = 8

// IADDescriptor is an Interface Association Descriptor
// Used to bind 2 interfaces together in CDC composite device
type IADDescriptor struct {
	len              uint8 // 8
	dtype            uint8 // 11
	firstInterface   uint8
	interfaceCount   uint8
	functionClass    uint8
	funtionSubClass  uint8
	functionProtocol uint8
	iInterface       uint8
}

// usb_D_IAD(_firstInterface, _count, _class, _subClass, _protocol) \
// 	{ 8, 11, _firstInterface, _count, _class, _subClass, _protocol, 0 }
/* iadclasscode_r10.pdf, Table 9\96Z. Standard Interface Association Descriptor
 * bLength, bDescriptorType, bFirstInterface, bInterfaceCount, bFunctionClass, bFunctionSubClass, bFunctionProtocol, iFunction */
func NewIADDescriptor(firstInterface, count, class, subClass, protocol uint8) IADDescriptor {
	return IADDescriptor{iadDescriptorSize, 11, firstInterface, count, class, subClass, protocol, 0}
}

// Bytes returns IADDescriptor data
func (d IADDescriptor) Bytes() []byte {
	buf := bytes.NewBuffer(make([]byte, 0, iadDescriptorSize))
	binary.Write(buf, binary.LittleEndian, d.len)
	binary.Write(buf, binary.LittleEndian, d.dtype)
	binary.Write(buf, binary.LittleEndian, d.firstInterface)
	binary.Write(buf, binary.LittleEndian, d.interfaceCount)
	binary.Write(buf, binary.LittleEndian, d.functionClass)
	binary.Write(buf, binary.LittleEndian, d.funtionSubClass)
	binary.Write(buf, binary.LittleEndian, d.functionProtocol)
	binary.Write(buf, binary.LittleEndian, d.iInterface)
	return buf.Bytes()
}

// CDC interfaces
// CDC CS interface descriptor

const cdcCSInterfaceDescriptorSize = 5

type CDCCSInterfaceDescriptor struct {
	len     uint8 // 5
	dtype   uint8 // 0x24
	subtype uint8
	d0      uint8
	d1      uint8
}

// usb_D_CDCCS(_subtype,_d0,_d1)	{ 5, 0x24, _subtype, _d0, _d1 }
func NewCDCCSInterfaceDescriptor(subtype, d0, d1 uint8) CDCCSInterfaceDescriptor {
	return CDCCSInterfaceDescriptor{cdcCSInterfaceDescriptorSize, 0x24, subtype, d0, d1}
}

func (d CDCCSInterfaceDescriptor) Bytes() []byte {
	buf := bytes.NewBuffer(make([]byte, 0, cdcCSInterfaceDescriptorSize))
	binary.Write(buf, binary.LittleEndian, d.len)
	binary.Write(buf, binary.LittleEndian, d.dtype)
	binary.Write(buf, binary.LittleEndian, d.subtype)
	binary.Write(buf, binary.LittleEndian, d.d0)
	binary.Write(buf, binary.LittleEndian, d.d1)
	return buf.Bytes()
}

const cdcCSInterfaceDescriptor4Size = 4

type CDCCSInterfaceDescriptor4 struct {
	len     uint8 // 4
	dtype   uint8 // 0x24
	subtype uint8
	d0      uint8
}

func (d CDCCSInterfaceDescriptor4) Bytes() []byte {
	buf := bytes.NewBuffer(make([]byte, 0, cdcCSInterfaceDescriptor4Size))
	binary.Write(buf, binary.LittleEndian, d.len)
	binary.Write(buf, binary.LittleEndian, d.dtype)
	binary.Write(buf, binary.LittleEndian, d.subtype)
	binary.Write(buf, binary.LittleEndian, d.d0)
	return buf.Bytes()
}

// Functional Descriptor General Format

const cmFunctionalDescriptorSize = 5

type CMFunctionalDescriptor struct {
	len            uint8
	dtype          uint8 // 0x24
	subtype        uint8 // 1
	bmCapabilities uint8
	bDataInterface uint8
}

// D_CDCCS(_subtype,_d0,_d1)	{ 5, 0x24, _subtype, _d0, _d1 }
func NewCMFunctionalDescriptor(subtype, d0, d1 uint8) CMFunctionalDescriptor {
	return CMFunctionalDescriptor{5, 0x24, subtype, d0, d1}
}

func (d CMFunctionalDescriptor) Bytes() []byte {
	buf := bytes.NewBuffer(make([]byte, 0, cmFunctionalDescriptorSize))
	binary.Write(buf, binary.LittleEndian, d.len)
	binary.Write(buf, binary.LittleEndian, d.dtype)
	binary.Write(buf, binary.LittleEndian, d.subtype)
	binary.Write(buf, binary.LittleEndian, d.bmCapabilities)
	binary.Write(buf, binary.LittleEndian, d.bDataInterface)
	return buf.Bytes()
}

/* bFunctionLength, bDescriptorType, bDescriptorSubtype, function specific data0, functional specific data N-1
 * CS_INTERFACE 24h */

const acmFunctionalDescriptorSize = 4

type ACMFunctionalDescriptor struct {
	len            uint8
	dtype          uint8 // 0x24
	subtype        uint8 // 1
	bmCapabilities uint8
}

// #define D_CDCCS4(_subtype,_d0)		{ 4, 0x24, _subtype, _d0 }
func NewACMFunctionalDescriptor(subtype, d0 uint8) ACMFunctionalDescriptor {
	return ACMFunctionalDescriptor{4, 0x24, subtype, d0}
}

func (d ACMFunctionalDescriptor) Bytes() []byte {
	buf := bytes.NewBuffer(make([]byte, 0, acmFunctionalDescriptorSize))
	binary.Write(buf, binary.LittleEndian, d.len)
	binary.Write(buf, binary.LittleEndian, d.dtype)
	binary.Write(buf, binary.LittleEndian, d.subtype)
	binary.Write(buf, binary.LittleEndian, d.bmCapabilities)
	return buf.Bytes()
}

type CDCDescriptor struct {
	//	IAD
	iad IADDescriptor // Only needed on compound device
	//	Control
	cif    InterfaceDescriptor
	header CDCCSInterfaceDescriptor

	controlManagement    ACMFunctionalDescriptor  // ACM
	functionalDescriptor CDCCSInterfaceDescriptor // CDC_UNION
	callManagement       CMFunctionalDescriptor   // Call Management
	cifin                EndpointDescriptor

	//	Data
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
	inp EndpointDescriptor,
	outp EndpointDescriptor) CDCDescriptor {
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

// Bytes returns CDCDescriptor data
func (d CDCDescriptor) Bytes() []byte {
	buf := bytes.NewBuffer(make([]byte, 0, cdcSize))
	buf.Write(d.iad.Bytes())
	buf.Write(d.cif.Bytes())
	buf.Write(d.header.Bytes())
	buf.Write(d.controlManagement.Bytes())
	buf.Write(d.functionalDescriptor.Bytes())
	buf.Write(d.callManagement.Bytes())
	buf.Write(d.cifin.Bytes())
	buf.Write(d.dif.Bytes())
	buf.Write(d.in.Bytes())
	buf.Write(d.out.Bytes())
	return buf.Bytes()
}

// static CDCDescriptor _cdcInterface = {
// 	D_IAD(0, 2, CDC_COMMUNICATION_INTERFACE_CLASS, CDC_ABSTRACT_CONTROL_MODEL, 0),

// 	// CDC communication interface
// 	D_INTERFACE(CDC_ACM_INTERFACE, 1, CDC_COMMUNICATION_INTERFACE_CLASS, CDC_ABSTRACT_CONTROL_MODEL, 0),
// 	D_CDCCS(CDC_HEADER, CDC_V1_10 & 0xFF, (CDC_V1_10>>8) & 0x0FF), // Header (1.10 bcd)

// 	D_CDCCS4(CDC_ABSTRACT_CONTROL_MANAGEMENT, 6), // SET_LINE_CODING, GET_LINE_CODING, SET_CONTROL_LINE_STATE supported
// 	D_CDCCS(CDC_UNION, CDC_ACM_INTERFACE, CDC_DATA_INTERFACE), // Communication interface is master, data interface is slave 0
// 	D_CDCCS(CDC_CALL_MANAGEMENT, 1, 1), // Device handles call management (not)
// 	D_ENDPOINT(USB_ENDPOINT_IN(CDC_ENDPOINT_ACM), USB_ENDPOINT_TYPE_INTERRUPT, 0x10, 0x10),

// 	// CDC data interface
// 	D_INTERFACE(CDC_DATA_INTERFACE, 2, CDC_DATA_INTERFACE_CLASS, 0, 0),
// 	D_ENDPOINT(USB_ENDPOINT_OUT(CDC_ENDPOINT_OUT), USB_ENDPOINT_TYPE_BULK, EPX_SIZE, 0),
// 	D_ENDPOINT(USB_ENDPOINT_IN (CDC_ENDPOINT_IN ), USB_ENDPOINT_TYPE_BULK, EPX_SIZE, 0)
// };

// MSCDescriptor is not used yet.
type MSCDescriptor struct {
	msc InterfaceDescriptor
	in  EndpointDescriptor
	out EndpointDescriptor
}

// _Pragma("pack(1)")
// static volatile LineInfo _usbLineInfo = {
// 	115200, // dWDTERate
// 	0x00,   // bCharFormat
// 	0x00,   // bParityType
// 	0x08,   // bDataBits
// 	0x00    // lineState
// };

type cdcLineInfo struct {
	dwDTERate   uint32
	bCharFormat uint8
	bParityType uint8
	bDataBits   uint8
	lineState   uint8
}

var (
	// TODO: allow setting these
	usb_STRING_LANGUAGE     = [2]uint16{(3 << 8) | (2 + 2), 0x0409} // English
	usb_STRING_PRODUCT      = "Arduino Zero"
	usb_STRING_MANUFACTURER = "Arduino"

	usb_VID uint16 = 0x2341
	usb_PID uint16 = 0x004d
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

	usb_EPX_SIZE = 64 // 64 for Full Speed, EPT size max is 1024
	usb_EPT_NUM  = 7

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
	ADDR      sam.RegValue
	PCKSIZE   sam.RegValue
	EXTREG    sam.RegValue16
	STATUS_BK sam.RegValue8
	_reserved [5]sam.RegValue8
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
	buf := bytes.NewBuffer(data)
	u := usbSetup{}
	binary.Read(buf, binary.LittleEndian, &(u.bmRequestType))
	binary.Read(buf, binary.LittleEndian, &(u.bRequest))
	binary.Read(buf, binary.LittleEndian, &(u.wValueL))
	binary.Read(buf, binary.LittleEndian, &(u.wValueH))
	binary.Read(buf, binary.LittleEndian, &(u.wIndex))
	binary.Read(buf, binary.LittleEndian, &(u.wLength))
	return u
}

func (u usbSetup) Print() {
	println(u.bmRequestType, u.bRequest, u.wValueL, u.wValueH, u.wIndex, u.wLength)
}
