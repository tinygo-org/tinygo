//go:build stm32f4 && (stm32f407 || stm32f469)
// +build stm32f4
// +build stm32f407 stm32f469

package runtime

import (
	"device/stm32"
	"machine"
)

func init() {
	initCLK()

	machine.Serial.Configure(machine.UARTConfig{})

	initTickTimer(&machine.TIM2)
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
	// Reset clock registers
	// Set HSION
	stm32.RCC.CR.SetBits(stm32.RCC_CR_HSION)
	for !stm32.RCC.CR.HasBits(stm32.RCC_CR_HSIRDY) {
	}

	// Reset CFGR
	stm32.RCC.CFGR.Set(0x00000000)
	// Reset HSEON, CSSON and PLLON
	stm32.RCC.CR.ClearBits(stm32.RCC_CR_HSEON | stm32.RCC_CR_CSSON | stm32.RCC_CR_PLLON)
	// Reset PLLCFGR
	stm32.RCC.PLLCFGR.Set(0x24003010)
	// Reset HSEBYP
	stm32.RCC.CR.ClearBits(stm32.RCC_CR_HSEBYP)
	// Disable all interrupts
	stm32.RCC.CIR.Set(0x00000000)

	// Set up the clock
	var startupCounter uint32 = 0

	// Enable HSE
	stm32.RCC.CR.Set(stm32.RCC_CR_HSEON)

	// Wait till HSE is ready and if timeout is reached exit
	for {
		startupCounter++
		if stm32.RCC.CR.HasBits(stm32.RCC_CR_HSERDY) || (startupCounter == HSE_STARTUP_TIMEOUT) {
			break
		}
	}
	if stm32.RCC.CR.HasBits(stm32.RCC_CR_HSERDY) {
		// Enable high performance mode, configure maximum system frequency.
		stm32.RCC.APB1ENR.SetBits(stm32.RCC_APB1ENR_PWREN)
		stm32.PWR.CR.SetBits(0x4000) // PWR_CR_VOS
		// HCLK = SYSCLK / 1
		stm32.RCC.CFGR.SetBits(stm32.RCC_CFGR_HPRE_Div1 << stm32.RCC_CFGR_HPRE_Pos)
		// PCLK2 = HCLK / 2
		stm32.RCC.CFGR.SetBits(stm32.RCC_CFGR_PPRE2_Div2 << stm32.RCC_CFGR_PPRE2_Pos)
		// PCLK1 = HCLK / 4
		stm32.RCC.CFGR.SetBits(stm32.RCC_CFGR_PPRE1_Div4 << stm32.RCC_CFGR_PPRE1_Pos)
		// Configure the main PLL
		stm32.RCC.PLLCFGR.Set(PLL_CFGR)
		// Enable main PLL
		stm32.RCC.CR.SetBits(stm32.RCC_CR_PLLON)
		// Wait till the main PLL is ready
		for (stm32.RCC.CR.Get() & stm32.RCC_CR_PLLRDY) == 0 {
		}
		// Configure Flash prefetch, Instruction cache, Data cache and wait state
		stm32.FLASH.ACR.Set(stm32.FLASH_ACR_ICEN | stm32.FLASH_ACR_DCEN | (5 << stm32.FLASH_ACR_LATENCY_Pos))
		// Select the main PLL as system clock source
		stm32.RCC.CFGR.ClearBits(stm32.RCC_CFGR_SW_Msk)
		stm32.RCC.CFGR.SetBits(stm32.RCC_CFGR_SW_PLL << stm32.RCC_CFGR_SW_Pos)
		for (stm32.RCC.CFGR.Get() & stm32.RCC_CFGR_SWS_Msk) != (stm32.RCC_CFGR_SWS_PLL << stm32.RCC_CFGR_SWS_Pos) {
		}

	} else {
		// If HSE failed to start up, the application will have wrong clock configuration
		for {
		}
	}

	// Enable the CCM RAM clock
	stm32.RCC.AHB1ENR.SetBits(1 << 20)
}
