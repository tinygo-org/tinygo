//go:build sam && atsamd51 && atsamd51p20
// +build sam,atsamd51,atsamd51p20

package runtime

import (
	"tinygo.org/x/device/sam"
)

func initSERCOMClocks() {
	// Turn on clock to SERCOM0 for UART0
	sam.MCLK.APBAMASK.SetBits(sam.MCLK_APBAMASK_SERCOM0_)
	sam.GCLK.PCHCTRL[sam.PCHCTRL_GCLK_SERCOM0_CORE].Set((sam.GCLK_PCHCTRL_GEN_GCLK1 << sam.GCLK_PCHCTRL_GEN_Pos) |
		sam.GCLK_PCHCTRL_CHEN)

	// sets the "slow" clock shared by all SERCOM
	sam.GCLK.PCHCTRL[sam.PCHCTRL_GCLK_SERCOMX_SLOW].Set((sam.GCLK_PCHCTRL_GEN_GCLK1 << sam.GCLK_PCHCTRL_GEN_Pos) |
		sam.GCLK_PCHCTRL_CHEN)

	// Turn on clock to SERCOM1
	sam.MCLK.APBAMASK.SetBits(sam.MCLK_APBAMASK_SERCOM1_)
	sam.GCLK.PCHCTRL[sam.PCHCTRL_GCLK_SERCOM1_CORE].Set((sam.GCLK_PCHCTRL_GEN_GCLK1 << sam.GCLK_PCHCTRL_GEN_Pos) |
		sam.GCLK_PCHCTRL_CHEN)

	// Turn on clock to SERCOM2
	sam.MCLK.APBBMASK.SetBits(sam.MCLK_APBBMASK_SERCOM2_)
	sam.GCLK.PCHCTRL[sam.PCHCTRL_GCLK_SERCOM2_CORE].Set((sam.GCLK_PCHCTRL_GEN_GCLK1 << sam.GCLK_PCHCTRL_GEN_Pos) |
		sam.GCLK_PCHCTRL_CHEN)

	// Turn on clock to SERCOM3
	sam.MCLK.APBBMASK.SetBits(sam.MCLK_APBBMASK_SERCOM3_)
	sam.GCLK.PCHCTRL[sam.PCHCTRL_GCLK_SERCOM3_CORE].Set((sam.GCLK_PCHCTRL_GEN_GCLK1 << sam.GCLK_PCHCTRL_GEN_Pos) |
		sam.GCLK_PCHCTRL_CHEN)

	// Turn on clock to SERCOM4
	sam.MCLK.APBDMASK.SetBits(sam.MCLK_APBDMASK_SERCOM4_)
	sam.GCLK.PCHCTRL[sam.PCHCTRL_GCLK_SERCOM4_CORE].Set((sam.GCLK_PCHCTRL_GEN_GCLK1 << sam.GCLK_PCHCTRL_GEN_Pos) |
		sam.GCLK_PCHCTRL_CHEN)

	// Turn on clock to SERCOM5
	sam.MCLK.APBDMASK.SetBits(sam.MCLK_APBDMASK_SERCOM5_)
	sam.GCLK.PCHCTRL[sam.PCHCTRL_GCLK_SERCOM5_CORE].Set((sam.GCLK_PCHCTRL_GEN_GCLK1 << sam.GCLK_PCHCTRL_GEN_Pos) |
		sam.GCLK_PCHCTRL_CHEN)

	// Turn on clock to SERCOM6
	sam.MCLK.APBDMASK.SetBits(sam.MCLK_APBDMASK_SERCOM6_)
	sam.GCLK.PCHCTRL[sam.PCHCTRL_GCLK_SERCOM6_CORE].Set((sam.GCLK_PCHCTRL_GEN_GCLK1 << sam.GCLK_PCHCTRL_GEN_Pos) |
		sam.GCLK_PCHCTRL_CHEN)

	// Turn on clock to SERCOM7
	sam.MCLK.APBDMASK.SetBits(sam.MCLK_APBDMASK_SERCOM7_)
	sam.GCLK.PCHCTRL[sam.PCHCTRL_GCLK_SERCOM7_CORE].Set((sam.GCLK_PCHCTRL_GEN_GCLK1 << sam.GCLK_PCHCTRL_GEN_Pos) |
		sam.GCLK_PCHCTRL_CHEN)
}
