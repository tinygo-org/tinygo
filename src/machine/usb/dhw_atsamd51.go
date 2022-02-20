//go:build (sam && atsamd51) || (sam && atsame5x)
// +build sam,atsamd51 sam,atsame5x

package usb

// Implementation of USB device controller hardware abstraction (dhw) for
// Microchip SAMx51.

import (
	"device/arm"
	"device/sam"
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

	bus *sam.USB_DEVICE_Type // USB core registers

	irqEVT interrupt.Interrupt // USB IRQs
	irqSOF interrupt.Interrupt
	irqTC0 interrupt.Interrupt
	irqTC1 interrupt.Interrupt

	speed Speed

	ready bool // has init() been called

	ep [descMaxEndpoints]dhwEPAddrStatus

	setup   dcdSetup
	stage   dcdStage
	address uint16
}

func deleteCache(addr, size uintptr) {}
func flushCache(addr, size uintptr)  {}
func runBootloader()                 {}

// allocDHW returns a reference to the USB hardware abstraction for the given
// device controller driver. Should be called only one time and during device
// controller initialization.
func allocDHW(port, instance int, speed Speed, dc *dcd) *dhw {
	switch port {
	case 0:
		dhwInstance[instance].dcd = dc
		dhwInstance[instance].bus = sam.USB_DEVICE
		dhwInstance[instance].irqEVT =
			interrupt.New(sam.IRQ_USB_OTHER,
				func(interrupt.Interrupt) { coreInstance[0].dc.interrupt() })
		dhwInstance[instance].irqSOF =
			interrupt.New(sam.IRQ_USB_SOF_HSOF,
				func(interrupt.Interrupt) { coreInstance[0].dc.interrupt() })
		dhwInstance[instance].irqTC0 =
			interrupt.New(sam.IRQ_USB_TRCPT0,
				func(interrupt.Interrupt) { coreInstance[0].dc.interrupt() })
		dhwInstance[instance].irqTC1 =
			interrupt.New(sam.IRQ_USB_TRCPT1,
				func(interrupt.Interrupt) { coreInstance[0].dc.interrupt() })
	}

	// SAMx51 has only one USB PHY, which is full-speed
	if speed == 0 {
		speed = FullSpeed
	}
	dhwInstance[instance].speed = speed
	dhwInstance[instance].ready = false

	return &dhwInstance[instance]
}

// Calibrate DP/DM pads using value from NVM. Based on the following from
// Atmel's CMSIS 1.2.2 for SAMD51:
//
//	 ... NOTE: These register defines are used to obtain calibration parameters
//  |
//  | #define NVMCTRL_SW0                   (0x00800080UL) /**< \brief (NVMCTRL) SW0 Base Address *
//  |
//   ...
//  |
//  | #define USB_FUSES_TRANSN_ADDR       (NVMCTRL_SW0 + 4)
//  | #define USB_FUSES_TRANSN_Pos        0            /**< \brief (NVMCTRL_SW0) USB pad Transn calibration */
//  | #define USB_FUSES_TRANSN_Msk        (_Ul(0x1F) << USB_FUSES_TRANSN_Pos)
//  | #define USB_FUSES_TRANSN(value)     (USB_FUSES_TRANSN_Msk & ((value) << USB_FUSES_TRANSN_Pos))
//  |
//  | #define USB_FUSES_TRANSP_ADDR       (NVMCTRL_SW0 + 4)
//  | #define USB_FUSES_TRANSP_Pos        5            /**< \brief (NVMCTRL_SW0) USB pad Transp calibration */
//  | #define USB_FUSES_TRANSP_Msk        (_Ul(0x1F) << USB_FUSES_TRANSP_Pos)
//  | #define USB_FUSES_TRANSP(value)     (USB_FUSES_TRANSP_Msk & ((value) << USB_FUSES_TRANSP_Pos))
//  |
//  | #define USB_FUSES_TRIM_ADDR         (NVMCTRL_SW0 + 4)
//  | #define USB_FUSES_TRIM_Pos          10           /**< \brief (NVMCTRL_SW0) USB pad Trim calibration */
//  | #define USB_FUSES_TRIM_Msk          (_Ul(0x7) << USB_FUSES_TRIM_Pos)
//  | #define USB_FUSES_TRIM(value)       (USB_FUSES_TRIM_Msk & ((value) << USB_FUSES_TRIM_Pos))
//  |
//   ...
//  |
//  | typedef union {
//  |   struct {
//  |     uint16_t TRANSP:5;         /*!< bit:  0.. 4  USB Pad Transp calibration         */
//  |     uint16_t :1;               /*!< bit:      5  Reserved                           */
//  |     uint16_t TRANSN:5;         /*!< bit:  6..10  USB Pad Transn calibration         */
//  |     uint16_t :1;               /*!< bit:     11  Reserved                           */
//  |     uint16_t TRIM:3;           /*!< bit: 12..14  USB Pad Trim calibration           */
//  |     uint16_t :1;               /*!< bit:     15  Reserved                           */
//  |   } bit;                       /*!< Structure used for bit  access                  */
//  |   uint16_t reg;                /*!< Type      used for register access              */
//  | } USB_PADCAL_Type;
//  |
//   ... NOTE: The following is where USB pad calibration actually occurrs:
//  |
//  | USB->DEVICE.PADCAL.bit.TRANSP = (*((uint32_t*) USB_FUSES_TRANSP_ADDR) & USB_FUSES_TRANSP_Msk) >> USB_FUSES_TRANSP_Pos;
//  | USB->DEVICE.PADCAL.bit.TRANSN = (*((uint32_t*) USB_FUSES_TRANSN_ADDR) & USB_FUSES_TRANSN_Msk) >> USB_FUSES_TRANSN_Pos;
//  | USB->DEVICE.PADCAL.bit.TRIM   = (*((uint32_t*) USB_FUSES_TRIM_ADDR) & USB_FUSES_TRIM_Msk) >> USB_FUSES_TRIM_Pos;
//  |
//   ...
//
func (d *dhw) calibrate() {
	const reg = 0x00800080 + 4 // NVMCTRL_SW0 + 4
	cal := *(*uint16)(unsafe.Pointer(uintptr(reg)))
	msk := uint16(sam.USB_DEVICE_PADCAL_TRANSP_Msk |
		sam.USB_DEVICE_PADCAL_TRANSN_Msk |
		sam.USB_DEVICE_PADCAL_TRIM_Msk)
	d.bus.PADCAL.ReplaceBits(cal, msk, 0)
}

// init configures the USB port for device mode operation by initializing all
// endpoint and transfer descriptor data structures, initializing core registers
// and interrupts, resetting the USB PHY, and enabling power on the bus.
func (d *dhw) init() status {

	// Enable USB clocks
	// const clockGenerator = sam.GCLK_PCHCTRL_GEN_GCLK10
	const clockGenerator = sam.PCHCTRL_GCLK_USB
	sam.MCLK.APBBMASK.SetBits(sam.MCLK_APBBMASK_USB_)
	sam.MCLK.AHBMASK.SetBits(sam.MCLK_AHBMASK_USB_)
	sam.GCLK.PCHCTRL[clockGenerator].Set(
		(sam.GCLK_PCHCTRL_GEN_GCLK1 << sam.GCLK_PCHCTRL_GEN_Pos) |
			sam.GCLK_PCHCTRL_CHEN)

	// Reset USB peripheral
	for d.bus.SYNCBUSY.HasBits(sam.USB_DEVICE_SYNCBUSY_SWRST) {
	}
	d.bus.CTRLA.Set(sam.USB_DEVICE_CTRLA_SWRST)
	for d.bus.SYNCBUSY.HasBits(sam.USB_DEVICE_SYNCBUSY_SWRST) {
	}
	d.calibrate()

	// Initialize USB interrupt priorities
	d.irqEVT.SetPriority(dhwInterruptPriority)
	d.irqSOF.SetPriority(dhwInterruptPriority)
	d.irqTC0.SetPriority(dhwInterruptPriority)
	d.irqTC1.SetPriority(dhwInterruptPriority)

	// Clear interrupts
	m := arm.DisableInterrupts() &
		^uintptr(sam.IRQ_USB_OTHER|sam.IRQ_USB_SOF_HSOF|
			sam.IRQ_USB_TRCPT0|sam.IRQ_USB_TRCPT1)
	arm.EnableInterrupts(m)

	// USB Quality of Service: High Quality (3)
	d.bus.QOSCTRL.Set((3 << sam.USB_DEVICE_QOSCTRL_CQOS_Pos) |
		(3 << sam.USB_DEVICE_QOSCTRL_DQOS_Pos))

	// Install USB endpoint descriptor table (USB_DEVICE.DESCADD)
	var addr uintptr
	switch d.cc.id {
	case classDeviceCDCACM:
		addr = uintptr(unsafe.Pointer(&descCDCACM[d.cc.config-1].ed[0]))
	case classDeviceHID:
		addr = uintptr(unsafe.Pointer(&descHID[d.cc.config-1].ed[0]))
	}
	d.bus.DESCADD.Set(uint32(addr))

	// Configure bus speed (always full-speed (FS)), device mode, enable PHY, and
	// put finite-state machine (FSM) in standby.
	d.bus.CTRLB.Set(sam.USB_DEVICE_CTRLB_SPDCONF_FS)
	d.bus.CTRLA.Set(sam.USB_DEVICE_CTRLA_MODE_DEVICE |
		sam.USB_DEVICE_CTRLA_ENABLE | sam.USB_DEVICE_CTRLA_RUNSTDBY)
	for d.bus.SYNCBUSY.HasBits(sam.USB_DEVICE_SYNCBUSY_ENABLE) {
	}

	// Clear and enable interrupts in USB core
	d.bus.INTFLAG.Set(d.bus.INTFLAG.Get())
	d.bus.INTENSET.Set( /*sam.USB_DEVICE_INTENSET_SOF |*/
		sam.USB_DEVICE_INTENSET_EORST)

	// Ensure D+ pulled down long enough for host to detect a previous disconnect
	udelay(5000)

	d.ready = true

	return statusOK
}

// enable enables the USB interrupts, connects the device to the bus via
// internal D+/D- pullup resistors, and enters the normal runtime.
func (d *dhw) enable(enable bool) {
	if d.ready { // ensure init() has been called
		d.enableInterrupts(enable)
		d.connect(enable)
	}
}

// connect attaches the USB device by enabling/disabling the internal pullup
// resistor on D+/D-.
func (d *dhw) connect(connect bool) {
	if d.ready { // ensure init() has been called
		if connect {
			d.bus.CTRLB.ClearBits(sam.USB_DEVICE_CTRLB_DETACH)
		} else {
			d.bus.CTRLB.SetBits(sam.USB_DEVICE_CTRLB_DETACH)
		}
	}
}

// enableInterrupts enables/disables all interrupts on the receiver's USB port.
func (d *dhw) enableInterrupts(enable bool) {
	if d.ready { // ensure init() has been called
		if enable {
			d.irqEVT.Enable() // Enable USB interrupts
			d.irqSOF.Enable()
			d.irqTC0.Enable()
			d.irqTC1.Enable()
		} else {
			d.irqEVT.Disable() // Disable USB interrupts
			d.irqSOF.Disable()
			d.irqTC0.Disable()
			d.irqTC1.Disable()
		}
	}
}

// enableSOF enables or disables start-of-frame (SOF) interrupts on the given
// USB device interface.
func (d *dhw) enableSOF(enable bool, iface uint8) {
	// if changing enabled state, clear interrupt
	if enable != d.bus.INTENSET.HasBits(sam.USB_DEVICE_INTENSET_SOF) {
		d.bus.INTFLAG.Set(sam.USB_DEVICE_INTFLAG_SOF)
	}
	if enable {
		d.bus.INTENSET.Set(sam.USB_DEVICE_INTENSET_SOF)
	} else {
		d.bus.INTENCLR.Set(sam.USB_DEVICE_INTENCLR_SOF)
	}
}

// interrupt handles the USB hardware interrupt events on all four IRQ lines and
// notifies the device controller driver using a common "virtual interrupt"
// code.
func (d *dhw) interrupt() {

	status := d.bus.INTFLAG.Get() & d.bus.INTENSET.Get()

	if status&sam.USB_DEVICE_INTFLAG_SOF != 0 {
		d.bus.INTFLAG.Set(sam.USB_DEVICE_INTFLAG_SOF)

		// TBD: handle SOF?
	}

	// SAMD doesn't distinguish between SUSPEND and DISCONNECT states.
	// Both conditions will trigger the SUSPEND interrupt.
	// To prevent it triggering when D+/D- are not stable, the SUSPEND interrupt is
	// only enabled after receiving SET_ADDRESS request and is cleared on RESET.
	if status&sam.USB_DEVICE_INTFLAG_SUSPEND != 0 {
		d.bus.INTFLAG.Set(sam.USB_DEVICE_INTFLAG_SUSPEND)

		d.bus.INTFLAG.Set(sam.USB_DEVICE_INTFLAG_WAKEUP)
		d.bus.INTENSET.Set(sam.USB_DEVICE_INTENSET_WAKEUP)

		d.event(dcdEvent{id: dcdEventStatusSuspend})
	}

	if status&sam.USB_DEVICE_INTFLAG_WAKEUP != 0 {
		d.bus.INTFLAG.Set(sam.USB_DEVICE_INTFLAG_WAKEUP)
		d.bus.INTENCLR.Set(sam.USB_DEVICE_INTENCLR_WAKEUP)

		d.event(dcdEvent{id: dcdEventStatusResume})
	}

	if status&sam.USB_DEVICE_INTFLAG_EORST != 0 {
		d.bus.INTFLAG.Set(sam.USB_DEVICE_INTFLAG_EORST)
		d.bus.INTENCLR.Set(sam.USB_DEVICE_INTENCLR_WAKEUP |
			sam.USB_DEVICE_INTENCLR_SUSPEND)

		d.event(dcdEvent{id: dcdEventDeviceReady})
	}

	num := endpointNumber(d.controlEndpoint())
	if d.bus.DEVICE_ENDPOINT[num].EPINTFLAG.HasBits(
		sam.USB_DEVICE_ENDPOINT_EPINTFLAG_RXSTP) {
		d.bus.DEVICE_ENDPOINT[num].EPINTFLAG.Set(
			sam.USB_DEVICE_ENDPOINT_EPINTFLAG_RXSTP |
				sam.USB_DEVICE_ENDPOINT_EPINTFLAG_TRCPT0)

		// Parse the SETUP packet immediately, clearing room in the (one and only)
		// control buffer for the next SETUP packet received.
		sup := setupFrom(d.controlSetupBuffer())
		dir := sup.direction()

		// We've copied the SETUP packet elsewhere and are ready to receive another.
		d.prepareSetup()

		// Although there is only one control buffer, EP0 has two transfer queues:
		// 1×Rx(OUT) and 1×Tx(IN). First we decode the SETUP packet via setupFrom,
		// and based on its contained request's direction (IN vs OUT), we attempt
		// to enqueue a new transfer request in EP0's corresponding transfer queue.
		if ready, _ := d.ep[num][dir].scheduleSetup(sup); ready {
			// Begin processing the control packet immediately since there were no
			// pending transfers in the control EP0's IN/OUT transfer queue.
			d.controlTransferStart(packEndpoint(num, dir))
		} else {
			// The EP0 IN/OUT transfer queue is busy servicing a previous request.
			// Stall the endpoint.
			d.controlStall(true, dir)
		}
	}

	epints := d.bus.EPINTSMRY.Get() & ((1 << descMaxEndpoints) - 1)

	for epints != 0 {

		ep := uint8(bits.TrailingZeros16(epints))
		epints &^= 1 << ep

		intFlag := d.bus.DEVICE_ENDPOINT[ep].EPINTFLAG.Get()

		out, in := d.endpointDescriptors(ep)

		// handle Tx (IN) endpoint complete
		if intFlag&sam.USB_DEVICE_ENDPOINT_EPINTFLAG_TRCPT1 != 0 {
			pcksize := in.packetSize.Get()
			// number of bytes to be sent on next IN transaction
			count := (pcksize >> USB_DEVICE_PCKSIZE_BYTE_COUNT_Pos) &
				USB_DEVICE_PCKSIZE_BYTE_COUNT_Msk
			// total number of bytes sent
			total := (pcksize >> USB_DEVICE_PCKSIZE_MULTI_PACKET_SIZE_Pos) &
				USB_DEVICE_PCKSIZE_MULTI_PACKET_SIZE_Msk
			// maximum packet size
			size, _ := endpointSizeDecode((pcksize >> USB_DEVICE_PCKSIZE_SIZE_Pos) &
				USB_DEVICE_PCKSIZE_SIZE_Msk)

			d.bus.DEVICE_ENDPOINT[ep].EPINTFLAG.Set(
				sam.USB_DEVICE_ENDPOINT_EPINTFLAG_TRCPT1)

			if ep == d.controlEndpoint() {
				d.controlStall(false, descDirTx)
				// check if there is more data to transfer or if we need to notify the
				// upper-layer device driver of a control transfer completion event.
				if count == 0 || count < size {
					d.controlTransferComplete(txEndpoint(ep), count, total)
				} else {
					d.controlTransferContinue(txEndpoint(ep), count, total)
				}
			} else if nil != d.ep[ep][descDirTx].callback {
				// call our device class-specific callback, if defined, on endpoint
				// data transfer complete events.
				d.ep[ep][descDirTx].callback(txEndpoint(ep), count)
			}
		}

		// handle Rx (OUT) endpoint complete
		if intFlag&sam.USB_DEVICE_ENDPOINT_EPINTFLAG_TRCPT0 != 0 {
			pcksize := out.packetSize.Get()
			// number of bytes received on last OUT/SETUP transaction
			count := (pcksize >> USB_DEVICE_PCKSIZE_BYTE_COUNT_Pos) &
				USB_DEVICE_PCKSIZE_BYTE_COUNT_Msk
			// total data size for the complete transfer
			total := (pcksize >> USB_DEVICE_PCKSIZE_MULTI_PACKET_SIZE_Pos) &
				USB_DEVICE_PCKSIZE_MULTI_PACKET_SIZE_Msk
			// maximum packet size
			size, _ := endpointSizeDecode((pcksize >> USB_DEVICE_PCKSIZE_SIZE_Pos) &
				USB_DEVICE_PCKSIZE_SIZE_Msk)

			d.bus.DEVICE_ENDPOINT[ep].EPINTFLAG.Set(
				sam.USB_DEVICE_ENDPOINT_EPINTFLAG_TRCPT0)

			if ep == d.controlEndpoint() {
				d.controlStall(false, descDirRx)
				// check if there is more data to transfer or if we need to notify the
				// upper-layer device driver of a control transfer completion event.
				if count == 0 || count < size {
					d.controlTransferComplete(rxEndpoint(ep), count, total)
				} else {
					d.controlTransferContinue(rxEndpoint(ep), count, total)
				}
			} else if nil != d.ep[ep][descDirRx].callback {
				// call our device class-specific callback, if defined, on endpoint
				// data transfer complete events.
				d.ep[ep][descDirRx].callback(rxEndpoint(ep), count)
			}
		}
	}

}

// prepareSetup configures the buffer for setup packets received on control
// endpoint 0 Rx (OUT).
func (d *dhw) prepareSetup() {
	desc := d.endpointDescriptor(rxEndpoint(d.controlEndpoint()))
	// control buffer address is device class-specific
	desc.address.Set(uint32(d.controlSetupBuffer()))
	// overwrite the BYTE_COUNT and MULTI_PACKET_SIZE bitfields only (with 0 and
	// sizeof(dcdSetup), respectively).
	var mask uint32
	mask |= USB_DEVICE_PCKSIZE_BYTE_COUNT_Msk << USB_DEVICE_PCKSIZE_BYTE_COUNT_Pos
	mask |= USB_DEVICE_PCKSIZE_MULTI_PACKET_SIZE_Msk <<
		USB_DEVICE_PCKSIZE_MULTI_PACKET_SIZE_Pos
	desc.packetSize.ReplaceBits(pcksize(0, uint32(dcdSetupSize), 0, false), mask, 0)
}

func (d *dhw) setDeviceAddress(addr uint16) {

	// SAMx51 can only set address after status for this request is complete,
	// which is checked in (*dhw).controlStatusComplete(dcdSetup).

	// Save the device address to the receiver, because the SETUP packet
	// containing SET_ADDRESS request is not populating wValue correctly.
	d.address = addr

	// Enable SUSPEND interrupt since the bus signal D+/D- are stable now.
	d.bus.INTFLAG.Set(sam.USB_DEVICE_INTFLAG_SUSPEND)
	d.bus.INTENSET.Set(sam.USB_DEVICE_INTENSET_SUSPEND)
}

func (d *dhw) remoteWakeup() {
	d.bus.CTRLB.SetBits(sam.USB_DEVICE_CTRLB_UPRSM)
}

// =============================================================================
//  Control Endpoint 0
// =============================================================================

func (d *dhw) controlEndpoint() uint8 {
	switch d.cc.id {
	case classDeviceCDCACM:
		return descCDCACMEndpointCtrl
	case classDeviceHID:
		return descHIDEndpointCtrl
	}
	return descEndpointInvalid
}

func (d *dhw) controlSetupBuffer() uintptr {
	switch d.cc.id {
	case classDeviceCDCACM:
		return uintptr(unsafe.Pointer(&descCDCACM[d.cc.config-1].sx[0]))
	case classDeviceHID:
		return uintptr(unsafe.Pointer(&descHID[d.cc.config-1].sx[0]))
	}
	return 0
}

func (d *dhw) controlStatusBuffer(data []uint8) uintptr {
	switch d.cc.id {
	case classDeviceCDCACM:
		// reference to class configuration data
		c := descCDCACM[d.cc.config-1]
		for i := range c.cx {
			c.cx[i] = 0 // zero out the control reply buffer
		}
		// copy the given data into control reply buffer
		copy(c.cx[:], data)
		return uintptr(unsafe.Pointer(&c.cx[0]))

	case classDeviceHID:
		// reference to class configuration data
		c := descHID[d.cc.config-1]
		for i := range c.cx {
			c.cx[i] = 0 // zero out the control reply buffer
		}
		// copy the given data into control reply buffer
		copy(c.cx[:], data)
		return uintptr(unsafe.Pointer(&c.cx[0]))
	}
	return 0
}

// controlStall stalls a transfer on control endpoint 0. To stall a transfer on
// any other endpoint, use method endpointStall().
func (d *dhw) controlStall(stall bool, dir uint8) {
	// Argument dir will be either 0 = Rx (OUT), or 1 = Tx (IN).
	// We need to translate this to the USB standard, which is encoded as
	// direction D and endpoint number N with the 8-bit mask DxxxNNNN.
	// The value of direction D bit is the same as argument dir described above.
	endpoint := ((dir << descEndptAddrDirectionPos) & descEndptAddrDirectionMsk) |
		((d.controlEndpoint() << descEndptAddrNumberPos) & descEndptAddrNumberMsk)
	d.endpointStall(endpoint, stall)
}

func (d *dhw) controlStatusStart(endpoint uint8) {

	num, dir := unpackEndpoint(endpoint)

	// Swap direction of the given endpoint Rx->Tx and Tx->Rx
	switch dir {
	case descDirRx:
		endpoint = txEndpoint(num)
	case descDirTx:
		endpoint = rxEndpoint(num)
	}
	d.endpointTransfer(endpoint, 0, 0)
}

func (d *dhw) controlStatusComplete(endpoint uint8) {

	if (d.setup.bmRequestType&descRequestTypeTypeMsk == descRequestTypeTypeStandard) &&
		(d.setup.bmRequestType&(descRequestTypeRecipientMsk|descRequestTypeDirMsk) ==
			descRequestTypeRecipientDevice|descRequestTypeDirOut) &&
		(d.setup.bRequest == descRequestStandardSetAddress) {

		d.bus.DADD.SetBits((uint8(d.address) << sam.USB_DEVICE_DADD_DADD_Pos) &
			sam.USB_DEVICE_DADD_DADD_Msk)
		d.bus.DADD.SetBits(sam.USB_DEVICE_DADD_ADDEN)
		d.event(dcdEvent{id: dcdEventDeviceAddress})
	}
	d.prepareSetup()
}

func (d *dhw) controlTransferStart(endpoint uint8) {

	num, dir := unpackEndpoint(endpoint)

	// Dequeue the next transfer descriptor available.
	if xfer, ok := d.ep[num][dir].pendingTransfer(); ok {
		// Update the active transfer descriptor on the corresponding endpoint.
		d.ep[num][dir].setActiveTransfer(xfer)
		// Invoke the DCD event handler for SETUP processing, which will enqueue
		// any necessary response transactions, which are serviced immediately
		// because this is the sole active transfer on this endpoint.
		d.event(dcdEvent{
			id:    dcdEventControlSetup,
			setup: xfer.setup,
		})
	}
}

func (d *dhw) controlTransferContinue(endpoint uint8, count, total uint32) {

	num, dir := unpackEndpoint(endpoint)

	if xfer, ok := d.ep[num][dir].activeTransfer(); ok {
		data, size := xfer.packetComplete(count)
		d.endpointTransfer(endpoint, data, size)
	}
}

func (d *dhw) controlTransferComplete(endpoint uint8, count, total uint32) {

	num, dir := unpackEndpoint(endpoint)
	setupDir := d.setup.direction()
	setupAddress := packEndpoint(num, setupDir)

	// If endpoint direction is opposite the direction in the original SETUP
	// packet, then this is the end of the STATUS stage, i.e., end of transfer.
	if dir != setupDir {
		// Run any post-processing for this endpoint.
		d.controlStatusComplete(setupAddress)
		// Notify the upper-layer device driver.
		d.event(dcdEvent{id: dcdEventControlComplete})
		// Clear the active transfer descriptor on this endpoint.
		d.ep[num][setupDir].setActiveTransfer(nil)
		// Start processing any pending control transfers.
		d.controlTransferStart(setupAddress)
	} else {
		// Initiate ZLP transfer in opposite direction.
		d.controlStatusStart(endpoint)
	}
}

// controlReceive receives (Rx, OUT) the first data packet on control endpoint 0.
// If the given data pointer and size are both 0, then a zero-length status
// packet (ZLP) is transmitted (Tx, IN) on control endpoint 0.
func (d *dhw) controlReceive(data uintptr, size uint32, notify bool) {
	ep := d.controlEndpoint()
	if size > 0 && data > 0 {
		if xfer, ok := d.ep[ep][descDirRx].activeTransfer(); ok {
			next := xfer.packetStart(data, size)
			d.endpointTransfer(rxEndpoint(ep), data, next)
		}
	} else {
		d.endpointTransfer(txEndpoint(ep), 0, 0)
	}
}

// controlTransmit transmits (Tx, IN) the first data packet on control endpoint 0.
// If the given data pointer and size are both 0, then a zero-length status
// packet (ZLP) is received (Rx, OUT) on control endpoint 0.
func (d *dhw) controlTransmit(data uintptr, size uint32, notify bool) {
	ep := d.controlEndpoint()
	if size > 0 && data > 0 {
		if xfer, ok := d.ep[ep][descDirTx].activeTransfer(); ok {
			next := xfer.packetStart(data, size)
			d.endpointTransfer(txEndpoint(ep), data, next)
		}
	} else {
		d.endpointTransfer(rxEndpoint(ep), 0, 0)
	}
}

// =============================================================================
//  Endpoint Transfer Descriptor
// =============================================================================

type dhwTransfer struct {
	endpoint      uint8
	maxPacketSize uint32
	setup         dcdSetup
	data          uintptr
	size          uint32
	sent          uint32
}

// dhwTransferDepth defines the size of the dhwEPStatus.xferQueue buffered channel,
// which affects the number of transfers each endpoint can enqueue for processing.
const dhwTransferDepth = 8

type dhwTransferLUT [dhwTransferDepth]dhwTransfer

func (t *dhwTransfer) init(endpoint uint8, maxPacketSize uint32) {
	t.endpoint = endpoint
	t.maxPacketSize = maxPacketSize
	t.reset()
}

func (t *dhwTransfer) reset() {
	// Do not clear the endpoint field, as it is statically-assigned (during
	// program initialization) and is never intended to change.
	t.setup.set(0)
	t.data = 0
	t.size = 0
	t.sent = 0
}

func (t *dhwTransfer) packetStart(data uintptr, size uint32) (next uint32) {
	t.data = data
	t.size = size
	t.sent = 0
	if next = size; next > t.maxPacketSize {
		next = t.maxPacketSize
	}
	return next
}

func (t *dhwTransfer) packetComplete(sent uint32) (data uintptr, size uint32) {
	t.sent = sent
	if size = t.size - t.sent; size > t.maxPacketSize {
		size = t.maxPacketSize
	}
	return t.data + uintptr(t.sent), size
}

func (t *dhwTransfer) hasDataPayload() bool  { return t.data != 0 || t.size != 0 }
func (t *dhwTransfer) hasSetupPayload() bool { return t.setup.pack() != 0 }
func (t *dhwTransfer) hasPayload() bool      { return t.hasDataPayload() || t.hasSetupPayload() }

// =============================================================================
//  Endpoint Configuration and Status
// =============================================================================

// dhwEPStatus holds the status, completion callback of the configured class
// driver, and the transfer queue for a given endpoint.
//
// The transfer queue is structured as follows:
//
//   - The xferTable field is a statically-allocated, single-dimensional array
//     used as a buffer of transfer requests - known as transfer descriptors -
//     on a single, directional endpoint.
//
//   - The length of xferTable defines the maximum number of pending transfers
//     in a given direction on a single endpoint.
//
//   - Since the transfer descriptors are statically-allocated, we do not risk
//     heap allocation when requesting transfers in the USB interrupt handler.
//
//   - The xferQueue field is a buffered channel of uint8 with capacity equal to
//     the length of the transfer descriptor table xferTable.
//
//   - To enqueue a new transfer request, the xferTable is first scanned to find
//     the index of an unused transfer descriptor. The descriptor at this table
//     index is populated with the transfer details, and this table index is
//     written to the xferQueue channel.
//
//   - If no other transfer descriptors are queued, the transfer is immediatly
//     sent to the USB. Otherwise, the next descriptor in queue will be read
//     from the xferQueue channel upon the next transfer complete interrupt
//     triggered on this endpoint.
//
//   - Once the transfer descriptor's table index is read from the xferQueue
//     channel, the descriptor is cleared in the xferTable, marking it free for
//     use with a subsequent transfer request.
//
type dhwEPStatus struct {
	device     *dhw
	endpoint   uint8
	callback   func(endpoint uint8, size uint32)
	flags      volatile.Register8
	xferActive volatile.Register32
	xferFIFO   [dhwTransferDepth]uint8
	xferQueue  Queue
	xferTable  dhwTransferLUT
}

// dhwEPAddrStatus contains an endpoint number's dhwEPStatus for both IN + OUT
// directions.
type dhwEPAddrStatus [2]dhwEPStatus

// Bitmasks for each bitfield stored in volatile field dhwEPStatus.flags.
const (
	dhwEPStatusStatusBusy    = 0x1
	dhwEPStatusStatusStalled = 0x2
	dhwEPStatusStatusClaimed = 0x4
)

func (s *dhwEPStatus) init(dhw *dhw, endpoint uint8) {
	s.device = dhw
	s.endpoint = endpoint
	s.callback = nil
	s.flags.Set(0)
	fifo := s.xferFIFO[:]
	s.xferQueue.Init(&fifo, dhwTransferDepth, QueueFullDiscardLast)
	mps := dhw.endpointMaxPacketSize(endpoint)
	for i := range s.xferTable {
		s.xferTable[i].init(endpoint, mps)
	}
}

// Accessor methods to return the logical boolean value from the bit value
// stored in volatile field dhwEPStatus.flags.
func (s *dhwEPStatus) busy() bool    { return s.flags.HasBits(dhwEPStatusStatusBusy) }
func (s *dhwEPStatus) stalled() bool { return s.flags.HasBits(dhwEPStatusStatusStalled) }
func (s *dhwEPStatus) claimed() bool { return s.flags.HasBits(dhwEPStatusStatusClaimed) }

// Mutator methods to set the bit value from the logical boolean value stored in
// volatile field dhwEPStatus.flags.
func (s *dhwEPStatus) setBusy(set bool)    { s.setFlags(set, dhwEPStatusStatusBusy) }
func (s *dhwEPStatus) setStalled(set bool) { s.setFlags(set, dhwEPStatusStatusStalled) }
func (s *dhwEPStatus) setClaimed(set bool) { s.setFlags(set, dhwEPStatusStatusClaimed) }

// setFlags consolidates the common logic of each dhwEPStatus mutator method
// defined above.
func (s *dhwEPStatus) setFlags(set bool, mask uint8) {
	if set {
		s.flags.SetBits(mask)
	} else {
		s.flags.ClearBits(mask)
	}
}

// hasActiveTransfer returns true if and only if the receiver's active transfer
// descriptor is not nil.
//
// Note that the result of this call does not guarantee a subsequent call to
// activeTransfer will succeed, as the active transfer may have been cleared
// preemptively (from the USB interrupt handler) during the time between these
// two calls. Thus, you should always verify an active transfer descriptor was
// obtained with the bool value returned from activeTransfer.
func (s *dhwEPStatus) hasActiveTransfer() bool {
	_, ok := s.activeTransfer()
	return ok
}

// activeTransfer returns a pointer to the receiver's active transfer descriptor
// being processed in one of the transaction stages (SETUP, DATA, or STATUS).
// The bool value returned is true if and only if the receiver's active transfer
// descriptor is not nil.
//
// The pointer returned refers to an element in the receiver's xferTable, which
// is also used by the receiver's pending transfer queue (FIFO). Thus, you can
// (and should) use this object to reset transfer descriptors when processing
// has completed (using (*dhwTransfer).reset()). This frees the descriptor and
// allows new transfer requests to be scheduled.
// You may also use (*dhwEPStatus).setActiveTransfer(nil) to free the descriptor
// if the receiver's active transfer descriptor is not nil.
func (s *dhwEPStatus) activeTransfer() (*dhwTransfer, bool) {
	if active := s.xferActive.Get(); active != 0 {
		return (*dhwTransfer)(unsafe.Pointer(uintptr(active))), true
	}
	return nil, false
}

// setActiveTransfer sets or clears the receiver's active transfer descriptor.
// The receiver's active transfer descriptor is cleared if the given transfer
// descriptor is nil.
//
// If the given transfer descriptor is nil, and the receiver's active transfer
// descriptor is not nil, then the receiver's active transfer descriptor is
// reset, marking it free for use by the receiver's transfer queue (FIFO).
//
// The given transfer descriptor should be a pointer into the receiver's
// transfer table xferTable. This enables interaction with the receiver's
// transfer queue, allowing it to detect when a descriptor is busy or available
// for scheduling.
func (s *dhwEPStatus) setActiveTransfer(xfer *dhwTransfer) {
	if xfer == nil {
		// Clearing the active transfer. Check if an active descriptor exists.
		if actv, ok := s.activeTransfer(); ok {
			// Reset the descriptor, freeing it for use in the transfer queue (FIFO).
			actv.reset()
		}
		s.xferActive.Set(0)
	} else {
		s.xferActive.Set(uint32(uintptr(unsafe.Pointer(xfer))))
	}
}

// hasPendingTransfer returns true if and only if the number of pending
// transfers in the receiver's transfer queue is greater than zero.
//
// Note that the result of this call does not guarantee that calls to either
// pendingTransfer/scheduleSetup/scheduleTransfer will succeed, as new requests
// may be added/removed preemptively (from the USB interrupt handler) during the
// time between these two calls. Thus, you should always verify queue operations
// operations by inspecting the final bool value returned by each of these
// mentioned functions.
func (s *dhwEPStatus) hasPendingTransfer() bool {
	return s.xferQueue.Len() > 0
}

// pendingTransfer dequeues the table index - referring to the next transfer
// descriptor to be processed - from the receiver's xferQueue, returning the
// transfer descriptor at that index and true to indicate a pending transfer
// descriptor was successfully obtained.
//
// If the receiver's transfer queue is empty, then the returned values are nil
// and a false bool value to indicate failure to obtain a pending transfer
// descriptor.
func (s *dhwEPStatus) pendingTransfer() (*dhwTransfer, bool) {
	if s.hasPendingTransfer() {
		s.device.enableInterrupts(false)
		defer s.device.enableInterrupts(true)
		if i, ok := s.xferQueue.Deq(); ok {
			return &s.xferTable[i], true
		}
	}
	return nil, false
}

// claimSchedule disables interrupts and scans the receiver's transfer table
// for an unused transfer descriptor, returning its table index and true.
// If all transfer descriptors are already claimed, re-enables interrupts and
// returns -1 and false.
//
// -- ** IMPORTANT ** --
// Note that interrupts are NOT re-enabled when a transfer index is
// successfully found and returned. This ensures no race condition exists
// between locating a free transfer index and initializing the transfer at that
// index. These two events must not be preempted by another scheduling request
// from the USB interrupt handler.
// The caller must re-enable interrupts once the available transfer at the
// vacant index has been processed.
//
// ( Because of this potentially danerous behavior, claimSchedule should be
//   restricted to the scheduling methods — scheduleTransfer and scheduleSetup —
//   so it can be verified easily that interrupts get re-enabled in all cases. )
func (s *dhwEPStatus) claimSchedule() (int, bool) {
	// Disable interrupts while scanning the xferTable
	s.device.enableInterrupts(false)
	for i := range s.xferTable {
		// Check that transfer has no payloads
		if !s.xferTable[i].hasPayload() {
			// Return index into xferTable (leave interrupts disabled!)
			return i, true
		}
	}
	// All elements of xferTable have a payload, so we cannot schedule a new
	// transfer. This request will be ignored, and we can re-enable interrupts
	// immediately.
	//
	// Realistically, we should never encounter this condition with a
	// sufficiently-sized xferTable/xferQueue and a well-behaved USB host.
	//
	// If you do reach this point, check that the transfers are being cleaned
	// up properly (with (*dhwTransfer).reset()) in the respective transfer
	// completion event handler.
	s.device.enableInterrupts(true)
	return -1, false
}

// scheduleTransfer enqueues a new data transfer descriptor to the receiver's
// transfer queue.
//
// The first bool returned indicates if this transfer request is the the only
// request in the queue, no other active transfer exists, and is thus available
// for immediate processing.
// The second bool returned is true if and only if the transfer request was
// added to the queue successfully.
// If the receiver's transfer queue is full, the request is ignored and false is
// returned for both return values.
func (s *dhwEPStatus) scheduleTransfer(data uintptr, size uint32) (ready bool, ok bool) {
	var i int
	if i, ok = s.claimSchedule(); ok {
		defer s.device.enableInterrupts(true)
		s.xferTable[i].reset()
		s.xferTable[i].data = data
		s.xferTable[i].size = size
		return !s.hasActiveTransfer() && !s.hasPendingTransfer(),
			s.xferQueue.Enq(uint8(i))
	}
	return false, false
}

// scheduleSetup enqueues a new control SETUP transfer to the receiver's
// transfer queue.
//
// The first bool returned indicates if this transfer request is the the only
// request in the queue, no other active transfer exists, and is thus available
// for immediate processing.
// The second bool returned is true if and only if the transfer request was
// added to the queue successfully.
// If the receiver's transfer queue is full, the request is ignored and false is
// returned for both return values.
func (s *dhwEPStatus) scheduleSetup(setup dcdSetup) (ready bool, ok bool) {
	var i int
	if i, ok = s.claimSchedule(); ok {
		defer s.device.enableInterrupts(true)
		s.xferTable[i].reset()
		s.xferTable[i].setup = setup
		return !s.hasActiveTransfer() && !s.hasPendingTransfer(),
			s.xferQueue.Enq(uint8(i))
	}
	return false, false
}

// =============================================================================
//  Endpoint Descriptor
// =============================================================================

// dhwEPDesc defines a USB endpoint descriptor, used to inform the USB DMA
// controller the location of each endpoint transfer buffer.
//
// Access to these instances is controlled; i.e., you shouldn't need to use
// them directly. Instead, use the higher-level API on types dhwEPStatus and
// dhwTransfer, through the (*dhw).ep[num][dir] elements, for scheduling and
// inspecting endpoint transfers.
type dhwEPDesc struct {
	address    volatile.Register32
	packetSize volatile.Register32
	extToken   volatile.Register16
	bankStatus volatile.Register8
	_          [5]uint8
}

// dhwEPAddrDesc defines an endpoint address descriptor, representing both
// directions (IN + OUT) of a given endpoint descriptor.
type dhwEPAddrDesc [2]dhwEPDesc

// Constants defining bitfields in the endpoint descriptor hardware register
// PCKSIZE. These were left out of the SVD for some reason.
const (
	USB_DEVICE_PCKSIZE_BYTE_COUNT_Pos = 0
	USB_DEVICE_PCKSIZE_BYTE_COUNT_Msk = 0x3FFF

	USB_DEVICE_PCKSIZE_MULTI_PACKET_SIZE_Pos = 14
	USB_DEVICE_PCKSIZE_MULTI_PACKET_SIZE_Msk = 0x3FFF

	USB_DEVICE_PCKSIZE_SIZE_Pos = 28
	USB_DEVICE_PCKSIZE_SIZE_Msk = 0x7

	USB_DEVICE_PCKSIZE_AUTOZLP_Pos = 31
	USB_DEVICE_PCKSIZE_AUTOZLP_Msk = 0x1
)

// pcksize is a convenience routine that constructs the bitfields of the PCKSIZE
// register of the USB_DEVICE peripheral, whose Pos/Msk definitions were ommitted
// from the SVD-generated device file.
//go:inline
func pcksize(byteCount, multiPacketSize, size uint32, zlp bool) uint32 {
	var zlpMask uint32
	if zlp {
		zlpMask = USB_DEVICE_PCKSIZE_AUTOZLP_Msk << USB_DEVICE_PCKSIZE_AUTOZLP_Pos
	}
	return ((byteCount & USB_DEVICE_PCKSIZE_BYTE_COUNT_Msk) <<
		USB_DEVICE_PCKSIZE_BYTE_COUNT_Pos) |
		((multiPacketSize & USB_DEVICE_PCKSIZE_MULTI_PACKET_SIZE_Msk) <<
			USB_DEVICE_PCKSIZE_MULTI_PACKET_SIZE_Pos) |
		((size & USB_DEVICE_PCKSIZE_SIZE_Msk) << USB_DEVICE_PCKSIZE_SIZE_Pos) |
		(zlpMask)
}

var (
	// endpointSizeEnum is a constant-time lookup table for translating packet
	// sizes (bytes) to the corresponding register PCKSIZE.SIZE enumerated value.
	//
	// These tables are used instead of simple arithmetic (powers of 2) because of
	// the exceptional case with packet size = 1023.
	endpointSizeEnum = map[uint32]uint32{
		8: 0, 16: 1, 32: 2, 64: 3, 128: 4, 256: 5, 512: 6, 1023: 7,
	}

	// endpointEnumSize is a constant-time lookup table for translating the
	// register PCKSIZE.SIZE enumerated values to its packet size (bytes).
	//
	// These tables are used instead of simple arithmetic (powers of 2) because of
	// the exceptional case with packet size = 1023.
	endpointEnumSize = [8]uint32{
		/* 0= */ 8,
		/* 1= */ 16,
		/* 2= */ 32,
		/* 3= */ 64,
		/* 4= */ 128,
		/* 5= */ 256,
		/* 6= */ 512,
		/* 7= */ 1023,
	}
)

// endpointSizeEncode returns the register PCKSIZE.SIZE enumerated value for a
// given endpoint descriptor packet size (bytes).
//
// See documentation on endpoint descriptor bank SRAM register PCKSIZE, bit
// field SIZE for details.
//go:inline
func endpointSizeEncode(size uint32) (enum uint32, ok bool) {
	enum, ok = endpointSizeEnum[size]
	return
}

// endpointSizeDecode returns the endpoint descriptor packet size (bytes) for a
// given register PCKSIZE.SIZE enumerated value.
//
// See documentation on endpoint descriptor bank SRAM register PCKSIZE, bit
// field SIZE for details.
//go:inline
func endpointSizeDecode(enum uint32) (size uint32, ok bool) {
	if ok = int(enum) < len(endpointEnumSize); ok {
		size = endpointEnumSize[enum]
	}
	return
}

// endpointDescriptors returns the OUT + IN endpoint descriptors for the given
// endpoint number, encoded as direction D and endpoint number N with the 8-bit
// mask D000NNNN. The direction bit D is ignored.
//go:inline
func (d *dhw) endpointDescriptors(endpoint uint8) (out, in *dhwEPDesc) {
	// endpoint descriptor is device class-specific
	return d.endpointDescriptor(rxEndpoint(endpoint)),
		d.endpointDescriptor(txEndpoint(endpoint))
}

// endpointDescriptor returns the endpoint descriptor for the given endpoint
// address, encoded as direction D and endpoint number N with the 8-bit mask
// D000NNNN.
//go:inline
func (d *dhw) endpointDescriptor(endpoint uint8) *dhwEPDesc {
	// endpoint descriptor is device class-specific
	num, dir := unpackEndpoint(endpoint)
	switch d.cc.id {
	case classDeviceCDCACM:
		return &descCDCACM[d.cc.config-1].ed[num][dir]
	case classDeviceHID:
		return &descHID[d.cc.config-1].ed[num][dir]
	default:
		return nil
	}
}

func (d *dhw) endpointMaxPacketSize(endpoint uint8) uint32 {

	switch d.cc.id {
	case classDeviceCDCACM:

		switch endpointNumber(endpoint) {
		case descCDCACMEndpointCtrl:
			return descControlPacketSize

		case descCDCACMEndpointStatus:
			return descCDCACMStatusPacketSize

		case descCDCACMEndpointDataRx:
			return descCDCACMDataRxPacketSize

		case descCDCACMEndpointDataTx:
			return descCDCACMDataTxPacketSize
		}
	case classDeviceHID:

		switch endpointNumber(endpoint) {
		case descHIDEndpointCtrl:
			return descControlPacketSize

		case descHIDEndpointKeyboard:
			return descHIDKeyboardTxPacketSize

		case descHIDEndpointMouse:
			return descHIDMouseTxPacketSize

		case descHIDEndpointSerialRx: // == descHIDEndpointSerialTx
			switch endpoint {
			case rxEndpoint(endpoint):
				return descHIDSerialRxPacketSize
			case txEndpoint(endpoint):
				return descHIDSerialTxPacketSize
			}

		case descHIDEndpointJoystick:
			return descHIDJoystickTxPacketSize

		case descHIDEndpointMediaKey:
			return descHIDMediaKeyTxPacketSize
		}
	}
	return descControlPacketSize
}

func (d *dhw) endpointEnable(endpoint uint8, control bool, config uint32) {

	if control {

		// Configure control endpoint 0 Rx (bank 0, OUT) and Tx (bank 1, IN)
		out, in := d.endpointDescriptors(d.controlEndpoint())

		if enum, ok := endpointSizeEncode(descControlPacketSize); ok {

			num := endpointNumber(d.controlEndpoint())

			// Initialize IN and OUT transfer descriptors on control endpoint 0.
			d.ep[num][descDirRx].init(d, rxEndpoint(num))
			d.ep[num][descDirTx].init(d, txEndpoint(num))

			// Conigure packet size for control endpoints.
			out.packetSize.ReplaceBits(enum,
				USB_DEVICE_PCKSIZE_SIZE_Msk, USB_DEVICE_PCKSIZE_SIZE_Pos)
			in.packetSize.ReplaceBits(enum,
				USB_DEVICE_PCKSIZE_SIZE_Msk, USB_DEVICE_PCKSIZE_SIZE_Pos)

			// rxType/txType uses the same rationale as epType (defined below in the
			// else-branch that handles non-control endpoints).
			// Thus, we add +1 to the value below.
			//
			// See the comment above the previously-mentioned epType (below)
			rxType := uint8(descEndptTypeControl+1) <<
				sam.USB_DEVICE_ENDPOINT_EPCFG_EPTYPE0_Pos
			txType := uint8(descEndptTypeControl+1) <<
				sam.USB_DEVICE_ENDPOINT_EPCFG_EPTYPE1_Pos

			// Configure bank 0 Rx (SETUP/OUT) as CONTROL, bank 1 Tx (IN) as CONTROL.
			d.bus.DEVICE_ENDPOINT[num].EPCFG.Set(rxType | txType)
			// Enable transfer complete and SETUP received interrupts
			d.bus.DEVICE_ENDPOINT[num].EPINTENSET.Set(
				sam.USB_DEVICE_ENDPOINT_EPINTENSET_TRCPT0 |
					sam.USB_DEVICE_ENDPOINT_EPINTENSET_TRCPT1 |
					sam.USB_DEVICE_ENDPOINT_EPINTENSET_RXSTP)

			// Prepare to start processing SETUP packets
			d.prepareSetup()
		}

	} else {

		desc := d.endpointDescriptor(endpoint)

		if enum, ok := endpointSizeEncode(d.endpointMaxPacketSize(endpoint)); ok {

			num, dir := unpackEndpoint(endpoint)

			// Initialize transfer descriptors now that the device class configuration
			// has been defined, which affects maximum packet size.
			d.ep[num][dir].init(d, endpoint)

			desc.packetSize.ReplaceBits(enum,
				USB_DEVICE_PCKSIZE_SIZE_Msk, USB_DEVICE_PCKSIZE_SIZE_Pos)

			// config contains the bmAttributes field per USB standard EP descriptor,
			// i.e., ctrl=0, iso=1, bulk=2, int=3, which corresponds to the EPCFG
			// register's EPTYPE0/1 bitfield+1: ctrl=1, iso=2, bulk=3, int=4, dual=5.
			// Thus, we add +1 to the value below.

			switch endpoint {
			case rxEndpoint(endpoint):

				epType := ((config >> descEndptConfigAttrRxPos) &
					descEndptAttrSyncTypeMsk) >> descEndptAttrSyncTypePos

				d.bus.DEVICE_ENDPOINT[num].EPCFG.ReplaceBits(
					uint8(epType+1)<<sam.USB_DEVICE_ENDPOINT_EPCFG_EPTYPE0_Pos,
					sam.USB_DEVICE_ENDPOINT_EPCFG_EPTYPE0_Msk, 0)

				d.bus.DEVICE_ENDPOINT[num].EPSTATUSCLR.Set(
					sam.USB_DEVICE_ENDPOINT_EPSTATUSCLR_STALLRQ0 |
						sam.USB_DEVICE_ENDPOINT_EPSTATUSCLR_DTGLOUT)

				d.bus.DEVICE_ENDPOINT[num].EPINTENSET.Set(
					sam.USB_DEVICE_ENDPOINT_EPINTENSET_TRCPT0)

			case txEndpoint(endpoint):

				epType := ((config >> descEndptConfigAttrTxPos) &
					descEndptAttrSyncTypeMsk) >> descEndptAttrSyncTypePos

				d.bus.DEVICE_ENDPOINT[num].EPCFG.ReplaceBits(
					uint8(epType+1)<<sam.USB_DEVICE_ENDPOINT_EPCFG_EPTYPE1_Pos,
					sam.USB_DEVICE_ENDPOINT_EPCFG_EPTYPE1_Msk, 0)

				d.bus.DEVICE_ENDPOINT[num].EPSTATUSCLR.Set(
					sam.USB_DEVICE_ENDPOINT_EPSTATUSCLR_STALLRQ1 |
						sam.USB_DEVICE_ENDPOINT_EPSTATUSCLR_DTGLIN)

				d.bus.DEVICE_ENDPOINT[num].EPINTENSET.Set(
					sam.USB_DEVICE_ENDPOINT_EPINTENSET_TRCPT1)
			}
		}
	}
}

func (d *dhw) endpointConfigure(endpoint uint8, callback func(endpoint uint8, size uint32)) {
	num, dir := unpackEndpoint(endpoint)
	d.ep[num][dir].callback = callback
}

// endpointStall sets or clears a stall on the given endpoint.
func (d *dhw) endpointStall(endpoint uint8, stall bool) {

	if stall {
		switch endpoint {
		case rxEndpoint(endpoint):
			d.bus.DEVICE_ENDPOINT[endpointNumber(endpoint)].EPSTATUSSET.Set(
				sam.USB_DEVICE_ENDPOINT_EPSTATUSSET_STALLRQ0)
		case txEndpoint(endpoint):
			d.bus.DEVICE_ENDPOINT[endpointNumber(endpoint)].EPSTATUSSET.Set(
				sam.USB_DEVICE_ENDPOINT_EPSTATUSSET_STALLRQ1)
		}
	} else {
		switch endpoint {
		case rxEndpoint(endpoint):
			d.bus.DEVICE_ENDPOINT[endpointNumber(endpoint)].EPSTATUSCLR.Set(
				sam.USB_DEVICE_ENDPOINT_EPSTATUSCLR_STALLRQ0 |
					sam.USB_DEVICE_ENDPOINT_EPSTATUSCLR_DTGLOUT)
		case txEndpoint(endpoint):
			d.bus.DEVICE_ENDPOINT[endpointNumber(endpoint)].EPSTATUSCLR.Set(
				sam.USB_DEVICE_ENDPOINT_EPSTATUSCLR_STALLRQ1 |
					sam.USB_DEVICE_ENDPOINT_EPSTATUSCLR_DTGLIN)
		}
	}
	num, dir := unpackEndpoint(endpoint)
	d.ep[num][dir].setStalled(stall)
}

func (d *dhw) endpointStatus(endpoint uint8) (status uint16) {
	num, dir := unpackEndpoint(endpoint)
	if int(num) < len(d.ep) {
		ep := d.ep[num][dir]
		if ep.stalled() {
			status |= 0x0001
		}
	}
	return status
}

func (d *dhw) endpointSetFeature(endpoint uint8) {
	d.endpointStall(endpoint, true)
}

func (d *dhw) endpointClearFeature(endpoint uint8) {
	d.endpointStall(endpoint, false)
}

func (d *dhw) endpointTransfer(endpoint uint8, data uintptr, size uint32) {

	desc := d.endpointDescriptor(endpoint)
	desc.address.Set(uint32(data))

	switch num, dir := unpackEndpoint(endpoint); dir {

	case descDirRx: // OUT

		// overwrite the BYTE_COUNT and MULTI_PACKET_SIZE bitfields only (with 0 and
		// size, respectively).
		var mask uint32
		mask |= USB_DEVICE_PCKSIZE_BYTE_COUNT_Msk <<
			USB_DEVICE_PCKSIZE_BYTE_COUNT_Pos
		mask |= USB_DEVICE_PCKSIZE_MULTI_PACKET_SIZE_Msk <<
			USB_DEVICE_PCKSIZE_MULTI_PACKET_SIZE_Pos
		desc.packetSize.ReplaceBits(pcksize(0, size, 0, false), mask, 0)

		d.bus.DEVICE_ENDPOINT[num].EPSTATUSCLR.SetBits(
			sam.USB_DEVICE_ENDPOINT_EPSTATUSCLR_BK0RDY)
		d.bus.DEVICE_ENDPOINT[num].EPINTFLAG.SetBits(
			sam.USB_DEVICE_ENDPOINT_EPINTFLAG_TRFAIL0)

	case descDirTx: // IN

		// overwrite the BYTE_COUNT and MULTI_PACKET_SIZE bitfields only (with size
		// and 0, respectively).
		var mask uint32
		mask |= USB_DEVICE_PCKSIZE_BYTE_COUNT_Msk <<
			USB_DEVICE_PCKSIZE_BYTE_COUNT_Pos
		mask |= USB_DEVICE_PCKSIZE_MULTI_PACKET_SIZE_Msk <<
			USB_DEVICE_PCKSIZE_MULTI_PACKET_SIZE_Pos
		desc.packetSize.ReplaceBits(pcksize(size, 0, 0, false), mask, 0)

		d.bus.DEVICE_ENDPOINT[num].EPSTATUSSET.SetBits(
			sam.USB_DEVICE_ENDPOINT_EPSTATUSSET_BK1RDY)
		d.bus.DEVICE_ENDPOINT[num].EPINTFLAG.SetBits(
			sam.USB_DEVICE_ENDPOINT_EPINTFLAG_TRFAIL1)

	}
}

// endpointComplete handles transfer completion of a data endpoint.
func (d *dhw) endpointComplete(endpoint uint8, size uint32) {

}

// =============================================================================
//  General-Purpose (GP) Timer
// =============================================================================

func (d *dhw) timerConfigure(timer int, usec uint32, fn func()) {

}

func (d *dhw) timerOneShot(timer int) {

}

func (d *dhw) timerStop(timer int) {

}

// =============================================================================
//  [CDC-ACM] Serial UART (Virtual COM Port)
// =============================================================================

func (d *dhw) uartConfigure() {

	acm := &descCDCACM[d.cc.config-1]

	acm.setState(descCDCACMStateConfigured)

	// SAMx51 only supports USB full-speed (FS) operation
	acm.sxSize = descCDCACMStatusFSPacketSize
	acm.rxSize = descCDCACMDataRxFSPacketSize
	acm.txSize = descCDCACMDataTxFSPacketSize

	rq := acm.rxq[:]
	tq := acm.txq[:]

	// Rx gives priority to incoming data, Tx gives priority to outgoing data
	acm.rq.Init(&rq, int(acm.rxSize), QueueFullDiscardFirst)
	acm.tq.Init(&tq, int(acm.txSize), QueueFullDiscardLast)

	d.endpointEnable(txEndpoint(descCDCACMEndpointStatus),
		false, descCDCACMConfigAttrStatus)
	d.endpointEnable(rxEndpoint(descCDCACMEndpointDataRx),
		false, descCDCACMConfigAttrDataRx)
	d.endpointEnable(txEndpoint(descCDCACMEndpointDataTx),
		false, descCDCACMConfigAttrDataTx)

	d.endpointConfigure(txEndpoint(descCDCACMEndpointStatus),
		nil)
	d.endpointConfigure(rxEndpoint(descCDCACMEndpointDataRx),
		d.uartReceiveComplete)
	d.endpointConfigure(txEndpoint(descCDCACMEndpointDataTx),
		d.uartTransmitComplete)

	d.uartReceiveStart(rxEndpoint(descCDCACMEndpointDataRx))
}

func (d *dhw) uartSetLineState(state uint16) {
	acm := &descCDCACM[d.cc.config-1]
	acm.setState(descCDCACMStateLineState)
	if acm.ls.parse(state) {
		// TBD: respond to changes in line state?
	}
}

func (d *dhw) uartSetLineCoding(coding []uint8) {
	acm := &descCDCACM[d.cc.config-1]
	acm.setState(descCDCACMStateLineCoding)
	if acm.lc.parse(coding) {
		switch acm.lc.baud {
		case 1200:
			if acm.ls.dataTerminalReady {
				// reboot CPU
			}
		}
	}
}

func (d *dhw) uartReady() bool {
	acm := &descCDCACM[d.cc.config-1]
	// Ensure we have received SET_CONFIGURATION class request, and then both
	// SET_LINE_STATE and SET_LINE_CODING CDC requests (in that order).
	//
	// Many USB hosts will send a default SET_LINE_CODING prior to SET_LINE_STATE,
	// and then another SET_LINE_CODING containing the actual terminal settings.
	//
	// We do not want to start UART Rx/Tx transactions until after we have
	// received the final SET_LINE_CODING with the intended terminal settings.
	//
	// The "set" method on type descCDCACMState defines this incremental state
	// machine, with the UART's current state stored in the volatile.Register8
	// field "state" of descCDCACMClass.
	return d.state() == dcdStateConfigured && //acm.ls.dataTerminalReady &&
		acm.state.Get() == uint8(descCDCACMStateLineCoding)
}

func (d *dhw) uartReceiveStart(endpoint uint8) {
	acm := &descCDCACM[d.cc.config-1]
	num := uint16(endpoint) & descEndptAddrNumberMsk

	ready, _ := d.ep[num][descDirRx].scheduleTransfer(
		uintptr(unsafe.Pointer(&acm.rx[0])), acm.rxSize)
	if ready {
		if xfer, ok := d.ep[num][descDirRx].pendingTransfer(); ok {
			// Update the active transfer descriptor on the corresponding endpoint.
			d.ep[num][descDirRx].setActiveTransfer(xfer)
			d.endpointTransfer(endpoint, xfer.data, xfer.size)
		}
	}
}

func (d *dhw) uartReceiveComplete(endpoint uint8, size uint32) {
	acm := &descCDCACM[d.cc.config-1]
	num := uint16(endpoint) & descEndptAddrNumberMsk

	if xfer, ok := d.ep[num][descDirRx].activeTransfer(); ok {
		for ptr := xfer.data; ptr < xfer.data+uintptr(size); ptr++ {
			acm.rq.Enq(*(*uint8)(unsafe.Pointer(ptr)))
		}
	}
	d.ep[num][descDirRx].setActiveTransfer(nil)
	d.uartReceiveStart(endpoint)
}

func (d *dhw) uartTransmitStart(endpoint uint8) {
	acm := &descCDCACM[d.cc.config-1]
	num := uint16(endpoint) & descEndptAddrNumberMsk

	// BULK data endpoints can simply use a single time slot in the schedule, and
	// repeatedly transfer from the same transmit buffer (acm.tx) as soon as the
	// transaction complete callback has been called for a prior transaction.

	// Do not schedule another transfer if one is already active, or if our Tx
	// FIFO is currently empty.
	if d.ep[num][descDirTx].hasActiveTransfer() || acm.tq.Len() == 0 {
		return
	}

	if send, err := acm.tq.Read(acm.tx[:]); err == nil && send > 0 {
		ready, _ := d.ep[num][descDirTx].scheduleTransfer(
			uintptr(unsafe.Pointer(&acm.tx[0])), uint32(send))
		if ready {
			if xfer, ok := d.ep[num][descDirTx].pendingTransfer(); ok {
				d.ep[num][descDirTx].setActiveTransfer(xfer)
				d.endpointTransfer(endpoint, xfer.data, xfer.size)
			}
		}
	}
}

func (d *dhw) uartTransmitComplete(endpoint uint8, size uint32) {
	acm := &descCDCACM[d.cc.config-1]
	num := uint16(endpoint) & descEndptAddrNumberMsk
	if size > 0 && size%acm.txSize == 0 {
		// Send ZLP if transfer length is a multiple of max packet size.
		d.endpointTransfer(endpoint, 0, 0)
	}
	d.ep[num][descDirTx].setActiveTransfer(nil)
	d.uartTransmitStart(endpoint)
}

// uartFlush discards all buffered input (Rx) data.
func (d *dhw) uartFlush() {
	acm := &descCDCACM[d.cc.config-1]
	acm.rq.Reset(int(acm.rxSize))
}

func (d *dhw) uartAvailable() int {
	acm := &descCDCACM[d.cc.config-1]
	return acm.rq.Len()
}

func (d *dhw) uartPeek() (uint8, bool) {
	acm := &descCDCACM[d.cc.config-1]
	return acm.rq.Front()
}

func (d *dhw) uartReadByte() (uint8, bool) {
	acm := &descCDCACM[d.cc.config-1]
	return acm.rq.Deq()
}

func (d *dhw) uartRead(data []uint8) (int, error) {
	acm := &descCDCACM[d.cc.config-1]
	return acm.rq.Read(data)
}

func (d *dhw) uartWriteByte(c uint8) error {
	_, err := d.uartWrite([]uint8{c})
	return err
}

func (d *dhw) uartWrite(data []uint8) (int, error) {

	acm := &descCDCACM[d.cc.config-1]
	num := uint16(descCDCACMEndpointDataTx) & descEndptAddrNumberMsk

	var sent int
	var werr error
	for off := 0; off < len(data); off += int(acm.txSize) {

		cnt := len(data[off:])
		if cnt > int(acm.txSize) {
			cnt = int(acm.txSize)
		}

		// Block until we have room in the Tx FIFO. Space will become available once
		// the endpoint transaction complete interrupt is raised for the Tx BULK data
		// endpoint, and then the uartTransmitComplete callback has dequeued data from
		// the Tx FIFO (acm.tq) into the Tx transmit buffer (acm.tx).
		for acm.tq.Rem() < cnt {
		}

		// Add data to Tx FIFO
		add, err := acm.tq.Write(data[off : off+cnt])
		if err != nil {
			werr = err
			break
		}
		sent += add

		if d.ep[num][descDirTx].hasActiveTransfer() {
			// If there is already a transmit in-progress, wait for its callback to
			// detect new data in the FIFO and continue the transfer automatically.
		} else {
			// Otherwise, initiate a new data transfer.
			d.uartTransmitStart(txEndpoint(descCDCACMEndpointDataTx))
		}
	}

	return sent, werr
}

// =============================================================================
//  [HID] Serial
// =============================================================================

func (d *dhw) serialConfigure() {

	hid := &descHID[d.cc.config-1]

	// SAMx51 only supports USB full-speed (FS) operation
	hid.rxSerialSize = descHIDSerialRxFSPacketSize
	hid.txSerialSize = descHIDSerialTxFSPacketSize

	// Rx and Tx are on same endpoint
	d.endpointEnable(descHIDEndpointSerialRx,
		false, descHIDConfigAttrSerial)

	// d.endpointConfigureRx(descHIDEndpointSerialRx,
	// 	hid.rxSerialSize, false, d.serialNotify)
	// d.endpointConfigureTx(descHIDEndpointSerialTx,
	// 	hid.txSerialSize, false, nil)

	// for i := range hid.rdSerial {
	// 	d.serialReceive(uint8(i))
	// }

	d.timerConfigure(0, descHIDSerialTxSyncUs, d.serialSync)
}

func (d *dhw) serialReceive(endpoint uint8) {
	hid := &descHID[d.cc.config-1]
	num := uint16(endpoint) & descEndptAddrNumberMsk
	_, _ = hid, num // TODO(ardnew): elaborate stub
}

func (d *dhw) serialTransmit() {
	hid := &descHID[d.cc.config-1]
	_ = hid // TODO(ardnew): elaborate stub
}

func (d *dhw) serialNotify( /* transfer *dhwTransfer */ ) {
	// hid := &descHID[d.cc.config-1]
	// len := hid.rxSerialSize - (uint16(transfer.token>>16) & 0x7FFF)
	// _ = len // TODO(ardnew): elaborate stub
}

// serialFlush discards all buffered input (Rx) data.
func (d *dhw) serialFlush() {
	hid := &descHID[d.cc.config-1]
	_ = hid
}

func (d *dhw) serialSync() {

}

// =============================================================================
//  [HID] Keyboard
// =============================================================================

func (d *dhw) keyboard() *Keyboard { return descHID[d.cc.config-1].keyboard }

func (d *dhw) keyboardConfigure() {

	hid := &descHID[d.cc.config-1]

	// Initialize keyboard
	hid.keyboard.configure(d.dcd, hid)

	// SAMx51 only supports USB full-speed (FS) operation
	hid.txKeyboardSize = descHIDKeyboardTxFSPacketSize

	d.endpointEnable(descHIDEndpointKeyboard,
		false, descHIDConfigAttrKeyboard)
	d.endpointEnable(descHIDEndpointMediaKey,
		false, descHIDConfigAttrMediaKey)

	// d.endpointConfigureTx(descHIDEndpointKeyboard,
	// 	hid.txKeyboardSize, false, nil)
	// d.endpointConfigureTx(descHIDEndpointMediaKey,
	// 	hid.txKeyboardSize, false, nil)
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

	// hid := &descHID[d.cc.config-1]

	// size := uint16(len(data))
	// xfer := &hid.tdKeyboard[hid.txKeyboardHead]
	// when := ticks()
	// for {
	// 	if 0 == xfer.token&0x80 {
	// 		if 0 != xfer.token&0x68 {
	// 			// TODO: token contains error, how to handle?
	// 		}
	// 		hid.txKeyboardPrev = false
	// 		break
	// 	}
	// 	if hid.txKeyboardPrev {
	// 		return false
	// 	}
	// 	if ticks()-when > descHIDKeyboardTxTimeoutMs {
	// 		// Waited too long, assume host connection dropped
	// 		hid.txKeyboardPrev = true
	// 		return false
	// 	}
	// }
	// // Without this delay, the order packets are transmitted is seriously screwy.
	// udelay(60)
	// buff := hid.txKeyboard[hid.txKeyboardHead*descHIDKeyboardTxSize:]
	// _ = copy(buff, data)
	// d.transferPrepare(xfer, &buff[0], size, 0)
	// flushCache(uintptr(unsafe.Pointer(&buff[0])), descHIDKeyboardTxSize)
	// d.endpointTransmit(endpoint, xfer)
	// hid.txKeyboardHead += 1
	// if hid.txKeyboardHead >= descHIDKeyboardTDCount {
	// 	hid.txKeyboardHead = 0
	// }
	return true
}

// =============================================================================
//  [HID] Mouse
// =============================================================================

func (d *dhw) mouseConfigure() {

	hid := &descHID[d.cc.config-1]

	// SAMx51 only supports USB full-speed (FS) operation
	hid.txMouseSize = descHIDMouseTxFSPacketSize

	d.endpointEnable(descHIDEndpointMouse,
		false, descHIDConfigAttrMouse)

	// d.endpointConfigureTx(descHIDEndpointMouse,
	// 	hid.txMouseSize, false, nil)
}

// =============================================================================
//  [HID] Joystick
// =============================================================================

func (d *dhw) joystickConfigure() {

	hid := &descHID[d.cc.config-1]

	// SAMx51 only supports USB full-speed (FS) operation
	hid.txJoystickSize = descHIDJoystickTxFSPacketSize

	d.endpointEnable(descHIDEndpointJoystick,
		false, descHIDConfigAttrJoystick)

	// d.endpointConfigureTx(descHIDEndpointJoystick,
	// 	hid.txJoystickSize, false, nil)
}
