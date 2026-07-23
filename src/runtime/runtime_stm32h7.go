//go:build stm32 && stm32h7

package runtime

import (
	"device/stm32"
	"machine"
	_ "machine/usb/cdc"
)

func init() {
	initCLK()
	initMPU()

	machine.InitSerial()

	initTickTimer(&machine.TIM3)
}

func putchar(c byte) {
	machine.Serial.WriteByte(c)
}

func getchar() byte {
	for machine.Serial.Buffered() == 0 {
		Gosched()
	}
	v, _ := machine.Serial.ReadByte()
	return v
}

func buffered() int {
	return machine.Serial.Buffered()
}

func initCLK() {
	// 1. Enable SYSCFG
	stm32.RCC.APB4ENR.SetBits(stm32.RCC_APB4ENR_SYSCFGEN)

	// H743/H753 have no SMPS; the NUCLEO-H753ZI runs VCORE from the internal
	// LDO (the CR3 reset state). Supply bits in CR3 are write-once after POR,
	// so keep LDOEN/BYPASS untouched and only enable the USB 3.3V level
	// detector needed by the USB transceivers.
	stm32.PWR.CR3.SetBits(stm32.PWR_CR3_USB33DEN)

	// 3. Configure VOS1 (Scale 1)
	// RM0433 §6.8.4: ACTVOSRDY must be 1 (Run mode confirmed) before changing VOS.
	for stm32.PWR.CSR1.Get()&stm32.PWR_CSR1_ACTVOSRDY == 0 {
	}
	// RM0433: VOS1 is 0b11.
	stm32.PWR.D3CR.ReplaceBits(0b11<<stm32.PWR_D3CR_VOS_Pos, stm32.PWR_D3CR_VOS_Msk, 0)
	for stm32.PWR.D3CR.Get()&stm32.PWR_D3CR_VOSRDY == 0 {
	}

	// 4. Enable HSE
	if machine.HSEBypass() {
		stm32.RCC.CR.SetBits(stm32.RCC_CR_HSEBYP | stm32.RCC_CR_HSEON)
	} else {
		stm32.RCC.CR.SetBits(stm32.RCC_CR_HSEON)
	}
	for stm32.RCC.CR.Get()&stm32.RCC_CR_HSERDY == 0 {
	}

	// 5. Configure PLL1
	pll := machine.PLLParams400MHz()

	// Source: HSE (2)
	stm32.RCC.PLLCKSELR.ReplaceBits(stm32.RCC_PLLCKSELR_PLLSRC_HSE, stm32.RCC_PLLCKSELR_PLLSRC_Msk, 0)
	// DIVM1
	stm32.RCC.PLLCKSELR.ReplaceBits(pll.M<<stm32.RCC_PLLCKSELR_DIVM1_Pos, stm32.RCC_PLLCKSELR_DIVM1_Msk, 0)

	// PLL1CFGR: Wide VCO (0), Range based on pll.R (VCO input frequency)
	stm32.RCC.PLLCFGR.ReplaceBits(
		(stm32.RCC_PLLCFGR_PLL1VCOSEL_WideVCO<<stm32.RCC_PLLCFGR_PLL1VCOSEL_Pos)|
			(pll.R<<stm32.RCC_PLLCFGR_PLL1RGE_Pos),
		stm32.RCC_PLLCFGR_PLL1VCOSEL_Msk|stm32.RCC_PLLCFGR_PLL1RGE_Msk, 0)

	// PLL1DIVR: DIVN1=pll.N, DIVP1=pll.P, DIVQ1=pll.Q
	// PLL1P = (VCO VCO_input * N) / P
	// PLL1Q = (VCO VCO_input * N) / Q
	stm32.RCC.PLL1DIVR.ReplaceBits(
		(pll.N-1)<<stm32.RCC_PLL1DIVR_DIVN1_Pos|(pll.P-1)<<stm32.RCC_PLL1DIVR_DIVP1_Pos|(pll.Q-1)<<stm32.RCC_PLL1DIVR_DIVQ1_Pos,
		stm32.RCC_PLL1DIVR_DIVN1_Msk|stm32.RCC_PLL1DIVR_DIVP1_Msk|stm32.RCC_PLL1DIVR_DIVQ1_Msk, 0)

	// Enable PLL1P (SYSCLK=400MHz) and PLL1Q (SPI1/2/3 kernel=200MHz)
	stm32.RCC.PLLCFGR.SetBits(stm32.RCC_PLLCFGR_DIVP1EN | stm32.RCC_PLLCFGR_DIVQ1EN)

	// Enable PLL1
	stm32.RCC.CR.SetBits(stm32.RCC_CR_PLL1ON)
	for stm32.RCC.CR.Get()&stm32.RCC_CR_PLL1RDY == 0 {
	}

	// 6. Bus Prescalers
	// D1CPRE=1 (0), HPRE=2 (8) -> HCLK=200MHz, D1PPRE (APB3)=2 (4) -> PCLK3=100MHz
	stm32.RCC.D1CFGR.ReplaceBits(
		(stm32.RCC_D1CFGR_D1CPRE_Div1<<stm32.RCC_D1CFGR_D1CPRE_Pos)|
			(stm32.RCC_D1CFGR_HPRE_Div2<<stm32.RCC_D1CFGR_HPRE_Pos)|
			(stm32.RCC_D1CFGR_D1PPRE_Div2<<stm32.RCC_D1CFGR_D1PPRE_Pos),
		stm32.RCC_D1CFGR_D1CPRE_Msk|stm32.RCC_D1CFGR_HPRE_Msk|stm32.RCC_D1CFGR_D1PPRE_Msk, 0)

	// D2CFGR: D2PPRE1 (APB1)=2 (4) -> PCLK1=100MHz, D2PPRE2 (APB2)=2 (4) -> PCLK2=100MHz
	stm32.RCC.D2CFGR.ReplaceBits(
		(stm32.RCC_D2CFGR_D2PPRE1_Div2<<stm32.RCC_D2CFGR_D2PPRE1_Pos)|
			(stm32.RCC_D2CFGR_D2PPRE2_Div2<<stm32.RCC_D2CFGR_D2PPRE2_Pos),
		stm32.RCC_D2CFGR_D2PPRE1_Msk|stm32.RCC_D2CFGR_D2PPRE2_Msk, 0)

	// D3CFGR: D3PPRE (APB4)=2 (4) -> PCLK4=100MHz
	stm32.RCC.D3CFGR.ReplaceBits(
		stm32.RCC_D3CFGR_D3PPRE_Div2<<stm32.RCC_D3CFGR_D3PPRE_Pos,
		stm32.RCC_D3CFGR_D3PPRE_Msk, 0)

	// 7. Flash Latency
	// VOS1, 200MHz AXI clock -> 2 wait states, WRHIGHFREQ=2 (RM0433 Table 17).
	stm32.FLASH.ACR.ReplaceBits(2|2<<stm32.FLASH_ACR_WRHIGHFREQ_Pos,
		stm32.FLASH_ACR_LATENCY_Msk|stm32.FLASH_ACR_WRHIGHFREQ_Msk, 0)
	for stm32.FLASH.ACR.Get()&stm32.FLASH_ACR_LATENCY_Msk != 2 {
	}

	// 8. Switch to PLL1
	// SW: PLL1 (3)
	stm32.RCC.CFGR.ReplaceBits(3<<stm32.RCC_CFGR_SW_Pos, stm32.RCC_CFGR_SW_Msk, 0)
	for (stm32.RCC.CFGR.Get() & stm32.RCC_CFGR_SWS_Msk) != (3 << stm32.RCC_CFGR_SWS_Pos) {
	}

	// 9. Peripheral Kernel Clocks
	// I2C1,2,3 source: HSI_KER (2) to keep 64MHz timing compatibility.
	stm32.RCC.D2CCIP2R.ReplaceBits(stm32.RCC_D2CCIP2R_I2C123SEL_HSI_KER<<stm32.RCC_D2CCIP2R_I2C123SEL_Pos, stm32.RCC_D2CCIP2R_I2C123SEL_Msk, 0)
	// I2C4 source: HSI_KER (2)
	stm32.RCC.D3CCIPR.ReplaceBits(stm32.RCC_D3CCIPR_I2C4SEL_HSI_KER<<stm32.RCC_D3CCIPR_I2C4SEL_Pos, stm32.RCC_D3CCIPR_I2C4SEL_Msk, 0)

	// SPI1,2,3 source: PLL1_Q (0) -> 200MHz
	stm32.RCC.D2CCIP1R.ReplaceBits(stm32.RCC_D2CCIP1R_SPI123SEL_PLL1_Q<<stm32.RCC_D2CCIP1R_SPI123SEL_Pos, stm32.RCC_D2CCIP1R_SPI123SEL_Msk, 0)
	// SPI4,5 source: APB (0) -> PCLK2 = 100MHz (PLL1-derived)
	stm32.RCC.D2CCIP1R.ReplaceBits(stm32.RCC_D2CCIP1R_SPI45SEL_APB<<stm32.RCC_D2CCIP1R_SPI45SEL_Pos, stm32.RCC_D2CCIP1R_SPI45SEL_Msk, 0)
	// SPI6 source: PCLK4 (0) -> 100MHz (PLL1-derived)
	stm32.RCC.D3CCIPR.ReplaceBits(stm32.RCC_D3CCIPR_SPI6SEL_RCC_PCLK4<<stm32.RCC_D3CCIPR_SPI6SEL_Pos, stm32.RCC_D3CCIPR_SPI6SEL_Msk, 0)

	// 10. HSI48 — used as kernel clock for RNG and USB (RM0433 §33.3 requires ≤48 MHz).
	stm32.RCC.CR.SetBits(stm32.RCC_CR_HSI48ON)
	for stm32.RCC.CR.Get()&stm32.RCC_CR_HSI48RDY == 0 {
	}
	// RNGSEL and USBSEL reset value is 0x0 (HSI48 or PLL1_Q); set explicitly to HSI48.
	stm32.RCC.D2CCIP2R.ReplaceBits(
		stm32.RCC_D2CCIP2R_RNGSEL_HSI48<<stm32.RCC_D2CCIP2R_RNGSEL_Pos|
			stm32.RCC_D2CCIP2R_USBSEL_HSI48<<stm32.RCC_D2CCIP2R_USBSEL_Pos,
		stm32.RCC_D2CCIP2R_RNGSEL_Msk|stm32.RCC_D2CCIP2R_USBSEL_Msk, 0)

	// 11. Enable CRS (Clock Recovery System) for HSI48 stabilization via USB SOF.
	stm32.RCC.APB1HENR.SetBits(stm32.RCC_APB1HENR_CRSEN)
	stm32.CRS.CFGR.ReplaceBits(stm32.CRS_CFGR_SYNCSRC_USB_SOF<<stm32.CRS_CFGR_SYNCSRC_Pos, stm32.CRS_CFGR_SYNCSRC_Msk, 0)
	stm32.CRS.CR.SetBits(stm32.CRS_CR_CEN | stm32.CRS_CR_AUTOTRIMEN)

	// 12. Configure PLL2 for ADC (80MHz)
	// DIVM2 = pll.M
	stm32.RCC.PLLCKSELR.ReplaceBits(pll.M<<stm32.RCC_PLLCKSELR_DIVM2_Pos, stm32.RCC_PLLCKSELR_DIVM2_Msk, 0)

	// PLL2CFGR: Wide VCO (0), Range based on pll.R (VCO input frequency)
	stm32.RCC.PLLCFGR.ReplaceBits(
		(stm32.RCC_PLLCFGR_PLL2VCOSEL_WideVCO<<stm32.RCC_PLLCFGR_PLL2VCOSEL_Pos)|
			(pll.R<<stm32.RCC_PLLCFGR_PLL2RGE_Pos),
		stm32.RCC_PLLCFGR_PLL2VCOSEL_Msk|stm32.RCC_PLLCFGR_PLL2RGE_Msk, 0)

	// PLL2DIVR: DIVN2=pll.N, DIVP2=10 (Value 9)
	// PLL2P = (VCO VCO_input * N) / 10 = 80MHz
	stm32.RCC.PLL2DIVR.ReplaceBits(
		(pll.N-1)<<stm32.RCC_PLL2DIVR_DIVN2_Pos|9<<stm32.RCC_PLL2DIVR_DIVP2_Pos,
		stm32.RCC_PLL2DIVR_DIVN2_Msk|stm32.RCC_PLL2DIVR_DIVP2_Msk, 0)

	// Enable DIVP2EN
	stm32.RCC.PLLCFGR.SetBits(stm32.RCC_PLLCFGR_DIVP2EN)

	// Enable PLL2
	stm32.RCC.CR.SetBits(stm32.RCC_CR_PLL2ON)
	for stm32.RCC.CR.Get()&stm32.RCC_CR_PLL2RDY == 0 {
	}

	// 12. ADC kernel clock source: PLL2_P (0).
	stm32.RCC.D3CCIPR.ReplaceBits(stm32.RCC_D3CCIPR_ADCSEL_PLL2_P<<stm32.RCC_D3CCIPR_ADCSEL_Pos, stm32.RCC_D3CCIPR_ADCSEL_Msk, 0)
}
