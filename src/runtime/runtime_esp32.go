//go:build esp32
// +build esp32

package runtime

import (
	"device"
	"device/esp"
	"machine"
)

// This is the function called on startup right after the stack pointer has been
// set.
//export main
func main() {
	// Disable both watchdog timers that are enabled by default on startup.
	// Note that these watchdogs can be protected, but the ROM bootloader
	// doesn't seem to protect them.
	esp.RTCCNTL.WDTCONFIG0.Set(0)
	esp.TIMG0.WDTCONFIG0.Set(0)

	// Switch SoC clock source to PLL (instead of the default which is XTAL).
	// This switches the CPU (and APB) clock from 40MHz to 80MHz.
	// Options:
	//   RTCCNTL_CLK_CONF_SOC_CLK_SEL:       PLL    (default XTAL)
	//   RTCCNTL_CLK_CONF_CK8M_DIV_SEL:      2      (default)
	//   RTCCNTL_CLK_CONF_DIG_CLK8M_D256_EN: Enable (default)
	//   RTCCNTL_CLK_CONF_CK8M_DIV:          DIV256 (default)
	// The only real change made here is modifying RTCCNTL_CLK_CONF_SOC_CLK_SEL,
	// but setting a fixed value produces smaller code.
	esp.RTCCNTL.CLK_CONF.Set((esp.RTCCNTL_CLK_CONF_SOC_CLK_SEL_PLL << esp.RTCCNTL_CLK_CONF_SOC_CLK_SEL_Pos) |
		(2 << esp.RTCCNTL_CLK_CONF_CK8M_DIV_SEL_Pos) |
		(esp.RTCCNTL_CLK_CONF_DIG_CLK8M_D256_EN_Enable << esp.RTCCNTL_CLK_CONF_DIG_CLK8M_D256_EN_Pos) |
		(esp.RTCCNTL_CLK_CONF_CK8M_DIV_DIV256 << esp.RTCCNTL_CLK_CONF_CK8M_DIV_Pos))

	// Switch CPU from 80MHz to 160MHz. This doesn't affect the APB clock,
	// which is still running at 80MHz.
	esp.DPORT.CPU_PER_CONF.Set(esp.DPORT_CPU_PER_CONF_CPUPERIOD_SEL_SEL_160)

	// Clear .bss section. .data has already been loaded by the ROM bootloader.
	// Do this after increasing the CPU clock to possibly make startup slightly
	// faster.
	clearbss()

	// Initialize UART.
	machine.Serial.Configure(machine.UARTConfig{})

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
