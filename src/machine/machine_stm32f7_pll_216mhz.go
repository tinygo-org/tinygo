//go:build stm32 && stm32f7x2

package machine

// PLLParams216MHz returns the HSE PLL dividers needed to reach a 432MHz VCO
// (216MHz SYSCLK, P=2, Q=9) for the configured crystal frequency.
func PLLParams216MHz() PLLParams {
	switch xtalHz {
	case 8_000_000:
		return PLLParams{M: 4, N: 216, P: 2, Q: 9}
	case 12_000_000:
		return PLLParams{M: 6, N: 216, P: 2, Q: 9}
	case 16_000_000:
		return PLLParams{M: 8, N: 216, P: 2, Q: 9}
	default:
		panic("unsupported xtal frequency")
	}
}
