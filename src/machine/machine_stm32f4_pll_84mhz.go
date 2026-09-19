//go:build stm32f4 && stm32f401

package machine

// PLLParams84MHz returns the HSE PLL dividers needed to reach a 336MHz VCO
// (84MHz SYSCLK, P=4, Q=7) for the configured crystal frequency. M is chosen
// to bring the PLL input (HSE/M) to 2MHz, as recommended by the STM32F401
// reference manual (RM0368). With VCO=336MHz and P=4, SYSCLK=84MHz. Q=7
// produces the 48MHz USB clock (VCO/Q).
func PLLParams84MHz() PLLParams {
	switch xtalHz {
	case 8_000_000:
		return PLLParams{M: 4, N: 168, P: 4, Q: 7}
	case 16_000_000:
		return PLLParams{M: 8, N: 168, P: 4, Q: 7}
	default:
		panic("unsupported xtal frequency")
	}
}
