// +build mimxrt1062

package usb2

// Implementation of USB device controller interface (dci) for NXP iMXRT1062.

import (
	"device/arm"
	"device/nxp"
	"runtime/interrupt"
	"runtime/volatile"
	"strconv"
)

// dciCount defines the number of USB cores to configure for device mode. It is
// computed as the sum of all declared device configuration descriptors.
const dciCount = descCDCACMConfigCount

// dciInterruptPriority defines the priority for all USB device interrupts.
const dciInterruptPriority = 3

// deviceController implements USB device controller interface (dci).
type deviceController struct {
	core *core // Parent USB core this instance is attached to
	port int   // USB port index
	id   int   // deviceControllerInstance index

	bus *nxp.USB_Type
	phy *nxp.USBPHY_Type
	irq interrupt.Interrupt

	cri volatile.Register8 // set to 1 if in critical section, else 0
	ivm uintptr            // interrupt state when entering critical section
}

// deviceControllerInstance provides statically-allocated instances of each USB
// device controller configured on this platform.
var deviceControllerInstance [dciCount]deviceController

// initDCI initializes and assigns a free device controller instance to the
// given USB port. Returns the initialized device controller or nil if no free
// device controller instances remain.
func initDCI(port int) (dci, status) {
	if 0 == dciCount {
		return nil, statusInvalidArgument // must have defined device descriptors
	}
	// Return the first instance whose assigned core is currently nil.
	for i := range deviceControllerInstance {
		if nil == deviceControllerInstance[i].core {
			// Initialize device controller.
			deviceControllerInstance[i].core = &coreInstance[port]
			deviceControllerInstance[i].port = port
			deviceControllerInstance[i].id = i
			switch port {
			case 0:
				deviceControllerInstance[i].bus = nxp.USB1
				deviceControllerInstance[i].phy = nxp.USBPHY1
				deviceControllerInstance[i].irq =
					interrupt.New(nxp.IRQ_USB_OTG1,
						func(interrupt.Interrupt) {
							coreInstance[0].dc.interrupt()
						})

			case 1:
				deviceControllerInstance[i].bus = nxp.USB2
				deviceControllerInstance[i].phy = nxp.USBPHY2
				deviceControllerInstance[i].irq =
					interrupt.New(nxp.IRQ_USB_OTG2,
						func(interrupt.Interrupt) {
							//coreInstance[1].dc.interrupt()
						})
			}
			return &deviceControllerInstance[i], statusOK
		}
	}
	return nil, statusBusy // No free device controller instances available.
}

func (dc *deviceController) init() status {

	dc.bus.BURSTSIZE.Set(0x0404)

	// if dc.phy.PWD.HasBits((nxp.USBPHY_PWD_RXPWDRX | nxp.USBPHY_PWD_RXPWDDIFF |
	// 	nxp.USBPHY_PWD_RXPWD1PT1 | nxp.USBPHY_PWD_RXPWDENV |
	// 	nxp.USBPHY_PWD_TXPWDV2I | nxp.USBPHY_PWD_TXPWDIBIAS |
	// 	nxp.USBPHY_PWD_TXPWDFS)) ||
	// 	dc.bus.USBMODE.HasBits(nxp.USB_USBMODE_CM_Msk) {
	// 	// reset controller if it was already enabled
	// 	dc.phy.CTRL_SET.Set(nxp.USBPHY_CTRL_SFTRST)
	// 	dc.bus.USBCMD.SetBits(nxp.USB_USBCMD_RST)
	// 	for dc.bus.USBCMD.HasBits(nxp.USB_USBCMD_RST) {
	// 	}
	// 	// clear interrupts
	// 	m := arm.DisableInterrupts()
	// 	switch dc.port {
	// 	case 0:
	// 		arm.EnableInterrupts(m & ^uintptr(nxp.IRQ_USB_OTG1))
	// 	case 1:
	// 		arm.EnableInterrupts(m & ^uintptr(nxp.IRQ_USB_OTG2))
	// 	}
	// 	dc.phy.CTRL_CLR.Set(nxp.USBPHY_CTRL_SFTRST)
	// }

	// reset the controller
	dc.phy.CTRL_SET.Set(nxp.USBPHY_CTRL_SFTRST)
	dc.bus.USBCMD.SetBits(nxp.USB_USBCMD_RST)
	for dc.bus.USBCMD.HasBits(nxp.USB_USBCMD_RST) {
	}
	// clear interrupts
	m := arm.DisableInterrupts()
	switch dc.port {
	case 0:
		arm.EnableInterrupts(m & ^uintptr(nxp.IRQ_USB_OTG1))
	case 1:
		arm.EnableInterrupts(m & ^uintptr(nxp.IRQ_USB_OTG2))
	}
	dc.phy.CTRL_CLR.Set(nxp.USBPHY_CTRL_CLKGATE | nxp.USBPHY_CTRL_SFTRST)
	dc.phy.PWD.Set(0)

	// clear the controller mode field and set to device mode:
	//   controller mode (CM) 0x0=idle, 0x2=device-only, 0x3=host-only
	dc.bus.USBMODE.ReplaceBits(nxp.USB_USBMODE_CM_CM_2,
		nxp.USB_USBMODE_CM_Msk>>nxp.USB_USBMODE_CM_Pos, nxp.USB_USBMODE_CM_Pos)

	dc.bus.USBCMD.ClearBits(nxp.USB_USBCMD_ITC_Msk)  // no interrupt threshold
	dc.bus.USBMODE.SetBits(nxp.USB_USBMODE_SLOM_Msk) // disable setup lockout
	dc.bus.USBMODE.ClearBits(nxp.USB_USBMODE_ES_Msk) // use little-endianness

	// configure ENDPOINTLISTADDR

	// enable interrupts
	dc.bus.USBINTR.Set(
		nxp.USB_USBINTR_UE_Msk | // bus enable
			nxp.USB_USBINTR_UEE_Msk | // bus error
			nxp.USB_USBINTR_PCE_Msk | // port change detect
			nxp.USB_USBINTR_URE_Msk | // bus reset
			nxp.USB_USBINTR_SLE) // sleep enable

	// ensure D+ pulled down long enough for host to detect previous disconnect
	dc.udelay(5000)

	return statusOK
}

func (dc *deviceController) enable(enable bool) status {

	dc.irq.SetPriority(dciInterruptPriority)
	dc.irq.Enable()

	dc.bus.USBCMD.SetBits(nxp.USB_USBCMD_RS)

	return statusOK
}

func (dc *deviceController) critical(enter bool) status {
	if enter {
		// check if critical section already locked
		if dc.cri.Get() != 0 {
			return statusRetry
		}
		// lock critical section
		dc.cri.Set(1)
		// disable interrupts, storing state in receiver
		dc.ivm = arm.DisableInterrupts()
	} else {
		// ensure critical section is locked
		if dc.cri.Get() != 0 {
			// re-enable interrupts, using state stored in receiver
			arm.EnableInterrupts(dc.ivm)
			// unlock critical section
			dc.cri.Set(0)
		}
	}
	return statusOK
}

func (dc *deviceController) interrupt() {
	// read and clear the interrupts that fired
	status := dc.bus.USBSTS.Get() & dc.bus.USBINTR.Get()
	dc.bus.USBSTS.Set(status)

	println(strconv.FormatUint(uint64(status), 16))
}

// udelay waits for the given number of microseconds before returning.
// We cannot use the sleep timer from this context (import cycle), but we need
// an approximate method to spin CPU cycles for short periods of time.
//go:inline
func (dc *deviceController) udelay(microsec uint32) {
	n := cycles(microsec, descCPUFrequencyHz)
	for i := uint32(0); i < n; i++ {
		arm.Asm(`nop`)
	}
}
