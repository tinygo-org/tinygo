//go:build machine.psram && serial.usb

package runtime

// psramFailed stops without a message. The USB host cannot read one this
// early, and printing would link in the USB stack.
func psramFailed(err string) {
	abort()
}
