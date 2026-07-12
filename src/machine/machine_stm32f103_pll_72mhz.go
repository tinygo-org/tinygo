//go:build stm32 && stm32f103

package machine

import "device/stm32"

// F103PLLParams holds the HSE prescaler (PLLXTPRE) and PLL multiplier
// (PLLMUL) needed to reach 72MHz SYSCLK from a given crystal frequency. The
// F1 PLL has no dedicated input divider, only an optional /2 HSE prescaler.
type F103PLLParams struct {
	Prediv uint32
	Mul    uint32
}

func PLLParams72MHz() F103PLLParams {
	switch xtalHz {
	case 8_000_000:
		return F103PLLParams{Prediv: stm32.RCC_CFGR_PLLXTPRE_Div1, Mul: stm32.RCC_CFGR_PLLMUL_Mul9}
	case 12_000_000:
		return F103PLLParams{Prediv: stm32.RCC_CFGR_PLLXTPRE_Div1, Mul: stm32.RCC_CFGR_PLLMUL_Mul6}
	case 16_000_000:
		// 16MHz / 2 (PLLXTPRE) x9 = 72MHz.
		return F103PLLParams{Prediv: stm32.RCC_CFGR_PLLXTPRE_Div2, Mul: stm32.RCC_CFGR_PLLMUL_Mul9}
	default:
		panic("unsupported xtal frequency")
	}
}
