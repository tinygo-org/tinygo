// +build mimxrt1062

package usb

// descCPUFrequencyHz defines the target CPU frequency (Hz).
const descCPUFrequencyHz = 600000000

// descCDCACMCount defines the number of USB cores that may be configured as
// CDC-ACM (single) devices.
const descCDCACMCount = 0

// descHIDCount defines the number of USB cores that may be configured as a
// composite (keyboard + mouse + joystick) human interface device (HID).
const descHIDCount = 1

// General USB device identification constants.
const (
	descCommonVendorID  = 0x16C0
	descCommonProductID = 0x0483
	descCommonReleaseID = 0x0101 // BCD (1.1)

	descCommonLanguage     = descLanguageEnglish
	descCommonManufacturer = "TinyGo"
	descCommonProduct      = "USB"
	descCommonSerialNumber = "00000"
)

// Constants for USB CDC-ACM device classes.
const (
	descCDCACMMaxPower = 50 // 100 mA

	// CDC-ACM Control Buffers

	descCDCACMQHCount = 2 * (descCDCACMEndpointCount + 1)
	descCDCACMCxCount = 8

	// CDC-ACM Data Buffers

	descCDCACMRDCount = 2 * descCDCACMEndpointCount
	descCDCACMRxSize  = descCDCACMDataRxPacketSize
	descCDCACMRxCount = descCDCACMRxSize * descCDCACMRDCount

	descCDCACMTDCount = descCDCACMEndpointCount
	descCDCACMTxSize  = 4 * descCDCACMDataTxPacketSize
	descCDCACMTxCount = descCDCACMTxSize * descCDCACMTDCount

	descCDCACMTxTimeoutMs = 120 // millisec
	descCDCACMTxSyncUs    = 75  // microsec

	// Default CDC-ACM Endpoint Configurations (High-Speed)

	descCDCACMStatusInterval   = descCDCACMStatusHSInterval   // Status
	descCDCACMStatusPacketSize = descCDCACMStatusHSPacketSize //

	descCDCACMDataRxPacketSize = descCDCACMDataRxHSPacketSize // Data Rx

	descCDCACMDataTxPacketSize = descCDCACMDataTxHSPacketSize // Data Tx

	// CDC-ACM Endpoint Configurations for Full-Speed Device

	descCDCACMStatusFSInterval   = 5  // Status
	descCDCACMStatusFSPacketSize = 16 //  (full-speed)

	descCDCACMDataRxFSPacketSize = 64 // Data Rx (full-speed)

	descCDCACMDataTxFSPacketSize = 64 // Data Tx (full-speed)

	// CDC-ACM Endpoint Configurations for High-Speed Device

	descCDCACMStatusHSInterval   = 5  // Status
	descCDCACMStatusHSPacketSize = 16 //  (high-speed)

	descCDCACMDataRxHSPacketSize = 512 // Data Rx (high-speed)

	descCDCACMDataTxHSPacketSize = 512 // Data Tx (high-speed)
)

// Constants for USB HID (keyboard, mouse, joystick) device classes.
const (
	descHIDMaxPower = 50 // 100 mA

	// HID Control Buffers

	descHIDQHCount = 2 * (descHIDEndpointCount + 1)
	descHIDCxCount = 8

	// HID Serial Buffers

	descHIDSerialRDCount = 8
	descHIDSerialRxSize  = descHIDSerialRxPacketSize
	descHIDSerialRxCount = descHIDSerialRxSize * descHIDSerialRDCount

	descHIDSerialTDCount = 12
	descHIDSerialTxSize  = descHIDSerialTxPacketSize
	descHIDSerialTxCount = descHIDSerialTxSize * descHIDSerialTDCount

	descHIDSerialTxTimeoutMs = 50 // millisec
	descHIDSerialTxSyncUs    = 75 // microsec

	// HID Keyboard Buffers

	descHIDKeyboardTDCount = 12
	descHIDKeyboardTxSize  = 4 * descHIDKeyboardTxPacketSize
	descHIDKeyboardTxCount = descHIDKeyboardTxSize * descHIDKeyboardTDCount

	descHIDKeyboardTxTimeoutMs = 50 // millisec

	// HID Mouse Buffers

	descHIDMouseTDCount = 4
	descHIDMouseTxSize  = 4 * descHIDMouseTxPacketSize
	descHIDMouseTxCount = descHIDMouseTxSize * descHIDMouseTDCount

	descHIDMouseTxTimeoutMs = 30 // millisec

	// HID Joystick Buffers

	descHIDJoystickTDCount = 4
	descHIDJoystickTxSize  = 4 * descHIDJoystickTxPacketSize
	descHIDJoystickTxCount = descHIDJoystickTxSize * descHIDJoystickTDCount

	descHIDJoystickTxTimeoutMs = 30 // millisec

	// Default HID Endpoint Configurations (High-Speed)

	descHIDSerialRxInterval   = descHIDSerialRxHSInterval   // Serial Rx
	descHIDSerialRxPacketSize = descHIDSerialRxHSPacketSize //

	descHIDSerialTxInterval   = descHIDSerialTxHSInterval   // Serial Tx
	descHIDSerialTxPacketSize = descHIDSerialTxHSPacketSize //

	descHIDKeyboardTxInterval   = descHIDKeyboardTxHSInterval   // Keyboard
	descHIDKeyboardTxPacketSize = descHIDKeyboardTxHSPacketSize //

	descHIDMediaKeyTxInterval   = descHIDMediaKeyTxHSInterval   // Keyboard Media Keys
	descHIDMediaKeyTxPacketSize = descHIDMediaKeyTxHSPacketSize //

	descHIDMouseTxInterval   = descHIDMouseTxHSInterval   // Mouse
	descHIDMouseTxPacketSize = descHIDMouseTxHSPacketSize //

	descHIDJoystickTxInterval   = descHIDJoystickTxHSInterval   // Joystick
	descHIDJoystickTxPacketSize = descHIDJoystickTxHSPacketSize //

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

	descHIDSerialRxHSInterval   = 2  // Serial
	descHIDSerialRxHSPacketSize = 32 //  (high-speed)

	descHIDSerialTxHSInterval   = 1  // Serial Tx
	descHIDSerialTxHSPacketSize = 64 //  (high-speed)

	descHIDKeyboardTxHSInterval   = 1 // Keyboard
	descHIDKeyboardTxHSPacketSize = 8 //  (high-speed)

	descHIDMediaKeyTxHSInterval   = 4 // Keyboard Media Keys
	descHIDMediaKeyTxHSPacketSize = 8 //  (high-speed)

	descHIDMouseTxHSInterval   = 1 // Mouse
	descHIDMouseTxHSPacketSize = 8 //  (high-speed)

	descHIDJoystickTxHSInterval   = 2  // Joystick
	descHIDJoystickTxHSPacketSize = 12 //  (high-speed)
)

// descCDCACM0QH is an array of endpoint queue heads, which is where all
// transfers for a given endpoint are managed, for the default CDC-ACM (single)
// device class configuration (index 1).
//
// From the iMXRT1062 Reference Manual:
//
//   Software must ensure that no interface data structure reachable
//   by the Device Controller spans a 4K-page boundary.
//
//   The [queue head] is a 48-byte data structure, but must be aligned on
//   64-byte boundaries.
//
//   Endpoint queue heads are arranged in an array in a continuous area of
//   memory pointed to by the USB.ENDPOINTLISTADDR pointer. The even-numbered
//   device queue heads in the list support receive endpoints (OUT/SETUP) and
//   the odd-numbered queue heads in the list are used for transmit endpoints
//   (IN/INTERRUPT). The device controller will index into this array based upon
//   the endpoint number received from the USB bus. All information necessary to
//   respond to transactions for all primed transfers is contained in this list
//   so the Device Controller can readily respond to incoming requests without
//   having to traverse a linked list.
//go:align 4096
var descCDCACM0QH [descCDCACMQHCount]dhwEndpoint

// descCDCACM0CD is the transfer descriptor for messages transmitted or received
// on the status/control endpoint 0 for the default CDC-ACM (single) device
// class configuration (index 1).
//go:align 32
var descCDCACM0CD dhwTransfer

// descCDCACM0Cx is the buffer for control/status data received on endpoint 0 of
// the default CDC-ACM (single) device class configuration (index 1).
//go:align 32
var descCDCACM0Cx [descCDCACMCxCount]uint8

// descCDCACM0AD is the transfer descriptor for ackowledgement (ACK) messages
// transmitted or received on the status/control endpoint 0 for the default
// CDC-ACM (single) device class configuration (index 1).
//go:align 32
var descCDCACM0AD dhwTransfer

// descCDCACM0Dx is the transmit (Tx) buffer of descriptor data on endpoint 0
// for the default CDC-ACM (single) device class configuration (index 1).
//go:align 32
var descCDCACM0Dx [descCDCACMConfigSize]uint8

// descCDCACM0RD is an array of transfer descriptors for Rx (OUT) transfers,
// which describe to the device controller the location and quantity of data
// being received for a given transfer, for the default CDC-ACM (single) device
// class configuration (index 1).
//go:align 32
var descCDCACM0RD [descCDCACMRDCount]dhwTransfer

// descCDCACM0Rx is the receive (Rx) transfer buffer for the default CDC-ACM
// (single) device class configuration (index 1).
//go:align 32
var descCDCACM0Rx [descCDCACMRxCount]uint8

// descCDCACM0TD is an array of transfer descriptors for Tx (IN) transfers,
// which describe to the device controller the location and quantity of data
// being transmitted for a given transfer, for the default CDC-ACM (single)
// device class configuration (index 1).
//go:align 32
var descCDCACM0TD [descCDCACMTDCount]dhwTransfer

// descCDCACM0Tx is the transmit (Tx) transfer buffer for the default CDC-ACM
// (single) device class configuration (index 1).
//go:align 32
var descCDCACM0Tx [descCDCACMTxCount]uint8

var descCDCACM0RDNum [descCDCACMRDCount]uint16
var descCDCACM0RDIdx [descCDCACMRDCount]uint16
var descCDCACM0RDQue [(descCDCACMRDCount + 1)]uint16

// descCDCACMClassData holds the buffers and control states for all CDC-ACM
// (single) device class configurations, ordered by index (offset by -1), for
// iMXRT1062 targets only.
//
// Instances of this type (elements of descCDCACMData) are embedded in elements
// of the common/target-agnostic CDC-ACM class configurations (descCDCACM).
// Methods defined on this type implement target-specific functionality, and
// some of these methods are required by the common device controller driver.
// Thus, this type functions as a hardware abstraction layer (HAL).
type descCDCACMClassData struct {

	// CDC-ACM Control Buffers

	qh *[descCDCACMQHCount]dhwEndpoint // endpoint queue heads

	cd *dhwTransfer                 // control endpoint 0 Rx/Tx transfer descriptor
	cx *[descCDCACMCxCount]uint8    // control endpoint 0 Rx/Tx transfer buffer
	ad *dhwTransfer                 // control endpoint 0 Rx/Tx ACK transfer descriptor
	dx *[descCDCACMConfigSize]uint8 // control endpoint 0 Tx (IN) descriptor transfer buffer

	// CDC-ACM Data Buffers

	rd *[descCDCACMRDCount]dhwTransfer // bulk data endpoint Rx (OUT) transfer descriptors
	rx *[descCDCACMRxCount]uint8       // bulk data endpoint Rx (OUT) transfer buffer
	td *[descCDCACMTDCount]dhwTransfer // bulk data endpoint Tx (IN) transfer descriptors
	tx *[descCDCACMTxCount]uint8       // bulk data endpoint Tx (IN) transfer buffer

	rxCount *[descCDCACMRDCount]uint16
	rxIndex *[descCDCACMRDCount]uint16
	rxQueue *[(descCDCACMRDCount + 1)]uint16

	sxSize uint16
	rxSize uint16
	txSize uint16

	txHead uint8
	txFree uint16
	txPrev bool

	rxHead uint8
	rxTail uint8
	rxFree uint16
}

// descCDCACMData holds statically-allocated instances for each of the target-
// specific (iMXRT1062) CDC-ACM (single) device class configurations' control
// and data structures, ordered by configuration index (offset by -1). Each
// element is embedded in a corresponding element of descCDCACM.
//go:align 64
var descCDCACMData = [dcdCount]descCDCACMClassData{

	{ // -- CDC-ACM (single) Class Configuration Index 1 --

		// CDC-ACM Control Buffers

		qh: &descCDCACM0QH,

		cd: &descCDCACM0CD,
		cx: &descCDCACM0Cx,
		ad: &descCDCACM0AD,
		dx: &descCDCACM0Dx,

		// CDC-ACM Data Buffers

		rd: &descCDCACM0RD,
		rx: &descCDCACM0Rx,
		td: &descCDCACM0TD,
		tx: &descCDCACM0Tx,

		rxCount: &descCDCACM0RDNum,
		rxIndex: &descCDCACM0RDIdx,
		rxQueue: &descCDCACM0RDQue,

		sxSize: descCDCACMStatusPacketSize,
		rxSize: descCDCACMDataRxPacketSize,
		txSize: descCDCACMDataTxPacketSize,
	},
}

// descHID0QH is an array of endpoint queue heads, which is where all transfers
// for a given endpoint are managed, for the default HID device class
// configuration (index 1).
//
// From the iMXRT1062 Reference Manual:
//
//   Software must ensure that no interface data structure reachable
//   by the Device Controller spans a 4K-page boundary.
//
//   The [queue head] is a 48-byte data structure, but must be aligned on
//   64-byte boundaries.
//
//   Endpoint queue heads are arranged in an array in a continuous area of
//   memory pointed to by the USB.ENDPOINTLISTADDR pointer. The even-numbered
//   device queue heads in the list support receive endpoints (OUT/SETUP) and
//   the odd-numbered queue heads in the list are used for transmit endpoints
//   (IN/INTERRUPT). The device controller will index into this array based upon
//   the endpoint number received from the USB bus. All information necessary to
//   respond to transactions for all primed transfers is contained in this list
//   so the Device Controller can readily respond to incoming requests without
//   having to traverse a linked list.
//go:align 4096
var descHID0QH [descHIDQHCount]dhwEndpoint

// descHID0CD is the transfer descriptor for messages transmitted or received on
// the status/control endpoint 0 for the default HID device class configuration
// (index 1).
//go:align 32
var descHID0CD dhwTransfer

// descHID0Cx is the buffer for control/status data received on endpoint 0 of
// the default HID device class configuration (index 1).
//go:align 32
var descHID0Cx [descHIDCxCount]uint8

// descHID0AD is the transfer descriptor for ackowledgement (ACK) messages
// transmitted or received on the status/control endpoint 0 for the default HID
// device class configuration (index 1).
//go:align 32
var descHID0AD dhwTransfer

// descHID0Dx is the transmit (Tx) buffer of descriptor data on endpoint 0 for
// the default HID device class configuration (index 1).
//go:align 32
var descHID0Dx [descHIDConfigSize]uint8

// descHID0SerialRD is an array of transfer descriptors for serial Rx (OUT)
// transfers, which describe to the device controller the location and quantity
// of data being received for a given transfer, for the default HID device class
// configuration (index 1).
//go:align 32
var descHID0SerialRD [descHIDSerialRDCount]dhwTransfer

// descHID0SerialRx is the serial receive (Rx) transfer buffer for the default
// HID device class configuration (index 1).
//go:align 32
var descHID0SerialRx [descHIDSerialRxCount]uint8

// descHID0SerialTD is an array of transfer descriptors for serial Tx (IN)
// transfers, which describe to the device controller the location and quantity
// of data being transmitted for a given transfer, for the default HID device
// class configuration (index 1).
//go:align 32
var descHID0SerialTD [descHIDSerialTDCount]dhwTransfer

// descHID0SerialTx is the serial transmit (Tx) transfer buffer for the default
// HID device class configuration (index 1).
//go:align 32
var descHID0SerialTx [descHIDSerialTxCount]uint8

var descHID0SerialRDIdx [descHIDSerialRDCount]uint16
var descHID0SerialRDQue [(descHIDSerialRDCount + 1)]uint16

// descHID0KeyboardTD is an array of transfer descriptors for keyboard Tx (IN)
// transfers, which describe to the device controller the location and quantity
// of data being transmitted for a given transfer, for the default HID device
// class configuration (index 1).
//go:align 32
var descHID0KeyboardTD [descHIDKeyboardTDCount]dhwTransfer

// descHID0KeyboardTx is the keyboard transmit (Tx) transfer buffer for the
// default HID device class configuration (index 1).
//go:align 32
var descHID0KeyboardTx [descHIDKeyboardTxCount]uint8

// descHID0KeyboardTp is the keyboard HID report transmit (Tx) transfer buffer
// for the default HID device class configuration (index 1).
//go:align 32
var descHID0KeyboardTp [descHIDKeyboardTxPacketSize]uint8

//go:align 32
var descHID0KeyboardTxKey [hidKeyboardKeyCount]uint8

//go:align 32
var descHID0KeyboardTxCon [hidKeyboardConCount]uint16

//go:align 32
var descHID0KeyboardTxSys [hidKeyboardSysCount]uint8

// descHID0MouseTD is an array of transfer descriptors for mouse Tx (IN)
// transfers, which describe to the device controller the location and quantity
// of data being transmitted for a given transfer, for the default HID device
// class configuration (index 1).
//go:align 32
var descHID0MouseTD [descHIDMouseTDCount]dhwTransfer

// descHID0MouseTx is the mouse transmit (Tx) transfer buffer for the default
// HID device class configuration (index 1).
//go:align 32
var descHID0MouseTx [descHIDMouseTxCount]uint8

// descHID0JoystickTD is an array of transfer descriptors for joystick Tx (IN)
// transfers, which describe to the device controller the location and quantity
// of data being transmitted for a given transfer, for the default HID device
// class configuration (index 1).
//go:align 32
var descHID0JoystickTD [descHIDJoystickTDCount]dhwTransfer

// descHID0JoystickTx is the joystick transmit (Tx) transfer buffer for the
// default HID device class configuration (index 1).
//go:align 32
var descHID0JoystickTx [descHIDJoystickTxCount]uint8

// descHID0Keyboard is the Keyboard instance with which the user may interact
// when using the default HID device class configuration (index 1).
//go:align 64
var descHID0Keyboard = Keyboard{
	key: &descHID0KeyboardTxKey,
	con: &descHID0KeyboardTxCon,
	sys: &descHID0KeyboardTxSys,
}

// descHIDClassData holds the buffers and control states for all of the HID
// device class configurations, ordered by index (offset by -1), for iMXRT1062
// targets only.
//
// Instances of this type (elements of descHIDData) are embedded in elements
// of the common/target-agnostic HID class configurations (descHID).
// Methods defined on this type implement target-specific functionality, and
// some of these methods are required by the common device controller driver.
// Thus, this type functions as a hardware abstraction layer (HAL).
type descHIDClassData struct {

	// HID Control Buffers

	qh *[descHIDQHCount]dhwEndpoint // endpoint queue heads

	cd *dhwTransfer              // control endpoint 0 Rx/Tx transfer descriptor
	cx *[descHIDCxCount]uint8    // control endpoint 0 Rx/Tx transfer buffer
	ad *dhwTransfer              // control endpoint 0 Rx/Tx ACK transfer descriptor
	dx *[descHIDConfigSize]uint8 // control endpoint 0 Tx (IN) descriptor transfer buffer

	// HID Serial Buffers

	rdSerial *[descHIDSerialRDCount]dhwTransfer // interrupt endpoint serial Rx (OUT) transfer descriptors
	rxSerial *[descHIDSerialRxCount]uint8       // interrupt endpoint serial Rx (OUT) transfer buffer
	tdSerial *[descHIDSerialTDCount]dhwTransfer // interrupt endpoint serial Tx (IN) transfer descriptors
	txSerial *[descHIDSerialTxCount]uint8       // interrupt endpoint serial Tx (IN) transfer buffer

	rxSerialIndex *[descHIDSerialRDCount]uint16
	rxSerialQueue *[(descHIDSerialRDCount + 1)]uint16

	rxSerialSize uint16
	txSerialSize uint16

	txSerialHead uint8
	txSerialFree uint16
	txSerialPrev bool

	rxSerialHead uint8
	rxSerialTail uint8
	rxSerialFree uint16

	// HID Keyboard Buffers

	tdKeyboard *[descHIDKeyboardTDCount]dhwTransfer // interrupt endpoint keyboard Tx (IN) transfer descriptors
	txKeyboard *[descHIDKeyboardTxCount]uint8       // interrupt endpoint keyboard Tx (IN) transfer buffer
	tpKeyboard *[descHIDKeyboardTxPacketSize]uint8  // interrupt endpoint keyboard Tx (IN) HID report bbuffer

	txKeyboardSize uint16

	txKeyboardHead uint8
	txKeyboardPrev bool

	// HID Mouse Buffers

	tdMouse *[descHIDMouseTDCount]dhwTransfer // interrupt endpoint mouse Tx (IN) transfer descriptors
	txMouse *[descHIDMouseTxCount]uint8       // interrupt endpoint mouse Tx (IN) transfer buffer

	txMouseSize uint16

	txMouseHead uint8
	txMousePrev bool

	// HID Joystick Buffers

	tdJoystick *[descHIDJoystickTDCount]dhwTransfer // interrupt endpoint joystick Tx (IN) transfer descriptors
	txJoystick *[descHIDJoystickTxCount]uint8       // interrupt endpoint joystick Tx (IN) transfer buffer

	txJoystickSize uint16

	txJoystickHead uint8
	txJoystickPrev bool

	// HID Device Instances

	keyboard *Keyboard
}

// descHIDData holds statically-allocated instances for each of the target-
// specific (iMXRT1062) HID device class configurations' control and data
// structures, ordered by configuration index (offset by -1). Each element is
// embedded in a corresponding element of descHID.
//go:align 64
var descHIDData = [dcdCount]descHIDClassData{

	{ // -- HID Class Configuration Index 1 --

		// HID Control Buffers

		qh: &descHID0QH,

		cd: &descHID0CD,
		cx: &descHID0Cx,
		ad: &descHID0AD,
		dx: &descHID0Dx,

		// HID Serial Buffers

		rdSerial: &descHID0SerialRD,
		rxSerial: &descHID0SerialRx,
		tdSerial: &descHID0SerialTD,
		txSerial: &descHID0SerialTx,

		rxSerialIndex: &descHID0SerialRDIdx,
		rxSerialQueue: &descHID0SerialRDQue,

		rxSerialSize: descHIDSerialRxPacketSize,
		txSerialSize: descHIDSerialTxPacketSize,

		// HID Keyboard Buffers

		tdKeyboard: &descHID0KeyboardTD,
		txKeyboard: &descHID0KeyboardTx,
		tpKeyboard: &descHID0KeyboardTp,

		txKeyboardSize: descHIDKeyboardTxPacketSize,

		// HID Mouse Buffers

		tdMouse: &descHID0MouseTD,
		txMouse: &descHID0MouseTx,

		txMouseSize: descHIDMouseTxPacketSize,

		// HID Joystick Buffers

		tdJoystick: &descHID0JoystickTD,
		txJoystick: &descHID0JoystickTx,

		txJoystickSize: descHIDJoystickTxPacketSize,

		// HID Device Instances

		keyboard: &descHID0Keyboard,
	},
}
