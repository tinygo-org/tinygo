//go:build nrf && nrf52840

package runtime

// This package needs to be present so that the machine package can go:linkname
// EnableUSBCDC from it.
import _ "machine/usb/cdc"
