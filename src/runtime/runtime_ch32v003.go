//go:build ch32v003

package runtime

import "device/ch32"

var currentTicks timeUnit

// startup and general functionality mostly inspired by https://github.com/cnlohr/ch32fun

//export main
func main() {
	// setup flash latency for higher clock speeds
	ch32.FLASH.SetACTLR_LATENCY(1)
	// no scaling on AHB clock source
	ch32.RCC.SetCFGR0_HPRE(0)
	// set PLL clock source to undivided HSI
	ch32.RCC.SetCFGR0_PLLSRC(0)
	// enable high speed clock
	ch32.RCC.SetCTLR_HSION(1)
	// enable PLL clock
	ch32.RCC.SetCTLR_PLLON(1)
	// clear clock ready flags
	ch32.RCC.SetINTR_LSIRDYC(1)
	ch32.RCC.SetINTR_HSIRDYC(1)
	ch32.RCC.SetINTR_HSERDYC(1)
	ch32.RCC.SetINTR_PLLRDYC(1)
	ch32.RCC.SetINTR_CSSC(1)

	for ch32.RCC.GetCTLR_PLLRDY() == 0 {
		// wait for PLL to be ready
	}
	// set system clock source to PLL
	ch32.RCC.SetCFGR0_SW(0b10)

	for ch32.RCC.GetCFGR0_SWS() != 0b10 {
		// wait for system clock to switch to PLL
	}

	// enable clock on IO ports
	ch32.RCC.SetAPB2PCENR_AFIOEN(1)
	ch32.RCC.SetAPB2PCENR_IOPAEN(1)
	ch32.RCC.SetAPB2PCENR_IOPCEN(1)
	ch32.RCC.SetAPB2PCENR_IOPDEN(1)

	//Enable SysTick with HCLK/1
	ch32.SYSTICK.SetCTLR_STE(1)
	ch32.SYSTICK.SetCTLR_STCLK(1)

	run()
	exit(0)
}

func ticks() timeUnit {
	return currentTicks
}

func ticksToNanoseconds(ticks timeUnit) int64 {
	// Convert ticks to nanoseconds based on the core clock frequency (48 Mhz) 1_000_000_000 / 48_000_000 = 125 / 6.
	return int64(ticks) * 125 / 6
}

func nanosecondsToTicks(ns int64) timeUnit {
	// Convert nanoseconds to ticks based on the core clock frequency (48 Mhz) 48_000_000 / 1_000_000_000 = 6 / 125.
	return timeUnit(ns * 6 / 125)
}

func sleepTicks(ticks timeUnit) {
	target := ticks + timeUnit(ch32.SYSTICK.CNT.Get())
	for timeUnit(ch32.SYSTICK.CNT.Get())-target < 0 {
		// wait until the target tick count is reached
	}
}

func putchar(c byte) {
	// dummy, TODO
}

func exit(code int) {
	abort()
}

func abort() {
	for {
		// lock up forever
	}
}
