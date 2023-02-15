//go:build rp2040

package machine

import (
	"device/rp"
	"runtime/volatile"
	"unsafe"
)

type pll struct {
	cs       volatile.Register32
	pwr      volatile.Register32
	fbDivInt volatile.Register32
	prim     volatile.Register32
}

var (
	pllSys = (*pll)(unsafe.Pointer(rp.PLL_SYS))
	pllUSB = (*pll)(unsafe.Pointer(rp.PLL_USB))
)

// init initializes pll (Sys or USB) given the following parameters.
//
// Input clock divider, refdiv.
//
// Requested output frequency from the VCO (voltage controlled oscillator), vcoFreq.
//
// Post Divider 1, postDiv1 with range 1-7 and be >= postDiv2.
//
// Post Divider 2, postDiv2 with range 1-7.
func (pll *pll) init(refdiv, vcoFreq, postDiv1, postDiv2 uint32) {
	refFreq := xoscFreq / refdiv

	// What are we multiplying the reference clock by to get the vco freq
	// (The regs are called div, because you divide the vco output and compare it to the refclk)
	fbdiv := vcoFreq / (refFreq * MHz)

	// Check fbdiv range
	if !(fbdiv >= 16 && fbdiv <= 320) {
		panic("fbdiv should be in the range [16,320]")
	}

	// Check divider ranges
	if !((postDiv1 >= 1 && postDiv1 <= 7) && (postDiv2 >= 1 && postDiv2 <= 7)) {
		panic("postdiv1, postdiv1 should be in the range [1,7]")
	}

	// postDiv1 should be >= postDiv2
	// from appnote page 11
	// postdiv1 is designed to operate with a higher input frequency
	// than postdiv2
	if postDiv1 < postDiv2 {
		panic("postdiv1 should be greater than or equal to postdiv2")
	}

	// Check that reference frequency is no greater than vco / 16
	if refFreq > vcoFreq/16 {
		panic("reference frequency should not be greater than vco frequency divided by 16")
	}

	// div1 feeds into div2 so if div1 is 5 and div2 is 2 then you get a divide by 10
	pdiv := postDiv1<<rp.PLL_SYS_PRIM_POSTDIV1_Pos | postDiv2<<rp.PLL_SYS_PRIM_POSTDIV2_Pos

	if pll.cs.HasBits(rp.PLL_SYS_CS_LOCK) &&
		refdiv == pll.cs.Get()&rp.PLL_SYS_CS_REFDIV_Msk &&
		fbdiv == pll.fbDivInt.Get()&rp.PLL_SYS_FBDIV_INT_FBDIV_INT_Msk &&
		pdiv == pll.prim.Get()&(rp.PLL_SYS_PRIM_POSTDIV1_Msk&rp.PLL_SYS_PRIM_POSTDIV2_Msk) {
		// do not disrupt PLL that is already correctly configured and operating
		return
	}

	var pllRst uint32
	if pll == pllSys {
		pllRst = rp.RESETS_RESET_PLL_SYS
	} else {
		pllRst = rp.RESETS_RESET_PLL_USB
	}
	resetBlock(pllRst)
	unresetBlockWait(pllRst)

	// Load VCO-related dividers before starting VCO
	pll.cs.Set(refdiv)
	pll.fbDivInt.Set(fbdiv)

	// Turn on PLL
	pwr := uint32(rp.PLL_SYS_PWR_PD | rp.PLL_SYS_PWR_VCOPD)
	pll.pwr.ClearBits(pwr)

	// Wait for PLL to lock
	for !(pll.cs.HasBits(rp.PLL_SYS_CS_LOCK)) {
	}

	// Set up post dividers
	pll.prim.Set(pdiv)

	// Turn on post divider
	pll.pwr.ClearBits(rp.PLL_SYS_PWR_POSTDIVPD)

}

func (pll *pll) stop() {
	// Turn off PLL
	pwr := uint32(rp.PLL_SYS_PWR_PD | rp.PLL_SYS_PWR_VCOPD | rp.PLL_SYS_PWR_POSTDIVPD)
	pll.pwr.SetBits(pwr)
}
