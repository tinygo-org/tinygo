//go:build (sam && atsamd51) || (sam && atsame5x)
// +build sam,atsamd51 sam,atsame5x

package usb

// descCPUFrequencyHz defines the target CPU frequency (Hz).
const descCPUFrequencyHz = 120000000

// descCoreCount defines the number of USB PHY cores available on this platform,
// independent of the number of cores which shall be configured as TinyGo USB
// host/device controller instances.
const descCoreCount = 1 // SAMx51 has a single, full-speed USB PHY

// descCDCACMCount defines the number of USB cores that may be configured as
// CDC-ACM (single) devices.
const descCDCACMCount = 1

// descHIDCount defines the number of USB cores that may be configured as a
// composite (keyboard + mouse + joystick) human interface device (HID).
const descHIDCount = 0

// General USB device identification constants.
const (
	descCommonVendorID  = 0x03EB
	descCommonProductID = 0x2421
	descCommonReleaseID = 0x0101 // BCD (1.1)

	descCommonLanguage     = descLanguageEnglish
	descCommonManufacturer = "TinyGo"
	descCommonProduct      = "USB"
	descCommonSerialNumber = "00000"
)

// Constants for all USB device classes.
const (

	// USB endpoints parameters

	descMaxEndpoints = 8 // SAMx51 maximum number of endpoints

	descBankOut = 0 // descriptor bank 0 holds OUT endpoints
	descBankIn  = 1 // descriptor bank 1 holds IN endpoints

	descControlPacketSize = 16
)

// Constants for USB CDC-ACM device classes.
const (

	// USB Bus Configuration Attributes

	descCDCACMMaxPowerMa = 100 // Maximum current (mA) requested from host

	// CDC-ACM Endpoint Descriptor Buffers

	descCDCACMEDCount = descMaxEndpoints

	// Setup packet is only 8 bytes in length. However, under certain scenarios,
	// USB DMA controller may decide to overwrite/overflow the buffer with 2 extra
	// bytes of CRC. From datasheet's "Management of SETUP Transactions" section:
	//   | If the number of received data bytes is the maximum data payload
	//   | specified by PCKSIZE.SIZE minus one, only the first CRC data is written
	//   | to the data buffer. If the number of received data is equal or less
	//   | than the data payload specified by PCKSIZE.SIZE minus two, both CRC
	//   | data bytes are written to the data buffer.
	// Thus, we need to allocate 2 extra bytes for control endpoint 0 Rx (OUT).
	descCDCACMCxSize = 8 + 2

	// CDC-ACM Data Buffers

	descCDCACMRxSize = 4 * descCDCACMDataRxPacketSize
	descCDCACMTxSize = 4 * descCDCACMDataTxPacketSize

	descCDCACMTxTimeoutMs = 120 // millisec
	descCDCACMTxSyncUs    = 75  // microsec

	// Default CDC-ACM Endpoint Configurations (Full-Speed)

	descCDCACMStatusInterval   = descCDCACMStatusFSInterval   // Status
	descCDCACMStatusPacketSize = descCDCACMStatusFSPacketSize //

	descCDCACMDataRxPacketSize = descCDCACMDataRxFSPacketSize // Data Rx
	descCDCACMDataTxPacketSize = descCDCACMDataTxFSPacketSize // Data Tx

	// CDC-ACM Endpoint Configurations for Full-Speed Device

	descCDCACMStatusFSInterval   = 5  // Status
	descCDCACMStatusFSPacketSize = 16 //  (full-speed)

	descCDCACMDataRxFSPacketSize = 64 // Data Rx (full-speed)
	descCDCACMDataTxFSPacketSize = 64 // Data Tx (full-speed)

	// CDC-ACM Endpoint Configurations for High-Speed Device

	//  - N/A, SAMx51 only has a full-speed PHY
)

// Constants for USB HID (keyboard, mouse, joystick) device classes.
const (

	// USB Bus Configuration Attributes

	descHIDMaxPowerMa = 100 // Maximum current (mA) requested from host

	// HID Endpoint Descriptor Buffers

	descHIDEDCount = descMaxEndpoints

	// Setup packet is only 8 bytes in length. However, under certain scenarios,
	// USB DMA controller may decide to overwrite/overflow the buffer with 2 extra
	// bytes of CRC. From datasheet's "Management of SETUP Transactions" section:
	//   | If the number of received data bytes is the maximum data payload
	//   | specified by PCKSIZE.SIZE minus one, only the first CRC data is written
	//   | to the data buffer. If the number of received data is equal or less
	//   | than the data payload specified by PCKSIZE.SIZE minus two, both CRC
	//   | data bytes are written to the data buffer.
	// Thus, we need to allocate 2 extra bytes for control endpoint 0 Rx (OUT).
	descHIDCxSize = 8 + 2

	// HID Serial Buffers

	descHIDSerialRxSize = descHIDSerialRxPacketSize
	descHIDSerialTxSize = descHIDSerialTxPacketSize

	descHIDSerialTxTimeoutMs = 50 // millisec
	descHIDSerialTxSyncUs    = 75 // microsec

	// HID Keyboard Buffers

	descHIDKeyboardTxSize = 4 * descHIDKeyboardTxPacketSize

	descHIDKeyboardTxTimeoutMs = 50 // millisec

	// HID Mouse Buffers

	descHIDMouseTxSize = 4 * descHIDMouseTxPacketSize

	descHIDMouseTxTimeoutMs = 30 // millisec

	// HID Joystick Buffers

	descHIDJoystickTxSize = 4 * descHIDJoystickTxPacketSize

	descHIDJoystickTxTimeoutMs = 30 // millisec

	// Default HID Endpoint Configurations (Full-Speed)

	descHIDSerialRxInterval   = descHIDSerialRxFSInterval   // Serial Rx
	descHIDSerialRxPacketSize = descHIDSerialRxFSPacketSize //

	descHIDSerialTxInterval   = descHIDSerialTxFSInterval   // Serial Tx
	descHIDSerialTxPacketSize = descHIDSerialTxFSPacketSize //

	descHIDKeyboardTxInterval   = descHIDKeyboardTxFSInterval   // Keyboard
	descHIDKeyboardTxPacketSize = descHIDKeyboardTxFSPacketSize //

	descHIDMediaKeyTxInterval   = descHIDMediaKeyTxFSInterval   // Keyboard Media Keys
	descHIDMediaKeyTxPacketSize = descHIDMediaKeyTxFSPacketSize //

	descHIDMouseTxInterval   = descHIDMouseTxFSInterval   // Mouse
	descHIDMouseTxPacketSize = descHIDMouseTxFSPacketSize //

	descHIDJoystickTxInterval   = descHIDJoystickTxFSInterval   // Joystick
	descHIDJoystickTxPacketSize = descHIDJoystickTxFSPacketSize //

	// HID Endpoint Configurations for Full-Speed Device

	descHIDSerialRxFSInterval   = 2 // Serial Rx
	descHIDSerialRxFSPacketSize = 8 //  (full-speed)

	descHIDSerialTxFSInterval   = 1  // Serial Tx
	descHIDSerialTxFSPacketSize = 16 //  (full-speed)

	descHIDKeyboardTxFSInterval   = 4 // Keyboard
	descHIDKeyboardTxFSPacketSize = 8 //  (full-speed)

	descHIDMediaKeyTxFSInterval   = 4 // Keyboard Media Keys
	descHIDMediaKeyTxFSPacketSize = 8 //  (full-speed)

	descHIDMouseTxFSInterval   = 4 // Mouse
	descHIDMouseTxFSPacketSize = 8 //  (full-speed)

	descHIDJoystickTxFSInterval   = 4  // Joystick
	descHIDJoystickTxFSPacketSize = 12 //  (full-speed)

	// HID Endpoint Configurations for High-Speed Device

	//  - N/A, SAMx51 only has a full-speed PHY
)

// descCDCACM0ED is an array of endpoint descriptors, which describes to the USB
// DMA controller the buffer and transfer properties for each endpoint, for the
// default CDC-ACM (single) device class configuration (index 1).
//go:align 32
var descCDCACM0ED [descCDCACMEDCount]dhwEndptAddrDesc

// descCDCACM0Cx is the buffer for control/status data received on endpoint 0 of
// the default CDC-ACM (single) device class configuration (index 1).
//go:align 32
var descCDCACM0Cx [descCDCACMCxSize]uint8

// descCDCACM0Dx is the transmit (Tx) buffer of descriptor data on endpoint 0
// for the default CDC-ACM (single) device class configuration (index 1).
//go:align 32
var descCDCACM0Dx [descCDCACMConfigSize]uint8

// descCDCACM0Rx is the receive (Rx) transfer buffer for the default CDC-ACM
// (single) device class configuration (index 1).
//go:align 32
var descCDCACM0Rx [descCDCACMRxSize]uint8

// descCDCACM0Tx is the transmit (Tx) transfer buffer for the default CDC-ACM
// (single) device class configuration (index 1).
//go:align 32
var descCDCACM0Tx [descCDCACMTxSize]uint8

// descCDCACMClassData holds the buffers and control states for all CDC-ACM
// (single) device class configurations, ordered by index (offset by -1), for
// SAMx51 targets only.
//
// Instances of this type (elements of descCDCACMData) are embedded in elements
// of the common/target-agnostic CDC-ACM class configurations (descCDCACM).
// Methods defined on this type implement target-specific functionality, and
// some of these methods are required by the common device controller driver.
// Thus, this type functions as a hardware abstraction layer (HAL).
type descCDCACMClassData struct {

	// CDC-ACM Control Buffers

	ed *[descCDCACMEDCount]dhwEndptAddrDesc // endpoint descriptors

	cx *[descCDCACMCxSize]uint8     // control endpoint 0 Rx/Tx transfer buffer
	dx *[descCDCACMConfigSize]uint8 // control endpoint 0 Tx (IN) descriptor transfer buffer

	// CDC-ACM Data Buffers

	rx *[descCDCACMRxSize]uint8 // bulk data endpoint Rx (OUT) transfer buffer
	tx *[descCDCACMTxSize]uint8 // bulk data endpoint Tx (IN) transfer buffer

	sxSize uint16
	rxSize uint16
	txSize uint16
}

// descCDCACMData holds statically-allocated instances for each of the target-
// specific (SAMx51) CDC-ACM (single) device class configurations' control and
// data structures, ordered by configuration index (offset by -1). Each element
// is embedded in a corresponding element of descCDCACM.
var descCDCACMData = [dcdCount]descCDCACMClassData{

	{ // -- CDC-ACM (single) Class Configuration Index 1 --

		// CDC-ACM Control Buffers

		ed: &descCDCACM0ED,

		cx: &descCDCACM0Cx,
		dx: &descCDCACM0Dx,

		// CDC-ACM Data Buffers

		rx: &descCDCACM0Rx,
		tx: &descCDCACM0Tx,

		sxSize: descCDCACMStatusPacketSize,
		rxSize: descCDCACMDataRxPacketSize,
		txSize: descCDCACMDataTxPacketSize,
	},
}

// descHID0ED is an array of endpoint descriptors, which describes to the USB
// DMA controller the buffer and transfer properties for each endpoint, for the
// default HID device class configuration (index 1).
//go:align 32
var descHID0ED [descHIDEDCount]dhwEndptAddrDesc

// descHID0Cx is the buffer for control/status data received on endpoint 0 of
// the default HID device class configuration (index 1).
//go:align 32
var descHID0Cx [descHIDCxSize]uint8

// descHID0Dx is the transmit (Tx) buffer of descriptor data on endpoint 0 for
// the default HID device class configuration (index 1).
//go:align 32
var descHID0Dx [descHIDConfigSize]uint8

// descHID0SerialRx is the serial receive (Rx) transfer buffer for the default
// HID device class configuration (index 1).
//go:align 32
var descHID0SerialRx [descHIDSerialRxSize]uint8

// descHID0SerialTx is the serial transmit (Tx) transfer buffer for the default
// HID device class configuration (index 1).
//go:align 32
var descHID0SerialTx [descHIDSerialTxSize]uint8

// descHID0KeyboardTx is the keyboard transmit (Tx) transfer buffer for the
// default HID device class configuration (index 1).
//go:align 32
var descHID0KeyboardTx [descHIDKeyboardTxSize]uint8

// descHID0KeyboardTp is the keyboard HID report transmit (Tx) transfer buffer
// for the default HID device class configuration (index 1).
//go:align 32
var descHID0KeyboardTp [descHIDKeyboardTxPacketSize]uint8

// descHID0MouseTx is the mouse transmit (Tx) transfer buffer for the default
// HID device class configuration (index 1).
//go:align 32
var descHID0MouseTx [descHIDMouseTxSize]uint8

// descHID0JoystickTx is the joystick transmit (Tx) transfer buffer for the
// default HID device class configuration (index 1).
//go:align 32
var descHID0JoystickTx [descHIDJoystickTxSize]uint8

var descHID0KeyboardTxKey [hidKeyboardKeyCount]uint8
var descHID0KeyboardTxCon [hidKeyboardConCount]uint16
var descHID0KeyboardTxSys [hidKeyboardSysCount]uint8

// descHID0Keyboard is the Keyboard instance with which the user may interact
// when using the default HID device class configuration (index 1).
var descHID0Keyboard = Keyboard{
	key: &descHID0KeyboardTxKey,
	con: &descHID0KeyboardTxCon,
	sys: &descHID0KeyboardTxSys,
}

// descHIDClassData holds the buffers and control states for all of the HID
// device class configurations, ordered by index (offset by -1), for SAMx51
// targets only.
//
// Instances of this type (elements of descHIDData) are embedded in elements
// of the common/target-agnostic HID class configurations (descHID).
// Methods defined on this type implement target-specific functionality, and
// some of these methods are required by the common device controller driver.
// Thus, this type functions as a hardware abstraction layer (HAL).
type descHIDClassData struct {

	// HID Control Buffers

	ed *[descHIDEDCount]dhwEndptAddrDesc // endpoint descriptors

	cx *[descHIDCxSize]uint8     // control endpoint 0 Rx/Tx transfer buffer
	dx *[descHIDConfigSize]uint8 // control endpoint 0 Tx (IN) descriptor transfer buffer

	// HID Serial Buffers

	rxSerial *[descHIDSerialRxSize]uint8 // interrupt endpoint serial Rx (OUT) transfer buffer
	txSerial *[descHIDSerialTxSize]uint8 // interrupt endpoint serial Tx (IN) transfer buffer

	rxSerialSize uint16
	txSerialSize uint16

	// HID Keyboard Buffers

	txKeyboard *[descHIDKeyboardTxSize]uint8       // interrupt endpoint keyboard Tx (IN) transfer buffer
	tpKeyboard *[descHIDKeyboardTxPacketSize]uint8 // interrupt endpoint keyboard Tx (IN) HID report bbuffer

	txKeyboardSize uint16

	// HID Mouse Buffers

	txMouse *[descHIDMouseTxSize]uint8 // interrupt endpoint mouse Tx (IN) transfer buffer

	txMouseSize uint16

	// HID Joystick Buffers

	txJoystick *[descHIDJoystickTxSize]uint8 // interrupt endpoint joystick Tx (IN) transfer buffer

	txJoystickSize uint16

	// HID Device Instances

	//serial *Serial
	keyboard *Keyboard
	//mouse *Mouse
	//joystick *Joystick
}

// descHIDData holds statically-allocated instances for each of the target-
// specific (SAMx51) HID device class configurations' control and data
// structures, ordered by configuration index (offset by -1). Each element is
// embedded in a corresponding element of descHID.
var descHIDData = [dcdCount]descHIDClassData{

	{ // -- HID Class Configuration Index 1 --

		// HID Control Buffers

		ed: &descHID0ED,

		cx: &descHID0Cx,
		dx: &descHID0Dx,

		// HID Serial Buffers

		rxSerial: &descHID0SerialRx,
		txSerial: &descHID0SerialTx,

		rxSerialSize: descHIDSerialRxPacketSize,
		txSerialSize: descHIDSerialTxPacketSize,

		// HID Keyboard Buffers

		txKeyboard: &descHID0KeyboardTx,
		tpKeyboard: &descHID0KeyboardTp,

		txKeyboardSize: descHIDKeyboardTxPacketSize,

		// HID Mouse Buffers

		txMouse: &descHID0MouseTx,

		txMouseSize: descHIDMouseTxPacketSize,

		// HID Joystick Buffers

		txJoystick: &descHID0JoystickTx,

		txJoystickSize: descHIDJoystickTxPacketSize,

		// HID Device Instances

		//serial: &descHID0Serial,
		keyboard: &descHID0Keyboard,
		//mouse: &descHID0Mouse,
		//joystick: &descHID0Joystick,
	},
}
