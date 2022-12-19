//go:build esp8266

package runtime

import (
	"device"
	"device/esp"
	"machine"
	"unsafe"
)

type timeUnit int64

var currentTime timeUnit = 0

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

// Write to the internal control bus (using I2C?).
// Signature found here:
// https://github.com/espressif/ESP8266_RTOS_SDK/blob/14171de0/components/esp8266/include/esp8266/rom_functions.h#L54
//
//export rom_i2c_writeReg
func rom_i2c_writeReg(block, host_id, reg_add, data uint8)

//export main
func main() {
	// Clear .bss section. .data has already been loaded by the ROM bootloader.
	preinit()

	// Initialize PLL.
	// I'm not quite sure what this magic incantation means, but it does set the
	// esp8266 to the right clock speed. Without this, it is running too slow.
	rom_i2c_writeReg(103, 4, 1, 136)
	rom_i2c_writeReg(103, 4, 2, 145)

	// Initialize UART.
	machine.InitSerial()

	// Initialize timer. Bits:
	//  ENABLE:   timer enable
	//  ROLLOVER: automatically reload when hitting 0
	//  PRESCALE: divide by 256
	esp.TIMER.FRC1_CTRL.Set(
		esp.TIMER_FRC1_CTRL_TIMER_ENABLE | esp.TIMER_FRC1_CTRL_ROLLOVER | esp.TIMER_FRC1_CTRL_PRESCALE_DIVIDER_DEVIDED_BY_256<<esp.TIMER_FRC1_CTRL_PRESCALE_DIVIDER_Pos)
	esp.TIMER.FRC1_LOAD.Set(0x3fffff)  // set all 22 bits to 1
	esp.TIMER.FRC1_COUNT.Set(0x3fffff) // set all 22 bits to 1

	run()

	// Fallback: if main ever returns, hang the CPU.
	exit(0)
}

//go:extern _sbss
var _sbss [0]byte

//go:extern _ebss
var _ebss [0]byte

func preinit() {
	// Initialize .bss: zero-initialized global variables.
	ptr := unsafe.Pointer(&_sbss)
	for ptr != unsafe.Pointer(&_ebss) {
		*(*uint32)(ptr) = 0
		ptr = unsafe.Pointer(uintptr(ptr) + 4)
	}
}

func ticks() timeUnit {
	// Get the counter value of the timer. It is 22 bits and starts with all
	// ones (0x3fffff). To make it easier to work with, let it count upwards.
	count := 0x3fffff - esp.TIMER.FRC1_COUNT.Get()

	// Replace the lowest 22 bits of the current time with the counter.
	newTime := (currentTime &^ 0x3fffff) | timeUnit(count)

	// If there was an overflow, the new time will be lower than the current
	// time, so will need to add (1<<22).
	if newTime < currentTime {
		newTime += 0x400000
	}

	// Update the timestamp for the next call to ticks().
	currentTime = newTime

	return currentTime
}

const tickNanos = 3200 // time.Second / (80MHz / 256)

func ticksToNanoseconds(ticks timeUnit) int64 {
	return int64(ticks) * tickNanos
}

func nanosecondsToTicks(ns int64) timeUnit {
	return timeUnit(ns / tickNanos)
}

// sleepTicks busy-waits until the given number of ticks have passed.
func sleepTicks(d timeUnit) {
	sleepUntil := ticks() + d
	for ticks() < sleepUntil {
	}
}

func exit(code int) {
	abort()
}

func abort() {
	for {
		device.Asm("waiti 0")
	}
}
