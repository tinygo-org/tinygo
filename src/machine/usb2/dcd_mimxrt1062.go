// +build mimxrt1062

package usb2

// Implementation of USB device controller interface (dcd) for NXP iMXRT1062.

import (
	"device/arm"
	"device/nxp"
	"math/bits"
	"runtime/interrupt"
	"runtime/volatile"
	"unsafe"
)

// dcdCount defines the number of USB cores to configure for device mode. It is
// computed as the sum of all declared device configuration descriptors.
const dcdCount = descCDCACMCount

// dcdInterruptPriority defines the priority for all USB device interrupts.
const dcdInterruptPriority = 3

// deviceController implements USB device controller driver (dcd) interface.
type deviceController struct {
	core  *core // Parent USB core this instance is attached to
	port  int   // USB port index
	class class // USB device class
	id    int   // deviceControllerInstance index

	bus *nxp.USB_Type
	phy *nxp.USBPHY_Type
	irq interrupt.Interrupt

	cri volatile.Register8 // set to 1 if in critical section, else 0
	ivm uintptr            // interrupt state when entering critical section

	stat *dcdEndpoint // endpoint 0 Rx ("out" direction)
	ctrl *dcdEndpoint // endpoint 0 Tx ("in" direction)

	acm *descCDCACMClass

	timerInterrupt [2]func()
	controlNotify  uint32
	endpointNotify uint32
	sofUsage       uint8
	rebootTimer    uint8
	setup          dcdSetup
	controlReply   [8]uint8

	speed uint8 // bus speed (0=full, 1=low, 2=high, 4=super)
}

// deviceControllerInstance provides statically-allocated instances of each USB
// device controller configured on this platform.
var deviceControllerInstance [dcdCount]deviceController

// cycleCount uses the ARM debug cycle counter available on iMXRT1062 (enabled
// in runtime_mimxrt1062_time.go) to return the number of CPU cycles since boot.
//go:inline
func cycleCount() uint32 {
	return (*volatile.Register32)(unsafe.Pointer(uintptr(0xe0001004))).Get()
}

//go:linkname ticks runtime.ticks
func ticks() int64

// initDCD initializes and assigns a free device controller instance to the
// given USB port. Returns the initialized device controller or nil if no free
// device controller instances remain.
func initDCD(port int, class class) (dcd, status) {
	if 0 == dcdCount {
		return nil, statusInvalid // must have defined device controllers
	}
	switch class.id {
	case classDeviceCDCACM:
		if 0 == class.config || class.config > descCDCACMCount {
			return nil, statusInvalid // must have defined descriptors
		}
	default:
	}
	// Return the first instance whose assigned core is currently nil.
	for i := range deviceControllerInstance {
		if nil == deviceControllerInstance[i].core {
			// Initialize device controller.
			deviceControllerInstance[i].core = &coreInstance[port]
			deviceControllerInstance[i].port = port
			deviceControllerInstance[i].class = class
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
			switch class.id {
			case classDeviceCDCACM:
				deviceControllerInstance[i].acm = &descCDCACM[class.config-1]
			default:
			}
			return &deviceControllerInstance[i], statusOK
		}
	}
	return nil, statusBusy // No free device controller instances available.
}

func (dc *deviceController) init() status {
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

	dc.stat = dc.endpointQueueHead(rxEndpoint(0))
	dc.ctrl = dc.endpointQueueHead(txEndpoint(0))
	dc.stat.config = (descEndptMaxPktSize << 16) | (1 << 15)
	dc.ctrl.config = (descEndptMaxPktSize << 16)

	dc.bus.ASYNCLISTADDR.Set(uint32(uintptr(unsafe.Pointer(dc.stat))))

	// clear installed timer callbacks
	dc.timerInterrupt[0] = nil
	dc.timerInterrupt[1] = nil

	// enable interrupts
	dc.bus.USBINTR.Set(
		nxp.USB_USBINTR_UE_Msk | // bus enable
			nxp.USB_USBINTR_UEE_Msk | // bus error
			nxp.USB_USBINTR_PCE_Msk | // port change detect
			nxp.USB_USBINTR_URE_Msk | // bus reset
			nxp.USB_USBINTR_SLE) // sleep enable

	// ensure D+ pulled down long enough for host to detect previous disconnect
	udelay(5000)

	return statusOK
}

func (dc *deviceController) enable(enable bool) status {
	if enable {
		dc.irq.SetPriority(dcdInterruptPriority)
		dc.irq.Enable()
		dc.bus.USBCMD.SetBits(nxp.USB_USBCMD_RS)
	} else {
		dc.bus.USBCMD.ClearBits(nxp.USB_USBCMD_RS)
		dc.irq.Disable()
	}
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

	// USB Interrupt (USBINT) - R/WC
	// This bit is set by the Host/Device Controller when the cause of an
	// interrupt is a completion of a USB transaction where the Transfer
	// Descriptor (TD) has an interrupt on complete (IOC) bit set.
	// This bit is also set by the Host/Device Controller when a short packet is
	// detected. A short packet is when the actual number of bytes received was
	// less than the expected number of bytes.
	if 0 != status&nxp.USB_USBSTS_UI {

		setupStatus := dc.bus.ENDPTSETUPSTAT.Get()
		for 0 != setupStatus {
			dc.bus.ENDPTSETUPSTAT.Set(setupStatus)
			var setup dcdSetup
			ready := false
			for !ready {
				dc.bus.USBCMD.SetBits(nxp.USB_USBCMD_SUTW)
				setup = dc.stat.setup
				ready = dc.bus.USBCMD.HasBits(nxp.USB_USBCMD_SUTW)
			}
			dc.bus.USBCMD.ClearBits(nxp.USB_USBCMD_SUTW)
			// flush endpoint 0 (bit 0=Rx, 16=Tx)
			dc.bus.ENDPTFLUSH.Set(0x00010001)
			for dc.bus.ENDPTFLUSH.HasBits(0x00010001) {
			} // wait for flush to complete
			dc.controlNotify = 0
			dc.control(setup)
			setupStatus = dc.bus.ENDPTSETUPSTAT.Get()
		}

		completeStatus := dc.bus.ENDPTCOMPLETE.Get()
		if 0 != completeStatus {
			dc.bus.ENDPTCOMPLETE.Set(completeStatus)
			if 0 != completeStatus&dc.controlNotify {
				dc.controlNotify = 0
				dc.controlComplete()
			}
			completeStatus &= dc.endpointNotify
			if 0 != completeStatus {
				tx := completeStatus >> 16
				for 0 != tx {
					num := uint8(bits.TrailingZeros32(tx))
					dc.endpointComplete(txEndpoint(num))
					tx &^= 1 << num
				}
				rx := completeStatus & 0xFFFF
				for 0 != rx {
					num := uint8(bits.TrailingZeros32(rx))
					dc.endpointComplete(rxEndpoint(num))
					rx &^= 1 << num
				}
			}
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
		dc.bus.ENDPTSETUPSTAT.Set(dc.bus.ENDPTSETUPSTAT.Get())
		// clear all endpoint complete status
		dc.bus.ENDPTCOMPLETE.Set(dc.bus.ENDPTCOMPLETE.Get())
		// wait on any endpoint priming
		for 0 != dc.bus.ENDPTPRIME.Get() {
		}
		dc.bus.ENDPTFLUSH.Set(0xFFFFFFFF)
		// if dc.bus.PORTSC1.HasBits(nxp.USB_PORTSC1_PR) {
		// }
		switch dc.class.id {
		case classDeviceCDCACM:
			// TBD: reset CDC-ACM UART?
		default:
		}
		dc.endpointNotify = 0
	}

	// General Purpose Timer Interrupt 0(GPTINT0) - R/WC
	// This bit is set when the counter in the GPTIMER0CTRL register transitions
	// to zero, writing a one to this bit clears it.
	if 0 != status&nxp.USB_USBSTS_TI0 {
		if nil != dc.timerInterrupt[0] {
			dc.timerInterrupt[0]()
		}
	}

	// General Purpose Timer Interrupt 1(GPTINT1) - R/WC
	// This bit is set when the counter in the GPTIMER1CTRL register transitions
	// to zero, writing a one to this bit will clear it.
	if 0 != status&nxp.USB_USBSTS_TI1 {
		if nil != dc.timerInterrupt[1] {
			dc.timerInterrupt[1]()
		}
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
		if dc.bus.PORTSC1.HasBits(nxp.USB_PORTSC1_HSP) {
			dc.speed = descDeviceSpeedHigh // 480 Mbit/sec
		} else {
			dc.speed = descDeviceSpeedFull // 12 Mbit/sec
		}
	}

	// DCSuspend - R/WC
	// When a controller enters a suspend state from an active state, this bit
	// will be set to a one. The device controller clears the bit upon exiting
	// from a suspend state. Only used in device operation mode.
	if 0 != status&nxp.USB_USBSTS_SLI {
		//println("suspend")
	}

	// USB Error Interrupt (USBERRINT) - R/WC
	// When completion of a USB transaction results in an error condition, this
	// bit is set by the Host/Device Controller. This bit is set along with the
	// USBINT bit, if the TD on which the error interrupt occurred also had its
	// interrupt on complete (IOC) bit set.
	// The device controller detects resume signaling only.
	if 0 != status&nxp.USB_USBSTS_UEI {
		//println("error")
	}

	// SOF Received - R/WC
	// When the device controller detects a Start Of (micro) Frame, this bit will
	// be set to a one. When a SOF is extremely late, the device controller will
	// automatically set this bit to indicate that an SOF was expected.
	// Therefore, this bit will be set roughly every 1ms in device FS mode and
	// every 125ms in HS mode and will be synchronized to the actual SOF that is
	// received.
	// Because the device controller is initialized to FS before connect, this bit
	// will be set at an interval of 1ms during the prelude to connect and chirp.
	// In host mode, this bit will be set every 125us and can be used by host
	// controller driver as a time base. Software writes a 1 to this bit to clear
	// it.
	if dc.bus.USBINTR.HasBits(nxp.USB_USBINTR_SRE) &&
		0 != status&nxp.USB_USBSTS_SRI {
		if 0 != dc.rebootTimer {
			dc.rebootTimer -= 1
			if 0 == dc.rebootTimer {
				dc.enableSofInterrupts(false, descCDCACMInterfaceCount)
			}
		}
	}
}

func (dc *deviceController) enableSofInterrupts(enable bool, iface uint8) {
	if enable {
		ivm := arm.DisableInterrupts()
		dc.sofUsage |= 1 << iface
		if !dc.bus.USBINTR.HasBits(nxp.USB_USBINTR_SRE) {
			dc.bus.USBSTS.Set(nxp.USB_USBSTS_SRI)
			dc.bus.USBINTR.SetBits(nxp.USB_USBINTR_SRE)
		}
		arm.EnableInterrupts(ivm)
	} else {
		dc.sofUsage &^= 1 << iface
		if 0 == dc.sofUsage {
			dc.bus.USBINTR.ClearBits(nxp.USB_USBINTR_SRE)
		}
	}
}

// receive schedules a receive (Rx, OUT) transfer on the given endpoint.
func (dc *deviceController) receive(endpoint uint8, transfer *dcdTransfer) {
	if endpoint < descCDCACMEndpointStatus ||
		endpoint > descCDCACMEndpointCount {
		return
	}
	ep := dc.endpointQueueHead(rxEndpoint(endpoint))
	em := (uint32(1) << endpoint) << descCDCACMConfigAttrRxPos
	dc.transferSchedule(ep, em, transfer)
}

// transmit schedules a transmit (Tx, IN) transfer on the given endpoint.
func (dc *deviceController) transmit(endpoint uint8, transfer *dcdTransfer) {
	if endpoint < descCDCACMEndpointStatus ||
		endpoint > descCDCACMEndpointCount {
		return
	}
	ep := dc.endpointQueueHead(txEndpoint(endpoint))
	em := (uint32(1) << endpoint) << descCDCACMConfigAttrTxPos
	dc.transferSchedule(ep, em, transfer)
}

// control handles setup messages on control endpoint 0.
func (dc *deviceController) control(setup dcdSetup) {

	// println(strconv.FormatUint(setup.pack(), 16))

	// First, switch on the type of request (standard, class, or vendor)
	switch setup.bmRequestType & descRequestTypeTypeMsk {

	// === STANDARD REQUEST ===
	case descRequestTypeTypeStandard:

		// Switch on the recepient and direction of the request
		switch setup.bmRequestType &
			(descRequestTypeRecipientMsk | descRequestTypeDirMsk) {

		// --- DEVICE Rx (OUT) ---
		case descRequestTypeRecipientDevice | descRequestTypeDirOut:

			// Identify which request was received
			switch setup.bRequest {

			// SET ADDRESS (0x05):
			case descRequestStandardSetAddress:
				dc.controlReceive(dcdPointerNil, 0, false)
				dc.bus.DEVICEADDR.Set(nxp.USB_DEVICEADDR_USBADRA |
					((uint32(setup.wValue) << nxp.USB_DEVICEADDR_USBADR_Pos) &
						nxp.USB_DEVICEADDR_USBADR_Msk))
				return

			// SET CONFIGURATION (0x09):
			case descRequestStandardSetConfiguration:
				dc.class.config = int(setup.wValue)
				if 0 == dc.class.config || dc.class.config > dcdCount {
					// Use default if invalid index received
					dc.class.config = 1
				}

				// Respond based on our device class configuration
				switch dc.class.id {

				// CDC-ACM (single)
				case classDeviceCDCACM:
					dc.bus.ENDPTCTRL2.Set(descCDCACMConfigAttrStatus) // Status Tx
					dc.bus.ENDPTCTRL3.Set(descCDCACMConfigAttrDataRx) // Bulk data Rx
					dc.bus.ENDPTCTRL4.Set(descCDCACMConfigAttrDataTx) // Bulk data Tx
					dc.serialConfigure()
					dc.controlReceive(dcdPointerNil, 0, false)

				default:
					// Unhandled device class
				}
				return

			default:
				// Unhandled request
			}

		// --- DEVICE Tx (IN) ---
		case descRequestTypeRecipientDevice | descRequestTypeDirIn:

			// Identify which request was received
			switch setup.bRequest {

			// GET STATUS (0x00):
			case descRequestStandardGetStatus:
				dc.controlReply[0] = 0
				dc.controlReply[1] = 0
				dc.controlTransmit(
					uintptr(unsafe.Pointer(&dc.controlReply[0])), 2, false)
				return

			// GET DESCRIPTOR (0x06):
			case descRequestStandardGetDescriptor:
				dc.controlDescriptor(setup)
				return

			// GET CONFIGURATION (0x08):
			case descRequestStandardGetConfiguration:
				dc.controlReply[0] = uint8(dc.class.config)
				dc.controlTransmit(
					uintptr(unsafe.Pointer(&dc.controlReply[0])), 1, false)
				return

			default:
				// Unhandled request
			}

		// --- INTERFACE Tx (IN) ---
		case descRequestTypeRecipientInterface | descRequestTypeDirIn:

			// Identify which request was received
			switch setup.bRequest {

			// GET DESCRIPTOR (0x06):
			case descRequestStandardGetDescriptor:
				dc.controlDescriptor(setup)
				return

			default:
				// Unhandled request
			}

		// --- ENDPOINT Rx (OUT) ---
		case descRequestTypeRecipientEndpoint | descRequestTypeDirOut:

			// Identify which request was received
			switch setup.bRequest {

			// CLEAR FEATURE (0x01):
			case descRequestStandardClearFeature:
				// TODO

			// SET FEATURE (0x03):
			case descRequestStandardSetFeature:
				// TODO

			default:
				// Unhandled request
			}

		// --- ENDPOINT Tx (IN) ---
		case descRequestTypeRecipientEndpoint | descRequestTypeDirIn:

			// Identify which request was received
			switch setup.bRequest {

			// GET STATUS (0x00):
			case descRequestStandardGetStatus:
				num, dir := unpackEndpoint(uint8(setup.wIndex))
				var reg *volatile.Register32
				switch num {
				case 0:
					reg = &dc.bus.ENDPTCTRL0
				case 1:
					reg = &dc.bus.ENDPTCTRL1
				case 2:
					reg = &dc.bus.ENDPTCTRL2
				case 3:
					reg = &dc.bus.ENDPTCTRL3
				case 4:
					reg = &dc.bus.ENDPTCTRL4
				case 5:
					reg = &dc.bus.ENDPTCTRL5
				case 6:
					reg = &dc.bus.ENDPTCTRL6
				case 7:
					reg = &dc.bus.ENDPTCTRL7
				}
				if nil != reg {
					dc.controlReply[0] = 0
					dc.controlReply[1] = 0
					if ((0 != dir) && reg.HasBits(nxp.USB_ENDPTCTRL0_TXS)) ||
						((0 == dir) && reg.HasBits(nxp.USB_ENDPTCTRL0_RXS)) {
						dc.controlReply[0] = 1
					}
					dc.controlTransmit(
						uintptr(unsafe.Pointer(&dc.controlReply[0])), 2, false)
					return
				}

			default:
				// Unhandled request
			}

		default:
			// Unhandled request recepient or direction
		}

	// === CLASS REQUEST ===
	case descRequestTypeTypeClass:

		// Switch on the recepient and direction of the request
		switch setup.bmRequestType &
			(descRequestTypeRecipientMsk | descRequestTypeDirMsk) {

		// --- INTERFACE Rx (OUT) ---
		case descRequestTypeRecipientInterface | descRequestTypeDirOut:

			// Identify which request was received
			switch setup.bRequest {

			// CDC | SET LINE CODING (0x20):
			case descCDCRequestSetLineCoding:

				// Respond based on our device class configuration
				switch dc.class.id {

				// CDC-ACM (single)
				case classDeviceCDCACM:
					// line coding must contain exactly 7 bytes
					if descCDCACMCodingSize == setup.wLength {
						dc.setup = setup
						dc.controlReceive(
							uintptr(unsafe.Pointer(&descCDCACM[dc.class.config-1].cx[0])),
							descCDCACMCodingSize, true)
						return
					}

				default:
					// Unhandled device class
				}

			// CDC | SET CONTROL LINE STATE (0x22):
			case descCDCRequestSetControlLineState:

				// Respond based on our device class configuration
				switch dc.class.id {

				// CDC-ACM (single)
				case classDeviceCDCACM:

					// Determine interface destination of the notification
					switch setup.wIndex {

					// Control/status interface:
					case descCDCACMInterfaceCtrl:
						acm := &descCDCACM[dc.class.config-1]
						acm.cticks = ticks()
						acm.rtsdtr = uint8(setup.wValue)
						dc.controlReceive(dcdPointerNil, 0, false)
						return

					default:
						// Unhandled device interface
					}

				default:
					// Unhandled device class
				}

			// CDC | SEND BREAK (0x23):
			case descCDCRequestSendBreak:

				// Respond based on our device class configuration
				switch dc.class.id {

				// CDC-ACM (single)
				case classDeviceCDCACM:
					dc.controlReceive(dcdPointerNil, 0, false)
					return

				default:
					// Unhandled device class
				}

			default:
				// Unhandled request
			}

		default:
			// Unhandled request recepient or direction
		}

	case descRequestTypeTypeVendor:
	default:
		// Unhandled request type
	}

	dc.bus.ENDPTCTRL0.Set(0x00010001)
}

// controlTransfers returns the data and ackowledgement transfer descriptors for
// the control endpoint (i.e., endpoint 0).
//go:inline
func (dc *deviceController) controlTransfers() (dat, ack *dcdTransfer) {
	// control endpoint is device class-specific
	switch dc.class.id {
	case classDeviceCDCACM:
		return descCDCACM[dc.class.config-1].cd, descCDCACM[dc.class.config-1].ad
	default:
		return nil, nil
	}
}

func (dc *deviceController) controlDescriptor(setup dcdSetup) {

	// Respond based on our device class configuration
	switch dc.class.id {

	// CDC-ACM (single)
	case classDeviceCDCACM:
		acm := &descCDCACM[dc.class.config-1]
		dxn := uint8(0)

		// Determine the type of descriptor being requested
		switch setup.wValue >> 8 {

		// Device descriptor
		case descTypeDevice:
			dxn = descLengthDevice
			_ = copy(acm.dx[:], acm.device[:dxn])

		// Configuration descriptor
		case descTypeConfigure:
			dxn = uint8(descCDCACMConfigSize)
			_ = copy(acm.dx[:], acm.config[:dxn])

		// String descriptor
		case descTypeString:

			var s []uint8

			// Determine the string index requested
			switch uint8(setup.wValue) {

			// Language
			case 0:
				if int(setup.wIndex) < len(acm.locstr) {
					s = acm.locstr[setup.wIndex].index[0][:]
				}

			// Manufacturer
			case 1:
				for i := range acm.locstr {
					if acm.locstr[i].language == setup.wIndex {
						s = acm.locstr[i].index[1][:]
						// copy manufacturer string to uint8 buffer as UTF-16
						for n, c := range descManufacturer {
							s[2+2*n] = uint8(c)
							s[3+2*n] = 0
						}
						break
					}
				}

			// Product
			case 2:
				for i := range acm.locstr {
					if acm.locstr[i].language == setup.wIndex {
						s = acm.locstr[i].index[2][:]
						// copy product string to uint8 buffer as UTF-16
						for n, c := range descProduct {
							s[2+2*n] = uint8(c)
							s[3+2*n] = 0
						}
						break
					}
				}

			// Serial number
			case 3:
				for i := range acm.locstr {
					if acm.locstr[i].language == setup.wIndex {
						s = acm.locstr[i].index[3][:]
						// copy serial number string to uint8 buffer as UTF-16
						for n, c := range descSerialNumber {
							s[2+2*n] = uint8(c)
							s[3+2*n] = 0
						}
						break
					}
				}
			}

			if nil != s && len(s) > 0 {
				dxn = s[0]
				_ = copy(acm.dx[:], s[:dxn])
			}

		// Device qualification descriptor
		case descTypeQualification:
			dxn = descLengthQualification
			_ = copy(acm.dx[:], acm.qualif[:dxn])

		// Alternate configuration descriptor
		case descTypeOtherSpeedConfiguration:

		}

		if dxn > 0 {
			if dxn > uint8(setup.wLength) {
				dxn = uint8(setup.wLength)
			}
			nxp.FlushDeleteDcache(
				uintptr(unsafe.Pointer(&acm.dx[0])), uintptr(dxn))
			dc.controlTransmit(
				uintptr(unsafe.Pointer(&acm.dx[0])), uint32(dxn), false)
		}

	default:
		// Unhandled device class
	}

}

// controlReceive receives (Rx, OUT) data on control endpoint 0.
func (dc *deviceController) controlReceive(
	data uintptr, size uint32, notify bool) {
	const (
		rm = uint32(1 << descCDCACMConfigAttrRxPos)
		tm = uint32(1 << descCDCACMConfigAttrTxPos)
	)
	cd, ad := dc.controlTransfers()
	if size > 0 {
		cd.next = dcdTransferEOL
		cd.token = (size << 16) | (1 << 7)
		for i := range cd.pointer {
			cd.pointer[i] = data + uintptr(i)*4096
		}
		// linked list is empty
		qr := dc.endpointQueueHead(rxEndpoint(0))
		qr.transfer.next = cd
		qr.transfer.token = 0
		dc.bus.ENDPTPRIME.SetBits(rm)
		for 0 != dc.bus.ENDPTPRIME.Get() {
		} // wait for endpoint finish priming
	}
	ad.next = dcdTransferEOL
	ad.token = 1 << 7
	if notify {
		ad.token |= 1 << 15
	}
	ad.pointer[0] = 0
	qt := dc.endpointQueueHead(txEndpoint(0))
	qt.transfer.next = ad
	qt.transfer.token = 0
	dc.bus.ENDPTCOMPLETE.Set(rm | tm)
	dc.bus.ENDPTPRIME.SetBits(tm)
	if notify {
		dc.controlNotify = tm
	}
}

// controlTransmit transmits (Tx, IN) data on control endpoint 0.
func (dc *deviceController) controlTransmit(
	data uintptr, size uint32, notify bool) {
	const (
		rm = uint32(1 << descCDCACMConfigAttrRxPos)
		tm = uint32(1 << descCDCACMConfigAttrTxPos)
	)
	cd, ad := dc.controlTransfers()
	if size > 0 {
		cd.next = dcdTransferEOL
		cd.token = (size << 16) | (1 << 7)
		for i := range cd.pointer {
			cd.pointer[i] = data + uintptr(i)*4096
		}
		// linked list is empty
		qt := dc.endpointQueueHead(txEndpoint(0))
		qt.transfer.next = cd
		qt.transfer.token = 0
		dc.bus.ENDPTPRIME.SetBits(tm)
		for 0 != dc.bus.ENDPTPRIME.Get() {
		} // wait for endpoint finish priming
	}
	ad.next = dcdTransferEOL
	ad.token = 1 << 7
	if notify {
		ad.token |= 1 << 15
	}
	ad.pointer[0] = 0
	qr := dc.endpointQueueHead(rxEndpoint(0))
	qr.transfer.next = ad
	qr.transfer.token = 0
	dc.bus.ENDPTCOMPLETE.Set(rm | tm)
	dc.bus.ENDPTPRIME.SetBits(rm)
	if notify {
		dc.controlNotify = rm
	}
}

// controlComplete handles the setup completion of control endpoint 0.
func (dc *deviceController) controlComplete() {

	// First, switch on the type of request (standard, class, or vendor)
	switch dc.setup.bmRequestType & descRequestTypeTypeMsk {

	// === CLASS REQUEST ===
	case descRequestTypeTypeClass:

		// Switch on the recepient and direction of the request
		switch dc.setup.bmRequestType &
			(descRequestTypeRecipientMsk | descRequestTypeDirMsk) {

		// --- INTERFACE Rx (OUT) ---
		case descRequestTypeRecipientInterface | descRequestTypeDirOut:

			// Identify which request was received
			switch dc.setup.bRequest {

			// CDC | SET LINE CODING (0x20):
			case descCDCRequestSetLineCoding:

				// Respond based on our device class configuration
				switch dc.class.id {

				// CDC-ACM (single)
				case classDeviceCDCACM:

					// Determine interface destination of the notification
					switch dc.setup.wIndex {

					// Control/status interface:
					case descCDCACMInterfaceCtrl:

						_ = copy(descCDCACM[dc.class.config-1].coding[:],
							// descCDCACM[dc.class.config-1].costat[:descCDCACMCodingSize])
							descCDCACM[dc.class.config-1].cx[:])
						var coding descCDCACMLineCoding
						if coding.parse(descCDCACM[dc.class.config-1].coding[:]) {
							if 134 == coding.baud {
								dc.enableSofInterrupts(true, descCDCACMInterfaceCount)
								dc.rebootTimer = 80
							}
						}

					default:
						// Unhandled device interface
					}

				default:
					// Unhandled device class
				}

			default:
				// Unhandled request
			}

		default:
			// Unhandled recepient or direction
		}

	default:
		// Unhandled request type
	}

	// // determine interface destination of the notification
	// switch dc.setup.wIndex {
	// // communication/control interface:
	// case descCDCACMInterfaceCtrl:
	// 	// switch on the type and recepient of the request
	// 	switch dc.setup.bmRequestType &
	// 		(descRequestTypeTypeMsk | descRequestTypeRecipientMsk) {
	// 	// interface class request:
	// 	case descRequestTypeRecipientInterface | descRequestTypeTypeClass:
	// 		// identify which request was received
	// 		switch dc.setup.bRequest {
	// 		// CDC_SET_LINE_CODING:
	// 		case descCDCRequestSetLineCoding:
	// 			// respond according to our device class
	// 			switch dc.class.id {
	// 			// CDC-ACM (single)
	// 			case classDeviceCDCACM:
	// 				_ = copy(descCDCACM[dc.class.config-1].coding[:],
	// 					// descCDCACM[dc.class.config-1].costat[:descCDCACMCodingSize])
	// 					descCDCACM[dc.class.config-1].cx[:])
	// 				var coding descCDCACMLineCoding
	// 				if coding.parse(descCDCACM[dc.class.config-1].coding[:]) {
	// 					if 134 == coding.baud {
	// 						dc.enableSofInterrupts(true, descCDCACMInterfaceCount)
	// 						dc.rebootTimer = 80
	// 					}
	// 				}
	// 			default:
	// 				// unhandled device class
	// 			}
	// 		default:
	// 			// unhandled request
	// 		}
	// 	default:
	// 		// unhandled request type or recepient
	// 	}
	// default:
	// 	// unhandled interface
	// }
}

// endpointQueueHead returns the queue head for the given endpoint address,
// encoded as direction D and endpoint number N with the 8-bit mask DxxxNNNN.
//go:inline
func (dc *deviceController) endpointQueueHead(endpoint uint8) *dcdEndpoint {
	// endpoint queue head is device class-specific
	switch dc.class.id {
	case classDeviceCDCACM:
		return &descCDCACM[dc.class.config-1].qh[endpointIndex(endpoint)]
	default:
		return nil
	}
}

func (dc *deviceController) endpointConfigure(
	ep *dcdEndpoint, packetSize uint16, zlp bool, callback dcdTransferCallback) {

	ep.config = uint32(packetSize) << 16
	if !zlp {
		ep.config |= 1 << 29
	}
	ep.current = nil
	ep.transfer.next = dcdTransferEOL
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

func (dc *deviceController) endpointConfigureRx(
	endpoint uint8, packetSize uint16, zlp bool, callback dcdTransferCallback) {

	if endpoint < descCDCACMEndpointStatus ||
		endpoint > descCDCACMEndpointCount {
		return
	}
	ep := dc.endpointQueueHead(rxEndpoint(endpoint))
	dc.endpointConfigure(ep, packetSize, zlp, callback)
	if nil != callback {
		dc.endpointNotify |= (uint32(1) << endpoint) << descCDCACMConfigAttrRxPos
	}
}

func (dc *deviceController) endpointConfigureTx(
	endpoint uint8, packetSize uint16, zlp bool, callback dcdTransferCallback) {

	if endpoint < descCDCACMEndpointStatus ||
		endpoint > descCDCACMEndpointCount {
		return
	}
	ep := dc.endpointQueueHead(txEndpoint(endpoint))
	dc.endpointConfigure(ep, packetSize, zlp, callback)
	if nil != callback {
		dc.endpointNotify |= (uint32(1) << endpoint) << descCDCACMConfigAttrTxPos
	}
}

// endpointComplete handles transfer completion of a data endpoint.
func (dc *deviceController) endpointComplete(endpoint uint8) {
	ep := dc.endpointQueueHead(endpoint)
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
			if 0 != t.token&(1<<7) {
				// active transfer, new list begins here
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

func (dc *deviceController) transferPrepare(
	transfer *dcdTransfer, data *uint8, size uint16, param uint32) {
	transfer.next = dcdTransferEOL
	transfer.token = (uint32(size) << 16) | (1 << 7)
	addr := uintptr(unsafe.Pointer(data))
	for i := range transfer.pointer {
		transfer.pointer[i] = addr + uintptr(i)*4096
	}
	transfer.param = param
}

func (dc *deviceController) transferSchedule(
	endpoint *dcdEndpoint, mask uint32, transfer *dcdTransfer) {

	if nil != endpoint.callback {
		transfer.token |= 1 << 15
	}
	ivm := arm.DisableInterrupts()
	last := endpoint.last
	if nil != last {
		last.next = transfer
		if dc.bus.ENDPTPRIME.HasBits(mask) {
			goto endTransfer
		}
		start := cycleCount()
		estat := uint32(0)
		for !dc.bus.USBCMD.HasBits(nxp.USB_USBCMD_ATDTW) &&
			(cycleCount()-start < 2400) {
			dc.bus.USBCMD.SetBits(nxp.USB_USBCMD_ATDTW)
			estat = dc.bus.ENDPTSTAT.Get()
		}
		if 0 != estat&mask {
			goto endTransfer
		}
	}
	endpoint.transfer.next = transfer
	endpoint.transfer.token = 0
	dc.bus.ENDPTPRIME.SetBits(mask)
	endpoint.first = transfer
endTransfer:
	endpoint.last = transfer
	arm.EnableInterrupts(ivm)
}

func (dc *deviceController) timerConfigure(timer int, usec uint32, fn func()) {
	if timer < 0 || timer >= len(dc.timerInterrupt) {
		return
	}
	dc.timerInterrupt[timer] = fn
	switch timer {
	case 0:
		dc.bus.GPTIMER0CTRL.Set(0)
		dc.bus.GPTIMER0LD.Set(usec - 1)
		dc.bus.USBINTR.SetBits(nxp.USB_USBINTR_TIE0)
	case 1:
		dc.bus.GPTIMER1CTRL.Set(0)
		dc.bus.GPTIMER1LD.Set(usec - 1)
		dc.bus.USBINTR.SetBits(nxp.USB_USBINTR_TIE1)
	}
}

func (dc *deviceController) timerOneShot(timer int) {
	switch timer {
	case 0:
		dc.bus.GPTIMER0CTRL.Set(
			nxp.USB_GPTIMER0CTRL_GPTRUN | nxp.USB_GPTIMER0CTRL_GPTRST)
	case 1:
		dc.bus.GPTIMER1CTRL.Set(
			nxp.USB_GPTIMER1CTRL_GPTRUN | nxp.USB_GPTIMER1CTRL_GPTRST)
	}
}

func (dc *deviceController) timerStop(timer int) {
	switch timer {
	case 0:
		dc.bus.GPTIMER0CTRL.Set(0)
	case 1:
		dc.bus.GPTIMER1CTRL.Set(0)
	}
}

func (dc *deviceController) serialConfigure() {
	acm := &descCDCACM[dc.class.config-1]
	switch dc.speed {
	case descDeviceSpeedHigh:
		acm.rxSize = descCDCACMDataRxHSPacketSize
		acm.txSize = descCDCACMDataTxHSPacketSize
	default:
		acm.rxSize = descCDCACMDataRxFSPacketSize
		acm.txSize = descCDCACMDataTxFSPacketSize
	}
	acm.txHead = 0
	acm.txFree = 0
	acm.rxHead = 0
	acm.rxTail = 0
	acm.rxFree = 0
	dc.endpointConfigureTx(descCDCACMEndpointStatus,
		acm.cxSize, false, nil)
	dc.endpointConfigureRx(descCDCACMEndpointDataRx,
		acm.rxSize, false, dc.serialNotify)
	dc.endpointConfigureTx(descCDCACMEndpointDataTx,
		acm.txSize, true, nil)
	for i := range acm.rd {
		dc.serialReceive(uint8(i))
	}
	dc.timerConfigure(0, 75, dc.serialFlush)
}

func (dc *deviceController) serialNotify(transfer *dcdTransfer) {
	acm := &descCDCACM[dc.class.config-1]
	len := acm.rxSize - (uint16(transfer.token>>16) & 0x7FFF)
	p := transfer.param
	if 0 == len {
		// zero-length packet (ZLP)
		dc.serialReceive(uint8(p))
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
				dc.serialReceive(uint8(p))
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

func (dc *deviceController) serialReceive(endpoint uint8) {
	ivm := arm.DisableInterrupts()
	num := uint16(endpoint) & descEndptAddrNumberMsk
	acm := &descCDCACM[dc.class.config-1]
	buf := &acm.rx[num*descCDCACMRxSize]
	dc.transferPrepare(&acm.rd[num], buf, acm.rxSize, uint32(endpoint))
	nxp.DeleteDcache(uintptr(unsafe.Pointer(buf)), uintptr(acm.rxSize))
	dc.receive(descCDCACMEndpointDataRx, &acm.rd[num])
	arm.EnableInterrupts(ivm)
}

func (dc *deviceController) serialFlush() {
	const autoFlushTx = true
	if !autoFlushTx {
		return
	}
	acm := &descCDCACM[dc.class.config-1]
	if 0 == acm.txFree {
		return
	}
	xfer := &acm.td[acm.txHead]
	buff := &acm.tx[uint16(acm.txHead)*descCDCACMTxSize]
	size := descCDCACMTxSize - acm.txFree
	dc.transferPrepare(xfer, buff, size, 0)
	nxp.FlushDeleteDcache(uintptr(unsafe.Pointer(buff)), uintptr(size))
	dc.transmit(descCDCACMEndpointDataTx, xfer)
	acm.txHead += 1
	if acm.txHead >= descCDCACMTDCount {
		acm.txHead = 0
	}
	acm.txFree = 0
}
