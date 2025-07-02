//go:build rp2040 || rp2350

package machine

import (
	"device/rp"
	"errors"
	"math"
	"math/bits"
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
func (pll *pll) init(refdiv, fbdiv, postDiv1, postDiv2 uint32) {
	refFreq := xoscFreq / refdiv

	// What are we multiplying the reference clock by to get the vco freq
	// (The regs are called div, because you divide the vco output and compare it to the refclk)

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

	// Check that reference frequency is no greater than vcoFreq / 16
	vcoFreq := calcVCO(xoscFreq, fbdiv, refdiv)
	if refFreq > vcoFreq/16 {
		panic("reference frequency should not be greater than vco frequency divided by 16")
	}

	// div1 feeds into div2 so if div1 is 5 and div2 is 2 then you get a divide by 10
	pdiv := uint32(postDiv1)<<rp.PLL_SYS_PRIM_POSTDIV1_Pos | uint32(postDiv2)<<rp.PLL_SYS_PRIM_POSTDIV2_Pos

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

var errVCOOverflow = errors.New("VCO calculation overflow; use lower MHz")

// pllSearch enables searching for a good PLL configuration.
// Example for 12MHz crystal and RP2040:
//
//	fbdiv, refdiv, pd1, pd2, _ := pllSearch{LockRefDiv:1}.CalcDivs(12*MHz, 125*MHz, MHz)
//
// Example for 12MHz crystal and RP2350:
//
//	fbdiv, refdiv, pd1, pd2, _ := pllSearch{LockRefDiv:1}.CalcDivs(12*MHz, 150*MHz, MHz)
type pllSearch struct {
	LowerVCO   bool
	LockRefDiv uint8
}

func (ps pllSearch) CalcDivs(xoscRef, targetFreq, MHz uint64) (fbdiv uint64, refdiv, pd1, pd2 uint8, err error) {
	genTable()
	var bestFreq, bestFbdiv uint64
	var bestRefdiv, bestpd1, bestpd2 uint8
	maxVCO, minVCO := 1600*MHz, 750*MHz
	var bestMargin int64 = int64(maxVCO)
	iters := 0
	for refdiv = 1; refdiv < 64; refdiv++ {
		if ps.LockRefDiv != 0 && refdiv != ps.LockRefDiv {
			continue
		}
		firstFBDiv := minVCO * uint64(refdiv) / xoscRef
		for fbdiv = firstFBDiv; fbdiv < 321; fbdiv++ {
			overflow, vco := bits.Mul64(xoscRef, fbdiv)
			vco /= uint64(refdiv)
			if overflow != 0 {
				return fbdiv, refdiv, pd1, pd2, errVCOOverflow
			} else if vco > maxVCO {
				break
			}
			calcPD12 := vco / targetFreq
			if calcPD12 < 1 {
				calcPD12 = 1
			} else if calcPD12 > 49 {
				calcPD12 = 49
			}
			iters++
			pd1 = pdTable[calcPD12].hivco[0]
			pd2 = pdTable[calcPD12].hivco[1]
			fout, err := pllFreqOutPostdiv(xoscRef, fbdiv, MHz, refdiv, pd1, pd2)
			found := false
			margin := abs(int64(fout) - int64(targetFreq))
			if err == nil && margin <= bestMargin {
				found = true
				bestFreq = fout
				bestFbdiv = fbdiv
				bestpd1 = pd1
				bestpd2 = pd2
				bestRefdiv = refdiv
				bestMargin = margin
			}
			pd1 = pdTable[calcPD12].lovco[0]
			pd2 = pdTable[calcPD12].lovco[1]
			fout, err = pllFreqOutPostdiv(xoscRef, fbdiv, MHz, refdiv, pd1, pd2)
			margin = abs(int64(fout) - int64(targetFreq))
			if err == nil && margin <= bestMargin {
				found = true
				bestFreq = fout
				bestFbdiv = fbdiv
				bestpd1 = pd1
				bestpd2 = pd2
				bestRefdiv = refdiv
				bestMargin = margin
			}
			if found && ps.LowerVCO {
				break
			}
		}
	}
	if bestFreq == 0 {
		return fbdiv, refdiv, pd1, pd2, errors.New("no best frequency found")
	}
	return bestFbdiv, bestRefdiv, bestpd1, bestpd2, nil
}

func abs(a int64) int64 {
	if a == math.MinInt64 {
		return math.MaxInt64
	} else if a < 0 {
		return -a
	}
	return a
}

func pllFreqOutPostdiv(xosc, fbdiv, MHz uint64, refdiv, postdiv1, postdiv2 uint8) (foutpostdiv uint64, err error) {
	// testing grounds.
	const (
		mhz    = 1
		cfref  = 12 * mhz // given by crystal oscillator selection.
		crefd  = 1
		cfbdiv = 100
		cvco   = cfref * cfbdiv / crefd
		cpd1   = 6
		cpd2   = 2
		foutpd = (cfref / crefd) * cfbdiv / (cpd1 * cpd2)
	)
	refFreq := xosc / uint64(refdiv)
	overflow, vco := bits.Mul64(xosc, fbdiv)
	vco /= uint64(refdiv)
	foutpostdiv = vco / uint64(postdiv1*postdiv2)
	switch {
	case refdiv < 1 || refdiv > 63:
		err = errors.New("reference divider out of range")
	case fbdiv < 16 || fbdiv > 320:
		err = errors.New("feedback divider out of range")
	case postdiv1 < 1 || postdiv1 > 7:
		err = errors.New("postdiv1 out of range")
	case postdiv2 < 1 || postdiv2 > 7:
		err = errors.New("postdiv2 out of range")
	case postdiv1 < postdiv2:
		err = errors.New("user error: use higher value for postdiv1 for lower power consumption")
	case vco < 750*MHz || vco > 1600*MHz:
		err = errors.New("VCO out of range")
	case refFreq < 5*MHz:
		err = errors.New("minimum reference frequency breach")
	case refFreq > vco/16:
		err = errors.New("maximum reference frequency breach")
	case vco > 1200*MHz && vco < 1600*MHz && xosc < 75*MHz && refdiv != 1:
		err = errors.New("refdiv should be 1 for given VCO and reference frequency")
	case overflow != 0:
		err = errVCOOverflow
	}
	if err != nil {
		return 0, err
	}
	return foutpostdiv, nil
}

func calcVCO(xoscFreq, fbdiv, refdiv uint32) uint32 {
	const maxXoscMHz = math.MaxUint32 / 320 / MHz // 13MHz maximum xosc apparently.
	if fbdiv > 320 || xoscFreq > math.MaxUint32/320 {
		panic("invalid VCO calculation args")
	}
	return xoscFreq * fbdiv / refdiv
}

var pdTable = [50]struct {
	hivco [2]uint8
	lovco [2]uint8
}{}

func genTable() {
	if pdTable[1].hivco[1] != 0 {
		return // Already generated.
	}
	for product := 1; product < len(pdTable); product++ {
		bestProdhi := 255
		bestProdlo := 255
		for pd1 := 7; pd1 > 0; pd1-- {
			for pd2 := pd1; pd2 > 0; pd2-- {
				gotprod := pd1 * pd2
				if abs(int64(gotprod-product)) < abs(int64(bestProdlo-product)) {
					bestProdlo = gotprod
					pdTable[product].lovco[0] = uint8(pd1)
					pdTable[product].lovco[1] = uint8(pd2)
				}
			}
		}
		for pd1 := 1; pd1 < 8; pd1++ {
			for pd2 := 1; pd2 <= pd1; pd2++ {
				gotprod := pd1 * pd2
				if abs(int64(gotprod-product)) < abs(int64(bestProdhi-product)) {
					bestProdhi = gotprod
					pdTable[product].hivco[0] = uint8(pd1)
					pdTable[product].hivco[1] = uint8(pd2)
				}
			}
		}
	}
}
