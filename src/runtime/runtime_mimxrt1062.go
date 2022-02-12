//go:build mimxrt1062
// +build mimxrt1062

package runtime

import (
	"device/arm"
	"device/nxp"
	"machine"
	"machine/usb"
	"math/bits"
	"unsafe"
)

//go:extern _svectors
var _svectors [0]byte

//go:extern _flexram_cfg
var _flexram_cfg [0]byte

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

	// enable SysTick, GPIO, USB, and other peripherals
	initPeripherals()

	// reenable interrupts
	arm.EnableInterrupts(irq)

	run()
	exit(0)
}

func getRamSizeConfig(itcmKB, dtcmKB uint32) uint32 {
	const minKB, disabled = uint32(4), uint32(0)
	if itcmKB < minKB {
		itcmKB = disabled
	}
	if dtcmKB < minKB {
		dtcmKB = disabled
	}
	itcmKB = uint32(bits.Len(uint(itcmKB))) << nxp.IOMUXC_GPR_GPR14_CM7_CFGITCMSZ_Pos
	dtcmKB = uint32(bits.Len(uint(dtcmKB))) << nxp.IOMUXC_GPR_GPR14_CM7_CFGDTCMSZ_Pos
	return (itcmKB & nxp.IOMUXC_GPR_GPR14_CM7_CFGITCMSZ_Msk) |
		(dtcmKB & nxp.IOMUXC_GPR_GPR14_CM7_CFGDTCMSZ_Msk)
}

func initSystem() {

	// configure SRAM capacity (512K for both ITCM and DTCM)
	ramc := uintptr(unsafe.Pointer(&_flexram_cfg))
	nxp.IOMUXC_GPR.GPR17.Set(uint32(ramc))
	nxp.IOMUXC_GPR.GPR16.Set(0x00200007)
	nxp.IOMUXC_GPR.GPR14.Set(getRamSizeConfig(512, 512))

	// from Teensyduino
	nxp.PMU.MISC0_SET.Set(nxp.PMU_MISC0_REFTOP_SELFBIASOFF)

	// install vector table (TODO: initialize interrupt/exception table?)
	vtor := uintptr(unsafe.Pointer(&_svectors))
	nxp.SystemControl.VTOR.Set(uint32(vtor))

	const wdogUpdateKey = 0xD928C520

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
		nxp.RTWDOG.CNT.Set(wdogUpdateKey)
	} else {
		nxp.RTWDOG.CNT.Set((wdogUpdateKey >> 0) & 0xFFFF)
		nxp.RTWDOG.CNT.Set((wdogUpdateKey >> 16) & 0xFFFF)
	}
	nxp.RTWDOG.TOVAL.Set(0xFFFF)
	nxp.RTWDOG.CS.Set((nxp.RTWDOG.CS.Get() & ^uint32(nxp.RTWDOG_CS_EN_Msk)) | nxp.RTWDOG_CS_UPDATE_Msk)
}

func initPeripherals() {

	enableTimerClocks() // activate GPT/PIT clock gates
	initSysTick()       // enable SysTick
	initRTC()           // enable real-time clock

	enablePinClocks() // activate IOMUXC(_GPR)/GPIO clock gates
	initPins()        // configure GPIO

	enablePeripheralClocks() // activate peripheral clock gates
	initUSB()                // configure USB CDC-ACM (UART0)
	initUART()               // configure hardware UART (UART1)
}

func initPins() {
	// use fast GPIO for all pins (GPIO6-9)
	nxp.IOMUXC_GPR.GPR26.Set(0xFFFFFFFF)
	nxp.IOMUXC_GPR.GPR27.Set(0xFFFFFFFF)
	nxp.IOMUXC_GPR.GPR28.Set(0xFFFFFFFF)
	nxp.IOMUXC_GPR.GPR29.Set(0xFFFFFFFF)
}

func initUART() {
	machine.UART1.Configure(machine.UARTConfig{})
}

func initUSB() {
	// machine.HID0.Configure(usb.HIDConfig{})
	machine.UART0.Configure(usb.UARTConfig{})
}

func putchar(c byte) {
	machine.UART0.WriteByte(c) // print to USB UART
	// machine.UART1.WriteByte(c) // print to hardware UART
}

func exit(code int) {
	abort()
}

func abort() {
	for {
		arm.Asm("wfe")
	}
}

func waitForEvents() {
	arm.Asm("wfe")
}
