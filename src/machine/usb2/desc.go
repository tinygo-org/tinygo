package usb2

const descUSBSpecVersion = uint16(0x0200) // USB 2.0

const descLanguageEnglish = uint16(0x0409)

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

// Common configuration constants for the USB CDC-ACM (single) device class.
const (
	// String descriptor languages
	descCDCACMLanguageCount = 1
	// Interfaces for all CDC-ACM configurations.
	descCDCACMInterfaceCount = 2
	descCDCACMInterfaceCtrl  = 0
	descCDCACMInterfaceData  = 1
	// Endpoints for all CDC-ACM configurations.
	descCDCACMEndpointCount  = 4
	descCDCACMEndpointStatus = 2 // Communication/control interrupt input
	descCDCACMEndpointDataRx = 3 // Bulk data output
	descCDCACMEndpointDataTx = 4 // Bulk data input
	// Endpoint configuration attributes for all CDC-ACM configurations.
	descCDCACMConfigAttrStatus = (descCDCACMConfigAttrUnused << descCDCACMConfigAttrRxPos) |
		(descCDCACMConfigAttrInterrupt << descCDCACMConfigAttrTxPos)
	descCDCACMConfigAttrDataRx = (descCDCACMConfigAttrBulk << descCDCACMConfigAttrRxPos) |
		(descCDCACMConfigAttrUnused << descCDCACMConfigAttrTxPos)
	descCDCACMConfigAttrDataTx = (descCDCACMConfigAttrUnused << descCDCACMConfigAttrRxPos) |
		(descCDCACMConfigAttrBulk << descCDCACMConfigAttrTxPos)

	// Size of all CDC-ACM configuration descriptors.
	descCDCACMConfigSize = uint16(
		descLengthConfigure + // configuration
			descLengthInterface + // communication/control interface
			descCDCFuncLengthHeader + // CDC header
			descCDCFuncLengthCallManagement + // CDC call management
			descCDCFuncLengthAbstractControl + // CDC abstract control
			descCDCFuncLengthUnion + // CDC union
			descLengthEndpoint + // communication/control input endpoint
			descLengthInterface + // data interface
			descLengthEndpoint + // data input endpoint
			descLengthEndpoint) // data output endpoint
	// Attributes of all CDC-ACM configuration descriptors.
	descCDCACMConfigAttr = descConfigAttrD7Msk | // Bit 7: reserved (1)
		(1 << descConfigAttrSelfPoweredPos) | // Bit 6: self-powered
		(0 << descConfigAttrRemoteWakeupPos) | // Bit 5: remote wakeup
		0 // Bits 0-4: reserved (0)

	descCDCACMConfigAttrRxPos       = 0
	descCDCACMConfigAttrTxPos       = 16
	descCDCACMConfigAttrUnused      = 0x02 // TBD: what is this?
	descCDCACMConfigAttrIsochronous = descCDCACMConfigAttr | descEndptAttrSyncTypeAsync
	descCDCACMConfigAttrBulk        = descCDCACMConfigAttr | descEndptAttrSyncTypeAdaptive
	descCDCACMConfigAttrInterrupt   = descCDCACMConfigAttr | descEndptAttrSyncTypeSync
)

// descCDCACM0Device holds the default device descriptor for CDC-ACM[0], i.e.,
// configuration index 1.
var descCDCACM0Device = [descLengthDevice]uint8{
	descLengthDevice,          // Size of this descriptor in bytes
	descTypeDevice,            // Descriptor Type
	lsU8(descUSBSpecVersion),  // USB Specification Release Number in BCD (low)
	msU8(descUSBSpecVersion),  // USB Specification Release Number in BCD (high)
	descCDCTypeComm,           // Class code (assigned by the USB-IF).
	descCDCSubNone,            // Subclass code (assigned by the USB-IF).
	descCDCProtoNone,          // Protocol code (assigned by the USB-IF).
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
}

// descCDCACM0Qualif holds the default device qualification descriptor for
// CDC-ACM[0], i.e., configuration index 1.
var descCDCACM0Qualif = [descLengthQualification]uint8{
	descLengthQualification,  // Size of this descriptor in bytes
	descTypeQualification,    // Descriptor Type
	lsU8(descUSBSpecVersion), // USB Specification Release Number in BCD (low)
	msU8(descUSBSpecVersion), // USB Specification Release Number in BCD (high)
	descCDCTypeComm,          // Class code (assigned by the USB-IF).
	descCDCSubNone,           // Subclass code (assigned by the USB-IF).
	descCDCProtoNone,         // Protocol code (assigned by the USB-IF).
	descEndptMaxPktSize,      // Maximum packet size for endpoint zero (8, 16, 32, or 64)
	descCDCACMCount,          // Number of possible configurations
	0,                        // Reserved
}

// descCDCACM0Config holds the default configuration descriptors for CDC-ACM[0],
// i.e., configuration index 1.
var descCDCACM0Config = [descCDCACMConfigSize]uint8{
	descLengthConfigure,        // Size of this descriptor in bytes
	descTypeConfigure,          // Descriptor Type
	lsU8(descCDCACMConfigSize), // Total length of data returned for this configuration (low)
	msU8(descCDCACMConfigSize), // Total length of data returned for this configuration (high)
	descCDCACMInterfaceCount,   // Number of interfaces supported by this configuration
	1,                          // Value to use to select this configuration (1 = CDC-ACM[0])
	0,                          // Index of string descriptor describing this configuration
	descCDCACMConfigAttr,       // Configuration attributes
	descCDCACMMaxPower,         // Max power consumption when fully-operational (2 mA units)

	// Communication/Control Interface Descriptor
	descLengthInterface,       // Descriptor length
	descTypeInterface,         // Descriptor type
	descCDCACMInterfaceCtrl,   // Interface index
	0,                         // Alternate setting
	1,                         // Number of endpoints
	descCDCTypeComm,           // Class code
	descCDCSubAbstractControl, // Subclass code
	descCDCProtoNone,          // Protocol code (NOTE: Teensyduino defines this as 1 [AT V.250])
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
	16,                               // Polling Interval

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
}

// descCDCACMCodingSize defines the length of a CDC-ACM UART line coding buffer.
const descCDCACMCodingSize = 7

// descCDCACM0LineCoding holds the default UART line coding for CDC-ACM[0],
// i.e., configuration index 1.
var descCDCACM0LineCoding descCDCACMLineCoding

type descCDCACMLineCoding struct {
	baud     uint32
	stopBits uint8
	parity   uint8
	numBits  uint8
	rtsdtr   uint8
}

func (lc *descCDCACMLineCoding) parse(buffer []uint8) bool {
	if len(buffer) < descCDCACMCodingSize {
		return false
	}
	_ = copy(buffer[:], buffer)
	lc.baud = packU32(buffer[:])
	lc.stopBits = buffer[4]
	if 0 == lc.stopBits {
		lc.stopBits = 1
	}
	lc.parity = buffer[5]
	lc.numBits = buffer[6]
	return true
}

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

// descCDCACM0String holds the default string descriptors for CDC-ACM[0], i.e.,
// configuration index 1.
var descCDCACM0String = [descCDCACMLanguageCount]descStringLanguage{
	{ // US English string descriptors
		language: descLanguageEnglish,
		descriptor: descStringIndex{
			{ // Language (index 0)
				4,
				descTypeString,
				lsU8(descLanguageEnglish),
				msU8(descLanguageEnglish),
			},
			// Actual string descriptors (index > 0) are copied into here at runtime!
		},
	},
}
