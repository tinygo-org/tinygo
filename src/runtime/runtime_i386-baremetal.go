// +build 386,baremetal

package runtime

import (
	"device/x86"
	"machine"
)

type timeUnit int64

const tickMicros = 1

//export main
func main() {
	initConsole()

	// Run all init functions.
	initAll()

	// Run main.main().
	callMain()

	machine.Halt()
}

func initConsole() {
	machine.COM1.Configure()
}

func putchar(c byte) {
	machine.COM1.WriteByte(c)
}

func ticks() timeUnit {
	return 0
}

const asyncScheduler = false

func sleepTicks(d timeUnit) {
	// TODO
}

func abort() {
	// Try to shutdown.
	machine.Halt()
}

func getCurrentStackPointer() uintptr {
	return x86.ReadRegister("%esp")
}
