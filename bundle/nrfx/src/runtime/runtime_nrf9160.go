//go:build nrf9160
// +build nrf9160

package runtime

import (
	"device/arm"
	"device/nrf"
	"machine"
)

//export Reset_Handler
func main() {
	// Reset CONTROL register
	arm.AsmFull(`movs.n {}, #0
		     msr CONTROL, {}
		     isb`, nil)

	// Clear SPLIM registers
	arm.AsmFull(`movs.n {}, #0
		     msr MSPLIM, {}
		     msr PSPLIM, {}`, nil)

	// lock interrupts: will get unlocked when switch to main task
	arm.AsmFull(`movs.n {}, #0
		     msr BASEPRI, {}`, nil)
	if nrf.FPUPresent {
		arm.SCB.CPACR.Set(0) // disable FPU if it is enabled
	}
	systemInit()
	preinit()
	run()
	exit(0)
}

func init() {
	machine.InitSerial()
	initLFCLK()
	initRTC(nrf.RTC1_S)
}

func initLFCLK() {
	if machine.HasLowFrequencyCrystal {
		nrf.CLOCK_S.LFCLKSRC.Set(nrf.CLOCK_LFCLKSTAT_SRC_LFRC)
	}
	nrf.CLOCK_S.TASKS_LFCLKSTART.Set(1)
	for nrf.CLOCK_S.EVENTS_LFCLKSTARTED.Get() == 0 {
	}
	nrf.CLOCK_S.EVENTS_LFCLKSTARTED.Set(0)
}
