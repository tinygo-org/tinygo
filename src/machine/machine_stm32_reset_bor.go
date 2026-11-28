//go:build stm32f4 || stm32f7

package machine

// BORRSTF, RM0090 section 7.3.21 and RM0431 section 5.3.21.
const stm32ResetFlagBOR = 1 << 25
