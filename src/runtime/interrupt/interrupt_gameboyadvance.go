// +build gameboyadvance

package interrupt

import (
	"runtime/volatile"
	"unsafe"
)

var (
	regInterruptEnable       = (*volatile.Register16)(unsafe.Pointer(uintptr(0x4000200)))
	regInterruptRequestFlags = (*volatile.Register16)(unsafe.Pointer(uintptr(0x4000202)))
	regInterruptMasterEnable = (*volatile.Register16)(unsafe.Pointer(uintptr(0x4000208)))
)

// Enable enables this interrupt. Right after calling this function, the
// interrupt may be invoked if it was already pending.
func (irq Interrupt) Enable() {
	regInterruptEnable.SetBits(1 << uint(irq.num))
}

//export handleInterrupt
func handleInterrupt() {
	flags := regInterruptRequestFlags.Get()
	for i := 0; i < 14; i++ {
		if flags&(1<<uint(i)) != 0 {
			regInterruptRequestFlags.Set(1 << uint(i)) // acknowledge interrupt
			callInterruptHandler(i)
		}
	}
}

// callInterruptHandler is a compiler-generated function that calls the
// appropriate interrupt handler for the given interrupt ID.
//go:linkname callInterruptHandler runtime.callInterruptHandler
func callInterruptHandler(id int)
