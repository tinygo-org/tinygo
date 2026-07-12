//go:build stm32f4 && (stm32f405 || stm32f407)

package machine

// PLLParams168MHz returns the HSE PLL dividers needed to reach a 336MHz VCO
// (168MHz SYSCLK, P=2, Q=7) for the configured crystal frequency. M is chosen
// to bring the PLL input (HSE/M) to 2MHz, the value RM0090 pg. 95 recommends
// to minimize jitter; the previous hardcoded F407 table used 1MHz.
func PLLParams168MHz() PLLParams {
	switch xtalHz {
	case 8_000_000:
		return PLLParams{M: 4, N: 168, P: 2, Q: 7}
	case 12_000_000:
		return PLLParams{M: 6, N: 168, P: 2, Q: 7}
	case 16_000_000:
		return PLLParams{M: 8, N: 168, P: 2, Q: 7}
	default:
		panic("unsupported xtal frequency")
	}
}
