// +build rpi3

package rpi3

import (
	"runtime/volatile"
)

var interruptVector [16 * 0x80]byte

var unexpectedInterruptNames = []string{
	"synchronous el1t",
	"irq el1t",
	"fiq el1t",
	"error el1t",

	"synchronous el1h",
	"irq el1h",
	"fiq el1h",
	"error el1h",

	"synchronous el0 64bit",
	"irq el0 64bit",
	"fiq el0 64bit",
	"error el0 64bit",

	"synchronous el0 32bit",
	"irq el0 32bit",
	"fiq el0 32bit",
	"error el0 32bit",
}

//go:export abort
func Abort() {
	print("program aborted\n")
	for {

	}
}

// wait a given number of CPU cycles (at least)
func WaitCycles(n int) {
	for n > 0 {
		Asm("nop")
		n--
	}
}

func WaitMuSec(n uint32) {
	var f, t, r uint64
	AsmFull(`mrs x28, cntfrq_el0
		str x28,{f}
		mrs x27, cntpct_el0
		str x27,{t}`, map[string]interface{}{"f": &f, "t": &t})
	//expires at t
	t += ((f / 1000) * uint64(n)) / 1000
	for r < t {
		AsmFull(`mrs x27, cntpct_el0
			str x27,{r}`, map[string]interface{}{"r": &r})
	}
}

func SysTimer() uint64 {
	h := uint32(0xffffffff)
	var l uint32

	// the reads from MMIO are are two separate 32 bit reads
	h = volatile.LoadUint32((*uint32)(SYSTMR_HI))
	l = volatile.LoadUint32((*uint32)(SYSTMR_LO))
	//the read of h can fail
	again := volatile.LoadUint32((*uint32)(SYSTMR_HI))
	if h != again {
		h = volatile.LoadUint32((*uint32)(SYSTMR_HI))
		l = volatile.LoadUint32((*uint32)(SYSTMR_LO))
	}
	high := uint64(h << 32)
	return high | uint64(l)
}

/**
 * Wait N microsec (with BCM System Timer)
 */
func WaitMuSecST(n uint32) {
	t := SysTimer()
	// we must check if it's non-zero, because qemu does not emulate
	// system timer, and returning constant zero would mean infinite loop
	if t == 0 {
		return
	}
	for SysTimer() < t+uint64(n) {

	}
}

/*
 * Timer on IRQ
 */

var curVal uint32 = 0

func TimerInit(interval uint32) {
	curVal = volatile.LoadUint32((*uint32)(TIMER_CLO))
	curVal += interval
	volatile.StoreUint32((*uint32)(TIMER_C1), curVal)
}

func HandleTimerIRQ(interval uint32) {
	curVal += interval
	volatile.StoreUint32((*uint32)(TIMER_C1), curVal)
	volatile.StoreUint32((*uint32)(TIMER_CS), TIMER_CS_M1)
}

func EnableTimerIRQ() {
	Asm("msr daifclr, #2")
}

func DisableTimerIRQ() {
	Asm("msr daifset, #2")
}

func EnableInterruptController() {
	volatile.StoreUint32((*uint32)(ENABLE_IRQS_1), SYSTEM_TIMER_IRQ_1)
}

//go:export unexpectedInterrupt
func UnexpectedInterrupt(n int, esr uint64, address uint64) {
	if n < 0 || n >= len(unexpectedInterruptNames) {
		print("unexpected interrupt called with bad index", n, "\n")
		return
	}
	print("unexpected interrupt:", unexpectedInterruptNames[n], " esr:", esr, " address:")
	UART0Hex64(address)
}

// get the counter frequency
func QEMUCounterFreq() uint32 {
	var val uint32
	AsmFull(`mrs x28, cntfrq_el0
		str x28,{val}`, map[string]interface{}{"val": &val})
	return val
}

//returns 0 when it doesn't want the timer anymore
var userCallbackFunc func(uint32, uint64) uint32

// clears the interrupt and writes target value (which is a time distance from now)
func QEMUSetCounterTargetInterval(val uint32, fn func(uint32, uint64) uint32) {
	print("setting target value:")
	UART0Hex(val)
	AsmFull(`mov x28,{val}
		msr cntv_tval_el0, x28`, map[string]interface{}{"val": val})
	userCallbackFunc = fn
}

func QEMUCounterTargetVal() uint32 {
	var val uint32
	AsmFull(`mrs x28,cntv_tval_el0
		str x28,{val}`, map[string]interface{}{"val": &val})
	return val
}

func QEMUCore0CounterToCore0Irq() {
	volatile.StoreUint32((*uint32)(CORE0_TIMER_IRQCNTL), 0x08)
}

func QEMUEnableCounter() {
	Asm(`orr x29, xzr, #1
		msr cntv_ctl_el0, x29`)
}
func QEMUDisableCounter() {
	Asm("msr cntv_ctl_el0, #0")
}

func WaitForInterrupt() {
	Asm("wfi")
}

func QEMUCore0TimerPending() uint32 {
	return volatile.LoadUint32((*uint32)(CORE0_IRQ_SOURCE))
}

func QEMUVirtualTimer() uint64 {
	var val uint64
	AsmFull(`mrs x28, cntvct_el0
		str x28,{val}`, map[string]interface{}{"val": &val})
	return val
}

//go:export irq_el1h_handler
//go:extern
func QEMUHandleTimerInterrupt() {
	DisableTimerIRQ()
	pending := QEMUCore0TimerPending()
	if pending&0x08 != 0 {
		val := QEMUCounterTargetVal()
		virtual := QEMUVirtualTimer()
		next := userCallbackFunc(val, virtual)
		if next > 0 {
			QEMUSetCounterTargetInterval(next, userCallbackFunc)
		}
	}
	EnableTimerIRQ()
}

func QEMUTryExit() {
	Asm(`ldr x0,0x18
	ldr x1,=0x20026
	svc =0x00123456`)
}
