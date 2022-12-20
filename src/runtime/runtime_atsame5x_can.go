//go:build (sam && atsame51) || (sam && atsame54)

package runtime

import (
	"device/sam"
)

func init() {
	initCANClock()
}

func initCANClock() {
	// Turn on clocks for CAN0/CAN1.
	sam.MCLK.AHBMASK.SetBits(sam.MCLK_AHBMASK_CAN0_)
	sam.MCLK.AHBMASK.SetBits(sam.MCLK_AHBMASK_CAN1_)

	// Put Generic Clock Generator 1 as source for USB
	sam.GCLK.PCHCTRL[sam.PCHCTRL_GCLK_CAN0].Set((sam.GCLK_PCHCTRL_GEN_GCLK1 << sam.GCLK_PCHCTRL_GEN_Pos) |
		sam.GCLK_PCHCTRL_CHEN)
	sam.GCLK.PCHCTRL[sam.PCHCTRL_GCLK_CAN1].Set((sam.GCLK_PCHCTRL_GEN_GCLK1 << sam.GCLK_PCHCTRL_GEN_Pos) |
		sam.GCLK_PCHCTRL_CHEN)
}
