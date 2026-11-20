//go:build (rp2040 || rp2350) && (!boardpackage || (!serial.uart && !serial.usb))

package runtime

func serialNil() bool {
	return false
}
