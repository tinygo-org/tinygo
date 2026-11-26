//go:build machine.psram && !serial.usb

package runtime

// psramFailed prints err and stops. It does not panic, as the heap may not be ready.
func psramFailed(err string) {
	printstring("panic: ")
	printstring(err)
	printnl()
	abort()
}
