//go:build ch32v003

package runtime

import (
	"device/wch"
	"runtime/volatile"
)

//export main
func main() {
	// Initialize main clock.
	wch.RCC.CTLR.Set(wch.RCC_CTLR_HSION | wch.RCC_CTLR_PLLON)
	wch.RCC.CFGR0.Set(0<<wch.RCC_CFGR0_HPRE_Pos | 0<<wch.RCC_CFGR0_PLLSRC_Pos)
	wch.FLASH.ACTLR.Set(1 << wch.FLASH_ACTLR_LATENCY_Pos)
	wch.RCC.INTR.Set(0x009F0000)
	for (wch.RCC.CTLR.Get() & wch.RCC_CTLR_PLLRDY) == 0 {
	}
	wch.RCC.SetCFGR0_SW(0b10)
	for (wch.RCC.CFGR0.Get() & wch.RCC_CFGR0_SWS_Msk) != 0x08 {
	}

	// Enable all GPIO pins (port A, C and D).
	wch.RCC.APB2PCENR.SetBits(wch.RCC_APB2PCENR_IOPAEN | wch.RCC_APB2PCENR_IOPDEN | wch.RCC_APB2PCENR_IOPCEN)

	run()
	exit(0)
}

type timeUnit int64

var currentTicks timeUnit

func putchar(c byte) {
	// TODO
}

func exit(code int) {
	abort()
}

func ticks() timeUnit {
	return currentTicks
}

// ticksToNanoseconds converts RTC ticks (at 32768Hz) to nanoseconds.
func ticksToNanoseconds(ticks timeUnit) int64 {
	// The following calculation is actually the following, but with both sides
	// reduced to reduce the risk of overflow:
	//     ticks * 1e9 / 32768
	return int64(ticks) * 1953125 / 64
}

// nanosecondsToTicks converts nanoseconds to RTC ticks (running at 32768Hz).
func nanosecondsToTicks(ns int64) timeUnit {
	// The following calculation is actually the following, but with both sides
	// reduced to reduce the risk of overflow:
	//     ns * 32768 / 1e9
	return timeUnit(ns * 64 / 1953125)
}

var dummyVolatile volatile.Register32

func sleepTicks(ticks timeUnit) {
	// TODO: use a timer (using the low-speed clock)
	loops := ticks * 70
	for i := timeUnit(0); i < loops; i++ {
		dummyVolatile.Get()
	}
}

func abort() {
	for {
	}
}
