//go:build avr && !avrtiny

package runtime

import (
	"device/avr"
	"machine"
	"runtime/interrupt"
	"runtime/volatile"
	"unsafe"
)

const BOARD = "arduino"

// timeUnit in nanoseconds
type timeUnit int64

// Watchdog timer periods. These can be off by a large margin (hence the jump
// between 64ms and 125ms which is not an exact double), so don't rely on this
// for accurate time keeping.
const (
	WDT_PERIOD_16MS = iota
	WDT_PERIOD_32MS
	WDT_PERIOD_64MS
	WDT_PERIOD_125MS
	WDT_PERIOD_250MS
	WDT_PERIOD_500MS
	WDT_PERIOD_1S
	WDT_PERIOD_2S
)

const timerRecalibrateInterval = 6e7 // 1 minute

var nextTimerRecalibrate timeUnit

//go:extern _sbss
var _sbss [0]byte

//go:extern _ebss
var _ebss [0]byte

//export main
func main() {
	preinit()
	initHardware()
	run()
	exit(0)
}

func preinit() {
	// Initialize .bss: zero-initialized global variables.
	ptr := unsafe.Pointer(&_sbss)
	for ptr != unsafe.Pointer(&_ebss) {
		*(*uint8)(ptr) = 0
		ptr = unsafe.Add(ptr, 1)
	}
}

func initHardware() {
	initUART()
	initMonotonicTimer()
	nextTimerRecalibrate = ticks() + timerRecalibrateInterval

	// Enable interrupts after initialization.
	avr.Asm("sei")
}

func ticksToNanoseconds(ticks timeUnit) int64 {
	return int64(ticks)
}

func nanosecondsToTicks(ns int64) timeUnit {
	return timeUnit(ns)
}

// Sleep this number of ticks of nanoseconds.
func sleepTicks(d timeUnit) {
	waitTill := ticks() + d
	for {
		// wait for interrupt
		avr.Asm("sleep")
		if waitTill <= ticks() {
			// done waiting
			return
		}
		if hasScheduler {
			// The interrupt may have awoken a goroutine, so bail out early.
			return
		}
	}
}

func ticks() (ticksReturn timeUnit) {
	state := interrupt.Disable()
	// use volatile since ticksCount can be changed when running on multi-core boards.
	ticksReturn = timeUnit(volatile.LoadUint64((*uint64)(unsafe.Pointer(&ticksCount))))
	interrupt.Restore(state)
	return
}

func exit(code int) {
	abort()
}

func abort() {
	// Disable interrupts and go to sleep.
	// This can never be awoken except for reset, and is recognized as termination by simavr.
	avr.Asm("cli")
	for {
		avr.Asm("sleep")
	}
}

var ticksCount int64                // nanoseconds since start
var nanosecondsInTick int64 = 16000 // nanoseconds per each tick

func initMonotonicTimer() {
	ticksCount = 0

	interrupt.New(avr.IRQ_TIMER0_OVF, func(i interrupt.Interrupt) {
		// use volatile
		ticks := volatile.LoadUint64((*uint64)(unsafe.Pointer(&ticksCount)))
		ticks += uint64(nanosecondsInTick)
		volatile.StoreUint64((*uint64)(unsafe.Pointer(&ticksCount)), ticks)
	})

	// initial initialization of the Timer0
	// - Mask interrupt
	avr.TIMSK0.ClearBits(avr.TIMSK0_TOIE0 | avr.TIMSK0_OCIE0A | avr.TIMSK0_OCIE0B)

	// - Write new values to TCNT2, OCR2x, and TCCR2x.
	avr.TCNT0.Set(0)
	avr.OCR0A.Set(0xff)
	// - Set mode 3
	avr.TCCR0A.Set(avr.TCCR0A_WGM00 | avr.TCCR0A_WGM01)
	// - Set prescaler 1
	avr.TCCR0B.Set(avr.TCCR0B_CS00)

	// - Unmask interrupt
	avr.TIMSK0.SetBits(avr.TIMSK0_TOIE0)
}

//go:linkname adjustMonotonicTimer machine.adjustMonotonicTimer
func adjustMonotonicTimer() {
	// adjust the nanosecondsInTick using volatile
	mask := interrupt.Disable()
	volatile.StoreUint64((*uint64)(unsafe.Pointer(&nanosecondsInTick)), uint64(currentNanosecondsInTick()))
	interrupt.Restore(mask)
}

func currentNanosecondsInTick() int64 {
	// this time depends on clk_IO, prescale, mode and OCR0A
	// assuming the clock source is CPU clock
	prescaler := int64(avr.TCCR0B.Get() & 0x7)
	clock := (int64(1e12) / prescaler) / int64(machine.CPUFrequency())
	mode := avr.TCCR0A.Get() & 0x7

	/*
	 Mode WGM02 WGM01 WGM00 Timer/Counter       TOP  Update of  TOV Flag
	                        Mode of Operation        OCRx at    Set on
	 0    0     0     0     Normal              0xFF Immediate  MAX
	 1    0     0     1     PWM, Phase Correct  0xFF TOP        BOTTOM
	 2    0     1     0     CTC                 OCRA Immediate  MAX
	 3    0     1     1     Fast PWM            0xFF BOTTOM     MAX
	 5    1     0     1     PWM, Phase Correct  OCRA TOP        BOTTOM
	 7    1     1     1     Fast PWM            OCRA BOTTOM     TOP
	*/
	switch mode {
	case 0, 3:
		// normal & fast PWM
		// TOV0 Interrupt when moving from MAX (0xff) to 0x00
		return clock * 256 / 1000
	case 1:
		// Phase Correct PWM
		// TOV0 Interrupt when moving from MAX (0xff) to 0x00
		return clock * 256 * 2 / 1000
	case 2, 7:
		// CTC & fast PWM
		// TOV0 Interrupt when moving from MAX (OCRA) to 0x00
		return clock * int64(avr.OCR0A.Get()) / 1000
	case 5:
		// Phase Correct PWM
		// TOV0 Interrupt when moving from MAX (OCRA) to 0x00
		return clock * int64(avr.OCR0A.Get()) * 2 / 1000
	}

	return clock / 1000 // for unknown
}
