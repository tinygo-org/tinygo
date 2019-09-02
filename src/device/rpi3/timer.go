// +build rpi3

package rpi3

import "runtime/volatile"

func Abort() {
	print("program aborted\n")
	Asm("wfe")
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
