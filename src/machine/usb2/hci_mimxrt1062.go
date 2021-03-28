// +build mimxrt1062

package usb2

// Implementation of USB host controller interface (hci) for NXP iMXRT1062.

import (
	"device/arm"
	"device/nxp"
	"runtime/interrupt"
	"runtime/volatile"
)

// hciCount defines the number of USB cores to configure for host mode. It is
// computed as the sum of all declared host configuration descriptors.
const hciCount = 0

// hciInterruptPriority defines the priority for all USB host interrupts.
const hciInterruptPriority = 3

// hostController implements USB host controller interface (hci).
type hostController struct {
	core *core // Parent USB core this instance is attached to
	port int   // USB port index
	id   int   // hostControllerInstance index

	bus *nxp.USB_Type
	phy *nxp.USBPHY_Type
	irq interrupt.Interrupt

	cri volatile.Register8 // set to 1 if in critical section, else 0
	ivm uintptr            // interrupt state when entering critical section
}

// hostControllerInstance provides statically-allocated instances of each USB
// host controller configured on this platform.
var hostControllerInstance [hciCount]hostController

// initHCI initializes and assigns a free host controller instance to the given
// USB port. Returns the initialized host controller or nil if no free host
// controller instances remain.
func initHCI(port int) (hci, status) {
	if 0 == hciCount {
		return nil, statusInvalidArgument // must have defined host descriptors
	}
	// Return the first instance whose assigned core is currently nil.
	for i := range hostControllerInstance {
		if nil == hostControllerInstance[i].core {
			// Initialize host controller.
			hostControllerInstance[i].core = &coreInstance[port]
			hostControllerInstance[i].port = port
			hostControllerInstance[i].id = i
			switch port {
			case 0:
				hostControllerInstance[i].bus = nxp.USB1
				hostControllerInstance[i].phy = nxp.USBPHY1
				hostControllerInstance[i].irq =
					interrupt.New(nxp.IRQ_USB_OTG1,
						func(interrupt.Interrupt) {
							coreInstance[0].hc.interrupt()
						})

			case 1:
				hostControllerInstance[i].bus = nxp.USB2
				hostControllerInstance[i].phy = nxp.USBPHY2
				hostControllerInstance[i].irq =
					interrupt.New(nxp.IRQ_USB_OTG2,
						func(interrupt.Interrupt) {
							//coreInstance[1].hc.interrupt()
						})
			}
			return &hostControllerInstance[i], statusOK
		}
	}
	return nil, statusBusy // No free host controller instances available.
}

func (hc *hostController) init() status {

	return statusOK
}

func (hc *hostController) enable(enable bool) status {

	hc.irq.SetPriority(hciInterruptPriority)
	hc.irq.Enable()

	return statusOK
}

func (hc *hostController) critical(enter bool) status {
	if enter {
		// check if critical section already locked
		if hc.cri.Get() != 0 {
			return statusRetry
		}
		// lock critical section
		hc.cri.Set(1)
		// disable interrupts, storing state in receiver
		hc.ivm = arm.DisableInterrupts()
	} else {
		// ensure critical section is locked
		if hc.cri.Get() != 0 {
			// re-enable interrupts, using state stored in receiver
			arm.EnableInterrupts(hc.ivm)
			// unlock critical section
			hc.cri.Set(0)
		}
	}
	return statusOK
}

func (hc *hostController) interrupt() {

}

// udelay waits for the given number of microseconds before returning.
// We cannot use the sleep timer from this context (import cycle), but we need
// an approximate method to spin CPU cycles for short periods of time.
//go:inline
func (hc *hostController) udelay(microsec uint32) {
	n := cycles(microsec, descCPUFrequencyHz)
	for i := uint32(0); i < n; i++ {
		arm.Asm(`nop`)
	}
}
