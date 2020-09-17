// +build mimxrt1062

package runtime

import (
	"device/arm"
	"device/nxp"
	"unsafe"
)

const asyncScheduler = false

//go:extern _svectors
var _svectors [0]byte

//go:extern _flexram_cfg
var _flexram_cfg [0]byte

func postinit() {}

//export Reset_Handler
func main() {

	// disable interrupts
	irq := arm.DisableInterrupts()

	// initialize FPU and VTOR, reset watchdogs
	initSystem()

	// configure core and peripheral clocks/PLLs/PFDs
	initClocks()

	// copy data/bss sections from flash to RAM
	preinit()

	// initialize cache and MPU
	initCache()

	// enable SysTick, GPIO, and peripherals
	initPeripherals()

	// reenable interrupts
	arm.EnableInterrupts(irq)

	run()
	abort()
}

func initSystem() {

	// configure SRAM capacity
	ramc := uintptr(unsafe.Pointer(&_flexram_cfg))
	nxp.IOMUXC_GPR.GPR17.Set(uint32(ramc))
	nxp.IOMUXC_GPR.GPR16.Set(0x00200007)
	nxp.IOMUXC_GPR.GPR14.Set(0x00AA0000)

	// use bandgap-based bias currents for best performance (Page 1175) [Teensyduino]
	nxp.PMU.MISC0_SET.Set(1 << 3)

	// install vector table (TODO: initialize interrupt/exception table?)
	vtor := uintptr(unsafe.Pointer(&_svectors))
	nxp.SystemControl.VTOR.Set(uint32(vtor))

	// disable watchdog powerdown counter
	nxp.WDOG1.WMCR.ClearBits(nxp.WDOG_WMCR_PDE_Msk)
	nxp.WDOG2.WMCR.ClearBits(nxp.WDOG_WMCR_PDE_Msk)

	// disable watchdog
	if nxp.WDOG1.WCR.HasBits(nxp.WDOG_WCR_WDE_Msk) {
		nxp.WDOG1.WCR.ClearBits(nxp.WDOG_WCR_WDE_Msk)
	}
	if nxp.WDOG2.WCR.HasBits(nxp.WDOG_WCR_WDE_Msk) {
		nxp.WDOG2.WCR.ClearBits(nxp.WDOG_WCR_WDE_Msk)
	}
	if nxp.RTWDOG.CS.HasBits(nxp.RTWDOG_CS_CMD32EN_Msk) {
		nxp.RTWDOG.CNT.Set(0xD928C520) // 0xD928C520 is the update key
	} else {
		nxp.RTWDOG.CNT.Set(0xC520)
		nxp.RTWDOG.CNT.Set(0xD928)
	}
	nxp.RTWDOG.TOVAL.Set(0xFFFF)
	nxp.RTWDOG.CS.Set((nxp.RTWDOG.CS.Get() & ^uint32(nxp.RTWDOG_CS_EN_Msk)) | nxp.RTWDOG_CS_UPDATE_Msk)
}

func initPeripherals() {

	// enable FPU - set CP10, CP11 full access
	nxp.SystemControl.CPACR.SetBits((3 << (10 * 2)) | (3 << (11 * 2)))

	enableTimerClocks() // activate GPT/PIT clock gates
	initSysTick()       // enable SysTick

	enablePinClocks() // activate IOMUXC(_GPR)/GPIO clock gates
	initPins()        // configure GPIO

	enablePeripheralClocks() // activate peripheral clock gates
}

func initPins() {
	// use fast GPIO for all pins (GPIO6-9)
	nxp.IOMUXC_GPR.GPR26.Set(0xFFFFFFFF)
	nxp.IOMUXC_GPR.GPR27.Set(0xFFFFFFFF)
	nxp.IOMUXC_GPR.GPR28.Set(0xFFFFFFFF)
	nxp.IOMUXC_GPR.GPR29.Set(0xFFFFFFFF)
}

func putchar(c byte) {}

func abort() {
	for {
		arm.Asm("wfe")
	}
}

func waitForEvents() {
	arm.Asm("wfe")
}
