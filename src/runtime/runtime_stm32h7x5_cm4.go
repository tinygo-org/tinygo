// +build stm32h7x5_cm4

package runtime

import (
	"device/arm"
	"device/stm32"
	"machine"
	"unsafe"
)

func initCore() {

	// Nothing to do:
	//   - Core registers initialization is handled by M7 core.
}

func initVectors() {

	// Initialize VTOR with vectors in flash
	vtor := uintptr(unsafe.Pointer(&_svectors))
	arm.SCB.VTOR.Set(uint32(vtor))

	// TODO: copy vectors to SRAM? I don't believe the M4 core can access any of
	//       the TCM regions, so VTOR cannot point to ITCM/DTCM.
}

func initMemory() {

	// Nothing to do:
	//   - MPU disabled on both cores for now
	//   - Cache features are only available on Cortex-M7 (and M33) core.
}

func initSync() {

	// SEVONPEND enabled so that an interrupt coming from the CPU(n) interrupt
	// signal is detectable by the CPU after a WFI/WFE instruction.
	arm.SCB.SCR.SetBits(arm.SCB_SCR_SEVONPEND)

	// Enable hardware semaphore (HSEM) clock
	_ = machine.EnableClock(unsafe.Pointer(stm32.HSEM), true)

	// Enable FLASH access from D2 power domain
	_ = stm32.Allocate(unsafe.Pointer(stm32.FLASH), stm32.RCC_CORE2)

	// Verify we are booting with the M4 core clock gated by the M7 core.
	// We should always enter this block, unless the bootloader option bytes are
	// programmed incorrectly.
	// After power-on, the M7 core must release the M4 clock gate for the M4 core
	// to boot. This block is where the M4 boots and resumes execution.
	if !stm32.IsBootM4() {

		// Activate HSEM notification for M4 core
		stm32.HSEM_CORE2.IER.SetBits(1 << machine.SemSTOP)

		// Put M4 core in deep-sleep mode, waiting for M7 core to complete system
		// initialization.
		stm32.PWR.CR1.ClearBits(stm32.PWR_CR1_LPDS)
		stm32.PWR_CPU1CR.ClearBits(stm32.PWR_CPUCR_PDDS_D2)
		stm32.PWR_CPU2CR.ClearBits(stm32.PWR_CPUCR_PDDS_D2)
		arm.SCB.SCR.SetBits(arm.SCB_SCR_SLEEPDEEP_Msk)
		arm.AsmFull(`
			dsb 0xF
			isb 0xF
			wfe
		`, nil)

		arm.SCB.SCR.ClearBits(arm.SCB_SCR_SLEEPDEEP_Msk)
		stm32.HSEM_CORE2.IER.ClearBits(1 << machine.SemSTOP)
		stm32.HSEM_CORE2.ICR.Set(1 << machine.SemSTOP)
	}
}

func initAccel() {

	// Activate instruction cache via ART accelerator
	_ = machine.EnableClock(unsafe.Pointer(stm32.ART), true)
	p := uint32(uintptr(unsafe.Pointer(&_svectors)))
	stm32.ART.CTR.ReplaceBits((p>>12)&0x000FFF00, stm32.ART_CTR_PCACHEADDR_Msk, 0)
	stm32.ART.CTR.SetBits(stm32.ART_CTR_EN)
}

func initCoreClocks() {

	// Nothing to do:
	//   - Core clocks initialization is handled by M7 core.
}

// Do not use directly -- call initSysTick instead, which also initializes the
// DWT cycle counter.
func initCoreSysTick(clk stm32.RCC_CLK_Type) {

	// Determine counter top which will cause rollover when the source clock (our
	// MCU core frequency) has cycled as many times as desired SysTick frequency.
	top := clk.HCLK/tickFreqHz - 1

	if top > arm.SYST_RVR_RELOAD_Msk {
		return // invalid tick count
	}

	// Disable SysTick before reconfiguring.
	arm.SYST.SYST_CSR.ClearBits(arm.SYST_CSR_ENABLE)

	tickCount.Set(0)
	arm.SYST.SYST_RVR.Set(top)
	arm.SYST.SYST_CVR.Set(0)
	// Enable SysTick IRQ and SysTick Timer, use internal (core) clock source
	arm.SYST.SYST_CSR.Set(arm.SYST_CSR_CLKSOURCE_Msk |
		arm.SYST_CSR_TICKINT_Msk | arm.SYST_CSR_ENABLE_Msk)

	// set SysTick and PendSV priority to 32
	arm.SCB.SHPR3.Set((0x20 << arm.SCB_SHPR3_PRI_15_Pos) |
		(0x20 << arm.SCB_SHPR3_PRI_14_Pos))
}
