// +build sam,atsamd51,atsamd51j20

package runtime

import (
	"device/sam"
)

func initSERCOMClocks() {
	// Turn on clock to SERCOM0 for UART0
	sam.MCLK.APBAMASK.SetBits(sam.MCLK_APBAMASK_SERCOM0_)
	sam.GCLK.PCHCTRL[7].Set((sam.GCLK_PCHCTRL_GEN_GCLK1 << sam.GCLK_PCHCTRL_GEN_Pos) |
		sam.GCLK_PCHCTRL_CHEN)

	// sets the "slow" clock shared by all SERCOM
	sam.GCLK.PCHCTRL[3].Set((sam.GCLK_PCHCTRL_GEN_GCLK1 << sam.GCLK_PCHCTRL_GEN_Pos) |
		sam.GCLK_PCHCTRL_CHEN)

	// Turn on clock to SERCOM1
	sam.MCLK.APBAMASK.SetBits(sam.MCLK_APBAMASK_SERCOM1_)
	sam.GCLK.PCHCTRL[8].Set((sam.GCLK_PCHCTRL_GEN_GCLK1 << sam.GCLK_PCHCTRL_GEN_Pos) |
		sam.GCLK_PCHCTRL_CHEN)

	// Turn on clock to SERCOM2
	sam.MCLK.APBBMASK.SetBits(sam.MCLK_APBBMASK_SERCOM2_)
	sam.GCLK.PCHCTRL[23].Set((sam.GCLK_PCHCTRL_GEN_GCLK1 << sam.GCLK_PCHCTRL_GEN_Pos) |
		sam.GCLK_PCHCTRL_CHEN)

	// Turn on clock to SERCOM3
	sam.MCLK.APBBMASK.SetBits(sam.MCLK_APBBMASK_SERCOM3_)
	sam.GCLK.PCHCTRL[24].Set((sam.GCLK_PCHCTRL_GEN_GCLK1 << sam.GCLK_PCHCTRL_GEN_Pos) |
		sam.GCLK_PCHCTRL_CHEN)

	// Turn on clock to SERCOM4
	sam.MCLK.APBDMASK.SetBits(sam.MCLK_APBDMASK_SERCOM4_)
	sam.GCLK.PCHCTRL[34].Set((sam.GCLK_PCHCTRL_GEN_GCLK1 << sam.GCLK_PCHCTRL_GEN_Pos) |
		sam.GCLK_PCHCTRL_CHEN)

	// Turn on clock to SERCOM5
	sam.MCLK.APBDMASK.SetBits(sam.MCLK_APBDMASK_SERCOM5_)
	sam.GCLK.PCHCTRL[35].Set((sam.GCLK_PCHCTRL_GEN_GCLK1 << sam.GCLK_PCHCTRL_GEN_Pos) |
		sam.GCLK_PCHCTRL_CHEN)
}
