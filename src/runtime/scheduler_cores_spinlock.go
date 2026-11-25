//go:build (scheduler.cores && !rp2040 && !rp2350) || (tinygo.riscv && virt)

package runtime

import "sync/atomic"

var (
	printLock     spinLock
	schedulerLock schedulerSpinLock
	atomicsLock   spinLock
	futexLock     spinLock
)

type spinLock struct {
	atomic.Uint32
}

func (l *spinLock) Lock() {
	for !l.Uint32.CompareAndSwap(0, 1) {
		spinLoopWait()
	}
}

func (l *spinLock) Unlock() {
	if schedulerAsserts && l.Uint32.Load() != 1 {
		runtimeFatal("unlock of unlocked spinlock")
	}
	l.Uint32.Store(0)
}
