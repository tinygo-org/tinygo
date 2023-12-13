//go:build esp32s3

package runtime

import (
	"device/esp"
)

// This is the function called on startup after the flash (IROM/DROM) is
// initialized and the stack pointer has been set.
//
//export main
func main() {
	// This initialization configures the following things:
	// * It disables all watchdog timers. They might be useful at some point in
	//   the future, but will need integration into the scheduler. For now,
	//   they're all disabled.
	// * It sets the CPU frequency to 160MHz, which is the maximum speed allowed
	//   for this CPU. Lower frequencies might be possible in the future, but
	//   running fast and sleeping quickly is often also a good strategy to save
	//   power.
	// TODO: protect certain memory regions, especially the area below the stack
	// to protect against stack overflows. See
	// esp_cpu_configure_region_protection in ESP-IDF.

	// Disable RTC watchdog.
	esp.RTC_CNTL.RTC_WDTWPROTECT.Set(0x50D83AA1)
	esp.RTC_CNTL.RTC_WDTCONFIG0.Set(0)
	esp.RTC_CNTL.RTC_WDTWPROTECT.Set(0x0) // Re-enable write protect

	// Disable Timer 0 watchdog.
	esp.TIMG1.WDTWPROTECT.Set(0x50D83AA1) // write protect
	esp.TIMG1.WDTCONFIG0.Set(0)           // disable TG0 WDT
	esp.TIMG1.WDTWPROTECT.Set(0x0)        // Re-enable write protect

	esp.TIMG0.WDTWPROTECT.Set(0x50D83AA1) // write protect
	esp.TIMG0.WDTCONFIG0.Set(0)           // disable TG0 WDT
	esp.TIMG0.WDTWPROTECT.Set(0x0)        // Re-enable write protect

	// Disable super watchdog.
	esp.RTC_CNTL.RTC_SWD_WPROTECT.Set(0x8F1D312A)
	esp.RTC_CNTL.RTC_SWD_CONF.Set(esp.RTC_CNTL_RTC_SWD_CONF_SWD_DISABLE)
	esp.RTC_CNTL.RTC_SWD_WPROTECT.Set(0x0) // Re-enable write protect

	// // Change CPU frequency from 20MHz to 80MHz, by switching from the XTAL to
	// // the PLL clock source (see table "CPU Clock Frequency" in the reference
	// // manual).
	// esp.SYSTEM.SYSCLK_CONF.Set(1 << esp.SYSTEM_SYSCLK_CONF_SOC_CLK_SEL_Pos)

	// // Change CPU frequency from 80MHz to 160MHz by setting SYSTEM_CPUPERIOD_SEL
	// // to 1 (see table "CPU Clock Frequency" in the reference manual).
	// // Note: we might not want to set SYSTEM_CPU_WAIT_MODE_FORCE_ON to save
	// // power. It is set here to keep the default on reset.
	// esp.SYSTEM.CPU_PER_CONF.Set(esp.SYSTEM_CPU_PER_CONF_CPU_WAIT_MODE_FORCE_ON | esp.SYSTEM_CPU_PER_CONF_PLL_FREQ_SEL | 1<<esp.SYSTEM_CPU_PER_CONF_CPUPERIOD_SEL_Pos)

	clearbss()

	// Initialize main system timer used for time.Now.
	initTimer()

	// Initialize the heap, call main.main, etc.
	run()

	// Fallback: if main ever returns, hang the CPU.
	exit(0)
}

func abort() {
	// lock up forever
	print("abort called\n")
}

//go:extern _vector_table
var _vector_table [0]uintptr

//go:extern _sbss
var _sbss [0]byte

//go:extern _ebss
var _ebss [0]byte
