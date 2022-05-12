//go:build (sam && atsamd51) || (sam && atsame5x)
// +build sam,atsamd51 sam,atsame5x

package usb

// descCPUFrequencyHz defines the target CPU frequency (Hz).
const descCPUFrequencyHz = 120000000

// descCoreCount defines the number of USB PHY cores available on this platform,
// independent of the number of cores which shall be configured as TinyGo USB
// host/device controller instances.
const descCoreCount = 1 // SAMx51 has a single, full-speed USB PHY

// descCDCCount defines the number of USB cores that may be configured as
// CDC-ACM (single) devices.
const descCDCCount = 0

// descHIDCount defines the number of USB cores that may be configured as a
// composite (keyboard + mouse + joystick) human interface device (HID).
const descHIDCount = 1

// General USB device identification constants.
const (
	descCommonVendorID  = 0x03EB
	descCommonProductID = 0x2421
	descCommonReleaseID = 0x0101 // BCD (1.1)

	descCommonLanguage     = descLanguageEnglish
	descCommonManufacturer = "TinyGo"
	descCommonProduct      = "USB"
	descCommonSerialNumber = "00000"
)

// Constants for all USB device classes.
const (

	// USB endpoints parameters

	descMaxEndpoints = 8 // SAMx51 maximum number of endpoints

	descControlPacketSize = 64
)
