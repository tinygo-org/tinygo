// +build stm32h7x7_cm7

package runtime

import (
	"device/stm32"
	"machine"
	"runtime/volatile"
	"unsafe"
)

//go:extern _svectors
var _svectors [0]uint8

//go:extern _evectors
var _evectors [0]uint8

//go:extern _svtor
var _svtor [0]uint8 // vectors location in RAM

func initCore() {

	// Increasing CPU frequency
	if stm32.FLASH_LATENCY_DEFAULT > stm32.FLASH.ACR.Get()&stm32.FLASH_ACR_LATENCY_Msk {
		stm32.FLASH.ACR.ReplaceBits(stm32.FLASH_LATENCY_DEFAULT, stm32.FLASH_ACR_LATENCY_Msk, 0)
	}

	stm32.RCC.CR.SetBits(stm32.RCC_CR_HSION)
	stm32.RCC.CFGR.Set(0)
	stm32.RCC.CR.ClearBits(stm32.RCC_CR_HSEON | stm32.RCC_CR_HSECSSON |
		stm32.RCC_CR_CSION | stm32.RCC_CR_RC48ON | stm32.RCC_CR_CSIKERON |
		stm32.RCC_CR_PLL1ON | stm32.RCC_CR_PLL2ON | stm32.RCC_CR_PLL3ON)

	// Decreasing the number of wait states because of lower CPU frequency
	if stm32.FLASH_LATENCY_DEFAULT < stm32.FLASH.ACR.Get()&stm32.FLASH_ACR_LATENCY_Msk {
		stm32.FLASH.ACR.ReplaceBits(stm32.FLASH_LATENCY_DEFAULT, stm32.FLASH_ACR_LATENCY_Msk, 0)
	}

	stm32.RCC.D1CFGR.Set(0)
	stm32.RCC.D2CFGR.Set(0)
	stm32.RCC.D3CFGR.Set(0)
	stm32.RCC.PLLCKSELR.Set(0x02020200)
	stm32.RCC.PLLCFGR.Set(0x01FF0000)
	stm32.RCC.PLL1DIVR.Set(0x01010280)
	stm32.RCC.PLL1FRACR.Set(0)
	stm32.RCC.PLL2DIVR.Set(0x01010280)
	stm32.RCC.PLL2FRACR.Set(0)
	stm32.RCC.PLL3DIVR.Set(0x01010280)
	stm32.RCC.PLL3FRACR.Set(0)
	stm32.RCC.CR.ClearBits(stm32.RCC_CR_HSEBYP)

	stm32.RCC.CIER.Set(0) // Disable all interrupts

	// Enable Cortex-M7 HSEM EXTI line (line 78)
	stm32.EXTI_CORE2.EMR3.SetBits(0x4000)

	// Check if STM32H7 revision prior to revision B
	if stm32.DBG.MCURevision() < stm32.DBG_MCU_REVISION_B {
		// Change the switch matrix read issuing capability to 1 for the AXI SRAM
		// target (Target 7)
		((*volatile.Register32)(unsafe.Pointer(uintptr(0x51008108)))).Set(1)
	}

	// Disable FMC bank 1 (enabled after reset).
	// This prevents CPU speculation access on this bank, which blocks the use of
	// FMC during 24us. During this time, the other FMC masters (such as LTDC)
	// cannot use it!
	stm32.FMC.FMC_BCR1.Set(0x000030D2)
}

func initVectors() {
	if false {
		// Initialize VTOR with vectors in flash
		src := unsafe.Pointer(&_svectors)
		stm32.SCB.VTOR.Set(uint32(uintptr(src)))
	} else {
		// Copy vectors to ITCM (AXI bus) for fastest possible code accesses
		src := unsafe.Pointer(&_svectors)
		dst := unsafe.Pointer(&_svtor)
		for src != unsafe.Pointer(&_evectors) {
			*(*uint32)(dst) = *(*uint32)(src)
			src = unsafe.Pointer(uintptr(src) + 4)
			dst = unsafe.Pointer(uintptr(dst) + 4)
		}
		// Initialize VTOR with vectors in ITCM
		stm32.SCB.VTOR.Set(uint32(uintptr(unsafe.Pointer(&_svtor))))
	}
}

func initMemory() {

	// Disable MPU (for now). See the comment above the call to initSystem()
	// within runtime.main(), which outlines the steps required to enable support
	// for the MPU (...if its even desired or necessary?).
	_ = stm32.MPU.Enable(false)

	// Enable CPU instruction cache (I-CACHE)
	stm32.SCB.EnableICache(true)
	// Enable CPU data cache (D-CACHE)
	stm32.SCB.EnableDCache(true)
}

func initSync() {

	// SEVONPEND enabled so that an interrupt coming from the CPU(n) interrupt
	// signal is detectable by the CPU after a WFI/WFE instruction.
	stm32.SCB.SCR.SetBits(stm32.SCB_SCR_SEVEONPEND)

	_ = machine.EnableClock(unsafe.Pointer(stm32.HSEM), true)
}

func initAccel() {
	// TBD: Configure ART on Cortex-M7 core? Currently enabled by Cortex-M4 core.
}

func initCoreClocks() {

	// First enable SYSCFG and GPIO peripheral clocks
	_ = machine.EnableClock(unsafe.Pointer(stm32.SYSCFG), true)
	// _ = enableGPIO()

	initLowSpeedCrystal(lseDriveLow)

	// initialize SYSCLK (PLL1 with HSE)
	if freq, ok := initCoreFreq(true, false); ok {
		// re-initialize SysTick with increased frequencies
		initSysTick(freq)
	}

	// Turn off the LED enabled by bootloader
	machine.LEDG.Configure(machine.PinConfig{Mode: machine.PinOutputPushPull})
	machine.LEDG.High()

	stm32.BootM4()
	for !stm32.RCC_FLAG_D2CKRDY.Get() {
	}
}

const (
	lseDriveLow = uint32(iota) << stm32.RCC_BDCR_LSEDRV_Pos
	lseDriveMedLow
	lseDriveMedHigh
	lseDriveHigh
)

func initLowSpeedCrystal(drive uint32) {
	if !stm32.RCC.BDCR.HasBits(stm32.RCC_BDCR_LSERDY) {
		// LSE is in the backup domain (DBP) and write access is denied to DBP after
		// reset, so we first have to enable write access for DBP before configuring.
		stm32.PWR.PWR_CR1.SetBits(stm32.PWR_PWR_CR1_DBP)
		if (lseDriveMedLow <= drive && drive <= lseDriveMedHigh) &&
			stm32.DBG.MCURevision() <= stm32.DBG_MCU_REVISION_Y {
			drive = ^drive & stm32.RCC_BDCR_LSEDRV_Msk
		}
		stm32.RCC.BDCR.ReplaceBits(drive, stm32.RCC_BDCR_LSEDRV_Msk, 0)
	}
}

func initCoreFreq(bypass, lowSpeed bool) (stm32.RCC_CLK_Type, bool) {

	// First need to change SYSCLK source to CSI before modifying the main PLL
	if stm32.RCC_PLL_SRC_HSE == (stm32.RCC.PLLCKSELR.Get()&
		stm32.RCC_PLLCKSELR_PLLSRC_Msk)>>stm32.RCC_PLLCKSELR_PLLSRC_Pos {
		if !stm32.RCC.ConfigCLK(stm32.RCC_CLK_CFG_Type{
			CLK:    stm32.RCC_CLK_SYSCLK,
			SYSSrc: stm32.RCC_SYS_SRC_CSI,
		}, stm32.FLASH_ACR_LATENCY_1WS) {
			return stm32.RCC_CLK_Type{}, false
		}
	}

	// Enable oscillator pin
	machine.PH01.Configure(machine.PinConfig{
		Mode: machine.PinOutputPushPull | machine.PinPullUp | machine.PinLowSpeed,
	})
	machine.PH01.Set(true)

	// Configure main internal regulator voltage for increased CPU frequencies.
	scale := uint32(stm32.PWR_REGULATOR_VOLTAGE_SCALE1)
	if lowSpeed {
		scale = stm32.PWR_REGULATOR_VOLTAGE_SCALE3
	}
	if !stm32.PWR.Configure(stm32.PWR_SMPS_1V8_SUPPLIES_LDO, scale) {
		return stm32.RCC_CLK_Type{}, false
	}

	for !stm32.PWR_FLAG_VOSRDY.Get() {
	} // wait for voltage to stabilize

	// Enable HSE oscillator and activate PLL with HSE as source
	osc := stm32.RCC_OSC_CFG_Type{
		OSC:   stm32.RCC_OSC_HSE | stm32.RCC_OSC_HSI48,
		HSE:   stm32.RCC_CR_HSEON,
		HSI48: stm32.RCC_CR_RC48ON,
		PLL1: stm32.RCC_PLL_CFG_Type{
			PLL: stm32.RCC_PLL_ON,
			Src: stm32.RCC_PLL_SRC_HSE,
			Div: stm32.RCC_PLL_DIV_Type{
				M:  5,
				N:  192,
				Nf: 0,
				PQR: stm32.RCC_PLL_PQR_Type{
					P: 2,
					Q: 10,
					R: 2,
				},
				VCI: stm32.RCC_PLL1_VCI_RANGE_2,
				VCO: stm32.RCC_PLL1_VCO_WIDE,
			},
		},
	}
	if bypass {
		osc.HSE = stm32.RCC_CR_HSEBYP | stm32.RCC_CR_HSEON
	}
	if lowSpeed {
		osc.PLL1.Div.N = 40
	}

	if !stm32.RCC.ConfigOSC(osc) {
		return stm32.RCC_CLK_Type{}, false
	}

	latency := uint32(stm32.FLASH_ACR_LATENCY_4WS)
	if lowSpeed {
		latency = stm32.FLASH_ACR_LATENCY_0WS
	}

	if !stm32.RCC.ConfigCLK(
		stm32.RCC_CLK_CFG_Type{
			CLK: stm32.RCC_CLK_SYSCLK | stm32.RCC_CLK_HCLK |
				stm32.RCC_CLK_PCLK1 | stm32.RCC_CLK_PCLK2 |
				stm32.RCC_CLK_D1PCLK1 | stm32.RCC_CLK_D3PCLK1,
			SYSSrc:  stm32.RCC_SYS_SRC_PLL1,
			SYSDiv:  stm32.RCC_D1CFGR_D1CPRE_DIV1,
			AHBDiv:  stm32.RCC_D1CFGR_HPRE_DIV2,
			APB1Div: stm32.RCC_D2CFGR_D2PPRE1_DIV2,
			APB2Div: stm32.RCC_D2CFGR_D2PPRE2_DIV2,
			APB3Div: stm32.RCC_D1CFGR_D1PPRE_DIV2,
			APB4Div: stm32.RCC_D3CFGR_D3PPRE_DIV2,
		}, latency) {
		return stm32.RCC_CLK_Type{}, false
	}

	// TODO: enable USB regulator, VBUS detection

	// Return the clock frequencies derived from the active core configuration.
	return stm32.RCC.ClockFreq(), true
}

// Do not use directly -- call initSysTick instead, which also initializes the
// DWT cycle counter.
func initCoreSysTick(clk stm32.RCC_CLK_Type) {

	// Determine counter top which will cause rollover when the source clock (our
	// MCU core frequency) has cycled as many times as desired SysTick frequency.
	top := clk.SYSCLK/tickFreqHz - 1

	if top > stm32.STK_RVR_RELOAD_Msk {
		return // invalid tick count
	}

	// Disable SysTick before reconfiguring.
	stm32.STK.CSR.ClearBits(stm32.STK_CSR_ENABLE)

	tickCount.Set(0)
	stm32.STK.RVR.Set(top)
	stm32.STK.CVR.Set(0)
	// Enable SysTick IRQ and SysTick Timer, use internal (core) clock source
	stm32.STK.CSR.Set(stm32.STK_CSR_CLKSOURCE_Msk |
		stm32.STK_CSR_TICKINT_Msk | stm32.STK_CSR_ENABLE_Msk)

	// set SysTick and PendSV priority to 32
	stm32.SCB.SHPR3.Set((0x20 << stm32.SCB_SHPR3_PRI_15_Pos) |
		(0x20 << stm32.SCB_SHPR3_PRI_14_Pos))
}
