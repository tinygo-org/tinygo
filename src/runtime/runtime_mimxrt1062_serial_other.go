//go:build mimxrt1062 && !serial.usb

package runtime

// serialReady reports whether machine.Serial can be used. The UART, RTT,
// and null serial are static values, so they are always usable.
//
//go:inline
func serialReady() bool {
	return true
}
