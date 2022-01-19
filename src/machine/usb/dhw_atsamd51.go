//go:build (sam && atsamd51) || (sam && atsame5x)
// +build sam,atsamd51 sam,atsame5x

package usb

// Implementation of USB device controller hardware abstraction (dhw) for
// Microchip SAMx51.

import (
	"device/arm"
	"device/sam"
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

	controlReply [8]uint8
	controlMask  uint32
	endpointMask uint32
	setup        dcdSetup
	stage        dcdStage
}

func runBootloader() {}

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
				func(interrupt.Interrupt) { coreInstance[0].dc.startOfFrame() })
		dhwInstance[instance].irqTC0 =
			interrupt.New(sam.IRQ_USB_TRCPT0,
				func(interrupt.Interrupt) { coreInstance[0].dc.complete(0) })
		dhwInstance[instance].irqTC1 =
			interrupt.New(sam.IRQ_USB_TRCPT1,
				func(interrupt.Interrupt) { coreInstance[0].dc.complete(1) })
	}

	// SAMx51 has only one USB PHY, which is full-speed
	if 0 == speed {
		speed = FullSpeed
	}
	dhwInstance[instance].speed = speed

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
	const clockGenerator = sam.GCLK_PCHCTRL_GEN_GCLK10
	sam.MCLK.APBBMASK.SetBits(sam.MCLK_APBBMASK_USB_)
	sam.MCLK.AHBMASK.SetBits(sam.MCLK_AHBMASK_USB_)
	sam.GCLK.PCHCTRL[clockGenerator].Set(clockGenerator | sam.GCLK_PCHCTRL_CHEN)

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

	// Reset USB peripheral
	d.bus.CTRLA.Set(sam.USB_DEVICE_CTRLA_SWRST)
	for !d.bus.SYNCBUSY.HasBits(sam.USB_DEVICE_SYNCBUSY_SWRST) {
	}
	for d.bus.SYNCBUSY.HasBits(sam.USB_DEVICE_SYNCBUSY_SWRST) {
	}
	d.calibrate()

	// USB Quality of Service: High Quality (3)
	d.bus.QOSCTRL.Set((3 << sam.USB_DEVICE_QOSCTRL_CQOS_Pos) |
		(3 << sam.USB_DEVICE_QOSCTRL_DQOS_Pos))

	// Install USB endpoint descriptor table (USB_DEVICE.DESCADD)
	var addr uintptr
	switch d.cc.id {
	case classDeviceCDCACM:
		addr = uintptr(unsafe.Pointer(descCDCACM[d.cc.config-1].ed))
	case classDeviceHID:
		addr = uintptr(unsafe.Pointer(descHID[d.cc.config-1].ed))
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
	d.bus.INTENSET.Set(sam.USB_DEVICE_INTENSET_SOF | sam.USB_DEVICE_INTENSET_EORST)

	// Ensure D+ pulled down long enough for host to detect a previous disconnect
	udelay(5000)

	return statusOK
}

// enable causes the USB core to enter (or exit) the normal run state and
// enables/disables all interrupts on the receiver's USB port.
func (d *dhw) enable(enable bool) {
	d.enableInterrupts(enable)
}

// enableInterrupts enables/disables all interrupts on the receiver's USB port.
func (d *dhw) enableInterrupts(enable bool) {
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

// enableSOF enables or disables start-of-frame (SOF) interrupts on the given
// USB device interface.
func (d *dhw) enableSOF(enable bool, iface uint8) {
}

// interrupt handles the USB hardware interrupt events on the "OTHER" IRQ line
// and notifies the device controller driver using a common "virtual interrupt"
// code.
func (d *dhw) interrupt() {

	// read and clear the interrupts that fired
	status := d.bus.USBSTS.Get() & d.bus.USBINTR.Get()
	d.bus.USBSTS.Set(status)

}

// startOfFrame handles the USB hardware interrupt events on the "SOF_HSOF" IRQ
// lines and notifies the device controller driver using a common "virtual
// interrupt" code.
func (d *dhw) startOfFrame() {
}

// complete handles the USB hardware interrupt events on the "TRCPT0" and
// "TRCPT1" IRQ lines and notifies the device controller driver using a common
// "virtual interrupt" code.
//
// When bank is 0, the interrupt occurred on "TRCPT0". Otherwise, bank is 1, and
// the interrupt occurred on "TRCPT1".
func (d *dhw) complete(bank int) {
}

func (d *dhw) prepareSetup() {

	// Configure control endpoint 0 OUT only
	endpoint := d.controlEndpoint()
	out, _ := d.endpointAddressDescriptor(endpoint)

	out.address.Set(uint32(d.controlBuffer()))
	out.packetSize.ReplaceBits(
		pcksize(0, 0, uint32(dcdSetupSize)),
		(USB_DEVICE_PCKSIZE_MULTI_PACKET_SIZE_Msk<<USB_DEVICE_PCKSIZE_MULTI_PACKET_SIZE_Pos)|
			(USB_DEVICE_PCKSIZE_BYTE_COUNT_Msk<<USB_DEVICE_PCKSIZE_BYTE_COUNT_Pos), 0)

}

func (d *dhw) setDeviceAddress(addr uint16) {

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
	// Don't use zero, since that is usually the actual control endpoint, and we
	// want to indicate that no device class configuration has been defined - as
	// this would be a bad situation if the user is requesting the control EP.
	//
	// Also, we don't necessarily want to check for this condition all the time
	// and may choose to use a call to this function as argument, so I've chosen
	// to not add a second bool/error return value - placing the error condition
	// in-band with the return value is safe because no device supports a control
	// endpoint number of 255 (the direction bit would be masked).
	return 0xFF
}

func (d *dhw) controlBuffer() uintptr {
	switch d.cc.id {
	case classDeviceCDCACM:
		return uintptr(unsafe.Pointer(descCDCACM[d.cc.config-1].cx))
	case classDeviceHID:
		return uintptr(unsafe.Pointer(descHID[d.cc.config-1].cx))
	}
	return 0
}

func (d *dhw) controlReset() {

	// Configure control endpoint 0
	endpoint := d.controlEndpoint()
	out, in := d.endpointAddressDescriptor(endpoint)

	if enum, ok := endpointSizeEncode(descControlPacketSize); ok {

		// Conigure packet size for control endpoints.
		out.packetSize.ReplaceBits(enum,
			USB_DEVICE_PCKSIZE_SIZE_Msk, USB_DEVICE_PCKSIZE_SIZE_Pos)
		in.packetSize.ReplaceBits(enum,
			USB_DEVICE_PCKSIZE_SIZE_Msk, USB_DEVICE_PCKSIZE_SIZE_Pos)

		// Configure bank 0 as control SETUP/OUT, bank 1 as control IN.
		d.endpointRegister(endpoint, epRegConfig).Set(
			(0x1 << sam.USB_DEVICE_ENDPOINT_EPCFG_EPTYPE0_Pos) |
				(0x1 << sam.USB_DEVICE_ENDPOINT_EPCFG_EPTYPE1_Pos))
		// Enable transfer complete and SETUP received interrupts
		d.endpointRegister(endpoint, epRegIntEnSet).Set(
			sam.USB_DEVICE_ENDPOINT_EPINTENSET_TRCPT0 |
				sam.USB_DEVICE_ENDPOINT_EPINTENSET_TRCPT1 |
				sam.USB_DEVICE_ENDPOINT_EPINTENSET_RXSTP)

		// Prepare to start processing SETUP packets
		d.prepareSetup()
	}
}

// controlStall stalls a transfer on control endpoint 0. To stall a transfer on
// any other endpoint, use method endpointStall().
func (d *dhw) controlStall() {
	d.endpointStall(d.controlEndpoint())
}

// controlReceive receives (Rx, OUT) data on control endpoint 0.
func (d *dhw) controlReceive(
	data uintptr, size uint32, notify bool) {

}

// controlTransmit transmits (Tx, IN) data on control endpoint 0.
func (d *dhw) controlTransmit(
	data uintptr, size uint32, notify bool) {

}

// =============================================================================
//  Endpoint Descriptor
// =============================================================================

// dhwEndptDesc defines a USB endpoint descriptor, used to inform the USB DMA
// controller the location of each endpoint transfer buffer.
type dhwEndptDesc struct {
	address    volatile.Register32
	packetSize volatile.Register32
	extToken   volatile.Register16
	bankStatus volatile.Register8
	_          [5]uint8
}

// dhwEndptAddrDesc defines an endpoint address descriptor, representing both
// directions (IN + OUT) of a given endpoint descriptor.
type dhwEndptAddrDesc [2]dhwEndptDesc

// dhwEndptDescSize defines the size (bytes) of a structure containing a USB
// endpoint descriptor.
const dhwEndptDescSize = unsafe.Sizeof(dhwEndptDesc{}) // 16 bytes

// Constants defining bitfields in the endpoint descriptor hardware register
// PCKSIZE. These were left out of the SVD for some reason.
const (
	USB_DEVICE_PCKSIZE_BYTE_COUNT_Pos = 0
	USB_DEVICE_PCKSIZE_BYTE_COUNT_Msk = 0x3FFF

	USB_DEVICE_PCKSIZE_SIZE_Pos = 28
	USB_DEVICE_PCKSIZE_SIZE_Msk = 0x7

	USB_DEVICE_PCKSIZE_MULTI_PACKET_SIZE_Pos = 14
	USB_DEVICE_PCKSIZE_MULTI_PACKET_SIZE_Msk = 0x3FFF
)

func pcksize(byteCount, size, multiPacketSize uint32) uint32 {
	return ((byteCount & USB_DEVICE_PCKSIZE_BYTE_COUNT_Msk) << USB_DEVICE_PCKSIZE_BYTE_COUNT_Pos) |
		((size & USB_DEVICE_PCKSIZE_SIZE_Msk) << USB_DEVICE_PCKSIZE_SIZE_Pos) |
		((multiPacketSize & USB_DEVICE_PCKSIZE_MULTI_PACKET_SIZE_Msk) << USB_DEVICE_PCKSIZE_MULTI_PACKET_SIZE_Pos)
}

var (
	// endpointSizeEnum is a constant-time lookup table for translating packet
	// sizes (bytes) to the corresponding register PCKSIZE.SIZE enumerated value.
	endpointSizeEnum = map[uint32]uint32{
		8: 0, 16: 1, 32: 2, 64: 3, 128: 4, 256: 5, 512: 6, 1023: 7,
	}

	// endpointEnumSize is a constant-time lookup table for translating the
	// register PCKSIZE.SIZE enumerated values to its packet size (bytes).
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
func endpointSizeEncode(size uint32) (enum uint32, ok bool) {
	enum, ok = endpointSizeEnum[size]
	return
}

// endpointSizeDecode returns the endpoint descriptor packet size (bytes) for a
// given register PCKSIZE.SIZE enumerated value.
//
// See documentation on endpoint descriptor bank SRAM register PCKSIZE, bit
// field SIZE for details.
func endpointSizeDecode(enum uint32) (size uint32, ok bool) {
	if ok = int(enum) < len(endpointEnumSize); ok {
		size = endpointEnumSize[enum]
	}
	return
}

// endpointAddressDescriptor returns the IN+OUT endpoint descriptors for the
// given endpoint address, encoded as direction D and endpoint number N with the
// 8-bit mask DxxxNNNN. The direction bit D is ignored.
//go:inline
func (d *dhw) endpointAddressDescriptor(endpoint uint8) (out, in *dhwEndptDesc) {
	// endpoint descriptor is device class-specific
	num, _ := unpackEndpoint(endpoint)
	switch d.cc.id {
	case classDeviceCDCACM:
		return &descCDCACM[d.cc.config-1].ed[num][descBankOut],
			&descCDCACM[d.cc.config-1].ed[num][descBankIn]
	case classDeviceHID:
		return &descHID[d.cc.config-1].ed[num][descBankOut],
			&descHID[d.cc.config-1].ed[num][descBankIn]
	default:
		return nil, nil
	}
}

// endpointDescriptor returns the endpoint descriptor for the given endpoint
// address, encoded as direction D and endpoint number N with the 8-bit mask
// DxxxNNNN.
//go:inline
func (d *dhw) endpointDescriptor(endpoint uint8) *dhwEndptDesc {
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

type dhwEndpointRegister int

const (
	epRegNone dhwEndpointRegister = iota
	epRegConfig
	epRegStatusClr
	epRegStatusSet
	epRegStatus
	epRegIntFlag
	epRegIntEnClr
	epRegIntEnSet
)

func (d *dhw) endpointRegister(endpoint uint8,
	register dhwEndpointRegister) *volatile.Register8 {

	if num, _ := unpackEndpoint(endpoint); num < descMaxEndpoints {

		switch register {
		case epRegConfig:
			return &d.bus.DEVICE_ENDPOINT[num].EPCFG

		case epRegStatusClr:
			return &d.bus.DEVICE_ENDPOINT[num].EPSTATUSCLR

		case epRegStatusSet:
			return &d.bus.DEVICE_ENDPOINT[num].EPSTATUSSET

		case epRegStatus:
			return &d.bus.DEVICE_ENDPOINT[num].EPSTATUS

		case epRegIntFlag:
			return &d.bus.DEVICE_ENDPOINT[num].EPINTFLAG

		case epRegIntEnClr:
			return &d.bus.DEVICE_ENDPOINT[num].EPINTENCLR

		case epRegIntEnSet:
			return &d.bus.DEVICE_ENDPOINT[num].EPINTENSET
		}
	}
	return nil
}

func (d *dhw) endpointEnable(endpoint uint8, control bool, config uint32) {

}

func (d *dhw) endpointStatus(endpoint uint8) uint16 {
	status := uint16(0)
	switch endpoint {
	case rxEndpoint(endpoint):
	case txEndpoint(endpoint):
	}
	return status
}

// endpointStall stalls a transfer on the given endpoint.
func (d *dhw) endpointStall(endpoint uint8) {

}

func (d *dhw) endpointClearFeature(endpoint uint8) {

}

func (d *dhw) endpointSetFeature(endpoint uint8) {

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

}

// endpointComplete handles transfer completion of a data endpoint.
func (d *dhw) endpointComplete(endpoint uint8) {

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

func (d *dhw) transferPrepare(
	transfer *dhwTransfer, data *uint8, size uint16, param uint32) {

}

func (d *dhw) transferSchedule(
	endpoint *dhwEndptDesc, mask uint32, transfer *dhwTransfer) {

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

	// SAMx51 only supports USB full-speed (FS) operation
	acm.rxSize = descCDCACMDataRxFSPacketSize
	acm.txSize = descCDCACMDataTxFSPacketSize

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
}

func (d *dhw) uartSetLineCoding(coding descCDCACMLineCoding) {
	if 134 == coding.baud {
		d.enableSOF(true, descCDCACMInterfaceCount)
	}
}

func (d *dhw) uartReceive(endpoint uint8) {
	acm := &descCDCACM[d.cc.config-1]
	num := uint16(endpoint) & descEndptAddrNumberMsk
}

func (d *dhw) uartNotify(transfer *dhwTransfer) {
	acm := &descCDCACM[d.cc.config-1]

}

// uartFlush discards all buffered input (Rx) data.
func (d *dhw) uartFlush() {
	acm := &descCDCACM[d.cc.config-1]
}

func (d *dhw) uartAvailable() int {
	return 0
}

func (d *dhw) uartPeek() (uint8, bool) {
	acm := &descCDCACM[d.cc.config-1]
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
	return int(read)
}

func (d *dhw) uartWriteByte(c uint8) bool {
	return 1 == d.uartWrite([]uint8{c})
}

func (d *dhw) uartWrite(data []uint8) int {
	acm := &descCDCACM[d.cc.config-1]
	sent := 0
	size := len(data)
	return sent
}

func (d *dhw) uartSync() {

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
}

func (d *dhw) serialTransmit() {
	hid := &descHID[d.cc.config-1]
}

func (d *dhw) serialNotify(transfer *dhwTransfer) {
	hid := &descHID[d.cc.config-1]
	len := hid.rxSerialSize - (uint16(transfer.token>>16) & 0x7FFF)
}

// serialFlush discards all buffered input (Rx) data.
func (d *dhw) serialFlush() {
	hid := &descHID[d.cc.config-1]
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

	// SAMx51 only supports USB full-speed (FS) operation
	hid.txMouseSize = descHIDMouseTxFSPacketSize

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

	// SAMx51 only supports USB full-speed (FS) operation
	hid.txJoystickSize = descHIDJoystickTxFSPacketSize

	d.endpointEnable(descHIDEndpointJoystick,
		false, descHIDConfigAttrJoystick)

	d.endpointConfigureTx(descHIDEndpointJoystick,
		hid.txJoystickSize, false, nil)
}
