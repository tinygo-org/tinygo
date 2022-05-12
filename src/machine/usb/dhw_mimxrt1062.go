//go:build mimxrt1062
// +build mimxrt1062

package usb

// Implementation of USB device controller hardware abstraction (dhw) for NXP
// iMXRT1062.

import (
	"device/arm"
	"device/nxp"
	"math/bits"
	"runtime/interrupt"
	"runtime/volatile"
	"unsafe"
)

// dhwInterruptPriority defines the priority for all USB device interrupts.
const dhwInterruptPriority = 3

// dhw implements USB device controller hardware abstraction for iMXRT1062.
type dhw struct {
	*dcd // USB device controller driver

	bus *nxp.USB_Type       // USB core register
	phy *nxp.USBPHY_Type    // USB PHY register
	irq interrupt.Interrupt // USB IRQ, only a single interrupt on iMXRT1062

	stat *dhwEndpoint // endpoint 0 Rx ("out" direction)
	ctrl *dhwEndpoint // endpoint 0 Tx ("in" direction)

	speed Speed

	controlReply [8]uint8
	controlMask  uint32
	endpointMask uint32
	setup        dcdSetup
	stage        dcdStage

	timerInterrupt [2]func()
	timerReboot    uint8
	sofUsage       uint8
}

func runBootloader() { arm.Asm(`bkpt #251`) }

// cycleCount uses the ARM debug cycle counter available on iMXRT1062 (enabled
// in runtime_mimxrt1062_time.go) to return the number of CPU cycles since boot.
//go:inline
func cycleCount() uint32 {
	return (*volatile.Register32)(unsafe.Pointer(uintptr(0xe0001004))).Get()
}

// deleteCache deletes cached data without touching physical memory. Useful for
// receiving data via DMA, which writes directly to memory, as this will force
// subsequent reads to ignore cache and access physical memory.
func deleteCache(addr, size uintptr) { nxp.DeleteDcache(addr, size) }

// flushCache immediately flushes cached data to physical memory. Useful for
// transmitting data via DMA, which reads directly from memory, as this will
// immediately flush data currently in cache to physical memory. This also
// purges the data from cache, since we no longer need to access it after
// priming DMA for transmission.
func flushCache(addr, size uintptr) { nxp.FlushDeleteDcache(addr, size) }

// allocDHW returns a reference to the USB hardware abstraction for the given
// device controller driver. Should be called only one time and during device
// controller initialization.
func allocDHW(port, instance int, speed Speed, dc *dcd) *dhw {
	switch port {
	case 0:
		dhwInstance[instance].dcd = dc
		dhwInstance[instance].bus = nxp.USB1
		dhwInstance[instance].phy = nxp.USBPHY1
		dhwInstance[instance].irq =
			interrupt.New(nxp.IRQ_USB_OTG1,
				func(interrupt.Interrupt) {
					coreInstance[0].dc.interrupt()
				})

	case 1:
		dhwInstance[instance].dcd = dc
		dhwInstance[instance].bus = nxp.USB2
		dhwInstance[instance].phy = nxp.USBPHY2
		dhwInstance[instance].irq =
			interrupt.New(nxp.IRQ_USB_OTG2,
				func(interrupt.Interrupt) {
					coreInstance[1].dc.interrupt()
				})
	}

	// Both ports default to high-speed (480 Mbit/sec) on Teensy 4.x
	if 0 == speed {
		speed = HighSpeed
	}
	dhwInstance[instance].speed = speed

	return &dhwInstance[instance]
}

// init configures the USB port for device mode operation by initializing all
// endpoint and transfer descriptor data structures, initializing core registers
// and interrupts, resetting the USB PHY, and enabling power on the bus.
func (d *dhw) init() status {

	// Reset the controller
	d.phy.CTRL_SET.Set(nxp.USBPHY_CTRL_SFTRST)
	d.bus.USBCMD.SetBits(nxp.USB_USBCMD_RST)
	for d.bus.USBCMD.HasBits(nxp.USB_USBCMD_RST) {
	}

	// Initialize USB interrupt priorities
	d.irq.SetPriority(dhwInterruptPriority)

	// Clear interrupts
	m := arm.DisableInterrupts()
	switch d.port {
	case 0:
		m &^= uintptr(nxp.IRQ_USB_OTG1)
	case 1:
		m &^= uintptr(nxp.IRQ_USB_OTG2)
	}
	arm.EnableInterrupts(m)

	// Initiate reset, and enable PHY power supply
	d.phy.CTRL_CLR.Set(nxp.USBPHY_CTRL_CLKGATE | nxp.USBPHY_CTRL_SFTRST)
	d.phy.PWD.Set(0) // ["Power down"] 0 = Power enabled, 1 = Power disabled

	// Clear the controller mode field and set to device mode:
	//   Controller mode (CM) 0x0=idle, 0x2=device-only, 0x3=host-only
	d.bus.USBMODE.ReplaceBits(nxp.USB_USBMODE_CM_CM_2,
		nxp.USB_USBMODE_CM_Msk>>nxp.USB_USBMODE_CM_Pos, nxp.USB_USBMODE_CM_Pos)

	d.bus.USBCMD.ClearBits(nxp.USB_USBCMD_ITC_Msk)  // No interrupt threshold
	d.bus.USBMODE.SetBits(nxp.USB_USBMODE_SLOM_Msk) // Disable setup lockout
	d.bus.USBMODE.ClearBits(nxp.USB_USBMODE_ES_Msk) // Use little-endianness

	// Initialize control endpoint 0 (stat = Rx/OUT, ctrl = Tx/IN)
	d.stat = d.endpointQueueHead(rxEndpoint(0))
	d.ctrl = d.endpointQueueHead(txEndpoint(0))
	d.stat.config = (descEndptMaxPktSize << 16) | (1 << 15)
	d.ctrl.config = (descEndptMaxPktSize << 16)

	// Install base address of endpoints
	d.bus.ASYNCLISTADDR.Set(uint32(uintptr(unsafe.Pointer(d.stat))))

	// Clear installed timer callbacks
	d.timerInterrupt[0] = nil
	d.timerInterrupt[1] = nil

	// Enable interrupts in USB core
	d.bus.USBINTR.Set(
		nxp.USB_USBINTR_UE_Msk | // bus enable
			nxp.USB_USBINTR_UEE_Msk | // bus error
			nxp.USB_USBINTR_PCE_Msk | // port change detect
			nxp.USB_USBINTR_URE_Msk | // bus reset
			nxp.USB_USBINTR_SLE) // sleep enable

	// Ensure D+ pulled down long enough for host to detect previous disconnect
	udelay(5000)

	return statusOK
}

// enable causes the USB core to enter (or exit) the normal run state and
// enables/disables all interrupts on the receiver's USB port.
func (d *dhw) enable(enable bool) {
	if enable {
		d.irq.Enable()                          // Enable USB interrupts
		d.bus.USBCMD.SetBits(nxp.USB_USBCMD_RS) // Enable "run" state
	} else {
		d.bus.USBCMD.ClearBits(nxp.USB_USBCMD_RS) // Disable "run" state
		d.irq.Disable()                           // Disable USB interrupts
	}
}

// enableInterrupts enables/disables all interrupts on the receiver's USB port.
func (d *dhw) enableInterrupts(enable bool) {
	if enable {
		d.irq.Enable()
	} else {
		d.irq.Disable()
	}
}

// enableSOF enables or disables start-of-frame (SOF) interrupts on the given
// USB device interface.
func (d *dhw) enableSOF(enable bool, iface uint8) {
	if enable {
		ivm := arm.DisableInterrupts()
		d.sofUsage |= 1 << iface
		if !d.bus.USBINTR.HasBits(nxp.USB_USBINTR_SRE) {
			d.bus.USBSTS.Set(nxp.USB_USBSTS_SRI)
			d.bus.USBINTR.SetBits(nxp.USB_USBINTR_SRE)
		}
		arm.EnableInterrupts(ivm)
		d.timerReboot = 80
	} else {
		d.timerReboot = 0
		d.sofUsage &^= 1 << iface
		if 0 == d.sofUsage {
			d.bus.USBINTR.ClearBits(nxp.USB_USBINTR_SRE)
		}
	}
}

// interrupt handles the USB hardware interrupt events and notifies the device
// controller driver using a common "virtual interrupt" code.
func (d *dhw) interrupt() {

	// read and clear the interrupts that fired
	status := d.bus.USBSTS.Get() & d.bus.USBINTR.Get()
	d.bus.USBSTS.Set(status)

	// USB Interrupt (USBINT) - R/WC
	// This bit is set by the Host/Device Controller when the cause of an
	// interrupt is a completion of a USB transaction where the Transfer
	// Descriptor (TD) has an interrupt on complete (IOC) bit set.
	// This bit is also set by the Host/Device Controller when a short packet is
	// detected. A short packet is when the actual number of bytes received was
	// less than the expected number of bytes.
	if 0 != status&nxp.USB_USBSTS_UI {

		setupStatus := d.bus.ENDPTSETUPSTAT.Get()
		for 0 != setupStatus {
			d.bus.ENDPTSETUPSTAT.Set(setupStatus)
			var setup dcdSetup
			ready := false
			for !ready {
				d.bus.USBCMD.SetBits(nxp.USB_USBCMD_SUTW)
				setup = d.stat.setup
				ready = d.bus.USBCMD.HasBits(nxp.USB_USBCMD_SUTW)
			}
			d.bus.USBCMD.ClearBits(nxp.USB_USBCMD_SUTW)
			// flush endpoint 0 (bit 0=Rx, 16=Tx)
			d.bus.ENDPTFLUSH.Set(0x00010001)
			// wait for flush to complete
			for d.bus.ENDPTFLUSH.HasBits(0x00010001) {
			}
			// Reset notify mask for control endpoint 0
			d.controlMask = 0
			// Notify device controller driver
			d.event(dcdEvent{
				id:    dcdEventControlSetup,
				setup: setup,
			})
			setupStatus = d.bus.ENDPTSETUPSTAT.Get()
		}

		completeStatus := d.bus.ENDPTCOMPLETE.Get()
		if 0 != completeStatus {
			d.bus.ENDPTCOMPLETE.Set(completeStatus)
			if 0 != completeStatus&d.controlMask {
				// Clear notify mask for control endpoint 0
				d.controlMask = 0
				// Notify device controller driver, which invokes any appropriate
				// callback(s) for the current device class configuration.
				d.controlComplete()
			}
			completeStatus &= d.endpointMask
			if 0 != completeStatus {
				tx := completeStatus >> 16
				for 0 != tx {
					num := uint8(bits.TrailingZeros32(tx))
					d.endpointComplete(txEndpoint(num))
					tx &^= 1 << num
				}
				rx := completeStatus & 0xFFFF
				for 0 != rx {
					num := uint8(bits.TrailingZeros32(rx))
					d.endpointComplete(rxEndpoint(num))
					rx &^= 1 << num
				}
			}
			// Notify device controller driver
			d.event(dcdEvent{
				id:   dcdEventTransactComplete,
				mask: completeStatus,
			})
		}
	}

	// USB Reset Received - R/WC
	// When the device controller detects a USB Reset and enters the default
	// state, this bit will be set to a one.
	// Software can write a 1 to this bit to clear the USB Reset Received status
	// bit.
	// Only used in device operation mode.
	if 0 != status&nxp.USB_USBSTS_URI {
		// clear all setup tokens
		d.bus.ENDPTSETUPSTAT.Set(d.bus.ENDPTSETUPSTAT.Get())
		// clear all endpoint complete status
		d.bus.ENDPTCOMPLETE.Set(d.bus.ENDPTCOMPLETE.Get())
		// wait on any endpoint priming
		for 0 != d.bus.ENDPTPRIME.Get() {
		}
		d.bus.ENDPTFLUSH.Set(0xFFFFFFFF)
		d.event(dcdEvent{id: dcdEventStatusReset})
		d.endpointMask = 0
	}

	// General Purpose Timer Interrupt 0(GPTINT0) - R/WC
	// This bit is set when the counter in the GPTIMER0CTRL register transitions
	// to zero, writing a one to this bit clears it.
	if 0 != status&nxp.USB_USBSTS_TI0 {
		if nil != d.timerInterrupt[0] {
			d.timerInterrupt[0]()
		}
		d.event(dcdEvent{id: dcdEventTimer, mask: 0})
	}

	// General Purpose Timer Interrupt 1(GPTINT1) - R/WC
	// This bit is set when the counter in the GPTIMER1CTRL register transitions
	// to zero, writing a one to this bit will clear it.
	if 0 != status&nxp.USB_USBSTS_TI1 {
		if nil != d.timerInterrupt[1] {
			d.timerInterrupt[1]()
		}
		d.event(dcdEvent{id: dcdEventTimer, mask: 1})
	}

	// Port Change Detect - R/WC
	// The Host Controller sets this bit to a one when on any port a Connect
	// Status occurs, a Port Enable/Disable Change occurs, or the Force Port
	// Resume bit is set as the result of a J-K transition on the suspended port.
	// The Device Controller sets this bit to a one when the port controller
	// enters the full or high-speed operational state. When the port controller
	// exits the full or high-speed operation states due to Reset or Suspend
	// events, the notification mechanisms are the USB Reset Received bit and the
	// DCSuspend bits respectively.
	if 0 != status&nxp.USB_USBSTS_PCI {
		if d.bus.PORTSC1.HasBits(nxp.USB_PORTSC1_HSP) {
			d.speed = HighSpeed // 480 Mbit/sec
		} else {
			d.speed = FullSpeed // 12 Mbit/sec
		}
		d.event(dcdEvent{id: dcdEventStatusRun})
	}

	// DCSuspend - R/WC
	// When a controller enters a suspend state from an active state, this bit
	// will be set to a one. The device controller clears the bit upon exiting
	// from a suspend state. Only used in device operation mode.
	if 0 != status&nxp.USB_USBSTS_SLI {
		d.event(dcdEvent{id: dcdEventStatusSuspend})
	}

	// USB Error Interrupt (USBERRINT) - R/WC
	// When completion of a USB transaction results in an error condition, this
	// bit is set by the Host/Device Controller. This bit is set along with the
	// USBINT bit, if the TD on which the error interrupt occurred also had its
	// interrupt on complete (IOC) bit set.
	// The device controller detects resume signaling only.
	if 0 != status&nxp.USB_USBSTS_UEI {
		d.event(dcdEvent{id: dcdEventStatusError})
	}

	// SOF Received - R/WC
	// When the device controller detects a Start Of (micro) Frame, this bit will
	// be set to a one. When a SOF is extremely late, the device controller will
	// automatically set this bit to indicate that an SOF was expected.
	// Therefore, this bit will be set roughly every 1ms in device FS mode and
	// every 125us in HS mode and will be synchronized to the actual SOF that is
	// received.
	// Because the device controller is initialized to FS before connect, this bit
	// will be set at an interval of 1ms during the prelude to connect and chirp.
	// In host mode, this bit will be set every 125us and can be used by host
	// controller driver as a time base. Software writes a 1 to this bit to clear
	// it.
	if d.bus.USBINTR.HasBits(nxp.USB_USBINTR_SRE) &&
		0 != status&nxp.USB_USBSTS_SRI {
		if 0 != d.timerReboot {
			d.timerReboot -= 1
			if 0 == d.timerReboot {
				d.enableSOF(false, descCDCACMInterfaceCount)
				runBootloader()
			}
		}
	}
}

func (d *dhw) setDeviceAddress(addr uint16) {
	d.bus.DEVICEADDR.Set(nxp.USB_DEVICEADDR_USBADRA |
		((uint32(addr) << nxp.USB_DEVICEADDR_USBADR_Pos) &
			nxp.USB_DEVICEADDR_USBADR_Msk))
	d.event(dcdEvent{id: dcdEventDeviceAddress})
}

// =============================================================================
//  Control Endpoint 0
// =============================================================================

// controlStall stalls a transfer on control endpoint 0. To stall a transfer on
// any other endpoint, use method endpointStall().
func (d *dhw) controlStall(stall bool) {
	d.endpointStall(0, stall)
}

// controlReceive receives (Rx, OUT) data on control endpoint 0.
func (d *dhw) controlReceive(
	data uintptr, size uint32, notify bool) {
	const (
		rm = uint32(1 << descEndptConfigAttrRxPos)
		tm = uint32(1 << descEndptConfigAttrTxPos)
	)
	cd, ad := d.transferControl()
	if size > 0 {
		cd.next = dhwTransferEOL
		cd.token = (size << 16) | (1 << 7)
		for i := range cd.pointer {
			cd.pointer[i] = data + uintptr(i)*4096
		}
		// linked list is empty
		qr := d.endpointQueueHead(rxEndpoint(0))
		qr.transfer.next = cd
		qr.transfer.token = 0
		d.bus.ENDPTPRIME.SetBits(rm)
		for 0 != d.bus.ENDPTPRIME.Get() {
		} // wait for endpoint finish priming
	}
	ad.next = dhwTransferEOL
	ad.token = 1 << 7 // Bit 7: active transfer
	if notify {
		ad.token |= 1 << 15
	}
	ad.pointer[0] = 0
	qt := d.endpointQueueHead(txEndpoint(0))
	qt.transfer.next = ad
	qt.transfer.token = 0
	d.bus.ENDPTCOMPLETE.Set(rm | tm)
	d.bus.ENDPTPRIME.SetBits(tm)
	if notify {
		d.controlMask = tm
	}
	for 0 != d.bus.ENDPTPRIME.Get() {
	} // wait for endpoint finish priming
}

// controlTransmit transmits (Tx, IN) data on control endpoint 0.
func (d *dhw) controlTransmit(
	data uintptr, size uint32, notify bool) {
	const (
		rm = uint32(1 << descEndptConfigAttrRxPos)
		tm = uint32(1 << descEndptConfigAttrTxPos)
	)
	cd, ad := d.transferControl()
	if size > 0 {
		cd.next = dhwTransferEOL
		cd.token = (size << 16) | (1 << 7)
		for i := range cd.pointer {
			cd.pointer[i] = data + uintptr(i)*4096
		}
		// linked list is empty
		qt := d.endpointQueueHead(txEndpoint(0))
		qt.transfer.next = cd
		qt.transfer.token = 0
		d.bus.ENDPTPRIME.SetBits(tm)
		for 0 != d.bus.ENDPTPRIME.Get() {
		} // wait for endpoint finish priming
	}
	ad.next = dhwTransferEOL
	ad.token = 1 << 7 // Bit 7: active transfer
	if notify {
		ad.token |= 1 << 15
	}
	ad.pointer[0] = 0
	qr := d.endpointQueueHead(rxEndpoint(0))
	qr.transfer.next = ad
	qr.transfer.token = 0
	d.bus.ENDPTCOMPLETE.Set(rm | tm)
	d.bus.ENDPTPRIME.SetBits(rm)
	if notify {
		d.controlMask = rm
	}
	for 0 != d.bus.ENDPTPRIME.Get() {
	} // wait for endpoint finish priming
}

// =============================================================================
//  Endpoint Descriptor
// =============================================================================

// dhwEndpoint defines a USB standard endpoint, used as the general channel of
// communication between host and device.
type dhwEndpoint struct {
	config   uint32
	current  *dhwTransfer
	transfer dhwTransfer
	setup    dcdSetup
	// Endpoints are 48-byte data structures. The remaining data extends this to
	// 64-byte, and also makes it simpler to align contiguous endpoints on 64-byte
	// boundaries.
	first *dhwTransfer
	last  *dhwTransfer
	// After some discussion, perhaps the simplest change to support 64-bit (or
	// future TinyGo versions that don't use 8 bytes to refer to a function) is to
	// allocate a separate buffer of callbacks. The device controller would assign
	// callbacks to unused elements in that buffer, and only the index of that
	// callback would be stored here in the endpoint.
	callback func(transfer *dhwTransfer)
}

// dhwEndpointSize defines the size (bytes) of a structure containing a USB
// standard endpoint.
const dhwEndpointSize = 64 // bytes

// endpointQueueHead returns the queue head for the given endpoint address,
// encoded as direction D and endpoint number N with the 8-bit mask DxxxNNNN.
//go:inline
func (d *dhw) endpointQueueHead(endpoint uint8) *dhwEndpoint {
	// endpoint queue head is device class-specific
	switch d.cc.id {
	case classDeviceCDCACM:
		return &descCDCACM[d.cc.config-1].qh[endpointIndex(endpoint)]
	case classDeviceHID:
		return &descHID[d.cc.config-1].qh[endpointIndex(endpoint)]
	default:
		return nil
	}
}

func (d *dhw) endpointControlRegister(endpoint uint8) *volatile.Register32 {
	num, _ := unpackEndpoint(endpoint)
	switch num {
	case 0:
		return &d.bus.ENDPTCTRL0
	case 1:
		return &d.bus.ENDPTCTRL1
	case 2:
		return &d.bus.ENDPTCTRL2
	case 3:
		return &d.bus.ENDPTCTRL3
	case 4:
		return &d.bus.ENDPTCTRL4
	case 5:
		return &d.bus.ENDPTCTRL5
	case 6:
		return &d.bus.ENDPTCTRL6
	case 7:
		return &d.bus.ENDPTCTRL7
	}
	return nil
}

func (d *dhw) endpointEnable(endpoint uint8, control bool, config uint32) {
	// control endpoint 0 configured in dhw init
	if !control || 0 != endpoint&descEndptAddrNumberMsk {
		d.endpointControlRegister(endpoint & descEndptAddrNumberMsk).Set(config)
	}
}

func (d *dhw) endpointStatus(endpoint uint8) uint16 {
	status := uint16(0)
	switch endpoint {
	case rxEndpoint(endpoint):
		status |= uint16((d.endpointControlRegister(endpoint).Get() &
			nxp.USB_ENDPTCTRL0_RXS_Msk) >> nxp.USB_ENDPTCTRL0_RXS_Pos)
	case txEndpoint(endpoint):
		status |= uint16((d.endpointControlRegister(endpoint).Get() &
			nxp.USB_ENDPTCTRL0_TXS_Msk) >> nxp.USB_ENDPTCTRL0_TXS_Pos)
	}
	return status
}

// endpointStall stalls a transfer on the given endpoint.
func (d *dhw) endpointStall(endpoint uint8, stall bool) {
	// RXS and TXS bits at same position in all endpoint control registers.
	d.endpointControlRegister(endpoint).SetBits(
		nxp.USB_ENDPTCTRL0_RXS | nxp.USB_ENDPTCTRL0_TXS,
	)
}

func (d *dhw) endpointClearFeature(endpoint uint8) {
	switch endpoint {
	case rxEndpoint(endpoint):
		d.endpointControlRegister(endpoint).ClearBits(nxp.USB_ENDPTCTRL0_RXS)
	case txEndpoint(endpoint):
		d.endpointControlRegister(endpoint).ClearBits(nxp.USB_ENDPTCTRL0_TXS)
	}
}

func (d *dhw) endpointSetFeature(endpoint uint8) {
	switch endpoint {
	case rxEndpoint(endpoint):
		d.endpointControlRegister(endpoint).SetBits(nxp.USB_ENDPTCTRL0_RXS)
	case txEndpoint(endpoint):
		d.endpointControlRegister(endpoint).SetBits(nxp.USB_ENDPTCTRL0_TXS)
	}
}

// endpointConfigure configures the given bulk data endpoint for transfer.
func (d *dhw) endpointConfigure(
	ep *dhwEndpoint, packetSize uint16, zlp bool, callback func(transfer *dhwTransfer)) {

	ep.config = uint32(packetSize) << 16
	if !zlp {
		ep.config |= 1 << 29
	}
	ep.current = nil
	ep.transfer.next = dhwTransferEOL
	ep.transfer.token = 0
	for i := range ep.transfer.pointer {
		ep.transfer.pointer[i] = 0
	}
	ep.transfer.param = 0
	ep.setup.bmRequestType = 0
	ep.setup.bRequest = 0
	ep.setup.wValue = 0
	ep.setup.wIndex = 0
	ep.setup.wLength = 0
	ep.first = nil
	ep.last = nil
	ep.callback = callback
}

// endpointComplete handles transfer completion of a data endpoint.
func (d *dhw) endpointComplete(endpoint uint8) {
	ep := d.endpointQueueHead(endpoint)
	if nil == ep.first {
		return
	}
	count := 0
	first := ep.first
	for t, eol := first, false; !eol; t, eol = t.nextTransfer() {
		if eol {
			// reached end of list, new list empty
			ep.first = nil
			ep.last = nil
		} else {
			if 0 != t.token&(1<<7) { // Bit 7: active transfer
				// new list begins here
				ep.first = t
				break
			} else {
				count += 1
			}
		}
	}
	// invoke all callbacks
	for i := 0; i < count; i++ {
		next := first.next
		ep.callback(first)
		first = next
	}
}

// endpointConfigureRx configures the given bulk data receive (Rx, OUT) endpoint
// for transfer.
func (d *dhw) endpointConfigureRx(
	endpoint uint8, packetSize uint16, zlp bool, callback func(transfer *dhwTransfer)) {

	// Configure based on our device class configuration
	switch d.cc.id {

	// CDC-ACM (single)
	case classDeviceCDCACM:
		if endpoint < descCDCACMEndpointStatus ||
			endpoint > descCDCACMEndpointCount {
			return
		}

	// HID
	case classDeviceHID:
		if endpoint < descHIDEndpointSerialRx ||
			endpoint > descHIDEndpointCount {
			return
		}

	default:
		// Unhandled device class
	}

	ep := d.endpointQueueHead(rxEndpoint(endpoint))
	d.endpointConfigure(ep, packetSize, zlp, callback)
	if nil != callback {
		d.endpointMask |= (uint32(1) << endpoint) << descEndptConfigAttrRxPos
	}
}

// endpointConfigureTx configures the given bulk data transmit (Tx, IN) endpoint
// for transfer.
func (d *dhw) endpointConfigureTx(
	endpoint uint8, packetSize uint16, zlp bool, callback func(transfer *dhwTransfer)) {

	// Configure based on our device class configuration
	switch d.cc.id {

	// CDC-ACM (single)
	case classDeviceCDCACM:
		if endpoint < descCDCACMEndpointStatus ||
			endpoint > descCDCACMEndpointCount {
			return
		}

	// HID
	case classDeviceHID:
		if endpoint < descHIDEndpointSerialRx ||
			endpoint > descHIDEndpointCount {
			return
		}

	default:
		// Unhandled device class
	}

	ep := d.endpointQueueHead(txEndpoint(endpoint))
	d.endpointConfigure(ep, packetSize, zlp, callback)
	if nil != callback {
		d.endpointMask |= (uint32(1) << endpoint) << descEndptConfigAttrTxPos
	}
}

// endpointReceive schedules a receive (Rx, OUT) transfer on the given endpoint.
func (d *dhw) endpointReceive(endpoint uint8, transfer *dhwTransfer) {

	// Configure based on our device class configuration
	switch d.cc.id {

	// CDC-ACM (single)
	case classDeviceCDCACM:
		if endpoint < descCDCACMEndpointStatus ||
			endpoint > descCDCACMEndpointCount {
			return
		}

	// HID
	case classDeviceHID:
		if endpoint < descHIDEndpointSerialRx ||
			endpoint > descHIDEndpointCount {
			return
		}

	default:
		// Unhandled device class
	}

	ep := d.endpointQueueHead(rxEndpoint(endpoint))
	em := (uint32(1) << endpoint) << descEndptConfigAttrRxPos
	d.transferSchedule(ep, em, transfer)
}

// endpointTransmit schedules a transmit (Tx, IN) transfer on the given
// endpoint.
func (d *dhw) endpointTransmit(endpoint uint8, transfer *dhwTransfer) {

	// Configure based on our device class configuration
	switch d.cc.id {

	// CDC-ACM (single)
	case classDeviceCDCACM:
		if endpoint < descCDCACMEndpointStatus ||
			endpoint > descCDCACMEndpointCount {
			return
		}

	// HID
	case classDeviceHID:
		if endpoint < descHIDEndpointSerialRx ||
			endpoint > descHIDEndpointCount {
			return
		}

	default:
		// Unhandled device class
	}

	ep := d.endpointQueueHead(txEndpoint(endpoint))
	em := (uint32(1) << endpoint) << descEndptConfigAttrTxPos
	d.transferSchedule(ep, em, transfer)
}

// =============================================================================
//  Transfer Descriptor
// =============================================================================

// dhwTransfer describes the size and location of data to be transferred to or
// from a USB endpoint.
type dhwTransfer struct {
	next    *dhwTransfer
	token   uint32
	pointer [5]uintptr
	param   uint32
}

// dhwTransferSize defines the size (bytes) of a USB standard transfer packet.
const dhwTransferSize = 32 // bytes

// dhwTransferEOL is a sentinel value used to indicate the final node in a
// linked list of transfer descriptors.
var dhwTransferEOL = (*dhwTransfer)(unsafe.Pointer(uintptr(1)))

// nextTransfer returns the next transfer descriptor pointed to by the receiver
// transfer descriptor, and whether or not that next descriptor is the final
// descriptor in the list.
func (t dhwTransfer) nextTransfer() (*dhwTransfer, bool) {
	return t.next, 1 == uintptr(unsafe.Pointer(t.next))
}

// transferControl returns the data and ackowledgement transfer descriptors for
// the control endpoint (i.e., endpoint 0).
//go:inline
func (d *dhw) transferControl() (dat, ack *dhwTransfer) {
	// control endpoint is device class-specific
	switch d.cc.id {
	case classDeviceCDCACM:
		return descCDCACM[d.cc.config-1].cd, descCDCACM[d.cc.config-1].ad
	case classDeviceHID:
		return descHID[d.cc.config-1].cd, descHID[d.cc.config-1].ad
	default:
		return nil, nil
	}
}

func (d *dhw) transferPrepare(
	transfer *dhwTransfer, data *uint8, size uint16, param uint32) {

	transfer.next = dhwTransferEOL

	// Set 15-bit packet size and transfer active bit 7.
	transfer.token = (uint32(size&0x7FFF) << 16) | (1 << 7)
	addr := uintptr(unsafe.Pointer(data))
	for i := range transfer.pointer {
		transfer.pointer[i] = addr + uintptr(i)*4096
	}
	transfer.param = param
}

func (d *dhw) transferSchedule(
	endpoint *dhwEndpoint, mask uint32, transfer *dhwTransfer) {

	if nil != endpoint.callback {
		transfer.token |= 1 << 15
	}
	ivm := arm.DisableInterrupts()
	last := endpoint.last
	if nil != last {
		last.next = transfer
		if d.bus.ENDPTPRIME.HasBits(mask) {
			goto endTransfer
		}
		start := cycleCount()
		estat := uint32(0)
		for !d.bus.USBCMD.HasBits(nxp.USB_USBCMD_ATDTW) &&
			(cycleCount()-start < 2400) {
			d.bus.USBCMD.SetBits(nxp.USB_USBCMD_ATDTW)
			estat = d.bus.ENDPTSTAT.Get()
		}
		if 0 != estat&mask {
			goto endTransfer
		}
	}
	endpoint.transfer.next = transfer
	endpoint.transfer.token = 0
	d.bus.ENDPTPRIME.SetBits(mask)
	endpoint.first = transfer
endTransfer:
	endpoint.last = transfer
	arm.EnableInterrupts(ivm)
}

// =============================================================================
//  General-Purpose (GP) Timer
// =============================================================================

func (d *dhw) timerConfigure(timer int, usec uint32, fn func()) {
	if timer < 0 || timer >= len(d.timerInterrupt) {
		return
	}
	d.timerInterrupt[timer] = fn
	switch timer {
	case 0:
		d.bus.GPTIMER0CTRL.Set(0)
		d.bus.GPTIMER0LD.Set(usec - 1)
		d.bus.USBINTR.SetBits(nxp.USB_USBINTR_TIE0)
	case 1:
		d.bus.GPTIMER1CTRL.Set(0)
		d.bus.GPTIMER1LD.Set(usec - 1)
		d.bus.USBINTR.SetBits(nxp.USB_USBINTR_TIE1)
	}
}

func (d *dhw) timerOneShot(timer int) {
	switch timer {
	case 0:
		d.bus.GPTIMER0CTRL.Set(
			nxp.USB_GPTIMER0CTRL_GPTRUN | nxp.USB_GPTIMER0CTRL_GPTRST)
	case 1:
		d.bus.GPTIMER1CTRL.Set(
			nxp.USB_GPTIMER1CTRL_GPTRUN | nxp.USB_GPTIMER1CTRL_GPTRST)
	}
}

func (d *dhw) timerStop(timer int) {
	switch timer {
	case 0:
		d.bus.GPTIMER0CTRL.Set(0)
	case 1:
		d.bus.GPTIMER1CTRL.Set(0)
	}
}

// =============================================================================
//  [CDC-ACM] Serial UART (Virtual COM Port)
// =============================================================================

func (d *dhw) uartConfigure() {

	acm := &descCDCACM[d.cc.config-1]

	switch d.speed {
	case HighSpeed, SuperSpeed, DualSuperSpeed:
		acm.rxSize = descCDCACMDataRxHSPacketSize
		acm.txSize = descCDCACMDataTxHSPacketSize
	default:
		fallthrough
	case LowSpeed, FullSpeed:
		acm.rxSize = descCDCACMDataRxFSPacketSize
		acm.txSize = descCDCACMDataTxFSPacketSize
	}
	acm.txHead = 0
	acm.txFree = 0
	acm.txPrev = false
	acm.rxHead = 0
	acm.rxTail = 0
	acm.rxFree = 0

	d.endpointEnable(descCDCACMEndpointStatus,
		false, descCDCACMConfigAttrStatus)
	d.endpointEnable(descCDCACMEndpointDataRx,
		false, descCDCACMConfigAttrDataRx)
	d.endpointEnable(descCDCACMEndpointDataTx,
		false, descCDCACMConfigAttrDataTx)

	d.endpointConfigureTx(descCDCACMEndpointStatus,
		acm.sxSize, false, nil)
	d.endpointConfigureRx(descCDCACMEndpointDataRx,
		acm.rxSize, false, d.uartNotify)
	d.endpointConfigureTx(descCDCACMEndpointDataTx,
		acm.txSize, true, nil)
	for i := range acm.rd {
		d.uartReceive(uint8(i))
	}
	d.timerConfigure(0, descCDCACMTxSyncUs, d.uartSync)
}

func (d *dhw) uartSetLineState(dtr, rts bool) {
	// TBD: does the PHY need to handle on iMXRT1062 (e.g., Teensyduino Loader)?
}

func (d *dhw) uartSetLineCoding(coding descCDCACMLineCoding) {
	if 134 == coding.baud {
		d.enableSOF(true, descCDCACMInterfaceCount)
	}
}

func (d *dhw) uartReceive(endpoint uint8) {
	acm := &descCDCACM[d.cc.config-1]
	num := uint16(endpoint) & descEndptAddrNumberMsk
	buf := &acm.rx[num*descCDCACMRxSize]
	d.enableInterrupts(false)
	d.transferPrepare(&acm.rd[num], buf, acm.rxSize, uint32(endpoint))
	deleteCache(uintptr(unsafe.Pointer(buf)), uintptr(acm.rxSize))
	d.endpointReceive(descCDCACMEndpointDataRx, &acm.rd[num])
	d.enableInterrupts(true)
}

func (d *dhw) uartNotify(transfer *dhwTransfer) {
	acm := &descCDCACM[d.cc.config-1]
	len := acm.rxSize - (uint16(transfer.token>>16) & 0x7FFF)
	p := transfer.param
	if 0 == len {
		// zero-length packet (ZLP)
		d.uartReceive(uint8(p))
	} else {
		// data packet
		h := acm.rxHead
		if h != acm.rxTail {
			// previous packet is still buffered
			q := acm.rxQueue[h]
			n := acm.rxCount[q]
			if len <= descCDCACMRxSize-n {
				// previous buffer has enough free space for this packet's data
				_ = copy(acm.rx[q*descCDCACMRxSize+n:],
					acm.rx[p*descCDCACMRxSize:uint16(p)*descCDCACMRxSize+len])
				acm.rxCount[q] = n + len
				acm.rxFree += len
				d.uartReceive(uint8(p))
				return
			}
		}
		// add this packet to Rx buffer
		acm.rxCount[p] = len
		acm.rxIndex[p] = 0
		h += 1
		if h > descCDCACMRDCount { // should be >=
			h = 0
		}
		acm.rxQueue[h] = uint16(p)
		acm.rxHead = h
		acm.rxFree += len
	}
}

// uartFlush discards all buffered input (Rx) data.
func (d *dhw) uartFlush() {
	acm := &descCDCACM[d.cc.config-1]
	tail := acm.rxTail
	for tail != acm.rxHead {
		tail += 1
		if tail > descCDCACMRDCount {
			tail = 0
		}
		i := acm.rxQueue[tail]
		acm.rxFree -= acm.rxCount[i] - acm.rxIndex[i]
		d.uartReceive(uint8(i))
		acm.rxTail = tail
	}
}

func (d *dhw) uartAvailable() int {
	return int(descCDCACM[d.cc.config-1].rxFree)
}

func (d *dhw) uartPeek() (uint8, bool) {
	acm := &descCDCACM[d.cc.config-1]
	tail := acm.rxTail
	if tail == acm.rxHead {
		return 0, false
	}
	tail += 1
	if tail > descCDCACMRDCount {
		tail = 0
	}
	i := acm.rxQueue[tail]
	return acm.rx[i*descCDCACMRxSize+acm.rxIndex[i]], true
}

func (d *dhw) uartReadByte() (uint8, bool) {
	b := []uint8{0}
	ok := d.uartRead(b) > 0
	return b[0], ok
}

func (d *dhw) uartRead(data []uint8) int {
	acm := &descCDCACM[d.cc.config-1]
	read := uint16(0)
	size := uint16(len(data))
	tail := acm.rxTail
	dest := uint16(0)
	d.enableInterrupts(false)
	for read < size && tail != acm.rxHead {
		tail += 1
		if tail > descCDCACMRDCount {
			tail = 0
		}
		i := acm.rxQueue[tail]
		count := uint16(size - read)
		avail := acm.rxCount[i] - acm.rxIndex[i]
		start := i*descCDCACMRxSize + acm.rxIndex[i]
		if avail > count {
			// partially consume packet
			_ = copy(data[dest:], acm.rx[start:start+count])
			acm.rxFree -= count
			acm.rxIndex[i] += count
			read += count
		} else {
			// fully consume packet
			_ = copy(data[dest:], acm.rx[start:start+avail])
			dest += avail //* uint16(unsafe.Sizeof(&data[0]))
			read += avail
			acm.rxFree -= avail
			acm.rxTail = tail
			d.uartReceive(uint8(i))
		}
	}
	d.enableInterrupts(true)
	return int(read)
}

func (d *dhw) uartWriteByte(c uint8) bool {
	return 1 == d.uartWrite([]uint8{c})
}

func (d *dhw) uartWrite(data []uint8) int {
	acm := &descCDCACM[d.cc.config-1]
	sent := 0
	size := len(data)
	for size > 0 {
		xfer := &acm.td[acm.txHead]
		wait := false
		when := int64(0)
		for 0 == acm.txFree {
			if 0 == xfer.token&0x80 {
				if 0 != xfer.token&0x68 {
					// TODO: token contains error, how to handle?
				}
				acm.txFree = descCDCACMTxSize
				acm.txPrev = false
				break
			}
			if !wait {
				wait = true
				when = ticks()
			}
			if acm.txPrev {
				return sent
			}
			if ticks()-when > descCDCACMTxTimeoutMs {
				acm.txPrev = true
				return sent
			}
		}
		buff := acm.tx[(int(acm.txHead)*descCDCACMTxSize)+
			(descCDCACMTxSize-int(acm.txFree)):]
		if size > int(acm.txFree) {
			_ = copy(buff, data[sent:sent+int(acm.txFree)])
			tx := &acm.tx[int(acm.txHead)*descCDCACMTxSize]
			d.transferPrepare(xfer, tx, descCDCACMTxSize, 0)
			flushCache(uintptr(unsafe.Pointer(tx)), descCDCACMTxSize)
			d.endpointTransmit(descCDCACMEndpointDataTx, xfer)
			acm.txHead += 1
			if acm.txHead >= descCDCACMTDCount {
				acm.txHead = 0
			}
			size -= int(acm.txFree)
			sent += int(acm.txFree)
			acm.txFree = 0
			d.timerStop(0)
		} else {
			_ = copy(buff, data[:size])
			acm.txFree -= uint16(size)
			sent += size
			size = 0
			d.timerOneShot(0)
		}
	}
	return sent
}

func (d *dhw) uartSync() {
	const autoFlushTx = true
	if !autoFlushTx {
		return
	}
	acm := &descCDCACM[d.cc.config-1]
	if 0 == acm.txFree {
		return
	}
	xfer := &acm.td[acm.txHead]
	buff := &acm.tx[uint16(acm.txHead)*descCDCACMTxSize]
	size := descCDCACMTxSize - acm.txFree
	d.transferPrepare(xfer, buff, size, 0)
	flushCache(uintptr(unsafe.Pointer(buff)), uintptr(size))
	d.endpointTransmit(descCDCACMEndpointDataTx, xfer)
	acm.txHead += 1
	if acm.txHead >= descCDCACMTDCount {
		acm.txHead = 0
	}
	acm.txFree = 0
}

// =============================================================================
//  [HID] Serial
// =============================================================================

func (d *dhw) serialConfigure() {

	hid := &descHID[d.cc.config-1]

	switch d.speed {
	case HighSpeed, SuperSpeed, DualSuperSpeed:
		hid.rxSerialSize = descHIDSerialRxHSPacketSize
		hid.txSerialSize = descHIDSerialTxHSPacketSize
	default:
		fallthrough
	case LowSpeed, FullSpeed:
		hid.rxSerialSize = descHIDSerialRxFSPacketSize
		hid.txSerialSize = descHIDSerialTxFSPacketSize
	}

	// Rx and Tx are on same endpoint
	d.endpointEnable(descHIDEndpointSerialRx,
		false, descHIDConfigAttrSerial)

	d.endpointConfigureRx(descHIDEndpointSerialRx,
		hid.rxSerialSize, false, d.serialNotify)
	d.endpointConfigureTx(descHIDEndpointSerialTx,
		hid.txSerialSize, false, nil)
	for i := range hid.rdSerial {
		d.serialReceive(uint8(i))
	}
	d.timerConfigure(0, descHIDSerialTxSyncUs, d.serialSync)
}

func (d *dhw) serialReceive(endpoint uint8) {
	hid := &descHID[d.cc.config-1]
	num := uint16(endpoint) & descEndptAddrNumberMsk
	buf := &hid.rxSerial[num*descHIDSerialRxSize]
	d.enableInterrupts(false)
	d.transferPrepare(&hid.rdSerial[num], buf, hid.rxSerialSize, uint32(endpoint))
	deleteCache(uintptr(unsafe.Pointer(buf)), uintptr(hid.rxSerialSize))
	d.endpointReceive(descHIDEndpointSerialRx, &hid.rdSerial[num])
	d.enableInterrupts(true)
}

func (d *dhw) serialTransmit() {
	hid := &descHID[d.cc.config-1]
	xfer := &hid.tdSerial[hid.txSerialHead]
	buff := &hid.txSerial[uint16(hid.txSerialHead)*descHIDSerialTxSize]
	d.transferPrepare(xfer, buff, hid.txSerialSize, 0)
	flushCache(uintptr(unsafe.Pointer(buff)), uintptr(hid.txSerialSize))
	d.endpointTransmit(descHIDEndpointSerialTx, xfer)
	hid.txSerialHead += 1
	if hid.txSerialHead >= descHIDSerialTDCount {
		hid.txSerialHead = 0
	}
}

func (d *dhw) serialNotify(transfer *dhwTransfer) {
	hid := &descHID[d.cc.config-1]
	len := hid.rxSerialSize - (uint16(transfer.token>>16) & 0x7FFF)
	p := transfer.param
	if len == hid.rxSerialSize && 0 != hid.rxSerial[p*uint32(hid.rxSerialSize)] {
		// data packet
		hid.rxSerialIndex[p] = 0
		h := hid.rxSerialHead + 1
		if h > descHIDSerialRDCount { // should be >=
			h = 0
		}
		hid.rxSerialQueue[h] = uint16(p)
		hid.rxSerialHead = h
		hid.rxSerialFree += len
	} else {
		// short packet
		d.serialReceive(uint8(p))
	}
}

// serialFlush discards all buffered input (Rx) data.
func (d *dhw) serialFlush() {
	hid := &descHID[d.cc.config-1]
	tail := hid.rxSerialTail
	for tail != hid.rxSerialHead {
		tail += 1
		if tail > descHIDSerialRDCount {
			tail = 0
		}
		i := hid.rxSerialQueue[tail]
		d.serialReceive(uint8(i))
		hid.rxSerialTail = tail
	}
}

func (d *dhw) serialSync() {
	const autoFlushTx = true
	if !autoFlushTx {
		return
	}
	hid := &descHID[d.cc.config-1]
	if 0 == hid.txSerialFree {
		return
	}
	xfer := &hid.tdSerial[hid.txSerialHead]
	buff := &hid.txSerial[uint16(hid.txSerialHead)*descHIDSerialTxSize]
	size := descHIDSerialTxSize - hid.txSerialFree
	d.transferPrepare(xfer, buff, size, 0)
	flushCache(uintptr(unsafe.Pointer(buff)), uintptr(size))
	d.endpointTransmit(descHIDEndpointSerialTx, xfer)
	hid.txSerialHead += 1
	if hid.txSerialHead >= descHIDSerialTDCount {
		hid.txSerialHead = 0
	}
	hid.txSerialFree = 0
}

// =============================================================================
//  [HID] Keyboard
// =============================================================================

func (d *dhw) keyboard() *Keyboard { return descHID[d.cc.config-1].keyboard }

func (d *dhw) keyboardConfigure() {

	hid := &descHID[d.cc.config-1]

	// Initialize keyboard
	hid.keyboard.configure(d.dcd, hid)

	switch d.speed {
	case HighSpeed, SuperSpeed, DualSuperSpeed:
		hid.txKeyboardSize = descHIDKeyboardTxHSPacketSize
	default:
		fallthrough
	case LowSpeed, FullSpeed:
		hid.txKeyboardSize = descHIDKeyboardTxFSPacketSize
	}

	d.endpointEnable(descHIDEndpointKeyboard,
		false, descHIDConfigAttrKeyboard)
	d.endpointEnable(descHIDEndpointMediaKey,
		false, descHIDConfigAttrMediaKey)

	d.endpointConfigureTx(descHIDEndpointKeyboard,
		hid.txKeyboardSize, false, nil)
	d.endpointConfigureTx(descHIDEndpointMediaKey,
		hid.txKeyboardSize, false, nil)
}

func (d *dhw) keyboardSendKeys(consumer bool) bool {

	hid := &descHID[d.cc.config-1]

	if !consumer {

		hid.tpKeyboard[0] = hid.keyboard.mod
		hid.tpKeyboard[1] = 0
		hid.tpKeyboard[2] = hid.keyboard.key[0]
		hid.tpKeyboard[3] = hid.keyboard.key[1]
		hid.tpKeyboard[4] = hid.keyboard.key[2]
		hid.tpKeyboard[5] = hid.keyboard.key[3]
		hid.tpKeyboard[6] = hid.keyboard.key[4]
		hid.tpKeyboard[7] = hid.keyboard.key[5]

		return d.keyboardWrite(descHIDEndpointKeyboard, hid.tpKeyboard[:])

	} else {

		// 44444444 44333333 33332222 22222211 11111111  [ word ]
		// 98765432 10987654 32109876 54321098 76543210  [ index ]  (right-to-left)

		hid.tpKeyboard[1] = uint8((hid.keyboard.con[1] << 2) | ((hid.keyboard.con[0] >> 8) & 0x03))
		hid.tpKeyboard[2] = uint8((hid.keyboard.con[2] << 4) | ((hid.keyboard.con[1] >> 6) & 0x0F))
		hid.tpKeyboard[3] = uint8((hid.keyboard.con[3] << 6) | ((hid.keyboard.con[2] >> 4) & 0x3F))
		hid.tpKeyboard[4] = uint8(hid.keyboard.con[3] >> 2)
		hid.tpKeyboard[5] = hid.keyboard.sys[0]
		hid.tpKeyboard[6] = hid.keyboard.sys[1]
		hid.tpKeyboard[7] = hid.keyboard.sys[2]

		return d.keyboardWrite(descHIDEndpointMediaKey, hid.tpKeyboard[:])

	}
}

func (d *dhw) keyboardWrite(endpoint uint8, data []uint8) bool {

	hid := &descHID[d.cc.config-1]

	size := uint16(len(data))
	xfer := &hid.tdKeyboard[hid.txKeyboardHead]
	when := ticks()
	for {
		if 0 == xfer.token&0x80 {
			if 0 != xfer.token&0x68 {
				// TODO: token contains error, how to handle?
			}
			hid.txKeyboardPrev = false
			break
		}
		if hid.txKeyboardPrev {
			return false
		}
		if ticks()-when > descHIDKeyboardTxTimeoutMs {
			// Waited too long, assume host connection dropped
			hid.txKeyboardPrev = true
			return false
		}
	}
	// Without this delay, the order packets are transmitted is seriously screwy.
	udelay(60)
	buff := hid.txKeyboard[hid.txKeyboardHead*descHIDKeyboardTxSize:]
	_ = copy(buff, data)
	d.transferPrepare(xfer, &buff[0], size, 0)
	flushCache(uintptr(unsafe.Pointer(&buff[0])), descHIDKeyboardTxSize)
	d.endpointTransmit(endpoint, xfer)
	hid.txKeyboardHead += 1
	if hid.txKeyboardHead >= descHIDKeyboardTDCount {
		hid.txKeyboardHead = 0
	}
	return true
}

// =============================================================================
//  [HID] Mouse
// =============================================================================

func (d *dhw) mouseConfigure() {

	hid := &descHID[d.cc.config-1]

	switch d.speed {
	case HighSpeed, SuperSpeed, DualSuperSpeed:
		hid.txMouseSize = descHIDMouseTxHSPacketSize
	default:
		fallthrough
	case LowSpeed, FullSpeed:
		hid.txMouseSize = descHIDMouseTxFSPacketSize
	}

	d.endpointEnable(descHIDEndpointMouse,
		false, descHIDConfigAttrMouse)

	d.endpointConfigureTx(descHIDEndpointMouse,
		hid.txMouseSize, false, nil)
}

// =============================================================================
//  [HID] Joystick
// =============================================================================

func (d *dhw) joystickConfigure() {

	hid := &descHID[d.cc.config-1]

	switch d.speed {
	case HighSpeed, SuperSpeed, DualSuperSpeed:
		hid.txJoystickSize = descHIDJoystickTxHSPacketSize
	default:
		fallthrough
	case LowSpeed, FullSpeed:
		hid.txJoystickSize = descHIDJoystickTxFSPacketSize
	}

	d.endpointEnable(descHIDEndpointJoystick,
		false, descHIDConfigAttrJoystick)

	d.endpointConfigureTx(descHIDEndpointJoystick,
		hid.txJoystickSize, false, nil)
}
