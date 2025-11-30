//go:build py32

package runtime

import (
	"device/arm"
	"machine"

	"device/py32"
)

//export Reset_Handler
func main() {
	preinit()

	py32.RCC.SetICSCR_HSI_FS(py32.RCC_ICSCR_HSI_FS_Freq24MHz)

	ConfigureSystemTimer(24_000_000)
	machine.InitSerial()

	run()
	exit(0)
}

// Configure SysTick to fire every 1ms on given system frequency.
// This should be called after any changes to the system clock frequency.
func ConfigureSystemTimer(systemFrequencyHz uint32) {
	arm.SetupSystemTimer(systemFrequencyHz / 1000)
}

func ticksToNanoseconds(ticks timeUnit) int64 {
	return int64(ticks * 1000_000)
}

func nanosecondsToTicks(ns int64) timeUnit {
	return timeUnit(ns / 1000_000)
}

var tickCounter uint64

//go:linkname ticks runtime.ticks
func ticks() timeUnit {
	return timeUnit(tickCounter)
}

func sleepTicks(d timeUnit) {
	if d <= 0 {
		return
	}
	start := ticks()
	stop := start + d
	for ticks() < stop {
		arm.Asm("wfe")
	}
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

//export SysTick_Handler
func handleSysTick() {
	tickCounter = tickCounter + 1
}
