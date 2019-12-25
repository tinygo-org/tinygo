// +build fe310

// This file implements target-specific things for the FE310 chip as used in the
// HiFive1.

package runtime

import (
	"machine"
	"unsafe"

	"device/sifive"
)

type timeUnit int64

const tickMicros = 32768 // RTC runs at 32.768kHz

//go:extern _sbss
var _sbss unsafe.Pointer

//go:extern _ebss
var _ebss unsafe.Pointer

//go:extern _sdata
var _sdata unsafe.Pointer

//go:extern _sidata
var _sidata unsafe.Pointer

//go:extern _edata
var _edata unsafe.Pointer

//go:export main
func main() {
	preinit()
	initAll()
	callMain()
	abort()
}

func init() {
	pric_init()
	machine.UART0.Configure(machine.UARTConfig{})
}

func pric_init() {
	// Make sure the HFROSC is on
	sifive.PRCI.HFROSCCFG.SetBits(sifive.PRCI_HFROSCCFG_ENABLE)

	// Run off 16 MHz Crystal for accuracy.
	sifive.PRCI.PLLCFG.SetBits(sifive.PRCI_PLLCFG_REFSEL | sifive.PRCI_PLLCFG_BYPASS)
	sifive.PRCI.PLLCFG.SetBits(sifive.PRCI_PLLCFG_SEL)

	// Turn off HFROSC to save power
	sifive.PRCI.HFROSCCFG.ClearBits(sifive.PRCI_HFROSCCFG_ENABLE)

	// Enable the RTC.
	sifive.RTC.RTCCFG.Set(sifive.RTC_RTCCFG_ENALWAYS)
}

func preinit() {
	// Initialize .bss: zero-initialized global variables.
	ptr := unsafe.Pointer(&_sbss)
	for ptr != unsafe.Pointer(&_ebss) {
		*(*uint32)(ptr) = 0
		ptr = unsafe.Pointer(uintptr(ptr) + 4)
	}

	// Initialize .data: global variables initialized from flash.
	src := unsafe.Pointer(&_sidata)
	dst := unsafe.Pointer(&_sdata)
	for dst != unsafe.Pointer(&_edata) {
		*(*uint32)(dst) = *(*uint32)(src)
		dst = unsafe.Pointer(uintptr(dst) + 4)
		src = unsafe.Pointer(uintptr(src) + 4)
	}
}

func putchar(c byte) {
	machine.UART0.WriteByte(c)
}

const asyncScheduler = false
