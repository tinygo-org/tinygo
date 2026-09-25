//go:build esp32c6

package runtime

import (
	"device/esp"
)

// TIMG0 counts at 40MHz. The 80MHz timer clock is divided by 2.
const (
	timerDivider = 2
	nsPerTick    = 25
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

	// Change CPU frequency to 160MHz from SPLL (480MHz).
	//
	// Clock tree: SPLL (480MHz) → HP root → CPU / AHB / APB
	//
	// Set dividers BEFORE switching the clock source so the first PLL
	// cycle already arrives divided:
	//   HP root  = SPLL / (HS_DIV_NUM+1)      = 480 / 3 = 160 MHz
	//   CPU      = HP root / (CPU_HS_DIV_NUM+1) = 160 / 1 = 160 MHz
	//   AHB      = HP root / (AHB_HS_DIV_NUM+1) = 160 / 4 =  40 MHz
	//   APB      = AHB / (APB_HS_DIV_NUM+1)     =  40 / 1 =  40 MHz
	esp.PCR.CPU_FREQ_CONF.Set(0 << 8) // CPU_HS_DIV_NUM = 0 (div1)
	esp.PCR.AHB_FREQ_CONF.Set(3 << 8) // AHB_HS_DIV_NUM = 3 (div4)
	esp.PCR.APB_FREQ_CONF.Set(0 << 8) // APB_HS_DIV_NUM = 0 (div1)

	// Switch to PLL: SOC_CLK_SEL = 1 (SPLL), HS_DIV_NUM = 2 (div3).
	esp.PCR.SYSCLK_CONF.Set(1<<16 | 2<<8)

	// Select the Timer Group 0 timer clock source.
	//
	// The shared timekeeping code (runtime_esp32xx.go) assumes the TIMG0 timer
	// counts at 40MHz: an 80MHz source divided by the prescaler of 2 set in
	// initTimer, giving 25ns/tick. On the ESP32-C6 the timer group timer clock
	// is not the APB clock; it is selected via PCR and defaults to the 40MHz
	// XTAL (TG0_TIMER_CLK_SEL = 0). With the /2 prescaler that yields a 20MHz
	// tick, so every delay would be twice as long as intended (time.Sleep and
	// therefore blinky run at half speed).
	//
	// Select PLL_F80M (80MHz) as the source so the timer counts at 40MHz and
	// the 25ns/tick assumption holds. Clock source encoding (esp-idf
	// timer_ll_set_clock_source): 0 = XTAL, 1 = PLL_F80M, 2 = RC_FAST.
	esp.PCR.SetTIMERGROUP0_CONF_TG0_CLK_EN(1)
	esp.PCR.SetTIMERGROUP0_TIMER_CLK_CONF_TG0_TIMER_CLK_SEL(1)
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
