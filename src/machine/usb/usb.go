package usb

var (
	// TODO: allow setting these
	STRING_LANGUAGE = [2]uint16{(3 << 8) | (2 + 2), 0x0409} // English
)

const (
	DescriptorConfigCDC = 1 << iota
	DescriptorConfigHID
	DescriptorConfigMIDI
	DescriptorConfigJoystick
)

const (
	IMANUFACTURER = 1
	IPRODUCT      = 2
	ISERIAL       = 3

	ENDPOINT_TYPE_DISABLE     = 0xFF
	ENDPOINT_TYPE_CONTROL     = 0x00
	ENDPOINT_TYPE_ISOCHRONOUS = 0x01
	ENDPOINT_TYPE_BULK        = 0x02
	ENDPOINT_TYPE_INTERRUPT   = 0x03

	EndpointOut = 0x00
	EndpointIn  = 0x80

	EndpointPacketSize = 64 // 64 for Full Speed, EPT size max is 1024

	// standard requests
	GET_STATUS        = 0
	CLEAR_FEATURE     = 1
	SET_FEATURE       = 3
	SET_ADDRESS       = 5
	GET_DESCRIPTOR    = 6
	SET_DESCRIPTOR    = 7
	GET_CONFIGURATION = 8
	SET_CONFIGURATION = 9
	GET_INTERFACE     = 10
	SET_INTERFACE     = 11

	// non standard requests
	GET_REPORT      = 1
	GET_IDLE        = 2
	GET_PROTOCOL    = 3
	SET_REPORT      = 9
	SET_IDLE        = 10
	SET_PROTOCOL    = 11
	SET_REPORT_TYPE = 33

	DEVICE_CLASS_COMMUNICATIONS  = 0x02
	DEVICE_CLASS_HUMAN_INTERFACE = 0x03
	DEVICE_CLASS_STORAGE         = 0x08
	DEVICE_CLASS_VENDOR_SPECIFIC = 0xFF

	CONFIG_POWERED_MASK  = 0x40
	CONFIG_BUS_POWERED   = 0x80
	CONFIG_SELF_POWERED  = 0xC0
	CONFIG_REMOTE_WAKEUP = 0x20

	// Interface
	NumberOfInterfaces = 3
	CDC_ACM_INTERFACE  = 0 // CDC ACM
	CDC_DATA_INTERFACE = 1 // CDC Data
	CDC_FIRST_ENDPOINT = 1
	HID_INTERFACE      = 2 // HID

	// Endpoint
	CONTROL_ENDPOINT  = 0
	CDC_ENDPOINT_ACM  = 1
	CDC_ENDPOINT_OUT  = 2
	CDC_ENDPOINT_IN   = 3
	HID_ENDPOINT_IN   = 4 // for Interrupt In
	HID_ENDPOINT_OUT  = 5 // for Interrupt Out
	MIDI_ENDPOINT_IN  = 6 // for Bulk In
	MIDI_ENDPOINT_OUT = 7 // for Bulk Out
	NumberOfEndpoints = 8

	// bmRequestType
	REQUEST_HOSTTODEVICE = 0x00
	REQUEST_DEVICETOHOST = 0x80
	REQUEST_DIRECTION    = 0x80

	REQUEST_STANDARD = 0x00
	REQUEST_CLASS    = 0x20
	REQUEST_VENDOR   = 0x40
	REQUEST_TYPE     = 0x60

	REQUEST_DEVICE    = 0x00
	REQUEST_INTERFACE = 0x01
	REQUEST_ENDPOINT  = 0x02
	REQUEST_OTHER     = 0x03
	REQUEST_RECIPIENT = 0x1F

	REQUEST_DEVICETOHOST_CLASS_INTERFACE    = (REQUEST_DEVICETOHOST | REQUEST_CLASS | REQUEST_INTERFACE)
	REQUEST_HOSTTODEVICE_CLASS_INTERFACE    = (REQUEST_HOSTTODEVICE | REQUEST_CLASS | REQUEST_INTERFACE)
	REQUEST_DEVICETOHOST_STANDARD_INTERFACE = (REQUEST_DEVICETOHOST | REQUEST_STANDARD | REQUEST_INTERFACE)
)

type Setup struct {
	BmRequestType uint8
	BRequest      uint8
	WValueL       uint8
	WValueH       uint8
	WIndex        uint16
	WLength       uint16
}

func NewSetup(data []byte) Setup {
	u := Setup{}
	u.BmRequestType = uint8(data[0])
	u.BRequest = uint8(data[1])
	u.WValueL = uint8(data[2])
	u.WValueH = uint8(data[3])
	u.WIndex = uint16(data[4]) | (uint16(data[5]) << 8)
	u.WLength = uint16(data[6]) | (uint16(data[7]) << 8)
	return u
}

var (
	// VendorID aka VID is the officially assigned vendor number
	// for this USB device. Only set this if you know what you are doing,
	// since changing it can make it difficult to reflash some devices.
	VendorID uint16

	// ProductID aka PID is the product number associated with the officially assigned
	// vendor number for this USB device. Only set this if you know what you are doing,
	// since changing it can make it difficult to reflash some devices.
	ProductID uint16

	// Manufacturer is the manufacturer name displayed for this USB device.
	Manufacturer string

	// Product is the product name displayed for this USB device.
	Product string

	// Serial is the serial value displayed for this USB device. Assign a value to
	// transmit the serial to the host when requested.
	Serial string
)
