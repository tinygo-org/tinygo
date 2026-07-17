//go:build rp2040

package runtime

import (
	"device/arm"
	"device/rp"
	"runtime/interrupt"
	"runtime/volatile"
	"unsafe"
)

const (
	sioIrqFifoProc0 = rp.IRQ_SIO_IRQ_PROC0
	sioIrqFifoProc1 = rp.IRQ_SIO_IRQ_PROC1
)

const numSpinlocks = 32

// These spinlocks are needed by the runtime.
var (
	printLock     = spinLock{id: 20}
	schedulerLock = spinLock{id: 21}
	atomicsLock   = spinLock{id: 22}
	futexLock     = spinLock{id: 23}
)

func resetSpinLocks() {
	for i := uint8(0); i < numSpinlocks; i++ {
		l := &spinLock{id: i}
		l.spinlock().Set(0)
	}
}

// A hardware spinlock, one of the 32 spinlocks defined in the SIO peripheral.
type spinLock struct {
	id uint8
}

// Return the spinlock register: rp.SIO.SPINLOCKx
func (l *spinLock) spinlock() *volatile.Register32 {
	return (*volatile.Register32)(unsafe.Add(unsafe.Pointer(&rp.SIO.SPINLOCK0), l.id*4))
}

func (l *spinLock) Lock() {
	// Wait for the lock to be available.
	spinlock := l.spinlock()
	for spinlock.Get() == 0 {
		arm.Asm("wfe")
	}
}

func (l *spinLock) Unlock() {
	l.spinlock().Set(0)
	arm.Asm("sev")
}

// On RP2040, each core has its own SIO FIFO IRQ. Core0 enables
// IRQ_SIO_IRQ_PROC0 and Core1 enables IRQ_SIO_IRQ_PROC1, so each handler can
// use a fixed core ID.
func enableSIOFifoInterruptCore0() {
	intr := interrupt.New(sioIrqFifoProc0, handleSIOFifoInterruptCore0)
	intr.Enable()
	intr.SetPriority(0xff)
}

func enableSIOFifoInterruptCore1() {
	intr := interrupt.New(sioIrqFifoProc1, handleSIOFifoInterruptCore1)
	intr.Enable()
	intr.SetPriority(0xff)
}

func handleSIOFifoInterruptCore0(intr interrupt.Interrupt) {
	switch rp.SIO.FIFO_RD.Get() {
	case 1:
		gcInterruptHandler(0)
	}
}

func handleSIOFifoInterruptCore1(intr interrupt.Interrupt) {
	switch rp.SIO.FIFO_RD.Get() {
	case 1:
		gcInterruptHandler(1)
	}
}
