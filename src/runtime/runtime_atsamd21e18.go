//go:build sam && atsamd21 && atsamd21e18

package runtime

import (
	"device/sam"
)

func initSERCOMClocks() {
	// Turn on clock to SERCOM0 for UART0
	sam.PM.APBCMASK.SetBits(sam.PM_APBCMASK_SERCOM0_)

	// Use GCLK0 for SERCOM0 aka UART0
	// GCLK_CLKCTRL_ID( clockId ) | // Generic Clock 0 (SERCOMx)
	// GCLK_CLKCTRL_GEN_GCLK0 | // Generic Clock Generator 0 is source
	// GCLK_CLKCTRL_CLKEN ;
	sam.GCLK.CLKCTRL.Set((sam.GCLK_CLKCTRL_ID_SERCOM0_CORE << sam.GCLK_CLKCTRL_ID_Pos) |
		(sam.GCLK_CLKCTRL_GEN_GCLK0 << sam.GCLK_CLKCTRL_GEN_Pos) |
		sam.GCLK_CLKCTRL_CLKEN)
	waitForSync()

	// Turn on clock to SERCOM1
	sam.PM.APBCMASK.SetBits(sam.PM_APBCMASK_SERCOM1_)
	sam.GCLK.CLKCTRL.Set((sam.GCLK_CLKCTRL_ID_SERCOM1_CORE << sam.GCLK_CLKCTRL_ID_Pos) |
		(sam.GCLK_CLKCTRL_GEN_GCLK0 << sam.GCLK_CLKCTRL_GEN_Pos) |
		sam.GCLK_CLKCTRL_CLKEN)
	waitForSync()

	// Turn on clock to SERCOM2
	sam.PM.APBCMASK.SetBits(sam.PM_APBCMASK_SERCOM2_)
	sam.GCLK.CLKCTRL.Set((sam.GCLK_CLKCTRL_ID_SERCOM2_CORE << sam.GCLK_CLKCTRL_ID_Pos) |
		(sam.GCLK_CLKCTRL_GEN_GCLK0 << sam.GCLK_CLKCTRL_GEN_Pos) |
		sam.GCLK_CLKCTRL_CLKEN)
	waitForSync()

	// Turn on clock to SERCOM3
	sam.PM.APBCMASK.SetBits(sam.PM_APBCMASK_SERCOM3_)
	sam.GCLK.CLKCTRL.Set((sam.GCLK_CLKCTRL_ID_SERCOM3_CORE << sam.GCLK_CLKCTRL_ID_Pos) |
		(sam.GCLK_CLKCTRL_GEN_GCLK0 << sam.GCLK_CLKCTRL_GEN_Pos) |
		sam.GCLK_CLKCTRL_CLKEN)
	waitForSync()
}
