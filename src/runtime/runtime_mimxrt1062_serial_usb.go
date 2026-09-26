//go:build mimxrt1062 && serial.usb

package runtime

import "machine"

// serialReady reports whether machine.Serial can be used. With -serial usb
// it is a nil interface until machine.InitSerial runs from a package init.
//
//go:inline
func serialReady() bool {
	return machine.Serial != nil
}
