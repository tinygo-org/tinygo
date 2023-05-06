package descriptor

import (
	"encoding/binary"
)

const (
	hidUsagePage       = 0x05
	hidUsage           = 0x09
	hidLogicalMinimum  = 0x15
	hidLogicalMaximum  = 0x25
	hidUsageMinimum    = 0x19
	hidUsageMaximum    = 0x29
	hidPhysicalMinimum = 0x35
	hidPhysicalMaximum = 0x46
	hidUnitExponent    = 0x55
	hidUnit            = 0x65
	hidCollection      = 0xa1
	hidInput           = 0x81
	hidReportSize      = 0x75
	hidReportCount     = 0x95
	hidReportID        = 0x85
)

var (
	HIDUsagePageGenericDesktop     = []byte{hidUsagePage, 0x01}
	HIDUsagePageSimulationControls = []byte{hidUsagePage, 0x02}
	HIDUsagePageVRControls         = []byte{hidUsagePage, 0x03}
	HIDUsagePageSportControls      = []byte{hidUsagePage, 0x04}
	HIDUsagePageGameControls       = []byte{hidUsagePage, 0x05}
	HIDUsagePageGenericControls    = []byte{hidUsagePage, 0x06}
	HIDUsagePageKeyboard           = []byte{hidUsagePage, 0x07}
	HIDUsagePageLED                = []byte{hidUsagePage, 0x08}
	HIDUsagePageButton             = []byte{hidUsagePage, 0x09}
	HIDUsagePageOrdinal            = []byte{hidUsagePage, 0x0A}
	HIDUsagePageTelephony          = []byte{hidUsagePage, 0x0B}
	HIDUsagePageConsumer           = []byte{hidUsagePage, 0x0C}
	HIDUsagePageDigitizers         = []byte{hidUsagePage, 0x0D}
	HIDUsagePageHaptics            = []byte{hidUsagePage, 0x0E}
	HIDUsagePagePhysicalInput      = []byte{hidUsagePage, 0x0F}
	HIDUsagePageUnicode            = []byte{hidUsagePage, 0x10}
	HIDUsagePageSoC                = []byte{hidUsagePage, 0x11}
	HIDUsagePageEyeHeadTrackers    = []byte{hidUsagePage, 0x12}
	HIDUsagePageAuxDisplay         = []byte{hidUsagePage, 0x14}
	HIDUsagePageSensors            = []byte{hidUsagePage, 0x20}
	HIDUsagePageMedicalInstrument  = []byte{hidUsagePage, 0x40}
	HIDUsagePageBrailleDisplay     = []byte{hidUsagePage, 0x41}
	HIDUsagePageLighting           = []byte{hidUsagePage, 0x59}
	HIDUsagePageMonitor            = []byte{hidUsagePage, 0x80}
	HIDUsagePageMonitorEnum        = []byte{hidUsagePage, 0x81}
	HIDUsagePageVESA               = []byte{hidUsagePage, 0x82}
	HIDUsagePagePower              = []byte{hidUsagePage, 0x84}
	HIDUsagePageBatterySystem      = []byte{hidUsagePage, 0x85}
	HIDUsagePageBarcodeScanner     = []byte{hidUsagePage, 0x8C}
	HIDUsagePageScales             = []byte{hidUsagePage, 0x8D}
	HIDUsagePageMagneticStripe     = []byte{hidUsagePage, 0x8E}
	HIDUsagePageCameraControl      = []byte{hidUsagePage, 0x90}
	HIDUsagePageArcade             = []byte{hidUsagePage, 0x91}
	HIDUsagePageGaming             = []byte{hidUsagePage, 0x92}
)

var (
	HIDUsageDesktopPointer         = []byte{hidUsage, 0x01}
	HIDUsageDesktopMouse           = []byte{hidUsage, 0x02}
	HIDUsageDesktopJoystick        = []byte{hidUsage, 0x04}
	HIDUsageDesktopGamepad         = []byte{hidUsage, 0x05}
	HIDUsageDesktopKeyboard        = []byte{hidUsage, 0x06}
	HIDUsageDesktopKeypad          = []byte{hidUsage, 0x07}
	HIDUsageDesktopMultiaxis       = []byte{hidUsage, 0x08}
	HIDUsageDesktopTablet          = []byte{hidUsage, 0x09}
	HIDUsageDesktopWaterCooling    = []byte{hidUsage, 0x0A}
	HIDUsageDesktopChassis         = []byte{hidUsage, 0x0B}
	HIDUsageDesktopWireless        = []byte{hidUsage, 0x0C}
	HIDUsageDesktopPortable        = []byte{hidUsage, 0x0D}
	HIDUsageDesktopSystemMultiaxis = []byte{hidUsage, 0x0E}
	HIDUsageDesktopSpatial         = []byte{hidUsage, 0x0F}
	HIDUsageDesktopAssistive       = []byte{hidUsage, 0x10}
	HIDUsageDesktopDock            = []byte{hidUsage, 0x11}
	HIDUsageDesktopDockable        = []byte{hidUsage, 0x12}
	HIDUsageDesktopCallState       = []byte{hidUsage, 0x13}
	HIDUsageDesktopX               = []byte{hidUsage, 0x30}
	HIDUsageDesktopY               = []byte{hidUsage, 0x31}
	HIDUsageDesktopZ               = []byte{hidUsage, 0x32}
	HIDUsageDesktopRx              = []byte{hidUsage, 0x33}
	HIDUsageDesktopRy              = []byte{hidUsage, 0x34}
	HIDUsageDesktopRz              = []byte{hidUsage, 0x35}
	HIDUsageDesktopSlider          = []byte{hidUsage, 0x36}
	HIDUsageDesktopDial            = []byte{hidUsage, 0x37}
	HIDUsageDesktopWheel           = []byte{hidUsage, 0x38}
	HIDUsageDesktopHatSwitch       = []byte{hidUsage, 0x39}
	HIDUsageDesktopCountedBuffer   = []byte{hidUsage, 0x3A}
)

var (
	HIDUsageConsumerControl             = []byte{hidUsage, 0x01}
	HIDUsageConsumerNumericKeypad       = []byte{hidUsage, 0x02}
	HIDUsageConsumerProgrammableButtons = []byte{hidUsage, 0x03}
	HIDUsageConsumerMicrophone          = []byte{hidUsage, 0x04}
	HIDUsageConsumerHeadphone           = []byte{hidUsage, 0x05}
	HIDUsageConsumerGraphicEqualizer    = []byte{hidUsage, 0x06}
)

var (
	HIDCollectionPhysical    = []byte{hidCollection, 0x00}
	HIDCollectionApplication = []byte{hidCollection, 0x01}
	HIDCollectionEnd         = []byte{0xc0}
)

var (
	// Input (Data,Ary,Abs), Key arrays (6 bytes)
	HIDInputDataAryAbs = []byte{hidInput, 0x00}

	// Input (Data, Variable, Absolute), Modifier byte
	HIDInputDataVarAbs = []byte{hidInput, 0x02}

	// Input (Const,Var,Abs), Modifier byte
	HIDInputConstVarAbs = []byte{hidInput, 0x03}

	// Input (Data, Variable, Relative), 2 position bytes (X & Y)
	HIDInputDataVarRel = []byte{hidInput, 0x06}
)

func HIDReportSize(size int) []byte {
	return []byte{hidReportSize, byte(size)}
}

func HIDReportCount(count int) []byte {
	return []byte{hidReportCount, byte(count)}
}

func HIDReportID(id int) []byte {
	return []byte{hidReportID, byte(id)}
}

func HIDLogicalMinimum(min int) []byte {
	if min > 255 {
		result := []byte{hidLogicalMinimum + 1, 0x0, 0x0}
		binary.LittleEndian.PutUint16(result[1:3], uint16(min))
		return result
	}

	return []byte{hidLogicalMinimum, byte(min)}
}

func HIDLogicalMaximum(max int) []byte {
	if max > 255 {
		result := []byte{hidLogicalMaximum + 1, 0x0, 0x0}
		binary.LittleEndian.PutUint16(result[1:3], uint16(max))
		return result
	}

	return []byte{hidLogicalMaximum, byte(max)}
}

func HIDUsageMinimum(min int) []byte {
	if min > 255 {
		result := []byte{hidUsageMinimum + 1, 0x0, 0x0}
		binary.LittleEndian.PutUint16(result[1:3], uint16(min))
		return result
	}

	return []byte{hidUsageMinimum, byte(min)}
}

func HIDUsageMaximum(max int) []byte {
	if max > 255 {
		result := []byte{hidUsageMaximum + 1, 0x0, 0x0}
		binary.LittleEndian.PutUint16(result[1:3], uint16(max))
		return result
	}

	return []byte{hidUsageMaximum, byte(max)}
}

func HIDPhysicalMinimum(min int) []byte {
	if min > 255 {
		result := []byte{hidPhysicalMinimum + 1, 0x0, 0x0}
		binary.LittleEndian.PutUint16(result[1:3], uint16(min))
		return result
	}

	return []byte{hidPhysicalMinimum, byte(min)}
}

func HIDPhysicalMaximum(max int) []byte {
	if max > 255 {
		result := []byte{hidPhysicalMaximum + 1, 0x0, 0x0}
		binary.LittleEndian.PutUint16(result[1:3], uint16(max))
		return result
	}

	return []byte{hidPhysicalMaximum, byte(max)}
}

func HIDUnitExponent(exp int) []byte {
	return []byte{hidUnitExponent, byte(exp)}
}

func HIDUnit(unit int) []byte {
	return []byte{hidUnit, byte(unit)}
}
