// +build esp32

package runtime

import (
	"device"
	"device/esp"
	"machine"
	"unsafe"
)

type timeUnit int64

var currentTime timeUnit

func putchar(c byte) {
	machine.UART0.WriteByte(c)
}

func postinit() {}

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
	preinit()

	// Initialize UART.
	machine.UART0.Configure(machine.UARTConfig{})

	// Configure timer 0 in timer group 0, for timekeeping.
	//   EN:       Enable the timer.
	//   INCREASE: Count up every tick (as opposed to counting down).
	//   DIVIDER:  16-bit prescaler, set to 2 for dividing the APB clock by two
	//             (40MHz).
	esp.TIMG0.T0CONFIG.Set(esp.TIMG_T0CONFIG_T0_EN | esp.TIMG_T0CONFIG_T0_INCREASE | 2<<esp.TIMG_T0CONFIG_T0_DIVIDER_Pos)

	// Set the timer counter value to 0.
	esp.TIMG0.T0LOADLO.Set(0)
	esp.TIMG0.T0LOADHI.Set(0)
	esp.TIMG0.T0LOAD.Set(0) // value doesn't matter.

	run()

	// Fallback: if main ever returns, hang the CPU.
	abort()
}

//go:extern _sbss
var _sbss [0]byte

//go:extern _ebss
var _ebss [0]byte

func preinit() {
	// Initialize .bss: zero-initialized global variables.
	// The .data section has already been loaded by the ROM bootloader.
	ptr := unsafe.Pointer(&_sbss)
	for ptr != unsafe.Pointer(&_ebss) {
		*(*uint32)(ptr) = 0
		ptr = unsafe.Pointer(uintptr(ptr) + 4)
	}
}

func ticks() timeUnit {
	// First, update the LO and HI register pair by writing any value to the
	// register. This allows reading the pair atomically.
	esp.TIMG0.T0UPDATE.Set(0)
	// Then read the two 32-bit parts of the timer.
	return timeUnit(uint64(esp.TIMG0.T0LO.Get()) | uint64(esp.TIMG0.T0HI.Get())<<32)
}

const asyncScheduler = false

func nanosecondsToTicks(ns int64) timeUnit {
	// Calculate the number of ticks from the number of nanoseconds. At a 80MHz
	// APB clock, that's 25 nanoseconds per tick with a timer prescaler of 2:
	// 25 = 1e9 / (80MHz / 2)
	return timeUnit(ns / 25)
}

func ticksToNanoseconds(ticks timeUnit) int64 {
	// See nanosecondsToTicks.
	return int64(ticks) * 25
}

// sleepTicks busy-waits until the given number of ticks have passed.
func sleepTicks(d timeUnit) {
	sleepUntil := ticks() + d
	for ticks() < sleepUntil {
		// TODO: suspend the CPU to not burn power here unnecessarily.
	}
}

func abort() {
	for {
		device.Asm("waiti 0")
	}
}
