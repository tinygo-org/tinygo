package usb

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

// descEndpointInvalid represents an invalid endpoint address.
const descEndpointInvalid = ^uint8(descEndptAddrNumberMsk | descEndptAddrDirectionMsk)

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
