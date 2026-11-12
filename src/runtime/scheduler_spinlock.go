//go:build scheduler.cores || rp2040 || rp2350 || (tinygo.riscv && virt && qemu)

package runtime

import "runtime/interrupt"

// schedulerSpinLock keeps interrupts off while held. An interrupt handler on
// this core that wakes a goroutine would otherwise spin on a lock that the code
// it interrupted holds, and a core spinning with interrupts on could be parked
// by the GC while holding the lock.
type schedulerSpinLock struct {
	spinLock
	state interrupt.State
}

func (l *schedulerSpinLock) Lock() {
	state := interrupt.Disable()
	l.spinLock.Lock()
	l.state = state
}

func (l *schedulerSpinLock) Unlock() {
	state := l.state
	l.spinLock.Unlock()
	interrupt.Restore(state)
}

// unlockMasked releases the lock and leaves interrupts off. Pass the returned
// state to lockMasked.
func (l *schedulerSpinLock) unlockMasked() interrupt.State {
	state := l.state
	l.spinLock.Unlock()
	return state
}

// lockMasked takes the lock with interrupts already off and records the state
// that Unlock will restore.
func (l *schedulerSpinLock) lockMasked(state interrupt.State) {
	l.spinLock.Lock()
	l.state = state
}
