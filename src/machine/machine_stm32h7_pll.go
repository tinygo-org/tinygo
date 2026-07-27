//go:build stm32 && stm32h7

package machine

import "device/stm32"

// PLLParams400MHz returns the HSE PLL dividers needed to reach a 800MHz VCO
// (400MHz SYSCLK, P=2, Q=4) for the configured crystal frequency.
// It returns the appropriate PLL1RGE range value in the R field.
func PLLParams400MHz() PLLParams {
	var m, n, rge uint32
	switch xtalHz {
	case 8_000_000:
		m = 1
		n = 100
	case 16_000_000:
		m = 2
		n = 100
	case 24_000_000:
		m = 3
		n = 100
	case 25_000_000:
		m = 5
		n = 160
	default:
		panic("unsupported xtal frequency")
	}

	vcoIn := xtalHz / m
	if vcoIn < 2_000_000 {
		rge = stm32.RCC_PLLCFGR_PLL1RGE_Range1
	} else if vcoIn < 4_000_000 {
		rge = stm32.RCC_PLLCFGR_PLL1RGE_Range2
	} else if vcoIn < 8_000_000 {
		rge = stm32.RCC_PLLCFGR_PLL1RGE_Range4
	} else {
		rge = stm32.RCC_PLLCFGR_PLL1RGE_Range8
	}

	return PLLParams{M: m, N: n, P: 2, Q: 4, R: rge}
}

// HSEBypass returns whether the HSE clock is configured in bypass mode (external MCO clock).
func HSEBypass() bool {
	return hseBypass
}
