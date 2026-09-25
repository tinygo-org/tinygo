//go:build esp32h2

package runtime

import (
	"device/esp"
)

// TIMG0 counts at 8MHz. The 32MHz XTAL is divided by 4.
const (
	timerDivider = 4
	nsPerTick    = 125
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
	// * It sets the CPU frequency to 32MHz from the XTAL.

	// Disable Timer Group 0 watchdog (unlock first).
	esp.TIMG0.WDTWPROTECT.Set(0x50D83AA1)
	esp.TIMG0.WDTCONFIG0.Set(0)

	// Disable Timer Group 1 watchdog (unlock first).
	esp.TIMG1.WDTWPROTECT.Set(0x50D83AA1)
	esp.TIMG1.WDTCONFIG0.Set(0)

	// Disable LP watchdog (write-protect key first).
	esp.LP_WDT.WDTWPROTECT.Set(0x50D83AA1)
	esp.LP_WDT.WDTCONFIG0.Set(0)

	// Disable super watchdog.
	esp.LP_WDT.SWD_WPROTECT.Set(0x50D83AA1)
	esp.LP_WDT.SWD_CONF.SetBits(1 << 30) // SWD_DISABLE bit

	// Run CPU, AHB and APB at 32MHz from the XTAL. The field values are divider-1.
	// See ESP-IDF esp_hw_support/port/esp32h2/rtc_clk.c rtc_clk_cpu_freq_to_xtal.
	esp.PCR.SetCPU_FREQ_CONF_CPU_DIV_NUM(0)
	esp.PCR.SetAHB_FREQ_CONF_AHB_DIV_NUM(0)
	esp.PCR.SetAPB_FREQ_CONF_APB_DIV_NUM(0)
	esp.PCR.SetSYSCLK_CONF_SOC_CLK_SEL(0)
	esp.PCR.SetBUS_CLK_UPDATE_BUS_CLOCK_UPDATE(1)
	for esp.PCR.GetBUS_CLK_UPDATE_BUS_CLOCK_UPDATE() != 0 {
	}

	// Use the 32MHz XTAL as the TIMG0 clock. 0 = XTAL, 1 = RC_FAST, 2 = PLL_F48M.
	// See ESP-IDF esp_hal_timg/esp32h2/include/hal/timer_ll.h.
	esp.PCR.SetTIMERGROUP0_CONF_TG0_CLK_EN(1)
	esp.PCR.SetTIMERGROUP0_TIMER_CLK_CONF_TG0_TIMER_CLK_SEL(0)
	esp.PCR.SetTIMERGROUP0_TIMER_CLK_CONF_TG0_TIMER_CLK_EN(1)

	clearbss()

	// Configure interrupt handler
	interruptInit()

	// Initialize main system timer used for time.Now.
	initTimer()

	// Initialize timer alarm interrupt for the scheduler.
	initTimerInterrupt()

	// Initialize the heap, call main.main, etc.
	run()

	// Fallback: if main ever returns, hang the CPU.
	exit(0)
}
