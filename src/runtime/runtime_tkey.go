//go:build tkey

// This file implements target-specific things for the TKey.

package runtime

import (
	"device/tkey"
	"machine"
	"runtime/volatile"
)

//export main
func main() {
	preinit()
	initPeripherals()
	run()
	exit(0)
}

// initPeripherals configures peripherals the way the runtime expects them.
func initPeripherals() {
	// prescaler value that results in 0.00001-second timer-ticks.
	// given an 18 MHz processor, a millisecond is about 18,000 cycles.
	tkey.TIMER.PRESCALER.Set(18 * machine.MHz / 100000)

	machine.InitSerial()
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

var timestamp volatile.Register32

// ticks returns the current value of the timer in ticks.
func ticks() timeUnit {
	return timeUnit(timestamp.Get())
}

// sleepTicks sleeps for at least the duration d.
func sleepTicks(d timeUnit) {
	target := uint32(ticks() + d)

	tkey.TIMER.TIMER.Set(uint32(d))
	tkey.TIMER.CTRL.SetBits(tkey.TK1_MMIO_TIMER_CTRL_START)
	for tkey.TIMER.STATUS.Get() != 0 {
	}
	timestamp.Set(target)
}
