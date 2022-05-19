//go:build usb.hid
// +build usb.hid

package usb

// descCDCCount defines the number of USB cores that may be configured as
// CDC-ACM (single) devices.
const descCDCCount = 0

// USB HID constants defined per specification
const (
	// HID class
	descHIDType = 0x03

	// HID subclass
	descHIDSubNone = 0x00
	descHIDSubBoot = 0x01

	// HID protocol
	descHIDProtoNone     = 0x00
	descHIDProtoKeyboard = 0x01
	descHIDProtoMouse    = 0x02

	descHIDRequestGetReport            = 0x01 // HID request GET_REPORT
	descHIDRequestGetReportTypeInput   = 0x01 // HID request GET_REPORT type INPUT
	descHIDRequestGetReportTypeOutput  = 0x02 // HID request GET_REPORT type OUTPUT
	descHIDRequestGetReportTypeFeature = 0x03 // HID request GET_REPORT type FEATURE
	descHIDRequestGetIdle              = 0x02 // HID request GET_IDLE
	descHIDRequestGetProtocol          = 0x03 // HID request GET_PROTOCOL
	descHIDRequestSetReport            = 0x09 // HID request SET_REPORT
	descHIDRequestSetIdle              = 0x0A // HID request SET_IDLE
	descHIDRequestSetProtocol          = 0x0B // HID request SET_PROTOCOL
)

const (
	// Size of all HID configuration descriptors.
	descHIDConfigSize = uint16(
		descLengthConfigure + // Configuration Header
			descLengthInterface + // Keyboard Interface Descriptor
			descLengthInterface + // Keyboard HID Interface Descriptor
			descLengthEndpoint + // Keyboard Endpoint Descriptor
			64 +
			0)

	// Position of each HID interface descriptor as offsets into the configuration
	// descriptor. See comments in the configuration descriptor definition for the
	// incremental tally that computes these.
	descHIDConfigKeyboardPos = 18
	descHIDConfigMousePos    = 43
	descHIDConfigSerialPos   = 68
	descHIDConfigJoystickPos = 100
	descHIDConfigMediaKeyPos = 125
)

// Common configuration constants for the USB HID device class.
const (
	descHIDLanguageCount = 1 // String descriptor languages available

	descHIDInterfaceCount = 5 // Interfaces for all HID configurations.
	descHIDEndpointCount  = 6 // Endpoints for all HID configurations.

	descHIDEndpointCtrl = 0 // HID Control Endpoint 0

	descHIDInterfaceKeyboard  = 0 // HID Keyboard Interface
	descHIDEndpointKeyboard   = 3 // HID Keyboard IN Endpoint
	descHIDConfigAttrKeyboard = descEndptConfigAttrTxInterrupt | descEndptConfigAttrRxUnused

	descHIDInterfaceMouse  = 1 // HID Mouse Interface
	descHIDEndpointMouse   = 5 // HID Mouse IN Endpoint
	descHIDConfigAttrMouse = descEndptConfigAttrTxInterrupt | descEndptConfigAttrRxUnused

	descHIDInterfaceSerial  = 2 // HID Serial (UART emulation) Interface
	descHIDEndpointSerialRx = 2 // HID Serial OUT (Rx) Endpoint
	descHIDEndpointSerialTx = 2 // HID Serial IN (Tx) Endpoint
	descHIDConfigAttrSerial = descEndptConfigAttrTxInterrupt | descEndptConfigAttrRxInterrupt

	descHIDInterfaceJoystick  = 3 // HID Joystick Interface
	descHIDEndpointJoystick   = 6 // HID Joystick IN Endpoint
	descHIDConfigAttrJoystick = descEndptConfigAttrTxInterrupt | descEndptConfigAttrRxUnused

	descHIDInterfaceMediaKey  = 4 // HID Keyboard Media Keys Interface
	descHIDEndpointMediaKey   = 4 // HID Keyboard Media Keys IN Endpoint
	descHIDConfigAttrMediaKey = descEndptConfigAttrTxInterrupt | descEndptConfigAttrRxUnused
)

// descHIDClass holds references to all descriptors, buffers, and control
// structures for the USB HID device class.
type descHIDClass struct {
	*descHIDClassData // Target-defined, class-specific data

	locale *[descHIDLanguageCount]descStringLanguage // string descriptors
	device *[descLengthDevice]uint8                  // device descriptor
	qualif *[descLengthQualification]uint8           // device qualification descriptor
	config *[descHIDConfigSize]uint8                 // configuration descriptor
}

// descHID holds statically-allocated instances for each of the HID device class
// configurations, ordered by index (offset by -1).
var descHID = [dcdCount]descHIDClass{

	{ // HID class configuration index 1
		descHIDClassData: &descHIDData[0],

		locale: &[descHIDLanguageCount]descStringLanguage{

			{ // [0x0409] US English
				language: descLanguageEnglish,
				descriptor: descStringIndex{
					{ /* [0] Language */
						4,
						descTypeString,
						lsU8(descLanguageEnglish),
						msU8(descLanguageEnglish),
					},
					// Actual string descriptors (index > 0) are copied into here at runtime!
					// This allows for application- or even user-defined string descriptors.
					{ /* [1] Manufacturer */ },
					{ /* [2] Product */ },
					{ /* [3] Serial Number */ },
				},
			},
		},
		device: &[descLengthDevice]uint8{
			descLengthDevice,          // Size of this descriptor in bytes
			descTypeDevice,            // Descriptor Type
			lsU8(descUSBSpecVersion),  // USB Specification Release Number in BCD (low)
			msU8(descUSBSpecVersion),  // USB Specification Release Number in BCD (high)
			0,                         // Class code (assigned by the USB-IF).
			0,                         // Subclass code (assigned by the USB-IF).
			0,                         // Protocol code (assigned by the USB-IF).
			descEndptMaxPktSize,       // Maximum packet size for endpoint zero (8, 16, 32, or 64)
			lsU8(descCommonVendorID),  // Vendor ID (low) (assigned by the USB-IF)
			msU8(descCommonVendorID),  // Vendor ID (high) (assigned by the USB-IF)
			lsU8(descCommonProductID), // Product ID (low) (assigned by the manufacturer)
			msU8(descCommonProductID), // Product ID (high) (assigned by the manufacturer)
			lsU8(descCommonReleaseID), // Device release number in BCD (low)
			msU8(descCommonReleaseID), // Device release number in BCD (high)
			1,                         // Index of string descriptor describing manufacturer
			2,                         // Index of string descriptor describing product
			3,                         // Index of string descriptor describing the device's serial number
			descHIDCount,              // Number of possible configurations
		},
		qualif: &[descLengthQualification]uint8{
			descLengthQualification,  // Size of this descriptor in bytes
			descTypeQualification,    // Descriptor Type
			lsU8(descUSBSpecVersion), // USB Specification Release Number in BCD (low)
			msU8(descUSBSpecVersion), // USB Specification Release Number in BCD (high)
			0,                        // Class code (assigned by the USB-IF).
			0,                        // Subclass code (assigned by the USB-IF).
			0,                        // Protocol code (assigned by the USB-IF).
			descEndptMaxPktSize,      // Maximum packet size for endpoint zero (8, 16, 32, or 64)
			descHIDCount,             // Number of possible configurations
			0,                        // Reserved
		},
		config: &[descHIDConfigSize]uint8{
			// [0+9]
			descLengthConfigure,     // Size of this descriptor in bytes
			descTypeConfigure,       // Descriptor Type
			lsU8(descHIDConfigSize), // Total length of data returned for this configuration (low)
			msU8(descHIDConfigSize), // Total length of data returned for this configuration (high)
			1,                       // Number of interfaces supported by this configuration
			1,                       // Value to use to select this configuration (1 = CDC-ACM[0])
			0,                       // Index of string descriptor describing this configuration
			descEndptConfigAttr,     // Configuration attributes
			descHIDMaxPowerMa >> 1,  // Max power consumption when fully-operational (2 mA units)

			// [9+9] Keyboard Interface Descriptor
			descLengthInterface,      // Descriptor length
			descTypeInterface,        // Descriptor type
			descHIDInterfaceKeyboard, // Interface index
			0,                        // Alternate setting
			1,                        // Number of endpoints
			descHIDType,              // Class code (HID = 0x03)
			descHIDSubBoot,           // Subclass code (Boot = 0x01)
			descHIDProtoKeyboard,     // Protocol code (Keyboard = 0x01)
			0,                        // Interface Description String Index

			// [18+9] Keyboard HID Interface Descriptor
			descLengthInterface,                      // Descriptor length
			descTypeHID,                              // Descriptor type
			0x11,                                     // HID BCD (low)
			0x01,                                     // HID BCD (high)
			0,                                        // Country code
			1,                                        // Number of descriptors
			descTypeHIDReport,                        // Descriptor type
			lsU8(uint16(len(descHIDReportKeyboard))), // Descriptor length (low)
			msU8(uint16(len(descHIDReportKeyboard))), // Descriptor length (high)

			// [27+7] Keyboard Endpoint Descriptor
			descLengthEndpoint, // Size of this descriptor in bytes
			descTypeEndpoint,   // Descriptor Type
			descHIDEndpointKeyboard | // Endpoint address
				descEndptAddrDirectionIn,
			descEndptTypeInterrupt,            // Attributes
			lsU8(descHIDKeyboardTxPacketSize), // Max packet size (low)
			msU8(descHIDKeyboardTxPacketSize), // Max packet size (high)
			descHIDKeyboardTxInterval,         // Polling Interval
		},
	},
}

var descHIDReportSerial = [...]uint8{
	0x06, 0xC9, 0xFF, // Usage Page 0xFFC9 (vendor defined)
	0x09, 0x04, // Usage 0x04
	0xA1, 0x5C, // Collection 0x5C
	0x75, 0x08, //   report size = 8 bits (global)
	0x15, 0x00, //   logical minimum = 0 (global)
	0x26, 0xFF, 0x00, //   logical maximum = 255 (global)
	0x95, descHIDSerialTxPacketSize, //   report count (global)
	0x09, 0x75, //   usage (local)
	0x81, 0x02, //   Input
	0x95, descHIDSerialRxPacketSize, //   report count (global)
	0x09, 0x76, //   usage (local)
	0x91, 0x02, //   Output
	0x95, 0x04, //   report count (global)
	0x09, 0x76, //   usage (local)
	0xB1, 0x02, //   Feature
	0xC0, // end collection
}

var descHIDReportKeyboard = [...]uint8{
	0x05, 0x01, // Usage Page (Generic Desktop)
	0x09, 0x06, // Usage (Keyboard)
	0xA1, 0x01, // Collection (Application)
	0x75, 0x01, //   Report Size (1)
	0x95, 0x08, //   Report Count (8)
	0x05, 0x07, //   Usage Page (Key Codes)
	0x19, 0xE0, //   Usage Minimum (224)
	0x29, 0xE7, //   Usage Maximum (231)
	0x15, 0x00, //   Logical Minimum (0)
	0x25, 0x01, //   Logical Maximum (1)
	0x81, 0x02, //   Input (Data, Variable, Absolute) [Modifier keys]
	0x95, 0x01, //   Report Count (1)
	0x75, 0x08, //   Report Size (8)
	0x81, 0x03, //   Input (Constant) [Reserved byte]
	0x95, 0x05, //   Report Count (5)
	0x75, 0x01, //   Report Size (1)
	0x05, 0x08, //   Usage Page (LEDs)
	0x19, 0x01, //   Usage Minimum (1)
	0x29, 0x05, //   Usage Maximum (5)
	0x91, 0x02, //   Output (Data, Variable, Absolute) [LED report]
	0x95, 0x01, //   Report Count (1)
	0x75, 0x03, //   Report Size (3)
	0x91, 0x03, //   Output (Constant) [LED report padding]
	0x95, 0x06, //   Report Count (6)
	0x75, 0x08, //   Report Size (8)
	0x15, 0x00, //   Logical Minimum (0)
	0x25, 0x7F, //   Logical Maximum(104)
	0x05, 0x07, //   Usage Page (Key Codes)
	0x19, 0x00, //   Usage Minimum (0)
	0x29, 0x7F, //   Usage Maximum (104)
	0x81, 0x00, //   Input (Data, Array) [Normal keys]
	0xC0, // End Collection
}

var descHIDReportMediaKey = [...]uint8{
	0x05, 0x0C, // Usage Page (Consumer)
	0x09, 0x01, // Usage (Consumer Controls)
	0xA1, 0x01, // Collection (Application)
	0x75, 0x0A, //   Report Size (10)
	0x95, 0x04, //   Report Count (4)
	0x19, 0x00, //   Usage Minimum (0)
	0x2A, 0x9C, 0x02, //   Usage Maximum (0x29C)
	0x15, 0x00, //   Logical Minimum (0)
	0x26, 0x9C, 0x02, //   Logical Maximum (0x29C)
	0x81, 0x00, //   Input (Data, Array)
	0x05, 0x01, //   Usage Page (Generic Desktop)
	0x75, 0x08, //   Report Size (8)
	0x95, 0x03, //   Report Count (3)
	0x19, 0x00, //   Usage Minimum (0)
	0x29, 0xB7, //   Usage Maximum (0xB7)
	0x15, 0x00, //   Logical Minimum (0)
	0x26, 0xB7, 0x00, //   Logical Maximum (0xB7)
	0x81, 0x00, //   Input (Data, Array)
	0xC0, // End Collection
}

var descHIDReportMouse = [...]uint8{
	0x05, 0x01, // Usage Page (Generic Desktop)
	0x09, 0x02, // Usage (Mouse)
	0xA1, 0x01, // Collection (Application)
	0x85, 0x01, //   REPORT_ID (1)
	0x05, 0x09, //   Usage Page (Button)
	0x19, 0x01, //   Usage Minimum (Button #1)
	0x29, 0x08, //   Usage Maximum (Button #8)
	0x15, 0x00, //   Logical Minimum (0)
	0x25, 0x01, //   Logical Maximum (1)
	0x95, 0x08, //   Report Count (8)
	0x75, 0x01, //   Report Size (1)
	0x81, 0x02, //   Input (Data, Variable, Absolute)
	0x05, 0x01, //   Usage Page (Generic Desktop)
	0x09, 0x30, //   Usage (X)
	0x09, 0x31, //   Usage (Y)
	0x09, 0x38, //   Usage (Wheel)
	0x15, 0x81, //   Logical Minimum (-127)
	0x25, 0x7F, //   Logical Maximum (127)
	0x75, 0x08, //   Report Size (8),
	0x95, 0x03, //   Report Count (3),
	0x81, 0x06, //   Input (Data, Variable, Relative)
	0x05, 0x0C, //   Usage Page (Consumer)
	0x0A, 0x38, 0x02, //   Usage (AC Pan)
	0x15, 0x81, //   Logical Minimum (-127)
	0x25, 0x7F, //   Logical Maximum (127)
	0x75, 0x08, //   Report Size (8),
	0x95, 0x01, //   Report Count (1),
	0x81, 0x06, //   Input (Data, Variable, Relative)
	0xC0,       // End Collection
	0x05, 0x01, // Usage Page (Generic Desktop)
	0x09, 0x02, // Usage (Mouse)
	0xA1, 0x01, // Collection (Application)
	0x85, 0x02, //   REPORT_ID (2)
	0x05, 0x01, //   Usage Page (Generic Desktop)
	0x09, 0x30, //   Usage (X)
	0x09, 0x31, //   Usage (Y)
	0x15, 0x00, //   Logical Minimum (0)
	0x26, 0xFF, 0x7F, //   Logical Maximum (32767)
	0x75, 0x10, //   Report Size (16),
	0x95, 0x02, //   Report Count (2),
	0x81, 0x02, //   Input (Data, Variable, Absolute)
	0xC0, // End Collection
}

var descHIDReportJoystick = [...]uint8{
	0x05, 0x01, // Usage Page (Generic Desktop)
	0x09, 0x04, // Usage (Joystick)
	0xA1, 0x01, // Collection (Application)
	0x15, 0x00, //   Logical Minimum (0)
	0x25, 0x01, //   Logical Maximum (1)
	0x75, 0x01, //   Report Size (1)
	0x95, 0x20, //   Report Count (32)
	0x05, 0x09, //   Usage Page (Button)
	0x19, 0x01, //   Usage Minimum (Button #1)
	0x29, 0x20, //   Usage Maximum (Button #32)
	0x81, 0x02, //   Input (variable,absolute)
	0x15, 0x00, //   Logical Minimum (0)
	0x25, 0x07, //   Logical Maximum (7)
	0x35, 0x00, //   Physical Minimum (0)
	0x46, 0x3B, 0x01, //   Physical Maximum (315)
	0x75, 0x04, //   Report Size (4)
	0x95, 0x01, //   Report Count (1)
	0x65, 0x14, //   Unit (20)
	0x05, 0x01, //   Usage Page (Generic Desktop)
	0x09, 0x39, //   Usage (Hat switch)
	0x81, 0x42, //   Input (variable,absolute,null_state)
	0x05, 0x01, //   Usage Page (Generic Desktop)
	0x09, 0x01, //   Usage (Pointer)
	0xA1, 0x00, //   Collection ()
	0x15, 0x00, //     Logical Minimum (0)
	0x26, 0xFF, 0x03, //     Logical Maximum (1023)
	0x75, 0x0A, //     Report Size (10)
	0x95, 0x04, //     Report Count (4)
	0x09, 0x30, //     Usage (X)
	0x09, 0x31, //     Usage (Y)
	0x09, 0x32, //     Usage (Z)
	0x09, 0x35, //     Usage (Rz)
	0x81, 0x02, //     Input (variable,absolute)
	0xC0,       //   End Collection
	0x15, 0x00, //   Logical Minimum (0)
	0x26, 0xFF, 0x03, //   Logical Maximum (1023)
	0x75, 0x0A, //   Report Size (10)
	0x95, 0x02, //   Report Count (2)
	0x09, 0x36, //   Usage (Slider)
	0x09, 0x36, //   Usage (Slider)
	0x81, 0x02, //   Input (variable,absolute)
	0xC0, // End Collection
}
