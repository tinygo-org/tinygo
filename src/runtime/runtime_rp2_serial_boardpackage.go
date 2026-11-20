//go:build (rp2040 || rp2350) && boardpackage && (serial.uart || serial.usb)

package runtime

import "machine"

func serialNil() bool {
	return machine.Serial == nil
}
