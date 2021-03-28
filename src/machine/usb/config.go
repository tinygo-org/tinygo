package usb

// The following constants must be defined for USB 2.0 host/device support.
//
// However, some or all of these constants may be unused in the core driver
// code, depending on which components are allocated by the USB driver.
//
// These values are __PLATFORM-AGNOSTIC__, and they serve two primary roles:
//   1. Aggregate and export values derived from platform-specific constants.
//   2. Configure core attributes defined by the USB 2.0 specification.
//
const (
	// ConfigPortCount defines the number of USB cores supported on this platform.
	ConfigPortCount = configDeviceCount + configHostCount
)

// ConfigDeviceDescriptor defines the fields used to populate host-enumerated
// device descriptors.
type ConfigDeviceDescriptor struct {
	Manufacturer string
	Product      string
	VID          uint16
	PID          uint16
	BCD          uint16
}
