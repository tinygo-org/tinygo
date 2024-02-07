//go:build esp32

package runtime

import (
	"device"
	"device/esp"
	"machine"
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

	// Initialize UART.
	machine.InitSerial()

	// Initialize main system timer used for time.Now.
	initTimer()

	// Initialize the heap, call main.main, etc.
	run()

	// Fallback: if main ever returns, hang the CPU.
	exit(0)
}

//go:extern _sbss
var _sbss [0]byte

//go:extern _ebss
var _ebss [0]byte

func abort() {
	for {
		device.Asm("waiti 0")
	}
}
