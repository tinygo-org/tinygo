// +build rpi3_qemu

package runtime

import (
	"device/arm"
	"unsafe"
)

const tickMicros = int64(100000) //centiseconds is provided by clock

type timeUnit int64

//go:export sleepticks sleepticks
func sleepTicks(n timeUnit) {
	start := Semihostingv2Call(uint64(Semihostingv2OpClock), 0)
	centis := uint64(n) / uint64(tickMicros)
	current := start
	for current-start < centis { //busy wait
		for i := 0; i < 20; i++ {
			arm.Asm("nop")
		}
		current = Semihostingv2Call(uint64(Semihostingv2OpClock), 0)
	}
}

func ticks() timeUnit {
	current := Semihostingv2Call(uint64(Semihostingv2OpClock), 0)
	return timeUnit(current * uint64(tickMicros))
}

//go:export main
func main() {
	run()
	Exit()
}

func putchar(c byte) {
	Semihostingv2Call(uint64(Semihostingv2OpWriteC), uint64(uintptr(unsafe.Pointer(&c))))

}

//go:extern _sbss
var _sbss [0]byte

//go:extern _ebss
var _ebss [0]byte

func abort() {
	Semihostingv2Call(uint64(Semihostingv2OpExit), uint64(Semihostingv2StopRuntimeErrorUnknown))
}

func Exit() {
	Semihostingv2Call(uint64(Semihostingv2OpExit), uint64(Semihostingv2StopApplicationExit))
}

func postinit() {
	// Initialize .bss: zero-initialized global variables.
	ptr := unsafe.Pointer(&_sbss)
	for ptr != unsafe.Pointer(&_ebss) {
		*(*uint32)(ptr) = 0
		ptr = unsafe.Pointer(uintptr(ptr) + 4)
	}
}

const asyncScheduler = false

type SemiHostingOp uint64

const (
	Semihostingv2OpExit   SemiHostingOp = 0x18
	Semihostingv2OpClock  SemiHostingOp = 0x10
	Semihostingv2OpWriteC SemiHostingOp = 0x03
)

type SemihostingStopCode int

const (
	Semihostingv2StopBreakpoint          SemihostingStopCode = 0x20020
	Semihostingv2StopWatchpoint          SemihostingStopCode = 0x20021
	Semihostingv2StopStepComplete        SemihostingStopCode = 0x20022
	Semihostingv2StopRuntimeErrorUnknown SemihostingStopCode = 0x20023
	Semihostingv2StopInternalError       SemihostingStopCode = 0x20024
	Semihostingv2StopUserInterruption    SemihostingStopCode = 0x20025
	Semihostingv2StopApplicationExit     SemihostingStopCode = 0x20026
	Semihostingv2StopStackOverflow       SemihostingStopCode = 0x20027
	Semihostingv2StopDivisionByZero      SemihostingStopCode = 0x20028
	Semihostingv2StopOSSpecific          SemihostingStopCode = 0x20029
)

//go:linkname semihosting_param_block semihosting.semiHostingParamBlock
var semihostingParamBlock uint64

//semihosting_call (the second param may be a value or a pointer)
//if it is a pointer, it will point to semihosting_param_block
//export semihosting_call
func Semihostingv2Call(op uint64, param uint64) uint64

//Semihostingv2ClockMicros is supposed to return micros since the program
//started. However, the underlying call is supposed to return centiseconds
//and it appears to be wildly inaccurate.
func Semihostingv2ClockMicros() uint64 {
	centis := Semihostingv2Call(uint64(Semihostingv2OpClock), 0)
	return centis * 100000
}
