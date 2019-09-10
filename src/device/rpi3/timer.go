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
	// QEMUTryExit()
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
 * This code refers to a separate time on the Broadcom chip.

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
}*/

func EnableTimerIRQ() {
	Asm("msr daifclr, #2")
}

func DisableTimerIRQ() {
	Asm("msr daifset, #2")
}

func EnableInterruptController() {
	volatile.StoreUint32((*uint32)(ENABLE_IRQS_1), SYSTEM_TIMER_IRQ_1)
}

// the exported name is what connects it to the interrupt vector
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
func CounterFreq() uint32 {
	var val uint32
	AsmFull(`mrs x28, cntfrq_el0
		str x28,{val}`, map[string]interface{}{"val": &val})
	return val
}

//returns 0 when it doesn't want the timer anymore
var userCallbackFunc func(uint32, uint64) uint32

// clears the interrupt and writes target value (which is a time distance from now)
func SetCounterTargetInterval(val uint32, fn func(uint32, uint64) uint32) {
	AsmFull(`mov x28,{val}
		msr cntv_tval_el0, x28`, map[string]interface{}{"val": val})
	userCallbackFunc = fn
}

func CounterTargetVal() uint32 {
	var val uint32
	AsmFull(`mrs x28,cntv_tval_el0
		str x28,{val}`, map[string]interface{}{"val": &val})
	return val
}

func Core0CounterToCore0Irq() {
	volatile.StoreUint32((*uint32)(CORE0_TIMER_IRQCNTL), 0x08)
}

func EnableCounter() {
	Asm(`orr x29, xzr, #1
		msr cntv_ctl_el0, x29`)
}
func UDisableCounter() {
	Asm("msr cntv_ctl_el0, #0")
}

func WaitForInterrupt() {
	Asm("wfi")
}

func Core0TimerPending() uint32 {
	return volatile.LoadUint32((*uint32)(CORE0_IRQ_SOURCE))
}

func VirtualTimer() uint64 {
	var val uint64
	AsmFull(`mrs x28, cntvct_el0
		str x28,{val}`, map[string]interface{}{"val": &val})
	return val
}

// the export name is what connects it to the list of interrupt handlers
//go:export irq_el1h_handler
//go:extern
func HandleTimerInterrupt() {
	DisableTimerIRQ()
	//defer EnableTimerIRQ() --->  crashes the compiler
	//defer func() {
	// EnableTimerIRQ()
	//}()                    ---> also crashes the compiler
	// error:
	// tinygo build --target=rpi3_bl -o kernel8.elf .
	// Assertion failed: (From->getType() == To->getType()), function replaceDominatedUsesWith, file /Users/iansmith/tinygo.src/src/github.com/tinygo/tinygo/llvm-project/llvm/lib/Transforms/Utils/Local.cpp, line 2401.
	// SIGABRT: abort
	// PC=0x7fff6e16c2c6 m=3 sigcode=0
	pending := Core0TimerPending()
	if pending&0x08 != 0 {
		val := CounterTargetVal()
		virtual := VirtualTimer()
		if userCallbackFunc == nil {
			print("no int handler set for timer, aborting----\n")
			Abort()
		} else {
			next := userCallbackFunc(val, virtual)
			if next > 0 {
				SetCounterTargetInterval(next, userCallbackFunc)
			}
		}
	}
	EnableTimerIRQ()
}

// appears to be a bug in the compiler (LLVM) because it sys the svc code must be
// a 16 bit value, but the ARM docs say it can be 24 bits (which would allow this one)
// func QEMUTryExit() {
// 	Asm(`ldr x0,0x18
// 	ldr x1,=0x20026
// 	svc 0x123456`)
// }
