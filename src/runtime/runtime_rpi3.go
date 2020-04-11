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
	machine.Abort()
}

func putchar(c byte) {
	machine.MiniUART.WriteByte(c)
}


// abort is called by panic().
func abort() {
	machine.Abort()
}


// called once the memory allocator is ready and heap initialized
// this turns off the AuxInterrupt so we can configure the UART
func postinit() {
	// Initialize .bss: zero-initialized global variables.
	//ptr := unsafe.Pointer(&_sbss)
	//for ptr != unsafe.Pointer(&_ebss) {
	//	*(*uint32)(ptr) = 0
	//	ptr = unsafe.Pointer(uintptr(ptr) + 4)
	//}
}

//wasm workaround -- we will get coroutines
const asyncScheduler = false

//go:extern _sbss
var _sbss unsafe.Pointer

//go:extern _ebss
var _ebss unsafe.Pointer
