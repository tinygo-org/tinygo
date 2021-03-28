package usb2

const descUSBSpecVersion = 0x0200 // USB 2.0

// USB constants defined per specification.
const (

	// Descriptor length
	descLengthDevice                   = 18
	descLengthConfigure                = 9
	descLengthInterface                = 9
	descLengthEndpoint                 = 7
	descLengthDeviceQualitier          = 10
	descLengthOTG                      = 5
	descLengthBOS                      = 5
	descLengthEndpointCompanion        = 6
	descLengthDevCapTypeUSB20Extension = 7
	descLengthDevCapTypeSuperspeed     = 10

	// Descriptor type
	descTypeDevice                  = 0x01
	descTypeConfigure               = 0x02
	descTypeString                  = 0x03
	descTypeInterface               = 0x04
	descTypeEndpoint                = 0x05
	descTypeDeviceQualitier         = 0x06
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
	descEndptAttrTypeMsk                   = 0x03
	descEndptAttrNumberPos                 = 0
	descEndptAttrSyncTypeMsk               = 0x0C
	descEndptAttrSyncTypePos               = 2
	descEndptAttrSyncTypeNoSync            = 0x00
	descEndptAttrSyncTypeAsync             = 0x04
	descEndptAttrSyncTypeAdaptive          = 0x08
	descEndptAttrSyncTypeSync              = 0x0C
	descEndptAttrUsageTypeMsk              = 0x30
	descEndptAttrUsageTypePos              = 4
	descEndptAttrUsageTypeDataEndpoint     = 0x00
	descEndptAttrUsageTypeFeedEndpoint     = 0x10
	descEndptAttrUsageTypeFeedDataEndpoint = 0x20

	// Endpoint max packet size
	descEndptMaxPktSizeSizeMsk      = 0x07FF
	descEndptMaxPktSizeMultTransMsk = 0x1800
	descEndptMaxPktSizeMultTransPos = 11
	descEndptMaxPktSizeMaximum      = 64

	// OTG attributes
	descOTGAttrSRPMsk = 0x01
	descOTGAttrHNPMsk = 0x02
	descOTGAttrADPMsk = 0x04

	// Device capability type
	descDevCapTypeWireless       = 0x01
	descDevCapTypeUSB20Extension = 0x02
	descDevCapTypeSuperspeed     = 0x03

	// Device capability attributes (USB 2.0 extension)
	descDevCapExtAttrLPMMsk  = 0x02
	descDevCapExtAttrLPMPos  = 1
	descDevCapExtAttrBESLMsk = 0x04
	descDevCapExtAttrBESLPos = 2
)

// USB CDC constants defined per specification.
const (

	// Device class
	descCDCComm = 0x02 // communication/control
	descCDCData = 0x0A // data

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
	descCDCLengthFuncHeader          = 5
	descCDCLengthFuncCallManagement  = 5
	descCDCLengthFuncAbstractControl = 4
	descCDCLengthFuncUnion           = 5

	// Functional descriptor type
	descCDCTypeFuncHeader             = 0x00
	descCDCTypeFuncCallManagement     = 0x01
	descCDCTypeFuncAbstractControl    = 0x02
	descCDCTypeFuncDirectLine         = 0x03
	descCDCTypeFuncTelephoneRinger    = 0x04
	descCDCTypeFuncTelephoneReport    = 0x05
	descCDCTypeFuncUnion              = 0x06
	descCDCTypeFuncCountrySelect      = 0x07
	descCDCTypeFuncTelephoneModes     = 0x08
	descCDCTypeFuncTerminal           = 0x09
	descCDCTypeFuncNetworkChannel     = 0x0A
	descCDCTypeFuncProtocolUnit       = 0x0B
	descCDCTypeFuncExtensionUnit      = 0x0C
	descCDCTypeFuncMultiChannel       = 0x0D
	descCDCTypeFuncCAPIControl        = 0x0E
	descCDCTypeFuncEthernetNetworking = 0x0F
	descCDCTypeFuncATMNetworking      = 0x10
	descCDCTypeFuncWirelessControl    = 0x11
	descCDCTypeFuncMobileDirectLine   = 0x12
	descCDCTypeFuncMDLMDetail         = 0x13
	descCDCTypeFuncDeviceManagement   = 0x14
	descCDCTypeFuncOBEX               = 0x15
	descCDCTypeFuncCommandSet         = 0x16
	descCDCTypeFuncCommandSetDetail   = 0x17
	descCDCTypeFuncTelephoneControl   = 0x18
	descCDCTypeFuncOBEXServiceID      = 0x19
)

// Common configuration constants for the USB CDC-ACM (single) device class.
const (
	// Interfaces for the first CDC-ACM device (index 1).
	descCDCACM0InterfaceCount = 2
	descCDCACM0InterfaceCtrl  = 0
	descCDCACM0InterfaceData  = 1
	// Endpoints for the first CDC-ACM device (index 1).
	descCDCACM0EndpointCount  = 4
	descCDCACM0EndpointStatus = 2 // Communication/control interrupt input
	descCDCACM0EndpointDataRx = 3 // Bulk data output
	descCDCACM0EndpointDataTx = 4 // Bulk data input

	// Size of all CDC-ACM configuration descriptors.
	descCDCACMConfigSize = uint16(
		descLengthConfigure + // configuration
			descLengthInterface + // communication/control interface
			descCDCLengthFuncHeader + // CDC header
			descCDCLengthFuncCallManagement + // CDC call management
			descCDCLengthFuncAbstractControl + // CDC abstract control
			descCDCLengthFuncUnion + // CDC union
			descLengthEndpoint + // communication/control input endpoint
			descLengthInterface + // data interface
			descLengthEndpoint + // data input endpoint
			descLengthEndpoint) // data output endpoint
	// Attributes of all CDC-ACM configuration descriptors.
	descCDCACMConfigAttr = descConfigAttrD7Msk | // Bit 7: reserved (1)
		(1 << descConfigAttrSelfPoweredPos) | // Bit 6: self-powered
		(0 << descConfigAttrRemoteWakeupPos) | // Bit 5: remote wakeup
		0 // Bits 0-4: reserved (0)
)

// Common descriptors for the USB CDC-ACM (single) device class.
var (
	// descDeviceCDCACM holds the default device descriptors for CDC-ACM devices.
	descDeviceCDCACM = [descCDCACMConfigCount][descLengthDevice]uint8{
		{
			descLengthDevice,           // Size of this descriptor in bytes
			descTypeDevice,             // DEVICE Descriptor Type
			lsU8(descUSBSpecVersion),   // USB Specification Release Number in BCD (low)
			msU8(descUSBSpecVersion),   // USB Specification Release Number in BCD (high)
			descCDCComm,                // Class code (assigned by the USB-IF).
			descCDCSubNone,             // Subclass code (assigned by the USB-IF).
			descCDCProtoNone,           // Protocol code (assigned by the USB-IF).
			descEndptMaxPktSizeMaximum, // Maximum packet size for endpoint zero (8, 16, 32, or 64)
			lsU8(descVendorID),         // Vendor ID (low) (assigned by the USB-IF)
			msU8(descVendorID),         // Vendor ID (high) (assigned by the USB-IF)
			lsU8(descProductID),        // Product ID (low) (assigned by the manufacturer)
			msU8(descProductID),        // Product ID (high) (assigned by the manufacturer)
			lsU8(descReleaseID),        // Device release number in BCD (low)
			msU8(descReleaseID),        // Device release number in BCD (high)
			1,                          // Index of string descriptor describing manufacturer
			2,                          // Index of string descriptor describing product
			0,                          // Index of string descriptor describing the device's serial number
			descCDCACMConfigCount,      // Number of possible configurations
		},
	}

	// descConfigCDCACM holds the default configuration descriptors for CDC-ACM
	// devices.
	descConfigCDCACM = [descCDCACMConfigCount][descCDCACMConfigSize]uint8{
		{
			descLengthConfigure,        // Size of this descriptor in bytes
			descTypeConfigure,          // Descriptor Type
			lsU8(descCDCACMConfigSize), // Total length of data returned for this configuration (low)
			msU8(descCDCACMConfigSize), // Total length of data returned for this configuration (high)
			descCDCACM0InterfaceCount,  // Number of interfaces supported by this configuration
			1,                          // Value to use to select this configuration (1 = CDC-ACM[0])
			0,                          // Index of string descriptor describing this configuration
			descCDCACMConfigAttr,       // Configuration attributes
			descCDCACMMaxPower,         // Max power consumption when fully-operational (2 mA units)

			// Communication/Control Interface Descriptor
			descLengthInterface,       // Descriptor length
			descTypeInterface,         // Descriptor type
			descCDCACM0InterfaceCtrl,  // Interface index
			0,                         // Alternate setting
			1,                         // Number of endpoints
			descCDCComm,               // Class code
			descCDCSubAbstractControl, // Subclass code
			descCDCProtoNone,          // Protocol code (NOTE: Teensyduino defines this as 1 [AT V.250])
			0,                         // Interface Description String Index

			// CDC Header Functional Descriptor
			descCDCLengthFuncHeader, // Size of this descriptor in bytes
			descTypeCDCInterface,    // Descriptor Type
			descCDCTypeFuncHeader,   // Descriptor Subtype
			0x10,                    // USB CDC specification version 1.10 (low)
			0x01,                    // USB CDC specification version 1.10 (high)

			// CDC Call Management Functional Descriptor
			descCDCLengthFuncCallManagement, // Size of this descriptor in bytes
			descTypeCDCInterface,            // Descriptor Type
			descCDCTypeFuncCallManagement,   // Descriptor Subtype
			0x01,                            // Capabilities
			descCDCACM0InterfaceData,        // Data Interface

			// CDC Abstract Control Management Functional Descriptor
			descCDCLengthFuncAbstractControl, // Size of this descriptor in bytes
			descTypeCDCInterface,             // Descriptor Type
			descCDCTypeFuncAbstractControl,   // Descriptor Subtype
			0x06,                             // Capabilities

			// CDC Union Functional Descriptor
			descCDCLengthFuncUnion,   // Size of this descriptor in bytes
			descTypeCDCInterface,     // Descriptor Type
			descCDCTypeFuncUnion,     // Descriptor Subtype
			descCDCACM0InterfaceCtrl, // Controlling interface index
			descCDCACM0InterfaceData, // Controlled interface index

			// Communication/Control Notification Endpoint descriptor
			descLengthEndpoint, // Size of this descriptor in bytes
			descTypeEndpoint,   // Descriptor Type
			descCDCACM0EndpointStatus | // Endpoint address
				descEndptAddrDirectionIn,
			descEndptTypeInterrupt,           // Attributes
			lsU8(descCDCACMStatusPacketSize), // Max packet size (low)
			msU8(descCDCACMStatusPacketSize), // Max packet size (high)
			8,                                // Polling Interval

			// Data Interface Descriptor
			descLengthInterface,      // Interface length
			descTypeInterface,        // Interface type
			descCDCACM0InterfaceData, // Interface index
			0,                        // Alternate setting
			2,                        // Number of endpoints
			descCDCData,              // Class code
			descCDCSubNone,           // Subclass code
			descCDCProtoNone,         // Protocol code
			0,                        // Interface Description String Index

			// Data Bulk Rx Endpoint descriptor
			descLengthEndpoint, // Size of this descriptor in bytes
			descTypeEndpoint,   // Descriptor Type
			descCDCACM0EndpointDataRx | // Endpoint address
				descEndptAddrDirectionOut,
			descEndptTypeBulk,                // Attributes
			lsU8(descCDCACMDataRxPacketSize), // Max packet size (low)
			msU8(descCDCACMDataRxPacketSize), // Max packet size (high)
			0,                                // Polling Interval

			// Data Bulk Tx Endpoint descriptor
			descLengthEndpoint, // Size of this descriptor in bytes
			descTypeEndpoint,   // Descriptor Type
			descCDCACM0EndpointDataTx | // Endpoint address
				descEndptAddrDirectionIn,
			descEndptTypeBulk,                // Attributes
			lsU8(descCDCACMDataTxPacketSize), // Max packet size (low)
			msU8(descCDCACMDataTxPacketSize), // Max packet size (high)
			0,                                // Polling Interval
		},
	}
)
