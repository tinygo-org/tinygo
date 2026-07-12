//go:build stm32

package machine

// PLLParams holds the HSE main-PLL dividers/multipliers (RCC_PLLCFGR M/N/P/Q/R
// fields) needed to reach a chip's target VCO/SYSCLK frequency from a given
// crystal frequency. R is left zero on chips without a PLLR output.
type PLLParams struct {
	M, N, P, Q, R uint32
}
