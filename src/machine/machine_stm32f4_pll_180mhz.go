//go:build stm32f4 && stm32f469

package machine

// PLLParams180MHz returns the HSE PLL dividers needed to reach a 360MHz VCO
// (180MHz SYSCLK, P=2, Q=7, R=6) for the configured crystal frequency.
func PLLParams180MHz() PLLParams {
	switch xtalHz {
	case 8_000_000:
		return PLLParams{M: 4, N: 180, P: 2, Q: 7, R: 6}
	case 12_000_000:
		return PLLParams{M: 6, N: 180, P: 2, Q: 7, R: 6}
	case 16_000_000:
		return PLLParams{M: 8, N: 180, P: 2, Q: 7, R: 6}
	default:
		panic("unsupported xtal frequency")
	}
}
