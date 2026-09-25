//go:build esp32 || esp32c3 || esp32c6

package runtime

import (
	"device/esp"
	"machine"
	"unsafe"
)

// Initialize .bss: zero-initialized global variables.
// The .data section has already been loaded by the ROM bootloader.
func clearbss() {
	ptr := unsafe.Pointer(&_sbss)
	for ptr != unsafe.Pointer(&_ebss) {
		*(*uint32)(ptr) = 0
		ptr = unsafe.Add(ptr, 4)
	}
}

func initTimer() {
	// Configure timer 0 in timer group 0, for timekeeping.
	//   EN:       Enable the timer.
	//   INCREASE: Count up every tick (as opposed to counting down).
	//   DIVIDER:  16-bit prescaler. Each chip sets timerDivider.
	// esp.TIMG0.T0CONFIG.Set(0 << esp.TIMG_T0CONFIG_T0_EN_Pos)
	esp.TIMG0.T0CONFIG.Set(esp.TIMG_T0CONFIG_EN | esp.TIMG_T0CONFIG_INCREASE | timerDivider<<esp.TIMG_T0CONFIG_DIVIDER_Pos)
	// esp.TIMG0.T0CONFIG.Set(1 << esp.TIMG_T0CONFIG_T0_DIVCNT_RST_Pos)
	// esp.TIMG0.T0CONFIG.Set(esp.TIMG_T0CONFIG_T0_EN)

	// Set the timer counter value to 0.
	esp.TIMG0.T0LOADLO.Set(0)
	esp.TIMG0.T0LOADHI.Set(0)
	esp.TIMG0.T0LOAD.Set(0) // value doesn't matter.
}

func ticks() timeUnit {
	// First, update the LO and HI register pair by writing any value to the
	// register. This allows reading the pair atomically.
	esp.TIMG0.T0UPDATE.Set(0)
	// Then read the two 32-bit parts of the timer.
	return timeUnit(uint64(esp.TIMG0.T0LO.Get()) | uint64(esp.TIMG0.T0HI.Get())<<32)
}

func nanosecondsToTicks(ns int64) timeUnit {
	// Calculate the number of ticks from the number of nanoseconds. Each chip
	// sets nsPerTick.
	return timeUnit(ns / nsPerTick)
}

func ticksToNanoseconds(ticks timeUnit) int64 {
	// See nanosecondsToTicks.
	return int64(ticks) * nsPerTick
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
