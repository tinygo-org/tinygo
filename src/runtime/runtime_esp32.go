//go:build esp32

package runtime

import (
	"device"
	"device/esp"
	"machine"
	"unsafe"
)

// This is the function called on startup right after the stack pointer has been
// set.
//
//export main
func main() {
	// Disable the protection on the watchdog timer (needed when started from
	// the bootloader).
	esp.RTC_CNTL.WDTWPROTECT.Set(0x050D83AA1)

	// Disable both watchdog timers that are enabled by default on startup.
	// Note that these watchdogs can be protected, but the ROM bootloader
	// doesn't seem to protect them.
	esp.RTC_CNTL.WDTCONFIG0.Set(0)
	esp.TIMG0.WDTCONFIG0.Set(0)

	// Switch SoC clock source to PLL (instead of the default which is XTAL).
	// This switches the CPU (and APB) clock from 40MHz to 80MHz.
	// Options:
	//   RTC_CNTL_CLK_CONF_SOC_CLK_SEL:       PLL (1)       (default XTAL)
	//   RTC_CNTL_CLK_CONF_CK8M_DIV_SEL:      2             (default)
	//   RTC_CNTL_CLK_CONF_DIG_CLK8M_D256_EN: Enable        (default)
	//   RTC_CNTL_CLK_CONF_CK8M_DIV:          divide by 256 (default)
	// The only real change made here is modifying RTC_CNTL_CLK_CONF_SOC_CLK_SEL,
	// but setting a fixed value produces smaller code.
	esp.RTC_CNTL.CLK_CONF.Set((1 << esp.RTC_CNTL_CLK_CONF_SOC_CLK_SEL_Pos) |
		(2 << esp.RTC_CNTL_CLK_CONF_CK8M_DIV_SEL_Pos) |
		(1 << esp.RTC_CNTL_CLK_CONF_DIG_CLK8M_D256_EN_Pos) |
		(1 << esp.RTC_CNTL_CLK_CONF_CK8M_DIV_Pos))

	// Switch CPU from 80MHz to 160MHz. This doesn't affect the APB clock,
	// which is still running at 80MHz.
	esp.DPORT.CPU_PER_CONF.Set(1) // PLL_CLK / 2, see table 3-3 in the reference manual

	// Clear .bss section. .data has already been loaded by the ROM bootloader.
	// Do this after increasing the CPU clock to possibly make startup slightly
	// faster.
	clearbss()

	// Initialize main system timer used for time.Now.
	initTimer()

	// Set up the Xtensa interrupt vector table. This zeroes INTENABLE, so it
	// must run before any peripheral (UART, timer, etc) enables its own CPU
	// interrupt line - otherwise that enable would be wiped out here.
	interruptInit()

	// Initialize timer alarm interrupt for the scheduler.
	initTimerInterrupt()

	// Initialize UART. This enables the UART RX interrupt, which must happen
	// after interruptInit so the INTENABLE bit is not cleared again.
	machine.InitSerial()

	// Initialize the heap, call main.main, etc.
	run()

	// Fallback: if main ever returns, hang the CPU.
	exit(0)
}

//go:extern _sbss
var _sbss [0]byte

//go:extern _ebss
var _ebss [0]byte

//go:extern _vector_table
var _vector_table [0]uintptr

// interruptInit installs the Xtensa vector table by writing its address
// to the VECBASE special register and ensures all CPU interrupts are
// initially disabled.
func interruptInit() {
	// Disable all CPU interrupts while we configure.
	device.AsmFull("wsr {zero}, INTENABLE", map[string]interface{}{
		"zero": uintptr(0),
	})

	// Write the vector table address to VECBASE (SR 231).
	vecbase := uintptr(unsafe.Pointer(&_vector_table))
	device.AsmFull("wsr {vecbase}, VECBASE", map[string]interface{}{
		"vecbase": vecbase,
	})

	// Clear PS.EXCM and PS.INTLEVEL so that level-1 interrupts can fire.
	// The ROM bootloader leaves PS.EXCM=1 (exception mode), which masks
	// all interrupts at level ≤ EXCMLEVEL (level 1 on ESP32).
	// PS.INTLEVEL may also be non-zero. Both must be 0 for peripheral
	// interrupts to trigger.
	//
	// We also set PS.UM=1 (bit 5) so that level-1 interrupts route to
	// the User exception vector at VECBASE+0x340, where our handler lives.
	// With PS.UM=0 (the ROM default), they would go to the Kernel exception
	// vector at VECBASE+0x300 which is a reset stub.
	ps := uintptr(device.AsmFull("rsr {}, PS", nil))
	ps &^= 0x1F // clear INTLEVEL (bits 0-3) and EXCM (bit 4)
	ps |= 0x20  // set PS.UM (bit 5) — use User exception vector
	device.AsmFull("wsr {ps}, PS", map[string]interface{}{
		"ps": ps,
	})

	// Synchronize pipeline after writing special registers.
	device.Asm("rsync")
}
