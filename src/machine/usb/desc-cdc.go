//go:build baremetal && usb.cdc
// +build baremetal,usb.cdc

package usb

// USB CDC constants defined per specification.
const (

	// Device class
	descCDCTypeComm = 0x02 // communication/control
	descCDCTypeData = 0x0A // data

	// Communication/control subclass
	descCDCSubNone                      = 0x00
	descCDCSubDirectLineControl         = 0x01
	descCDCSubAbstractControl           = 0x02
	descCDCSubTelephoneControl          = 0x03
	descCDCSubMultiChannelControl       = 0x04
	descCDCSubCAPIControl               = 0x05
	descCDCSubEthernetNetworkingControl = 0x06
	descCDCSubATMNetworkingControl      = 0x07
	descCDCSubWirelessHandsetControl    = 0x08
	descCDCSubDeviceManagement          = 0x09
	descCDCSubMobileDirectLine          = 0x0A
	descCDCSubOBEX                      = 0x0B
	descCDCSubEthernetEmulation         = 0x0C

	// Communication/control protocol
	descCDCProtoNone              = 0x00 // also for data class
	descCDCProtoAT250             = 0x01
	descCDCProtoATPCCA101         = 0x02
	descCDCProtoATPCCA101AnnexO   = 0x03
	descCDCProtoATGSM707          = 0x04
	descCDCProtoAT3GPP27007       = 0x05
	descCDCProtoATTIACDMA         = 0x06
	descCDCProtoEthernetEmulation = 0x07
	descCDCProtoExternal          = 0xFE
	descCDCProtoVendorSpecific    = 0xFF // also for data class

	// Data protocol
	descCDCProtoPyhsicalInterface     = 0x30
	descCDCProtoHDLC                  = 0x31
	descCDCProtoTransparent           = 0x32
	descCDCProtoManagement            = 0x50
	descCDCProtoDataLinkQ931          = 0x51
	descCDCProtoDataLinkQ921          = 0x52
	descCDCProtoDataCompressionV42BIS = 0x90
	descCDCProtoEuroISDN              = 0x91
	descCDCProtoRateAdaptionISDNV24   = 0x92
	descCDCProtoCAPICommands          = 0x93
	descCDCProtoHostBasedDriver       = 0xFD
	descCDCProtoUnitFunctional        = 0xFE

	// Functional descriptor length
	descCDCFuncLengthHeader          = 5
	descCDCFuncLengthCallManagement  = 5
	descCDCFuncLengthAbstractControl = 4
	descCDCFuncLengthUnion           = 5

	// Functional descriptor type
	descCDCFuncTypeHeader             = 0x00
	descCDCFuncTypeCallManagement     = 0x01
	descCDCFuncTypeAbstractControl    = 0x02
	descCDCFuncTypeDirectLine         = 0x03
	descCDCFuncTypeTelephoneRinger    = 0x04
	descCDCFuncTypeTelephoneReport    = 0x05
	descCDCFuncTypeUnion              = 0x06
	descCDCFuncTypeCountrySelect      = 0x07
	descCDCFuncTypeTelephoneModes     = 0x08
	descCDCFuncTypeTerminal           = 0x09
	descCDCFuncTypeNetworkChannel     = 0x0A
	descCDCFuncTypeProtocolUnit       = 0x0B
	descCDCFuncTypeExtensionUnit      = 0x0C
	descCDCFuncTypeMultiChannel       = 0x0D
	descCDCFuncTypeCAPIControl        = 0x0E
	descCDCFuncTypeEthernetNetworking = 0x0F
	descCDCFuncTypeATMNetworking      = 0x10
	descCDCFuncTypeWirelessControl    = 0x11
	descCDCFuncTypeMobileDirectLine   = 0x12
	descCDCFuncTypeMDLMDetail         = 0x13
	descCDCFuncTypeDeviceManagement   = 0x14
	descCDCFuncTypeOBEX               = 0x15
	descCDCFuncTypeCommandSet         = 0x16
	descCDCFuncTypeCommandSetDetail   = 0x17
	descCDCFuncTypeTelephoneControl   = 0x18
	descCDCFuncTypeOBEXServiceID      = 0x19

	// Standard request
	descCDCRequestSendEncapsulatedCommand     = 0x00 // CDC request SEND_ENCAPSULATED_COMMAND
	descCDCRequestGetEncapsulatedResponse     = 0x01 // CDC request GET_ENCAPSULATED_RESPONSE
	descCDCRequestSetCommFeature              = 0x02 // CDC request SET_COMM_FEATURE
	descCDCRequestGetCommFeature              = 0x03 // CDC request GET_COMM_FEATURE
	descCDCRequestClearCommFeature            = 0x04 // CDC request CLEAR_COMM_FEATURE
	descCDCRequestSetAuxLineState             = 0x10 // CDC request SET_AUX_LINE_STATE
	descCDCRequestSetHookState                = 0x11 // CDC request SET_HOOK_STATE
	descCDCRequestPulseSetup                  = 0x12 // CDC request PULSE_SETUP
	descCDCRequestSendPulse                   = 0x13 // CDC request SEND_PULSE
	descCDCRequestSetPulseTime                = 0x14 // CDC request SET_PULSE_TIME
	descCDCRequestRingAuxJack                 = 0x15 // CDC request RING_AUX_JACK
	descCDCRequestSetLineCoding               = 0x20 // CDC request SET_LINE_CODING
	descCDCRequestGetLineCoding               = 0x21 // CDC request GET_LINE_CODING
	descCDCRequestSetControlLineState         = 0x22 // CDC request SET_CONTROL_LINE_STATE
	descCDCRequestSendBreak                   = 0x23 // CDC request SEND_BREAK
	descCDCRequestSetRingerParams             = 0x30 // CDC request SET_RINGER_PARAMS
	descCDCRequestGetRingerParams             = 0x31 // CDC request GET_RINGER_PARAMS
	descCDCRequestSetOperationParam           = 0x32 // CDC request SET_OPERATION_PARAM
	descCDCRequestGetOperationParam           = 0x33 // CDC request GET_OPERATION_PARAM
	descCDCRequestSetLineParams               = 0x34 // CDC request SET_LINE_PARAMS
	descCDCRequestGetLineParams               = 0x35 // CDC request GET_LINE_PARAMS
	descCDCRequestDialDigits                  = 0x36 // CDC request DIAL_DIGITS
	descCDCRequestSetUnitParameter            = 0x37 // CDC request SET_UNIT_PARAMETER
	descCDCRequestGetUnitParameter            = 0x38 // CDC request GET_UNIT_PARAMETER
	descCDCRequestClearUnitParameter          = 0x39 // CDC request CLEAR_UNIT_PARAMETER
	descCDCRequestSetEthernetMulticastFilters = 0x40 // CDC request SET_ETHERNET_MULTICAST_FILTERS
	descCDCRequestSetEthernetPowPatternFilter = 0x41 // CDC request SET_ETHERNET_POW_PATTER_FILTER
	descCDCRequestGetEthernetPowPatternFilter = 0x42 // CDC request GET_ETHERNET_POW_PATTER_FILTER
	descCDCRequestSetEthernetPacketFilter     = 0x43 // CDC request SET_ETHERNET_PACKET_FILTER
	descCDCRequestGetEthernetStatistic        = 0x44 // CDC request GET_ETHERNET_STATISTIC
	descCDCRequestSetATMDataFormat            = 0x50 // CDC request SET_ATM_DATA_FORMAT
	descCDCRequestGetATMDeviceStatistics      = 0x51 // CDC request GET_ATM_DEVICE_STATISTICS
	descCDCRequestSetATMDefaultVC             = 0x52 // CDC request SET_ATM_DEFAULT_VC
	descCDCRequestGetATMVCStatistics          = 0x53 // CDC request GET_ATM_VC_STATISTICS
	descCDCRequestMDLMSpecificRequestsMask    = 0x7F // CDC request MDLM_SPECIFIC_REQUESTS_MASK

	// Notification type
	descCDCNotifyNetworkConnection     = 0x00 // CDC notify NETWORK_CONNECTION
	descCDCNotifyResponseAvail         = 0x01 // CDC notify RESPONSE_AVAIL
	descCDCNotifyAuxJackHookState      = 0x08 // CDC notify AUX_JACK_HOOK_STATE
	descCDCNotifyRingDetect            = 0x09 // CDC notify RING_DETECT
	descCDCNotifySerialState           = 0x20 // CDC notify SERIAL_STATE
	descCDCNotifyCallStateChange       = 0x28 // CDC notify CALL_STATE_CHANGE
	descCDCNotifyLineStateChange       = 0x29 // CDC notify LINE_STATE_CHANGE
	descCDCNotifyConnectionSpeedChange = 0x2A // CDC notify CONNECTION_SPEED_CHANGE

	// Feature select
	descCDCFeatureAbstractState  = 0x01 // CDC feature select ABSTRACT_STATE
	descCDCFeatureCountrySetting = 0x02 // CDC feature select COUNTRY_SETTING

	// Control signal
	descCDCControlSigBitmapCarrierActivation = 0x02 // CDC control signal CARRIER_ACTIVATION
	descCDCControlSigBitmapDTEPresence       = 0x01 // CDC control signal DTE_PRESENCE

	// UART emulated state
	descCDCUARTStateRxCarrier  = 0x01 // UART state RX_CARRIER
	descCDCUARTStateTxCarrier  = 0x02 // UART state TX_CARRIER
	descCDCUARTStateBreak      = 0x04 // UART state BREAK
	descCDCUARTStateRingSignal = 0x08 // UART state RING_SIGNAL
	descCDCUARTStateFraming    = 0x10 // UART state FRAMING
	descCDCUARTStateParity     = 0x20 // UART state PARITY
	descCDCUARTStateOverrun    = 0x40 // UART state OVERRUN
)

const (
	// Size of all CDC-ACM configuration descriptors.
	descCDCConfigSize = uint16(
		descLengthConfigure + // Configuration Header
			descLengthInterface + // CDC Interface Descriptor
			descCDCFuncLengthHeader + // CDC Header
			descCDCFuncLengthCallManagement + // CDC Call Management Func Descriptor
			descCDCFuncLengthAbstractControl + // CDC Abstract Control Func Descriptor
			descCDCFuncLengthUnion + // CDC Union Functional Descriptor
			descLengthEndpoint + // CDC Status IN Endpoint Descriptor
			descLengthInterface + // CDC Data Interface Descriptor
			descLengthEndpoint + // CDC Data IN Endpoint Descriptor
			descLengthEndpoint) // CDC Data OUT Endpoint Descriptor
)

// descCDCLineCodingSize defines the length of a CDC-ACM UART line coding
// buffer. Note that the actual buffer may be padded for alignment; but for
// Rx/Tx transfer purposes, descCDCLineCodingSize defines the number of bytes
// that are transferred following a control SETUP request.
const descCDCLineCodingSize = 7

// descCDCLineCoding represents an emulated UART's line configuration.
//
// Use descCDCLineCodingSize instead of unsafe.Sizeof(descCDCLineCoding)
// in any transfer requests, because the actual struct is padded for alignment.
type descCDCLineCoding struct {
	baud     uint32
	stopBits uint8
	parity   uint8
	numBits  uint8
	_        uint8
}

// parse initializes the receiver descCDCLineCoding from the given []uint8 v.
// Argument v is a Rx transfer buffer, filled following the completion of a
// control transfer from a CDC SET_LINE_CODING (0x20) request
func (s *descCDCLineCoding) parse(v []uint8) bool {
	if len(v) >= descCDCLineCodingSize {
		s.baud = packU32(v[:])
		s.stopBits = v[4]
		s.parity = v[5]
		s.numBits = v[6]
		return true
	}
	return false
}

// descCDCLineState represents an emulated UART's line state.
type descCDCLineState struct {
	// dataTerminalReady indicates if DTE is present or not.
	// Corresponds to V.24 signal 108/2 and RS-232 signal DTR.
	dataTerminalReady bool // DTR
	// requestToSend is the carrier control for half-duplex modems.
	// Corresponds to V.24 signal 105 and RS-232 signal RTS.
	requestToSend bool // RTS
}

// parse initializes the receiver descCDCLineState from the given uint16 v.
// Argument v corresponds to the wValue field in a control SETUP packet, which
// carries the line state from a CDC SET_CONTROL_LINE_STATE (0x22) request.
func (s *descCDCLineState) parse(v uint16) bool {
	s.dataTerminalReady = 0 != v&0x1
	s.requestToSend = 0 != v&0x2
	return true
}

// Common configuration constants for the USB CDC-ACM (single) device class.
const (
	descCDCLanguageCount = 1 // String descriptor languages available

	descCDCInterfaceCount = 2 // Interfaces for all CDC-ACM configurations.
	descCDCEndpointCount  = 4 // Endpoints for all CDC-ACM configurations.

	descCDCEndpointCtrl = 0 // CDC-ACM Control Endpoint 0

	descCDCInterfaceCtrl    = 0 // CDC-ACM Control Interface
	descCDCEndpointStatus   = 1 // CDC-ACM Interrupt IN Endpoint
	descCDCConfigAttrStatus = descEndptConfigAttrRxUnused | descEndptConfigAttrTxInterrupt

	descCDCInterfaceData    = 1 // CDC-ACM Data Interface
	descCDCEndpointDataRx   = 2 // CDC-ACM Bulk Data OUT (Rx) Endpoint
	descCDCConfigAttrDataRx = descEndptConfigAttrRxBulk | descEndptConfigAttrTxUnused
	descCDCEndpointDataTx   = 3 // CDC-ACM Bulk Data IN (Tx) Endpoint
	descCDCConfigAttrDataTx = descEndptConfigAttrRxUnused | descEndptConfigAttrTxBulk
)

// descCDCClass holds references to all descriptors, buffers, and control
// structures for the USB CDC-ACM (single) device class.
type descCDCClass struct {
	*descCDCClassData // Target-defined, class-specific data

	locale *[descCDCLanguageCount]descStringLanguage // string descriptors
	device *[descLengthDevice]uint8                  // device descriptor
	qualif *[descLengthQualification]uint8           // device qualification descriptor
	config *[descCDCConfigSize]uint8                 // configuration descriptor
}

// descCDC holds statically-allocated instances for each of the CDC-ACM
// (single) device class configurations, ordered by index (offset by -1).
var descCDC = [dcdCount]descCDCClass{

	{ // CDC-ACM (single) class configuration index 1
		descCDCClassData: &descCDCData[0],

		locale: &[descCDCLanguageCount]descStringLanguage{

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
			descCDCACMCount,           // Number of possible configurations
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
			descCDCACMCount,          // Number of possible configurations
			0,                        // Reserved
		},
		config: &[descCDCConfigSize]uint8{
			descLengthConfigure,     // Size of this descriptor in bytes
			descTypeConfigure,       // Descriptor Type
			lsU8(descCDCConfigSize), // Total length of data returned for this configuration (low)
			msU8(descCDCConfigSize), // Total length of data returned for this configuration (high)
			descCDCInterfaceCount,   // Number of interfaces supported by this configuration
			1,                       // Value to use to select this configuration (1 = CDC-ACM[0])
			0,                       // Index of string descriptor describing this configuration
			descEndptConfigAttr,     // Configuration attributes
			descCDCMaxPowerMa >> 1,  // Max power consumption when fully-operational (2 mA units)

			// Communication/Control Interface Descriptor
			descLengthInterface,       // Descriptor length
			descTypeInterface,         // Descriptor type
			descCDCInterfaceCtrl,      // Interface index
			0,                         // Alternate setting
			1,                         // Number of endpoints
			descCDCTypeComm,           // Class code
			descCDCSubAbstractControl, // Subclass code
			descCDCProtoAT250,         // Protocol code (NOTE: Teensyduino & Arduino-Mbed define this as 1 [AT V.250])
			0,                         // Interface Description String Index

			// CDC Header Functional Descriptor
			descCDCFuncLengthHeader, // Size of this descriptor in bytes
			descTypeCDCInterface,    // Descriptor Type
			descCDCFuncTypeHeader,   // Descriptor Subtype
			0x10,                    // USB CDC specification version 1.10 (low)
			0x01,                    // USB CDC specification version 1.10 (high)

			// CDC Call Management Functional Descriptor
			descCDCFuncLengthCallManagement, // Size of this descriptor in bytes
			descTypeCDCInterface,            // Descriptor Type
			descCDCFuncTypeCallManagement,   // Descriptor Subtype
			0x01,                            // Capabilities
			descCDCInterfaceData,            // Data Interface

			// CDC Abstract Control Management Functional Descriptor
			descCDCFuncLengthAbstractControl, // Size of this descriptor in bytes
			descTypeCDCInterface,             // Descriptor Type
			descCDCFuncTypeAbstractControl,   // Descriptor Subtype
			0x06,                             // Capabilities

			// CDC Union Functional Descriptor
			descCDCFuncLengthUnion, // Size of this descriptor in bytes
			descTypeCDCInterface,   // Descriptor Type
			descCDCFuncTypeUnion,   // Descriptor Subtype
			descCDCInterfaceCtrl,   // Controlling interface index
			descCDCInterfaceData,   // Controlled interface index

			// Communication/Control Notification Endpoint descriptor
			descLengthEndpoint, // Size of this descriptor in bytes
			descTypeEndpoint,   // Descriptor Type
			descCDCEndpointStatus | // Endpoint address
				descEndptAddrDirectionIn,
			descEndptTypeInterrupt,        // Attributes
			lsU8(descCDCStatusPacketSize), // Max packet size (low)
			msU8(descCDCStatusPacketSize), // Max packet size (high)
			descCDCStatusInterval,         // Polling Interval

			// Data Interface Descriptor
			descLengthInterface,  // Interface length
			descTypeInterface,    // Interface type
			descCDCInterfaceData, // Interface index
			0,                    // Alternate setting
			2,                    // Number of endpoints
			descCDCTypeData,      // Class code
			descCDCSubNone,       // Subclass code
			descCDCProtoNone,     // Protocol code
			0,                    // Interface Description String Index

			// Data Bulk Rx Endpoint descriptor
			descLengthEndpoint, // Size of this descriptor in bytes
			descTypeEndpoint,   // Descriptor Type
			descCDCEndpointDataRx | // Endpoint address
				descEndptAddrDirectionOut,
			descEndptTypeBulk,             // Attributes
			lsU8(descCDCDataRxPacketSize), // Max packet size (low)
			msU8(descCDCDataRxPacketSize), // Max packet size (high)
			0,                             // Polling Interval

			// Data Bulk Tx Endpoint descriptor
			descLengthEndpoint, // Size of this descriptor in bytes
			descTypeEndpoint,   // Descriptor Type
			descCDCEndpointDataTx | // Endpoint address
				descEndptAddrDirectionIn,
			descEndptTypeBulk,             // Attributes
			lsU8(descCDCDataTxPacketSize), // Max packet size (low)
			msU8(descCDCDataTxPacketSize), // Max packet size (high)
			0,                             // Polling Interval
		},
	},
}
