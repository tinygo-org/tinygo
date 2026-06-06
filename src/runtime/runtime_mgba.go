//go:build gameboyadvance && mgbadebug

package runtime

import (
	_ "runtime/interrupt" // make sure the interrupt handler is defined
	"runtime/volatile"
	"unsafe"
)

var (
	// Setting this memory address to 0xC0DE enables mGBA's debug printing
	debugEnable = (*volatile.Register16)(unsafe.Pointer(uintptr(0x4FFF780)))
	// mGBA supports log levels from 0x100(fatal) to 0x104(debug)
	logLevel uint16 = 0x104 // use the debug log level
	// Once we are ready to output we set the debug flags register to logLevel
	// mGBA will output the text and then clear the text buffer when set
	debugFlags = (*volatile.Register16)(unsafe.Pointer(uintptr(0x4FFF700)))

	textBuffer = (*[255]byte)(unsafe.Pointer(uintptr(0x4FFF600)))
	index      = 0
)

func putchar(c byte) {
	if c == '\n' || index >= len(textBuffer) {
		debugFlags.Set(logLevel)
		index = 0

		// mGBA automatically prints a new line so we can ignore it
		if c == '\n' {
			return
		}
	}

	textBuffer[index] = c
	index++
}

func getchar() byte {
	// dummy, TODO
	return 0
}

func buffered() int {
	// dummy, TODO
	return 0
}

//go:extern _sbss
var _sbss [0]byte

//go:extern _ebss
var _ebss [0]byte

//go:extern _sdata
var _sdata [0]byte

//go:extern _sidata
var _sidata [0]byte

//go:extern _edata
var _edata [0]byte

// Entry point for Go. Initialize all packages and call main.main().
//
//export main
func main() {
	// Initialize .data and .bss sections.
	preinit()

	// Enable mGBA debugging.
	debugEnable.Set(0xC0DE)

	// Run program.
	run()
}

func preinit() {
	// Initialize .bss: zero-initialized global variables.
	ptr := unsafe.Pointer(&_sbss)
	for ptr != unsafe.Pointer(&_ebss) {
		*(*uint32)(ptr) = 0
		ptr = unsafe.Add(ptr, 4)
	}

	// Initialize .data: global variables initialized from flash.
	src := unsafe.Pointer(&_sidata)
	dst := unsafe.Pointer(&_sdata)
	for dst != unsafe.Pointer(&_edata) {
		*(*uint32)(dst) = *(*uint32)(src)
		dst = unsafe.Add(dst, 4)
		src = unsafe.Add(src, 4)
	}
}

func ticksToNanoseconds(ticks timeUnit) int64 {
	return int64(ticks)
}

func nanosecondsToTicks(ns int64) timeUnit {
	return timeUnit(ns)
}

func ticks() timeUnit {
	// TODO
	return 0
}

func sleepTicks(d timeUnit) {
	// TODO
}

func exit(code int) {
	abort()
}

func abort() {
	// TODO
	for {
	}
}
