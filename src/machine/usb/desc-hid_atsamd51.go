//go:build usb.hid && (atsamd51 || atsame5x)
// +build usb.hid
// +build atsamd51 atsame5x

package usb

// descHIDCount defines the number of USB cores that may be configured as a
// composite (keyboard + mouse + joystick) human interface device (HID).
const descHIDCount = 1

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
	descHIDSxSize = 8 + 2
	descHIDCxSize = descControlPacketSize

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

// descHID0ED is an array of endpoint descriptors, which describes to the USB
// DMA controller the buffer and transfer properties for each endpoint, for the
// default HID device class configuration (index 1).
//go:align 32
var descHID0ED [descHIDEDCount]dhwEPAddrDesc

// descHID0Sx is the receive (Rx) buffer for setup packets on control endpoint 0
// of the default HID device class configuration (index 1).
//go:align 32
var descHID0Sx [descHIDSxSize]uint8

// descHID0Cx is the transmit (Tx) buffer for control/status packets on control
// endpoint 0 of the default CDC-ACM (single) device class configuration (index 1).
//go:align 32
var descHID0Cx [descHIDCxSize]uint8

// descHID0Dx is the transmit (Tx) buffer of descriptor data on endpoint 0 for
// the default HID device class configuration (index 1).
//go:align 32
var descHID0Dx [descHIDConfigSize]uint8

// descHID0SerialRx is the serial receive (Rx) transfer buffer for the default
// HID device class configuration (index 1).
//go:align 32
// var descHID0SerialRx [descHIDSerialRxSize]uint8

// descHID0SerialTx is the serial transmit (Tx) transfer buffer for the default
// HID device class configuration (index 1).
//go:align 32
// var descHID0SerialTx [descHIDSerialTxSize]uint8

// descHID0KeyboardTx is the keyboard HID report transmit (Tx) transfer buffer
// for the default HID device class configuration (index 1).
//go:align 32
var descHID0KeyboardTx [descHIDKeyboardTxPacketSize]uint8

// descHID0KeyboardTq is the keyboard transmit (Tx) transfer buffer for the
// default HID device class configuration (index 1).
//go:align 32
// var descHID0KeyboardTq [descHIDKeyboardTxSize]uint8

// descHID0MouseTx is the mouse transmit (Tx) transfer buffer for the default
// HID device class configuration (index 1).
//go:align 32
// var descHID0MouseTx [descHIDMouseTxSize]uint8

// descHID0JoystickTx is the joystick transmit (Tx) transfer buffer for the
// default HID device class configuration (index 1).
//go:align 32
// var descHID0JoystickTx [descHIDJoystickTxSize]uint8

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

	ed *[descHIDEDCount]dhwEPAddrDesc // endpoint descriptors

	sx *[descHIDSxSize]uint8     // control endpoint 0 Rx (OUT) setup packets
	cx *[descHIDCxSize]uint8     // control endpoint 0 Tx (IN) control/status packets
	dx *[descHIDConfigSize]uint8 // control endpoint 0 Tx (IN) descriptor transfer buffer

	// HID Serial Buffers

	// rxSerial *[descHIDSerialRxSize]uint8 // interrupt endpoint serial Rx (OUT) transfer buffer
	// txSerial *[descHIDSerialTxSize]uint8 // interrupt endpoint serial Tx (IN) transfer buffer

	// rxSerialSize uint16
	// txSerialSize uint16

	// HID Keyboard Buffers

	txKeyboard *[descHIDKeyboardTxPacketSize]uint8 // interrupt endpoint keyboard Tx (IN) HID report buffer
	// txqKeyboard *[descHIDKeyboardTxSize]uint8       // interrupt endpoint keyboard Tx (IN) transfer FIFO
	// tqKeyboard  *Queue

	txKeyboardSize uint16

	// HID Mouse Buffers

	// txMouse *[descHIDMouseTxSize]uint8 // interrupt endpoint mouse Tx (IN) transfer buffer

	// txMouseSize uint16

	// HID Joystick Buffers

	// txJoystick *[descHIDJoystickTxSize]uint8 // interrupt endpoint joystick Tx (IN) transfer buffer

	// txJoystickSize uint16

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

		sx: &descHID0Sx,
		cx: &descHID0Cx,
		dx: &descHID0Dx,

		// HID Serial Buffers

		// rxSerial: &descHID0SerialRx,
		// txSerial: &descHID0SerialTx,

		// rxSerialSize: descHIDSerialRxPacketSize,
		// txSerialSize: descHIDSerialTxPacketSize,

		// HID Keyboard Buffers

		txKeyboard: &descHID0KeyboardTx,
		// txqKeyboard: &descHID0KeyboardTq,
		// tqKeyboard:  &Queue{},

		txKeyboardSize: descHIDKeyboardTxPacketSize,

		// HID Mouse Buffers

		// txMouse: &descHID0MouseTx,

		// txMouseSize: descHIDMouseTxPacketSize,

		// HID Joystick Buffers

		// txJoystick: &descHID0JoystickTx,

		// txJoystickSize: descHIDJoystickTxPacketSize,

		// HID Device Instances

		//serial: &descHID0Serial,
		keyboard: &descHID0Keyboard,
		//mouse: &descHID0Mouse,
		//joystick: &descHID0Joystick,
	},
}
