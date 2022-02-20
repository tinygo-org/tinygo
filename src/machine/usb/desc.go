package usb

import "runtime/volatile"

const descUSBSpecVersion = uint16(0x0200) // USB 2.0

const descLanguageEnglish = uint16(0x0409) // (US) English

// USB constants defined per specification.
const (

	// Descriptor length
	descLengthDevice            = 18
	descLengthConfigure         = 9
	descLengthInterface         = 9
	descLengthEndpoint          = 7
	descLengthQualification     = 10
	descLengthOTG               = 5
	descLengthBOS               = 5
	descLengthEndpointCompanion = 6
	descLengthUSB20Extension    = 7
	descLengthSuperspeed        = 10

	// Descriptor type
	descTypeDevice                  = 0x01
	descTypeConfigure               = 0x02
	descTypeString                  = 0x03
	descTypeInterface               = 0x04
	descTypeEndpoint                = 0x05
	descTypeQualification           = 0x06
	descTypeOtherSpeedConfiguration = 0x07
	descTypeInterfacePower          = 0x08
	descTypeOTG                     = 0x09
	descTypeInterfaceAssociation    = 0x0B
	descTypeBOS                     = 0x0F
	descTypeDeviceCapability        = 0x10
	descTypeHID                     = 0x21
	descTypeHIDReport               = 0x22
	descTypeHIDPhysical             = 0x23
	descTypeCDCInterface            = 0x24
	descTypeCDCEndpoint             = 0x25
	descTypeEndpointCompanion       = 0x30

	// Standard request type
	descRequestTypeDirMsk             = 0x80
	descRequestTypeDirPos             = 7
	descRequestTypeDirOut             = 0x00
	descRequestTypeDirIn              = 0x80
	descRequestTypeTypeMsk            = 0x60
	descRequestTypeTypePos            = 5
	descRequestTypeTypeStandard       = 0
	descRequestTypeTypeClass          = 0x20
	descRequestTypeTypeVendor         = 0x40
	descRequestTypeRecipientMsk       = 0x1F
	descRequestTypeRecipientPos       = 0
	descRequestTypeRecipientDevice    = 0x00
	descRequestTypeRecipientInterface = 0x01
	descRequestTypeRecipientEndpoint  = 0x02
	descRequestTypeRecipientOther     = 0x03

	// Standard request
	descRequestStandardGetStatus        = 0x00
	descRequestStandardClearFeature     = 0x01
	descRequestStandardSetFeature       = 0x03
	descRequestStandardSetAddress       = 0x05
	descRequestStandardGetDescriptor    = 0x06
	descRequestStandardSetDescriptor    = 0x07
	descRequestStandardGetConfiguration = 0x08
	descRequestStandardSetConfiguration = 0x09
	descRequestStandardGetInterface     = 0x0A
	descRequestStandardSetInterface     = 0x0B
	descRequestStandardSynchFrame       = 0x0C

	// Configuration attributes
	descConfigAttrD7Msk           = 0x80
	descConfigAttrD7Pos           = 7
	descConfigAttrSelfPoweredMsk  = 0x40
	descConfigAttrSelfPoweredPos  = 6
	descConfigAttrRemoteWakeupMsk = 0x20
	descConfigAttrRemoteWakeupPos = 5

	// Endpoint type
	descEndptTypeControl     = 0x00
	descEndptTypeIsochronous = 0x01
	descEndptTypeBulk        = 0x02
	descEndptTypeInterrupt   = 0x03

	// Endpoint address
	descEndptAddrNumberMsk    = 0x0F
	descEndptAddrNumberPos    = 0
	descEndptAddrDirectionMsk = 0x80
	descEndptAddrDirectionPos = 7
	descEndptAddrDirectionOut = 0
	descEndptAddrDirectionIn  = 0x80

	// Endpoint attributes
	descEndptAttrTypeMsk           = 0x03
	descEndptAttrNumberPos         = 0
	descEndptAttrSyncTypeMsk       = 0x0C
	descEndptAttrSyncTypePos       = 2
	descEndptAttrSyncTypeNoSync    = 0x00
	descEndptAttrSyncTypeAsync     = 0x04
	descEndptAttrSyncTypeAdaptive  = 0x08
	descEndptAttrSyncTypeSync      = 0x0C
	descEndptAttrUsageTypeMsk      = 0x30
	descEndptAttrUsageTypePos      = 4
	descEndptAttrUsageTypeData     = 0x00
	descEndptAttrUsageTypeFeed     = 0x10
	descEndptAttrUsageTypeFeedData = 0x20

	// Endpoint max packet size
	descEndptMaxPktSizeMsk     = 0x07FF
	descEndptMaxPktSize        = 64
	descEndptMaxPktSizeMultMsk = 0x1800
	descEndptMaxPktSizeMultPos = 11

	// OTG attributes
	descOTGAttrSRPMsk = 0x01
	descOTGAttrHNPMsk = 0x02
	descOTGAttrADPMsk = 0x04

	// Device bus speed
	descDeviceSpeedFull  = 0x00
	descDeviceSpeedLow   = 0x01
	descDeviceSpeedHigh  = 0x02
	descDeviceSpeedSuper = 0x04

	// Device capability type
	descDeviceCapTypeWireless       = 0x01
	descDeviceCapTypeUSB20Extension = 0x02
	descDeviceCapTypeSuperspeed     = 0x03

	// Device capability attributes (USB 2.0 extension)
	descDeviceCapExtAttrLPMMsk  = 0x02
	descDeviceCapExtAttrLPMPos  = 1
	descDeviceCapExtAttrBESLMsk = 0x04
	descDeviceCapExtAttrBESLPos = 2
)

const (
	descDirOut = descRequestTypeDirOut >> descRequestTypeDirPos
	descDirIn  = descRequestTypeDirIn >> descRequestTypeDirPos

	descDirRx = descDirOut // "IN" and "OUT" terms are from host's perspective,
	descDirTx = descDirIn  // which is opposite from USB device. Kinda awkward.
)

// device returns the enumerated device descriptor value, defined per USB
// specification, for the receiver Speed s.
func (s Speed) device() uint32 {
	switch s {
	case LowSpeed:
		return descDeviceSpeedLow
	case FullSpeed:
		return descDeviceSpeedFull
	case HighSpeed:
		return descDeviceSpeedHigh
	case SuperSpeed, DualSuperSpeed:
		return descDeviceSpeedSuper
	default: // unrecognized Speed defaults to full-speed
		return descDeviceSpeedFull
	}
}

const (
	// Common attributes for all endpoint descriptor configurations.
	descEndptConfigAttr = descConfigAttrD7Msk | // Bit 7: reserved (1)
		(0 << descConfigAttrSelfPoweredPos) | // Bit 6: self-powered
		(0 << descConfigAttrRemoteWakeupPos) | // Bit 5: remote wakeup
		0 // Bits 0-4: reserved (0)

	descEndptConfigAttrRxPos = 0
	descEndptConfigAttrTxPos = 16
	descEndptConfigAttrRxMsk = (descEndptAttrSyncTypeMsk | descEndptConfigAttr) << descEndptConfigAttrRxPos
	descEndptConfigAttrTxMsk = (descEndptAttrSyncTypeMsk | descEndptConfigAttr) << descEndptConfigAttrTxPos

	descEndptConfigAttrRxUnused      = 0x02 << descEndptConfigAttrRxPos
	descEndptConfigAttrTxUnused      = 0x02 << descEndptConfigAttrTxPos
	descEndptConfigAttrRxIsochronous = (descEndptAttrSyncTypeAsync | descEndptConfigAttr) << descEndptConfigAttrRxPos
	descEndptConfigAttrTxIsochronous = (descEndptAttrSyncTypeAsync | descEndptConfigAttr) << descEndptConfigAttrTxPos
	descEndptConfigAttrRxBulk        = (descEndptAttrSyncTypeAdaptive | descEndptConfigAttr) << descEndptConfigAttrRxPos
	descEndptConfigAttrTxBulk        = (descEndptAttrSyncTypeAdaptive | descEndptConfigAttr) << descEndptConfigAttrTxPos
	descEndptConfigAttrRxInterrupt   = (descEndptAttrSyncTypeSync | descEndptConfigAttr) << descEndptConfigAttrRxPos
	descEndptConfigAttrTxInterrupt   = (descEndptAttrSyncTypeSync | descEndptConfigAttr) << descEndptConfigAttrTxPos
)

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
	// Size of all CDC-ACM configuration descriptors.
	descCDCACMConfigSize = uint16(
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

	// Size of all HID configuration descriptors.
	descHIDConfigSize = uint16(
		descLengthConfigure + // Configuration Header
			descLengthInterface + // Keyboard Interface Descriptor
			descLengthInterface + // Keyboard HID Interface Descriptor
			descLengthEndpoint + // Keyboard Endpoint Descriptor
			descLengthInterface + // Mouse Interface Descriptor
			descLengthInterface + // Mouse HID Interface Descriptor
			descLengthEndpoint + // Mouse Endpoint Descriptor
			descLengthInterface + // Serial Interface Descriptor
			descLengthInterface + // Serial HID Interface Descriptor
			descLengthEndpoint + // Serial Tx Endpoint Descriptor
			descLengthEndpoint + // Serial Rx Endpoint Descriptor
			descLengthInterface + // Joystick Interface Descriptor
			descLengthInterface + // Joystick HID Interface Descriptor
			descLengthEndpoint + // Joystick Endpoint Descriptor
			descLengthInterface + // Keyboard Media Keys Interface Descriptor
			descLengthInterface + // Keyboard Media Keys HID Interface Descriptor
			descLengthEndpoint) // Keyboard Media Keys Endpoint Descriptor

	// Position of each HID interface descriptor as offsets into the configuration
	// descriptor. See comments in the configuration descriptor definition for the
	// incremental tally that computes these.
	descHIDConfigKeyboardPos = 18
	descHIDConfigMousePos    = 43
	descHIDConfigSerialPos   = 68
	descHIDConfigJoystickPos = 100
	descHIDConfigMediaKeyPos = 125
)

type (
	// descString is the actual byte array used to hold string descriptors. The
	// first two bytes are a USB-specified header (0=length, 1=type), and the
	// remaining bytes are UTF-16 code points, ordered low byte-first. If you just
	// want to use UTF-8 (or even ASCII), you still need to reserve 2 bytes for
	// each symbol, but you can set all of their high bytes 0.
	descString [descStringSize]uint8
	// descStringIndex defines an indexed collection of string descriptors for a
	// given language.
	descStringIndex [descStringIndexCount]descString
	// descStringLanguage contains a language code and an indexed collection of
	// string descriptors encoded in that language.
	descStringLanguage struct {
		language   uint16
		descriptor descStringIndex
	}
)

const (
	descStringIndexCount = 4  // Language, Manufacturer, Product, Serial Number
	descStringSize       = 64 // (64-2)/2 = 31 chars each (UTF-16 code points)
	// The maximum allowable string descriptor size is 255, or (255-2)/2 = 126
	// available UTF-16 code points. Considering we are allocating this storage at
	// compile-time, it seems like an awful waste of space (255*4 = ~1 KiB) just
	// to store four strings, which, in all likelihood, will not be modified by
	// anyone other than TinyGo devs; 64*4 = 256 B (i.e., 31 UTF-16 code points
	// for each string) seems a good compromise.
)

// descEndpointInvalid represents an invalid endpoint address.
const descEndpointInvalid = ^uint8(descEndptAddrNumberMsk | descEndptAddrDirectionMsk)

// descCDCACMLineCodingSize defines the length of a CDC-ACM UART line coding
// buffer. Note that the actual buffer may be padded for alignment; but for
// Rx/Tx transfer purposes, descCDCACMLineCodingSize defines the number of bytes
// that are transferred following a control SETUP request.
const descCDCACMLineCodingSize = 7

// descCDCACMLineCoding represents an emulated UART's line configuration.
//
// Use descCDCACMLineCodingSize instead of unsafe.Sizeof(descCDCACMLineCoding)
// in any transfer requests, because the actual struct is padded for alignment.
type descCDCACMLineCoding struct {
	baud     uint32
	stopBits uint8
	parity   uint8
	numBits  uint8
	_        uint8
}

// parse initializes the receiver descCDCACMLineCoding from the given []uint8 v.
// Argument v is a Rx transfer buffer, filled following the completion of a
// control transfer from a CDC SET_LINE_CODING (0x20) request
func (s *descCDCACMLineCoding) parse(v []uint8) bool {
	if len(v) >= descCDCACMLineCodingSize {
		s.baud = packU32(v[:])
		s.stopBits = v[4]
		s.parity = v[5]
		s.numBits = v[6]
		return true
	}
	return false
}

// descCDCACMLineState represents an emulated UART's line state.
type descCDCACMLineState struct {
	// dataTerminalReady indicates if DTE is present or not.
	// Corresponds to V.24 signal 108/2 and RS-232 signal DTR.
	dataTerminalReady bool // DTR
	// requestToSend is the carrier control for half-duplex modems.
	// Corresponds to V.24 signal 105 and RS-232 signal RTS.
	requestToSend bool // RTS
}

// parse initializes the receiver descCDCACMLineState from the given uint16 v.
// Argument v corresponds to the wValue field in a control SETUP packet, which
// carries the line state from a CDC SET_CONTROL_LINE_STATE (0x22) request.
func (s *descCDCACMLineState) parse(v uint16) bool {
	s.dataTerminalReady = 0 != v&0x1
	s.requestToSend = 0 != v&0x2
	return true
}

type descCDCACMState uint8

const (
	descCDCACMStateConfigured descCDCACMState = iota // Received SET_CONFIGURATION class request
	descCDCACMStateLineState                         // Received SET_LINE_STATE after Configured state
	descCDCACMStateLineCoding                        // Received SET_LINE_CODING after LineState state
)

func (s *descCDCACMState) set(state descCDCACMState) {
	if state > *s {
		// state must be incremented in-order. Otherwise, reset to initial state.
		if state == *s+1 {
			*s = state
		} else {
			var init descCDCACMState // Reset to zero-value of type.
			*s = init
		}
	}
}

// Common configuration constants for the USB CDC-ACM (single) device class.
const (
	descCDCACMLanguageCount = 1 // String descriptor languages available

	descCDCACMInterfaceCount = 2 // Interfaces for all CDC-ACM configurations.
	descCDCACMEndpointCount  = 4 // Endpoints for all CDC-ACM configurations.

	descCDCACMEndpointCtrl = 0 // CDC-ACM Control Endpoint 0

	descCDCACMInterfaceCtrl    = 0 // CDC-ACM Control Interface
	descCDCACMEndpointStatus   = 1 // CDC-ACM Interrupt IN Endpoint
	descCDCACMConfigAttrStatus = descEndptConfigAttrRxUnused | descEndptConfigAttrTxInterrupt

	descCDCACMInterfaceData    = 1 // CDC-ACM Data Interface
	descCDCACMEndpointDataRx   = 2 // CDC-ACM Bulk Data OUT (Rx) Endpoint
	descCDCACMConfigAttrDataRx = descEndptConfigAttrRxBulk | descEndptConfigAttrTxUnused
	descCDCACMEndpointDataTx   = 3 // CDC-ACM Bulk Data IN (Tx) Endpoint
	descCDCACMConfigAttrDataTx = descEndptConfigAttrRxUnused | descEndptConfigAttrTxBulk
)

// descCDCACMClass holds references to all descriptors, buffers, and control
// structures for the USB CDC-ACM (single) device class.
type descCDCACMClass struct {
	*descCDCACMClassData // Target-defined, class-specific data

	locale *[descCDCACMLanguageCount]descStringLanguage // string descriptors
	device *[descLengthDevice]uint8                     // device descriptor
	qualif *[descLengthQualification]uint8              // device qualification descriptor
	config *[descCDCACMConfigSize]uint8                 // configuration descriptor

	state volatile.Register8
}

func (c *descCDCACMClass) setState(state descCDCACMState) {
	s := descCDCACMState(c.state.Get())
	s.set(state)
	c.state.Set(uint8(s))
}

// descCDCACM holds statically-allocated instances for each of the CDC-ACM
// (single) device class configurations, ordered by index (offset by -1).
var descCDCACM = [dcdCount]descCDCACMClass{

	{ // CDC-ACM (single) class configuration index 1
		descCDCACMClassData: &descCDCACMData[0],

		locale: &[descCDCACMLanguageCount]descStringLanguage{

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
		config: &[descCDCACMConfigSize]uint8{
			descLengthConfigure,        // Size of this descriptor in bytes
			descTypeConfigure,          // Descriptor Type
			lsU8(descCDCACMConfigSize), // Total length of data returned for this configuration (low)
			msU8(descCDCACMConfigSize), // Total length of data returned for this configuration (high)
			descCDCACMInterfaceCount,   // Number of interfaces supported by this configuration
			1,                          // Value to use to select this configuration (1 = CDC-ACM[0])
			0,                          // Index of string descriptor describing this configuration
			descEndptConfigAttr,        // Configuration attributes
			descCDCACMMaxPowerMa >> 1,  // Max power consumption when fully-operational (2 mA units)

			// Communication/Control Interface Descriptor
			descLengthInterface,       // Descriptor length
			descTypeInterface,         // Descriptor type
			descCDCACMInterfaceCtrl,   // Interface index
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
			descCDCACMInterfaceData,         // Data Interface

			// CDC Abstract Control Management Functional Descriptor
			descCDCFuncLengthAbstractControl, // Size of this descriptor in bytes
			descTypeCDCInterface,             // Descriptor Type
			descCDCFuncTypeAbstractControl,   // Descriptor Subtype
			0x06,                             // Capabilities

			// CDC Union Functional Descriptor
			descCDCFuncLengthUnion,  // Size of this descriptor in bytes
			descTypeCDCInterface,    // Descriptor Type
			descCDCFuncTypeUnion,    // Descriptor Subtype
			descCDCACMInterfaceCtrl, // Controlling interface index
			descCDCACMInterfaceData, // Controlled interface index

			// Communication/Control Notification Endpoint descriptor
			descLengthEndpoint, // Size of this descriptor in bytes
			descTypeEndpoint,   // Descriptor Type
			descCDCACMEndpointStatus | // Endpoint address
				descEndptAddrDirectionIn,
			descEndptTypeInterrupt,           // Attributes
			lsU8(descCDCACMStatusPacketSize), // Max packet size (low)
			msU8(descCDCACMStatusPacketSize), // Max packet size (high)
			descCDCACMStatusInterval,         // Polling Interval

			// Data Interface Descriptor
			descLengthInterface,     // Interface length
			descTypeInterface,       // Interface type
			descCDCACMInterfaceData, // Interface index
			0,                       // Alternate setting
			2,                       // Number of endpoints
			descCDCTypeData,         // Class code
			descCDCSubNone,          // Subclass code
			descCDCProtoNone,        // Protocol code
			0,                       // Interface Description String Index

			// Data Bulk Rx Endpoint descriptor
			descLengthEndpoint, // Size of this descriptor in bytes
			descTypeEndpoint,   // Descriptor Type
			descCDCACMEndpointDataRx | // Endpoint address
				descEndptAddrDirectionOut,
			descEndptTypeBulk,                // Attributes
			lsU8(descCDCACMDataRxPacketSize), // Max packet size (low)
			msU8(descCDCACMDataRxPacketSize), // Max packet size (high)
			0,                                // Polling Interval

			// Data Bulk Tx Endpoint descriptor
			descLengthEndpoint, // Size of this descriptor in bytes
			descTypeEndpoint,   // Descriptor Type
			descCDCACMEndpointDataTx | // Endpoint address
				descEndptAddrDirectionIn,
			descEndptTypeBulk,                // Attributes
			lsU8(descCDCACMDataTxPacketSize), // Max packet size (low)
			msU8(descCDCACMDataTxPacketSize), // Max packet size (high)
			0,                                // Polling Interval
		},
	},
}

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
			descHIDInterfaceCount,   // Number of interfaces supported by this configuration
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

			// [34+9] Mouse Interface Descriptor
			descLengthInterface,   // Descriptor length
			descTypeInterface,     // Descriptor type
			descHIDInterfaceMouse, // Interface index
			0,                     // Alternate setting
			1,                     // Number of endpoints
			descHIDType,           // Class code (HID = 0x03)
			descHIDSubBoot,        // Subclass code (Boot = 0x01)
			descHIDProtoMouse,     // Protocol code (Mouse = 0x02)
			0,                     // Interface Description String Index

			// [43+9] Mouse HID Interface Descriptor
			descLengthInterface,                   // Descriptor length
			descTypeHID,                           // Descriptor type
			0x11,                                  // HID BCD (low)
			0x01,                                  // HID BCD (high)
			0,                                     // Country code
			1,                                     // Number of descriptors
			descTypeHIDReport,                     // Descriptor type
			lsU8(uint16(len(descHIDReportMouse))), // Descriptor length (low)
			msU8(uint16(len(descHIDReportMouse))), // Descriptor length (high)

			// [52+7] Mouse Endpoint Descriptor
			descLengthEndpoint, // Size of this descriptor in bytes
			descTypeEndpoint,   // Descriptor Type
			descHIDEndpointMouse | // Endpoint address
				descEndptAddrDirectionIn,
			descEndptTypeInterrupt,         // Attributes
			lsU8(descHIDMouseTxPacketSize), // Max packet size (low)
			msU8(descHIDMouseTxPacketSize), // Max packet size (high)
			descHIDMouseTxInterval,         // Polling Interval

			// [59+9] Serial Interface Descriptor
			descLengthInterface,    // Descriptor length
			descTypeInterface,      // Descriptor type
			descHIDInterfaceSerial, // Interface index
			0,                      // Alternate setting
			2,                      // Number of endpoints
			descHIDType,            // Class code (HID = 0x03)
			descHIDSubNone,         // Subclass code
			descHIDProtoNone,       // Protocol code
			0,                      // Interface Description String Index

			// [68+9] Serial HID Interface Descriptor
			descLengthInterface,                    // Descriptor length
			descTypeHID,                            // Descriptor type
			0x11,                                   // HID BCD (low)
			0x01,                                   // HID BCD (high)
			0,                                      // Country code
			1,                                      // Number of descriptors
			descTypeHIDReport,                      // Descriptor type
			lsU8(uint16(len(descHIDReportSerial))), // Descriptor length (low)
			msU8(uint16(len(descHIDReportSerial))), // Descriptor length (high)

			// [77+7] Serial Tx Endpoint Descriptor
			descLengthEndpoint, // Size of this descriptor in bytes
			descTypeEndpoint,   // Descriptor Type
			descHIDEndpointSerialTx | // Endpoint address
				descEndptAddrDirectionIn,
			descEndptTypeInterrupt,          // Attributes
			lsU8(descHIDSerialTxPacketSize), // Max packet size (low)
			msU8(descHIDSerialTxPacketSize), // Max packet size (high)
			descHIDSerialTxInterval,         // Polling Interval

			// [84+7] Serial Rx Endpoint Descriptor
			descLengthEndpoint, // Size of this descriptor in bytes
			descTypeEndpoint,   // Descriptor Type
			descHIDEndpointSerialRx | // Endpoint address
				descEndptAddrDirectionOut,
			descEndptTypeInterrupt,          // Attributes
			lsU8(descHIDSerialRxPacketSize), // Max packet size (low)
			msU8(descHIDSerialRxPacketSize), // Max packet size (high)
			descHIDSerialRxInterval,         // Polling Interval

			// [91+9] Joystick Interface Descriptor
			descLengthInterface,      // Descriptor length
			descTypeInterface,        // Descriptor type
			descHIDInterfaceJoystick, // Interface index
			0,                        // Alternate setting
			1,                        // Number of endpoints
			descHIDType,              // Class code (HID = 0x03)
			descHIDSubNone,           // Subclass code
			descHIDProtoNone,         // Protocol code
			0,                        // Interface Description String Index

			// [100+9] Joystick HID Interface Descriptor
			descLengthInterface,                      // Descriptor length
			descTypeHID,                              // Descriptor type
			0x11,                                     // HID BCD (low)
			0x01,                                     // HID BCD (high)
			0,                                        // Country code
			1,                                        // Number of descriptors
			descTypeHIDReport,                        // Descriptor type
			lsU8(uint16(len(descHIDReportJoystick))), // Descriptor length (low)
			msU8(uint16(len(descHIDReportJoystick))), // Descriptor length (high)

			// [109+7] Joystick Endpoint Descriptor
			descLengthEndpoint, // Size of this descriptor in bytes
			descTypeEndpoint,   // Descriptor Type
			descHIDEndpointJoystick | // Endpoint address
				descEndptAddrDirectionIn,
			descEndptTypeInterrupt,            // Attributes
			lsU8(descHIDJoystickTxPacketSize), // Max packet size (low)
			msU8(descHIDJoystickTxPacketSize), // Max packet size (high)
			descHIDJoystickTxInterval,         // Polling Interval

			// [116+9] Keyboard Media Keys Interface Descriptor
			descLengthInterface,      // Descriptor length
			descTypeInterface,        // Descriptor type
			descHIDInterfaceMediaKey, // Interface index
			0,                        // Alternate setting
			1,                        // Number of endpoints
			descHIDType,              // Class code (HID = 0x03)
			descHIDSubNone,           // Subclass code
			descHIDProtoNone,         // Protocol code
			0,                        // Interface Description String Index

			// [125+9] Keyboard Media Keys HID Interface Descriptor
			descLengthInterface,                      // Descriptor length
			descTypeHID,                              // Descriptor type
			0x11,                                     // HID BCD (low)
			0x01,                                     // HID BCD (high)
			0,                                        // Country code
			1,                                        // Number of descriptors
			descTypeHIDReport,                        // Descriptor type
			lsU8(uint16(len(descHIDReportMediaKey))), // Descriptor length (low)
			msU8(uint16(len(descHIDReportMediaKey))), // Descriptor length (high)

			// [134+7] Keyboard Media Keys Endpoint Descriptor
			descLengthEndpoint, // Size of this descriptor in bytes
			descTypeEndpoint,   // Descriptor Type
			descHIDEndpointMediaKey | // Endpoint address
				descEndptAddrDirectionIn,
			descEndptTypeInterrupt,            // Attributes
			lsU8(descHIDMediaKeyTxPacketSize), // Max packet size (low)
			msU8(descHIDMediaKeyTxPacketSize), // Max packet size (high)
			descHIDMediaKeyTxInterval,         // Polling Interval
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
