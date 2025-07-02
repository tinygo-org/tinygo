package descriptor

import "math"

const (
	hidUsagePage       = 0x04
	hidUsage           = 0x08
	hidLogicalMinimum  = 0x14
	hidLogicalMaximum  = 0x24
	hidUsageMinimum    = 0x18
	hidUsageMaximum    = 0x28
	hidPhysicalMinimum = 0x34
	hidPhysicalMaximum = 0x44
	hidUnitExponent    = 0x54
	hidUnit            = 0x64
	hidCollection      = 0xA0
	hidInput           = 0x80
	hidOutput          = 0x90
	hidReportSize      = 0x74
	hidReportCount     = 0x94
	hidReportID        = 0x84
)

const (
	hidSizeValue0 = 0x00
	hidSizeValue1 = 0x01
	hidSizeValue2 = 0x02
	hidSizeValue4 = 0x03
)

var (
	HIDUsagePageGenericDesktop     = HIDUsagePage(0x01)
	HIDUsagePageSimulationControls = HIDUsagePage(0x02)
	HIDUsagePageVRControls         = HIDUsagePage(0x03)
	HIDUsagePageSportControls      = HIDUsagePage(0x04)
	HIDUsagePageGameControls       = HIDUsagePage(0x05)
	HIDUsagePageGenericControls    = HIDUsagePage(0x06)
	HIDUsagePageKeyboard           = HIDUsagePage(0x07)
	HIDUsagePageLED                = HIDUsagePage(0x08)
	HIDUsagePageButton             = HIDUsagePage(0x09)
	HIDUsagePageOrdinal            = HIDUsagePage(0x0A)
	HIDUsagePageTelephony          = HIDUsagePage(0x0B)
	HIDUsagePageConsumer           = HIDUsagePage(0x0C)
	HIDUsagePageDigitizers         = HIDUsagePage(0x0D)
	HIDUsagePageHaptics            = HIDUsagePage(0x0E)
	HIDUsagePagePhysicalInput      = HIDUsagePage(0x0F)
	HIDUsagePageUnicode            = HIDUsagePage(0x10)
	HIDUsagePageSoC                = HIDUsagePage(0x11)
	HIDUsagePageEyeHeadTrackers    = HIDUsagePage(0x12)
	HIDUsagePageAuxDisplay         = HIDUsagePage(0x14)
	HIDUsagePageSensors            = HIDUsagePage(0x20)
	HIDUsagePageMedicalInstrument  = HIDUsagePage(0x40)
	HIDUsagePageBrailleDisplay     = HIDUsagePage(0x41)
	HIDUsagePageLighting           = HIDUsagePage(0x59)
	HIDUsagePageMonitor            = HIDUsagePage(0x80)
	HIDUsagePageMonitorEnum        = HIDUsagePage(0x81)
	HIDUsagePageVESA               = HIDUsagePage(0x82)
	HIDUsagePagePower              = HIDUsagePage(0x84)
	HIDUsagePageBatterySystem      = HIDUsagePage(0x85)
	HIDUsagePageBarcodeScanner     = HIDUsagePage(0x8C)
	HIDUsagePageScales             = HIDUsagePage(0x8D)
	HIDUsagePageMagneticStripe     = HIDUsagePage(0x8E)
	HIDUsagePageCameraControl      = HIDUsagePage(0x90)
	HIDUsagePageArcade             = HIDUsagePage(0x91)
	HIDUsagePageGaming             = HIDUsagePage(0x92)
)

var (
	HIDUsageDesktopPointer         = HIDUsage(0x01)
	HIDUsageDesktopMouse           = HIDUsage(0x02)
	HIDUsageDesktopJoystick        = HIDUsage(0x04)
	HIDUsageDesktopGamepad         = HIDUsage(0x05)
	HIDUsageDesktopKeyboard        = HIDUsage(0x06)
	HIDUsageDesktopKeypad          = HIDUsage(0x07)
	HIDUsageDesktopMultiaxis       = HIDUsage(0x08)
	HIDUsageDesktopTablet          = HIDUsage(0x09)
	HIDUsageDesktopWaterCooling    = HIDUsage(0x0A)
	HIDUsageDesktopChassis         = HIDUsage(0x0B)
	HIDUsageDesktopWireless        = HIDUsage(0x0C)
	HIDUsageDesktopPortable        = HIDUsage(0x0D)
	HIDUsageDesktopSystemMultiaxis = HIDUsage(0x0E)
	HIDUsageDesktopSpatial         = HIDUsage(0x0F)
	HIDUsageDesktopAssistive       = HIDUsage(0x10)
	HIDUsageDesktopDock            = HIDUsage(0x11)
	HIDUsageDesktopDockable        = HIDUsage(0x12)
	HIDUsageDesktopCallState       = HIDUsage(0x13)
	HIDUsageDesktopX               = HIDUsage(0x30)
	HIDUsageDesktopY               = HIDUsage(0x31)
	HIDUsageDesktopZ               = HIDUsage(0x32)
	HIDUsageDesktopRx              = HIDUsage(0x33)
	HIDUsageDesktopRy              = HIDUsage(0x34)
	HIDUsageDesktopRz              = HIDUsage(0x35)
	HIDUsageDesktopSlider          = HIDUsage(0x36)
	HIDUsageDesktopDial            = HIDUsage(0x37)
	HIDUsageDesktopWheel           = HIDUsage(0x38)
	HIDUsageDesktopHatSwitch       = HIDUsage(0x39)
	HIDUsageDesktopCountedBuffer   = HIDUsage(0x3A)
)

var (
	HIDUsageConsumerControl             = HIDUsage(0x01)
	HIDUsageConsumerNumericKeypad       = HIDUsage(0x02)
	HIDUsageConsumerProgrammableButtons = HIDUsage(0x03)
	HIDUsageConsumerMicrophone          = HIDUsage(0x04)
	HIDUsageConsumerHeadphone           = HIDUsage(0x05)
	HIDUsageConsumerGraphicEqualizer    = HIDUsage(0x06)
)

var (
	HIDCollectionPhysical    = HIDCollection(0x00)
	HIDCollectionApplication = HIDCollection(0x01)
	HIDCollectionEnd         = []byte{0xC0}
)

var (
	// Input (Data,Ary,Abs), Key arrays (6 bytes)
	HIDInputDataAryAbs = HIDInput(0x00)

	// Input (Data, Variable, Absolute), Modifier byte
	HIDInputDataVarAbs = HIDInput(0x02)

	// Input (Const,Var,Abs), Modifier byte
	HIDInputConstVarAbs = HIDInput(0x03)

	// Input (Data, Variable, Relative), 2 position bytes (X & Y)
	HIDInputDataVarRel = HIDInput(0x06)

	// Output (Data, Variable, Absolute), Modifier byte
	HIDOutputDataVarAbs = HIDOutput(0x02)

	// Output (Const, Variable, Absolute), Modifier byte
	HIDOutputConstVarAbs = HIDOutput(0x03)
)

func hidShortItem(tag byte, value uint32) []byte {
	switch {
	case value <= math.MaxUint8:
		return []byte{tag | hidSizeValue1, byte(value)}
	case value <= math.MaxUint16:
		return []byte{tag | hidSizeValue2, byte(value), byte(value >> 8)}
	default:
		return []byte{tag | hidSizeValue4, byte(value), byte(value >> 8), byte(value >> 16), byte(value >> 24)}
	}
}

func hidShortItemSigned(tag byte, value int32) []byte {
	switch {
	case math.MinInt8 <= value && value <= math.MaxInt8:
		return []byte{tag | hidSizeValue1, byte(value)}
	case math.MinInt16 <= value && value <= math.MaxInt16:
		return []byte{tag | hidSizeValue2, byte(value), byte(value >> 8)}
	default:
		return []byte{tag | hidSizeValue4, byte(value), byte(value >> 8), byte(value >> 16), byte(value >> 24)}
	}
}

func HIDReportSize(size int) []byte {
	return hidShortItem(hidReportSize, uint32(size))
}

func HIDReportCount(count int) []byte {
	return hidShortItem(hidReportCount, uint32(count))
}

func HIDReportID(id int) []byte {
	return hidShortItem(hidReportID, uint32(id))
}

func HIDLogicalMinimum(min int) []byte {
	return hidShortItemSigned(hidLogicalMinimum, int32(min))
}

func HIDLogicalMaximum(max int) []byte {
	return hidShortItemSigned(hidLogicalMaximum, int32(max))
}

func HIDUsageMinimum(min int) []byte {
	return hidShortItem(hidUsageMinimum, uint32(min))
}

func HIDUsageMaximum(max int) []byte {
	return hidShortItem(hidUsageMaximum, uint32(max))
}

func HIDPhysicalMinimum(min int) []byte {
	return hidShortItemSigned(hidPhysicalMinimum, int32(min))
}

func HIDPhysicalMaximum(max int) []byte {
	return hidShortItemSigned(hidPhysicalMaximum, int32(max))
}

func HIDUnitExponent(exp int) []byte {
	// 4 Bit two's complement
	return hidShortItem(hidUnitExponent, uint32(exp&0xF))
}

func HIDUnit(unit uint32) []byte {
	return hidShortItem(hidUnit, unit)
}

func HIDUsagePage(id uint16) []byte {
	return hidShortItem(hidUsagePage, uint32(id))
}

func HIDUsage(id uint32) []byte {
	return hidShortItem(hidUsage, id)
}

func HIDCollection(id uint32) []byte {
	return hidShortItem(hidCollection, id)
}

func HIDInput(flags uint32) []byte {
	return hidShortItem(hidInput, flags)
}

func HIDOutput(flags uint32) []byte {
	return hidShortItem(hidOutput, flags)
}
