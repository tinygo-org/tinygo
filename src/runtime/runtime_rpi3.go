// +build rpi3

package runtime

import (
	"machine"
	"unsafe"
)

const tickMicros = int64(1)

type timeUnit int64

//go:export sleepticks sleepticks
func sleepTicks(n timeUnit) {

	machine.WaitMuSec(uint64(n))
}

func ticks() timeUnit {
	return timeUnit(machine.SystemTime())
}

//go:export main
func main() {
	run()
	Exit()
}

func putchar(c byte) {
	machine.MiniUART.WriteByte(c)
}

// abort is called by panic().
func abort() {
	machine.Abort()
}

func postinit() {
	// Initialize .bss: zero-initialized global variables.
	ptr := unsafe.Pointer(&_sbss)
	for ptr != unsafe.Pointer(&_ebss) {
		*(*uint32)(ptr) = 0
		ptr = unsafe.Pointer(uintptr(ptr) + 4)
	}
}

//go:extern _sbss
var _sbss [0]byte

//go:extern _ebss
var _ebss [0]byte

func Exit() {
	machine.MiniUART.WriteString("Program exited.\nDEADLOOP...")
}
