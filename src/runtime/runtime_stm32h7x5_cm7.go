// +build stm32h7x5_cm7

package runtime

import (
	"device/arm"
	"device/stm32"
	"machine"
	"runtime/volatile"
	"unsafe"
)

func initCore() {

	const revisionB = 0x2000 // STM32H7 revision B
	const defaultLatency = 7 // FLASH ACR: 7 wait states

	// Increasing CPU frequency
	if defaultLatency > stm32.FLASH.ACR.Get()&stm32.FLASH_ACR_LATENCY_Msk {
		stm32.FLASH.ACR.ReplaceBits(defaultLatency, stm32.FLASH_ACR_LATENCY_Msk, 0)
	}

	stm32.RCC.CR.SetBits(stm32.RCC_CR_HSION)
	stm32.RCC.CFGR.Set(0)
	stm32.RCC.CR.ClearBits(stm32.RCC_CR_HSEON | stm32.RCC_CR_HSECSSON |
		stm32.RCC_CR_CSION | stm32.RCC_CR_HSI48ON | stm32.RCC_CR_CSIKERON |
		stm32.RCC_CR_PLL1ON | stm32.RCC_CR_PLL2ON | stm32.RCC_CR_PLL3ON)

	// Decreasing the number of wait states because of lower CPU frequency
	if defaultLatency < stm32.FLASH.ACR.Get()&stm32.FLASH_ACR_LATENCY_Msk {
		stm32.FLASH.ACR.ReplaceBits(defaultLatency, stm32.FLASH_ACR_LATENCY_Msk, 0)
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
	stm32.EXTI.C2EMR3.SetBits(0x4000)

	// Check if STM32H7 revision prior to revision B
	if (stm32.DBGMCU.IDC.Get()&stm32.DBG_IDC_REV_ID_Msk)>>
		stm32.DBG_IDC_REV_ID_Pos < revisionB {
		// Change the switch matrix read issuing capability to 1 for the AXI SRAM
		// target (Target 7)
		((*volatile.Register32)(unsafe.Pointer(uintptr(0x51008108)))).Set(1)
	}

	// Disable FMC bank 1 (enabled after reset).
	// This prevents CPU speculation access on this bank, which blocks the use of
	// FMC during 24us. During this time, the other FMC masters (such as LTDC)
	// cannot use it!
	stm32.FMC.BCR1.Set(0x000030D2)
}

func initVectors() {

	if false {
		// Initialize VTOR with vectors in flash
		src := unsafe.Pointer(&_svectors)
		arm.SCB.VTOR.Set(uint32(uintptr(src)))
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
		arm.SCB.VTOR.Set(uint32(uintptr(unsafe.Pointer(&_svtor))))
	}
}

func initMemory() {

	// Disable MPU (for now). See the comment above the call to initSystem()
	// within runtime.main(), which outlines the steps required to enable support
	// for the MPU (...if its even desired or necessary?).
	_ = enableMPU(false)

	// Enable CPU instruction cache (I-CACHE)
	stm32.EnableICache(true)
	// Enable CPU data cache (D-CACHE)
	stm32.EnableDCache(true)
}

func initSync() {

	// SEVONPEND enabled so that an interrupt coming from the CPU(n) interrupt
	// signal is detectable by the CPU after a WFI/WFE instruction.
	arm.SCB.SCR.SetBits(arm.SCB_SCR_SEVONPEND)

	// Enable hardware semaphore (HSEM) clock
	_ = machine.EnableClock(unsafe.Pointer(stm32.HSEM), true)
}

func initAccel() {

	// TBD: Configure ART on Cortex-M7 core? Currently enabled by Cortex-M4 core.
}

func initCoreClocks() {

	// First enable SYSCFG clock
	_ = machine.EnableClock(unsafe.Pointer(stm32.SYSCFG), true)

	initLowSpeedCrystal(lseDriveLow)

	// initialize SYSCLK (PLL1 with HSE)
	if freq, ok := initCoreFreq(true, false); ok {
		// re-initialize SysTick with increased frequencies
		initSysTick(freq)
	}

	// Turn off the LED enabled by bootloader
	machine.LEDG.Configure(machine.PinConfig{Mode: machine.PinOutput})
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

	const revisionY = 0x1003 // STM32H7 revision Y

	if !stm32.RCC.BDCR.HasBits(stm32.RCC_BDCR_LSERDY) {
		// LSE is in the backup domain (DBP) and write access is denied to DBP after
		// reset, so we first have to enable write access for DBP before configuring.
		stm32.PWR.CR1.SetBits(stm32.PWR_CR1_DBP)
		if (lseDriveMedLow <= drive && drive <= lseDriveMedHigh) &&
			(stm32.DBGMCU.IDC.Get()&stm32.DBG_IDC_REV_ID_Msk)>>
				stm32.DBG_IDC_REV_ID_Pos <= revisionY {
			drive = ^drive & stm32.RCC_BDCR_LSEDRV_Msk
		}
		stm32.RCC.BDCR.ReplaceBits(drive, stm32.RCC_BDCR_LSEDRV_Msk, 0)
	}
}

func initCoreFreq(bypass, lowSpeed bool) (stm32.RCC_CLK_Type, bool) {

	// First need to change SYSCLK source to CSI before modifying the main PLL
	if stm32.RCC_PLL_SRC_HSE == (stm32.RCC.PLLCKSELR.Get()&
		stm32.RCC_PLLCKSELR_PLLSRC_Msk)>>stm32.RCC_PLLCKSELR_PLLSRC_Pos {
		config := stm32.RCC_CLK_CFG_Type{
			CLK:    stm32.RCC_CLK_SYSCLK,
			SYSSrc: stm32.RCC_SYS_SRC_CSI,
		}
		if !config.Set(1) {
			return stm32.RCC_CLK_Type{}, false
		}
	}

	// Enable oscillator pin
	machine.PH01.Configure(machine.PinConfig{Mode: machine.PinOutput})
	machine.PH01.Set(true)

	// Configure main internal regulator voltage for increased CPU frequencies.
	scale := uint32(0x03 << stm32.PWR_D3CR_VOS_Pos)
	if lowSpeed {
		scale = 0x01 << stm32.PWR_D3CR_VOS_Pos
	}
	if !stm32.SetPowerSupply(stm32.PWR_SMPS_1V8_SUPPLIES_LDO, scale) {
		return stm32.RCC_CLK_Type{}, false
	}

	for !stm32.PWR_FLAG_VOSRDY.Get() {
	} // wait for voltage to stabilize

	// Enable HSE oscillator and activate PLL with HSE as source
	osc := stm32.RCC_OSC_CFG_Type{
		OSC:   stm32.RCC_OSC_HSE | stm32.RCC_OSC_HSI48,
		HSE:   stm32.RCC_CR_HSEON,
		HSI48: stm32.RCC_CR_HSI48ON,
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

	if !osc.Set() {
		return stm32.RCC_CLK_Type{}, false
	}

	latency := uint32(4)
	if lowSpeed {
		latency = 0
	}

	config := stm32.RCC_CLK_CFG_Type{
		CLK: stm32.RCC_CLK_SYSCLK | stm32.RCC_CLK_HCLK |
			stm32.RCC_CLK_PCLK1 | stm32.RCC_CLK_PCLK2 |
			stm32.RCC_CLK_D1PCLK1 | stm32.RCC_CLK_D3PCLK1,
		SYSSrc:  stm32.RCC_SYS_SRC_PLL1,
		SYSDiv:  stm32.RCC_D1CFGR_D1CPRE_Div1 << stm32.RCC_D1CFGR_D1CPRE_Pos,
		AHBDiv:  stm32.RCC_D1CFGR_HPRE_Div2 << stm32.RCC_D1CFGR_HPRE_Pos,
		APB1Div: stm32.RCC_D2CFGR_D2PPRE1_Div2 << stm32.RCC_D2CFGR_D2PPRE1_Pos,
		APB2Div: stm32.RCC_D2CFGR_D2PPRE2_Div2 << stm32.RCC_D2CFGR_D2PPRE2_Pos,
		APB3Div: stm32.RCC_D1CFGR_D1PPRE_Div2 << stm32.RCC_D1CFGR_D1PPRE_Pos,
		APB4Div: stm32.RCC_D3CFGR_D3PPRE_Div2 << stm32.RCC_D3CFGR_D3PPRE_Pos,
	}

	if !config.Set(latency) {
		return stm32.RCC_CLK_Type{}, false
	}

	// TODO: enable USB regulator, VBUS detection

	// Return the clock frequencies derived from the active core configuration.
	return stm32.ClockFreq(), true
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
	arm.SCB.SHPR3.Set((0x20 << arm.SCB_SHPR3_PRI_15_Pos) |
		(0x20 << arm.SCB_SHPR3_PRI_14_Pos))
}

func enableMPU(enable bool) bool {
	if enable {
		// Enable the MPU with privileged access to default memory
		stm32.MPU.MPU_CTRL.Set(stm32.MPU_MPU_CTRL_PRIVDEFENA_Msk |
			stm32.MPU_MPU_CTRL_ENABLE_Msk)
		// Enable fault exceptions
		stm32.SCB_SHCSR.SetBits(stm32.SCB_SHCSR_MEMFAULTENA_Msk)
		// Ensure MPU setting take effects
		arm.AsmFull(`
				dsb 0xF
				isb 0xF
			`, nil)
	} else {
		// Make sure outstanding transfers are done
		arm.AsmFull(`
				dmb 0xF
			`, nil)
		// Disable fault exceptions
		stm32.SCB_SHCSR.ClearBits(stm32.SCB_SHCSR_MEMFAULTENA_Msk)
		// Disable the MPU and clear the control register
		stm32.MPU.MPU_CTRL.Set(0)
	}
	return true
}

type mpuRegion struct {
	Enable           uint8  // Specifies the status of the region.
	Number           uint8  // Specifies the number of the region to protect.
	BaseAddress      uint32 // Specifies the base address of the region to protect.
	Size             uint8  // Specifies the size of the region to protect.
	SubRegionDisable uint8  // Specifies the number of the subregion protection to disable.
	TypeExtField     uint8  // Specifies the TEX field level.
	AccessPermission uint8  // Specifies the region access permission type.
	DisableExec      uint8  // Specifies the instruction access status.
	IsShareable      uint8  // Specifies the shareability status of the protected region.
	IsCacheable      uint8  // Specifies the cacheable status of the region protected.
	IsBufferable     uint8  // Specifies the bufferable status of the protected region.
}

func (r mpuRegion) Set() bool {
	// Set the Region number
	stm32.MPU.MPU_RNR.Set(uint32(r.Number))
	if 0 != r.Enable {
		stm32.MPU.MPU_RBAR.Set(r.BaseAddress)
		stm32.MPU.MPU_RASR.Set((uint32(r.DisableExec) << stm32.MPU_MPU_RASR_XN_Pos) |
			(uint32(r.AccessPermission) << stm32.MPU_MPU_RASR_AP_Pos) |
			(uint32(r.TypeExtField) << stm32.MPU_MPU_RASR_TEX_Pos) |
			(uint32(r.IsShareable) << stm32.MPU_MPU_RASR_S_Pos) |
			(uint32(r.IsCacheable) << stm32.MPU_MPU_RASR_C_Pos) |
			(uint32(r.IsBufferable) << stm32.MPU_MPU_RASR_B_Pos) |
			(uint32(r.SubRegionDisable) << stm32.MPU_MPU_RASR_SRD_Pos) |
			(uint32(r.Size) << stm32.MPU_MPU_RASR_SIZE_Pos) |
			(uint32(r.Enable) << stm32.MPU_MPU_RASR_ENABLE_Pos))
	} else {
		stm32.MPU.MPU_RBAR.Set(0)
		stm32.MPU.MPU_RASR.Set(0)
	}
	return true
}
