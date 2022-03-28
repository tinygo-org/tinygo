//go:build stm32 && stm32l4
// +build stm32,stm32l4

package runtime

import (
	"device/stm32"
	"machine"
)

const (
	PWR_CR1_VOS_0                = 1 << stm32.PWR_CR1_VOS_Pos
	PWR_CR1_VOS_1                = 2 << stm32.PWR_CR1_VOS_Pos
	PWR_REGULATOR_VOLTAGE_SCALE1 = PWR_CR1_VOS_0
	PWR_REGULATOR_VOLTAGE_SCALE2 = PWR_CR1_VOS_1

	FLASH_LATENCY_0 = 0
	FLASH_LATENCY_1 = 1
	FLASH_LATENCY_2 = 2
	FLASH_LATENCY_3 = 3
	FLASH_LATENCY_4 = 4

	RCC_PLLP_DIV2 = 2
	RCC_PLLP_DIV7 = 7
	RCC_PLLQ_DIV2 = 2
	RCC_PLLR_DIV2 = 2

	RCC_CFGR_SWS_MSI = 0x0
	RCC_CFGR_SWS_PLL = 0xC

	RCC_PLLSOURCE_MSI = 1

	RCC_PLL_SYSCLK = stm32.RCC_PLLCFGR_PLLREN
)

type arrtype = uint32

func init() {
	initCLK()

	machine.Serial.Configure(machine.UARTConfig{})

	initTickTimer(&machine.TIM15)
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
	// PWR_CLK_ENABLE
	stm32.RCC.APB1ENR1.SetBits(stm32.RCC_APB1ENR1_PWREN)
	_ = stm32.RCC.APB1ENR1.Get()

	// Disable Backup domain protection
	if !stm32.PWR.CR1.HasBits(stm32.PWR_CR1_DBP) {
		stm32.PWR.CR1.SetBits(stm32.PWR_CR1_DBP)
		for !stm32.PWR.CR1.HasBits(stm32.PWR_CR1_DBP) {
		}
	}

	// Set LSE Drive to LOW
	stm32.RCC.BDCR.ReplaceBits(0, stm32.RCC_BDCR_LSEDRV_Msk, 0)

	// Initialize the High-Speed External Oscillator
	initOsc()

	// PWR_VOLTAGESCALING_CONFIG
	stm32.PWR.CR1.ReplaceBits(0, stm32.PWR_CR1_VOS_Msk, 0)
	_ = stm32.PWR.CR1.Get()

	// Set flash wait states (min 5 latency units) based on clock
	if (stm32.FLASH.ACR.Get() & 0xF) < 5 {
		stm32.FLASH.ACR.ReplaceBits(5, 0xF, 0)
	}

	// Ensure HCLK does not exceed max during transition
	stm32.RCC.CFGR.ReplaceBits(8<<stm32.RCC_CFGR_HPRE_Pos, stm32.RCC_CFGR_HPRE_Msk, 0)

	// Set SYSCLK source and wait
	// (3 = RCC_SYSCLKSOURCE_PLLCLK, 2=RCC_CFGR_SWS_Pos)
	stm32.RCC.CFGR.ReplaceBits(3, stm32.RCC_CFGR_SW_Msk, 0)
	for stm32.RCC.CFGR.Get()&(3<<2) != (3 << 2) {
	}

	// Set HCLK
	// (0 = RCC_SYSCLKSOURCE_PLLCLK)
	stm32.RCC.CFGR.ReplaceBits(0, stm32.RCC_CFGR_HPRE_Msk, 0)

	// Set flash wait states (max 5 latency units) based on clock
	if (stm32.FLASH.ACR.Get() & 0xF) > 5 {
		stm32.FLASH.ACR.ReplaceBits(5, 0xF, 0)
	}

	// Set APB1 and APB2 clocks (0 = DIV1)
	stm32.RCC.CFGR.ReplaceBits(0, stm32.RCC_CFGR_PPRE1_Msk, 0)
	stm32.RCC.CFGR.ReplaceBits(0, stm32.RCC_CFGR_PPRE2_Msk, 0)
}

func initOsc() {
	sysclkSource := stm32.RCC.CFGR.Get() & stm32.RCC_CFGR_SWS_Msk
	pllConfig := stm32.RCC.PLLCFGR.Get() & stm32.RCC_PLLCFGR_PLLSRC_Msk

	// Enable MSI, adjusting flash latency
	if sysclkSource == RCC_CFGR_SWS_MSI ||
		(sysclkSource == RCC_CFGR_SWS_PLL && pllConfig == RCC_PLLSOURCE_MSI) {
		if MSIRANGE > getMSIRange() {
			setFlashLatencyFromMSIRange(MSIRANGE)

			setMSIFreq(MSIRANGE, 0)
		} else {
			setMSIFreq(MSIRANGE, 0)

			if sysclkSource == RCC_CFGR_SWS_MSI {
				setFlashLatencyFromMSIRange(MSIRANGE)
			}
		}
	} else {
		stm32.RCC.CR.SetBits(stm32.RCC_CR_MSION)
		for !stm32.RCC.CR.HasBits(stm32.RCC_CR_MSIRDY) {
		}

		setMSIFreq(MSIRANGE, 0)
	}

	// Enable LSE, wait until ready
	stm32.RCC.BDCR.SetBits(stm32.RCC_BDCR_LSEON)
	for !stm32.RCC.BDCR.HasBits(stm32.RCC_BDCR_LSEON) {
	}

	// Disable the PLL, wait until disabled
	stm32.RCC.CR.ClearBits(stm32.RCC_CR_PLLON)
	for stm32.RCC.CR.HasBits(stm32.RCC_CR_PLLRDY) {
	}

	// Configure the PLL
	stm32.RCC.PLLCFGR.ReplaceBits(
		(1)| // 1 = RCC_PLLSOURCE_MSI
			(PLL_M-1)<<stm32.RCC_PLLCFGR_PLLM_Pos|
			(PLL_N<<stm32.RCC_PLLCFGR_PLLN_Pos)|
			(((PLL_Q>>1)-1)<<stm32.RCC_PLLCFGR_PLLQ_Pos)|
			(((PLL_R>>1)-1)<<stm32.RCC_PLLCFGR_PLLR_Pos)|
			(PLL_P<<stm32.RCC_PLLCFGR_PLLP_Pos),
		stm32.RCC_PLLCFGR_PLLSRC_Msk|stm32.RCC_PLLCFGR_PLLM_Msk|
			stm32.RCC_PLLCFGR_PLLN_Msk|stm32.RCC_PLLCFGR_PLLP_Msk|
			stm32.RCC_PLLCFGR_PLLR_Msk|stm32.RCC_PLLCFGR_PLLP_Msk,
		0)

	// Enable the PLL and PLL System Clock Output, wait until ready
	stm32.RCC.CR.SetBits(stm32.RCC_CR_PLLON)
	stm32.RCC.PLLCFGR.SetBits(stm32.RCC_PLLCFGR_PLLREN) // = RCC_PLL_SYSCLK
	for !stm32.RCC.CR.HasBits(stm32.RCC_CR_PLLRDY) {
	}

	// Enable system clock output
	stm32.RCC.PLLCFGR.SetBits(RCC_PLL_SYSCLK)
}

func getMSIRange() uint32 {
	if stm32.RCC.CR.HasBits(stm32.RCC_CR_MSIRGSEL) {
		return (stm32.RCC.CR.Get() & stm32.RCC_CR_MSIRANGE_Msk) >> stm32.RCC_CR_MSIRANGE_Pos
	}

	return (stm32.RCC.CSR.Get() & stm32.RCC_CSR_MSISRANGE_Msk) >> stm32.RCC_CSR_MSISRANGE_Pos
}

func setMSIFreq(r uint32, calibration uint32) {
	stm32.RCC.CR.SetBits(stm32.RCC_CR_MSIRGSEL)
	stm32.RCC.CR.ReplaceBits(r<<stm32.RCC_CR_MSIRANGE_Pos, stm32.RCC_CR_MSIRANGE_Msk, 0)

	stm32.RCC.ICSCR.ReplaceBits(calibration<<stm32.RCC_ICSCR_MSITRIM_Pos, stm32.RCC_ICSCR_MSITRIM_Msk, 0)
}

func setFlashLatencyFromMSIRange(r uint32) {
	var vos uint32
	if pwrIsClkEnabled() {
		vos = pwrExGetVoltageRange()
	} else {
		pwrClkEnable()
		vos = pwrExGetVoltageRange()
		pwrClkDisable()
	}

	latency := uint32(FLASH_LATENCY_0)
	if vos == PWR_REGULATOR_VOLTAGE_SCALE1 {
		if r > stm32.RCC_CR_MSIRANGE_Range16M {
			if r > stm32.RCC_CR_MSIRANGE_Range32M {
				latency = FLASH_LATENCY_2
			} else {
				latency = FLASH_LATENCY_1
			}
		}
	} else if r > stm32.RCC_CR_MSIRANGE_Range16M {
		latency = FLASH_LATENCY_3
	} else {
		if r == stm32.RCC_CR_MSIRANGE_Range16M {
			latency = FLASH_LATENCY_2
		} else if r == stm32.RCC_CR_MSIRANGE_Range8M {
			latency = FLASH_LATENCY_1
		}
	}

	stm32.FLASH.ACR.ReplaceBits(latency, stm32.Flash_ACR_LATENCY_Msk, 0)
}

func pwrIsClkEnabled() bool {
	return stm32.RCC.APB1ENR1.HasBits(stm32.RCC_APB1ENR1_PWREN)
}

func pwrClkEnable() {
	stm32.RCC.APB1ENR1.SetBits(stm32.RCC_APB1ENR1_PWREN)
}
func pwrClkDisable() {
	stm32.RCC.APB1ENR1.ClearBits(stm32.RCC_APB1ENR1_PWREN)
}

func pwrExGetVoltageRange() uint32 {
	return stm32.PWR.CR1.Get() & stm32.PWR_CR1_VOS_Msk
}
