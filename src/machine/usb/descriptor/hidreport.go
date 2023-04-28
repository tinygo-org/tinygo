package descriptor

var (
	HIDUsagePageGenericDesktop     = []byte{0x05, 0x01}
	HIDUsagePageSimulationControls = []byte{0x05, 0x02}
	HIDUsagePageVRControls         = []byte{0x05, 0x03}
	HIDUsagePageSportControls      = []byte{0x05, 0x04}
	HIDUsagePageGameControls       = []byte{0x05, 0x05}
	HIDUsagePageGenericControls    = []byte{0x05, 0x06}
	HIDUsagePageKeyboard           = []byte{0x05, 0x07}
	HIDUsagePageLED                = []byte{0x05, 0x08}
	HIDUsagePageButton             = []byte{0x05, 0x09}
	HIDUsagePageOrdinal            = []byte{0x05, 0x0A}
	HIDUsagePageTelephony          = []byte{0x05, 0x0B}
	HIDUsagePageConsumer           = []byte{0x05, 0x0C}
	HIDUsagePageDigitizers         = []byte{0x05, 0x0D}
	HIDUsagePageHaptics            = []byte{0x05, 0x0E}
	HIDUsagePagePhysicalInput      = []byte{0x05, 0x0F}
	HIDUsagePageUnicode            = []byte{0x05, 0x10}
	HIDUsagePageSoC                = []byte{0x05, 0x11}
	HIDUsagePageEyeHeadTrackers    = []byte{0x05, 0x12}
	HIDUsagePageAuxDisplay         = []byte{0x05, 0x14}
	HIDUsagePageSensors            = []byte{0x05, 0x20}
	HIDUsagePageMedicalInstrument  = []byte{0x05, 0x40}
	HIDUsagePageBrailleDisplay     = []byte{0x05, 0x41}
	HIDUsagePageLighting           = []byte{0x05, 0x59}
	HIDUsagePageMonitor            = []byte{0x05, 0x80}
	HIDUsagePageMonitorEnum        = []byte{0x05, 0x81}
	HIDUsagePageVESA               = []byte{0x05, 0x82}
	HIDUsagePagePower              = []byte{0x05, 0x84}
	HIDUsagePageBatterySystem      = []byte{0x05, 0x85}
	HIDUsagePageBarcodeScanner     = []byte{0x05, 0x8C}
	HIDUsagePageScales             = []byte{0x05, 0x8D}
	HIDUsagePageMagneticStripe     = []byte{0x05, 0x8E}
	HIDUsagePageCameraControl      = []byte{0x05, 0x90}
	HIDUsagePageArcade             = []byte{0x05, 0x91}
	HIDUsagePageGaming             = []byte{0x05, 0x92}
)

var (
	HIDUsageDesktopPointer         = []byte{0x09, 0x01}
	HIDUsageDesktopMouse           = []byte{0x09, 0x02}
	HIDUsageDesktopJoystick        = []byte{0x09, 0x04}
	HIDUsageDesktopGamepad         = []byte{0x09, 0x05}
	HIDUsageDesktopKeyboard        = []byte{0x09, 0x06}
	HIDUsageDesktopKeypad          = []byte{0x09, 0x07}
	HIDUsageDesktopMultiaxis       = []byte{0x09, 0x08}
	HIDUsageDesktopTablet          = []byte{0x09, 0x09}
	HIDUsageDesktopWaterCooling    = []byte{0x09, 0x0A}
	HIDUsageDesktopChassis         = []byte{0x09, 0x0B}
	HIDUsageDesktopWireless        = []byte{0x09, 0x0C}
	HIDUsageDesktopPortable        = []byte{0x09, 0x0D}
	HIDUsageDesktopSystemMultiaxis = []byte{0x09, 0x0E}
	HIDUsageDesktopSpatial         = []byte{0x09, 0x0F}
	HIDUsageDesktopAssistive       = []byte{0x09, 0x10}
	HIDUsageDesktopDock            = []byte{0x09, 0x11}
	HIDUsageDesktopDockable        = []byte{0x09, 0x12}
	HIDUsageDesktopCallState       = []byte{0x09, 0x13}
	HIDUsageDesktopX               = []byte{0x09, 0x30}
	HIDUsageDesktopY               = []byte{0x09, 0x31}
	HIDUsageDesktopZ               = []byte{0x09, 0x32}
	HIDUsageDesktopRx              = []byte{0x09, 0x33}
	HIDUsageDesktopRy              = []byte{0x09, 0x34}
	HIDUsageDesktopRz              = []byte{0x09, 0x35}
	HIDUsageDesktopSlider          = []byte{0x09, 0x36}
	HIDUsageDesktopDial            = []byte{0x09, 0x37}
	HIDUsageDesktopWheel           = []byte{0x09, 0x38}
	HIDUsageDesktopHatSwitch       = []byte{0x09, 0x39}
	HIDUsageDesktopCountedBuffer   = []byte{0x09, 0x3A}
)

var (
	HIDUsageConsumerControl             = []byte{0x09, 0x01}
	HIDUsageConsumerNumericKeypad       = []byte{0x09, 0x02}
	HIDUsageConsumerProgrammableButtons = []byte{0x09, 0x03}
	HIDUsageConsumerMicrophone          = []byte{0x09, 0x04}
	HIDUsageConsumerHeadphone           = []byte{0x09, 0x05}
	HIDUsageConsumerGraphicEqualizer    = []byte{0x09, 0x06}
)

var (
	HIDCollectionPhysical    = []byte{0xa1, 0x00}
	HIDCollectionApplication = []byte{0xa1, 0x01}
)

var (
	HIDEndCollection = []byte{0xc0}
)

var (
	// Input (Data,Ary,Abs), Key arrays (6 bytes)
	HIDInputDataAryAbs = []byte{0x81, 0x00}

	// Input (Data, Variable, Absolute), Modifier byte
	HIDInputDataVarAbs = []byte{0x81, 0x02}

	// Input (Const,Var,Abs), Modifier byte
	HIDInputConstVarAbs = []byte{0x81, 0x03}

	// Input (Data, Variable, Relative), 2 position bytes (X & Y)
	HIDInputDataVarRel = []byte{0x81, 0x06}
)

func HIDReportSize(size int) []byte {
	return []byte{0x75, byte(size)}
}

func HIDReportCount(count int) []byte {
	return []byte{0x95, byte(count)}
}

func HIDReportID(id int) []byte {
	return []byte{0x85, byte(id)}
}

func HIDLogicalMinimum(min int) []byte {
	return []byte{0x15, byte(min)}
}

func HIDLogicalMaximum(max int) []byte {
	return []byte{0x25, byte(max)}
}

func HIDUsageMinimum(min int) []byte {
	return []byte{0x19, byte(min)}
}

func HIDUsageMaximum(max int) []byte {
	return []byte{0x29, byte(max)}
}
