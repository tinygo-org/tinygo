//go:build nrf && nrf52840

package runtime

import (
	"device/arm"
	"device/nrf"
	"machine"
	"machine/usb/cdc"
)

//export Reset_Handler
func main() {
	if nrf.FPUPresent {
		arm.SCB.CPACR.Set(0) // disable FPU if it is enabled
	}
	systemInit()
	preinit()
	run()
	exit(0)
}

func init() {
	cdc.EnableUSBCDC()
	machine.USBDev.Configure(machine.UARTConfig{})
	machine.InitSerial()
	initLFCLK()
	initRTC(nrf.RTC1)
}

func initLFCLK() {
	if machine.HasLowFrequencyCrystal {
		nrf.CLOCK.LFCLKSRC.Set(nrf.CLOCK_LFCLKSTAT_SRC_Xtal)
	}
	nrf.CLOCK.TASKS_LFCLKSTART.Set(1)
	for nrf.CLOCK.EVENTS_LFCLKSTARTED.Get() == 0 {
	}
	nrf.CLOCK.EVENTS_LFCLKSTARTED.Set(0)
}
