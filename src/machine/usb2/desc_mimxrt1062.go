package usb2

// descCPUFrequencyHz defines the target CPU frequency (Hz).
const descCPUFrequencyHz = 600000000

// General USB device identification constants.
const (
	descVendorID  = 0xABCD
	descProductID = 0x1234
	descReleaseID = 0x0101

	descVendor  = "NXP Semiconductors"
	descProduct = "TinyGo USB"
)

// Constants for USB CDC-ACM device classes.
const (
	// descCDCACMConfigCount defines the number of USB cores that will be
	// configured as CDC-ACM (single) devices.
	descCDCACMConfigCount = 1

	descCDCACMMaxPower = 50 // 100 mA

	descCDCACMStatusPacketSize = 16
	descCDCACMDataRxPacketSize = descCDCACMDataRxFSPacketSize // full-speed
	descCDCACMDataTxPacketSize = descCDCACMDataTxFSPacketSize // full-speed

	descCDCACMDataRxFSPacketSize = 64  // full-speed
	descCDCACMDataTxFSPacketSize = 64  // full-speed
	descCDCACMDataRxHSPacketSize = 512 // high-speed
	descCDCACMDataTxHSPacketSize = 512 // high-speed
)
