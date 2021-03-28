// +build mimxrt1062

package usb

// Implementation of a port controller for USB device mode on NXP iMXRT1062.

import (
	"device/arm"
	"device/nxp"
	"runtime/interrupt"
	"runtime/volatile"
	"unsafe"
)

// deviceControl represents a USB device controller for NXP iMXRT1062.
// It implements the USB API's deviceController interface.
type deviceControl struct {
	port   uint8
	device *device
	irq    interrupt.Interrupt

	messages deviceControllerInterruptQueue

	interruptMask uintptr            // interrupt state upon entering critical section
	criticalState volatile.Register8 // set to 1 if in critical section, else 0

	bus           *nxp.USB_Type
	phy           *nxp.USBPHY_Type
	nc            *nxp.USBNC_Type
	qh            *deviceControllerQH     // The QH structure base address
	dtd           *deviceControllerDTD    // The DTD structure base address
	dtdFree       *deviceControllerDTD    // The idle DTD list head
	dtdHead       deviceControllerDTDList // The transferring DTD list head for each endpoint
	dtdTail       deviceControllerDTDList // The transferring DTD list tail for each endpoint
	dtdCount      uint8                   // The idle DTD node count
	endpointCount uint8                   // The endpoint number of EHCI
	isResetting   bool                    // Whether a PORT reset is occurring or not
	controllerId  uint8                   // Controller ID
	speed         uint8                   // Current speed of EHCI
	isSuspending  bool                    // Is suspending of the PORT
}

var (
	// deviceControlInstance holds instances for all USB device controllers
	// available on the platform.
	deviceControlInstance [configDeviceCount]deviceControl

	//go:align 2048
	deviceControllerQHBuffer [deviceControllerQHBufferSize]uint8
	//go:align 32
	deviceControllerDTDBuffer [deviceControllerDTDBufferSize]uint8
)

// We cannot use the sleep timer from this context (import cycle), but we need
// an approximate method to spin CPU cycles for short periods of time.
//go:inline
func delayMicrosec(microsec uint32) {
	n := cycles(microsec, configCPUFrequencyHz)
	for i := uint32(0); i < n; i++ {
		arm.Asm(`nop`)
	}
}

// initController returns a deviceController for the receiver USB device.
// It allocates the registers and installs/enables an interrupt handler for
// the receiver USB device. It also initializes the shared buffers used by the
// USB controller hardware.
func (d *device) initController() deviceController {

	dc := &deviceControlInstance[d.port]

	dc.port = d.port
	dc.device = d

	// based on the selected port, install interrupt handler and initialize
	// register references
	switch d.port {
	case 0:
		dc.irq = interrupt.New(nxp.IRQ_USB_OTG1, func(interrupt.Interrupt) {
			portInstance[0].device.controller.interrupt()
		})
		dc.bus = nxp.USB1
		dc.phy = nxp.USBPHY1
		dc.nc = nxp.USBNC1
	case 1:
		dc.irq = interrupt.New(nxp.IRQ_USB_OTG2, func(interrupt.Interrupt) {
			//portInstance[1].device.controller.interrupt()
		})
		dc.bus = nxp.USB2
		dc.phy = nxp.USBPHY2
		dc.nc = nxp.USBNC2
	}

	// get base address of QH and DTD buffers
	dc.qh = (*deviceControllerQH)(unsafe.Pointer(
		&deviceControllerQHBuffer[int(d.port)*configDeviceControllerQHAlign]))
	dc.dtd = (*deviceControllerDTD)(unsafe.Pointer(
		&deviceControllerDTDBuffer[int(d.port)*configDeviceControllerDTDAlign]))

	dc.irq.SetPriority(configInterruptPriority)
	dc.irq.Enable()

	return dc
}

// init initializes the USB device subsystem for the receiver deviceControl.
func (dc *deviceControl) init() status {

	// reset the controller
	dc.bus.USBCMD.SetBits(nxp.USB_USBCMD_RST)
	for dc.bus.USBCMD.HasBits(nxp.USB_USBCMD_RST) {
	}

	// get hardware's endpoint count
	dc.endpointCount = uint8(dc.bus.DCCPARAMS.Get() & nxp.USB_DCCPARAMS_DEN_Msk)
	if dc.endpointCount < configDeviceMaxEndpoints {
		return statusError
	}

	// clear the controller mode field and set to device mode:
	//   controller mode (CM) 0x0=idle, 0x2=device-only, 0x3=host-only
	dc.bus.USBMODE.ReplaceBits(nxp.USB_USBMODE_CM_CM_2,
		nxp.USB_USBMODE_CM_Msk>>nxp.USB_USBMODE_CM_Pos, nxp.USB_USBMODE_CM_Pos)

	// reset the constroller state to default
	return dc.resetState()
}

func (dc *deviceControl) deinit() status {
	return statusSuccess
}

func (dc *deviceControl) enable(enable bool) status {
	if enable {
		// ensure D+ pulled down long enough for host to detect previous disconnect
		delayMicrosec(5000)
		return dc.control(deviceControlRun, nil)
	} else {
		return dc.control(deviceControlStop, nil)
	}
}

// interrupt is the base interrupt handler for all USB device interrupts.
func (dc *deviceControl) interrupt() {

	// protect access to message queue
	for statusRetry == dc.critical(true) {
	}

	// read and clear the interrupts that fired
	status := dc.bus.USBSTS.Get() & dc.bus.USBINTR.Get()
	dc.bus.USBSTS.Set(status)

	// enqueue interrupts for runtime processing
	dc.messages.enq(uintptr(status))

	// release message queue
	_ = dc.critical(false)
}

func (dc *deviceControl) process() status {

	// protect access to message queue
	if dc.critical(true).OK() {

		// dequeue oldest interrupt in message queue
		status, ok := dc.messages.deq()

		// release message queue
		_ = dc.critical(false)

		// process message if queue was not empty
		if ok {

			if 0 != (status & nxp.USB_USBSTS_URI_Msk) { // USB reset
				dc.reset()
			}

			if 0 != (status & nxp.USB_USBSTS_UI_Msk) { // USB token done
				dc.tokenDone()
			}

			if 0 != (status & nxp.USB_USBSTS_PCI_Msk) { // USB port status change
				dc.portChange()
			}

			if 0 != (status & nxp.USB_USBSTS_SRI_Msk) { // USB start of frame (SOF)
				dc.frameStart()
			}
		}

		// message queue read and processed
		return statusSuccess
	}

	// could not acquire lock on message queue
	return statusBusy
}

func (dc *deviceControl) transfer(address uint8, buffer []uint8, length uint32) status {

	if dc.isResetting {
		return statusError
	}

	endpoint, direction := unpackEndpoint(address)
	endpointIndex := int((endpoint << 1) | direction)
	currentIndex := 0

	primeBit := uint32(1) << ((address & specDescriptorEndpointAddressNumberMsk) +
		((address & specDescriptorEndpointAddressDirectionMsk) >> 3))
	epStatus, qhIdle := primeBit, false

	qh := getQHBuffer(dc.port, 0, endpointIndex)
	if 0 == qh.endpointStatus&0x1 { // bit 0: isOpened
		return statusError
	}

	dtdRequestCount := (length + deviceControllerDTDTotalBytes - 1) /
		deviceControllerDTDTotalBytes
	if 0 == dtdRequestCount {
		dtdRequestCount = 1
	}

	if dtdRequestCount > uint32(dc.dtdCount) {
		return statusBusy
	}

	var (
		dtdHead    *deviceControllerDTD
		sendLength uint32
	)

	for {

		// limit transfer length to total DTD bytes
		sendLength = length
		if length > deviceControllerDTDTotalBytes {
			sendLength = deviceControllerDTDTotalBytes
		}
		length -= sendLength

		// select a free DTD
		dtd := dc.dtdFree
		dc.dtdFree = dtd.nextDTDPointer
		dc.dtdCount--

		// save DTD head when current active buffer offset is 0
		if 0 == currentIndex {
			dtdHead = dtd
		}

		// set DTD field
		dtd.nextDTDPointer = deviceControllerDTDTerminate
		dtd.bufferPointerPage[0] =
			uint32(uintptr(unsafe.Pointer(&buffer[0]))) + uint32(currentIndex)
		dtd.bufferPointerPage[1] =
			(dtd.bufferPointerPage[0] + deviceControllerDTDPageBlock) & deviceControllerDTDPageMsk
		dtd.bufferPointerPage[2] =
			dtd.bufferPointerPage[1] + deviceControllerDTDPageBlock
		dtd.bufferPointerPage[3] =
			dtd.bufferPointerPage[2] + deviceControllerDTDPageBlock
		dtd.bufferPointerPage[4] =
			dtd.bufferPointerPage[3] + deviceControllerDTDPageBlock

		// save original buffer and length to transfer
		dtd.originalBuffer = deviceControllerOriginalBuffer{
			originalBufferOffset: uint16(dtd.bufferPointerPage[0]) &
				deviceControllerDTDPageOffsetMsk,
			originalBufferLength: sendLength,
			dtdInvalid:           0,
		}.pack()

		// set IOC field in final DTD
		ioc := uint8(0)
		if 0 == length {
			ioc = 1
		}
		// set DTD active flag
		dtd.dtdToken = deviceControllerDTDToken{
			status:     deviceControllerDTDStatusActive,
			ioc:        ioc,
			totalBytes: uint16(sendLength),
		}.pack()

		// update buffer offset
		currentIndex += int(sendLength)

		// add DTD to in-use queue
		if nil != dc.dtdTail[endpointIndex] {
			dc.dtdTail[endpointIndex].nextDTDPointer = dtd
			dc.dtdTail[endpointIndex] = dtd
		} else {
			dc.dtdHead[endpointIndex] = dtd
			dc.dtdTail[endpointIndex] = dtd
			qhIdle = true
		}

		if 0 == length {
			break
		}
	}

	if specEndpointControl == endpoint && specIn == direction {
		// get last setup packet
		setupIndex := int(endpoint << 1)
		setupQH := getQHBuffer(dc.port, 0, setupIndex)
		setupMaxSize := uint32(setupQH.capabilities&0x07FF0000) >> 16 // bits 15-26: maxPacketSize
		var setup deviceSetup
		setup.parse(setupQH.setupBufferBack[:])
		if 0 != qh.endpointStatus&0x2 { // bit 1: ZLT
			if (0 != sendLength) && (sendLength < uint32(setup.wLength)) &&
				(0 == (sendLength % setupMaxSize)) {
				// enable ZLT (zlt==0)
				setupQH.capabilities &^= deviceControllerCapabilities{zlt: 1}.pack()
			}
		}
	}

	// check if QH is empty
	if !qhIdle {
		// if prime bit is set, nothing left to do
		if dc.bus.ENDPTPRIME.HasBits(primeBit) {
			return statusSuccess
		}
		// wait to safely transmit DTD
		for {
			dc.bus.USBCMD.SetBits(nxp.USB_USBCMD_ATDTW)
			_ = dc.bus.ENDPTSTAT.Get() // read-clear endpoint status register
			if dc.bus.USBCMD.HasBits(nxp.USB_USBCMD_ATDTW) {
				break
			}
		}
		dc.bus.USBCMD.ClearBits(nxp.USB_USBCMD_ATDTW)
	}

	// if QH is empty or the endpoint is not primed, need to link current DTD head
	// to the QH. if endpoint is not primed and qhIdle is false, QH is empty.
	if qhIdle || 0 == epStatus&primeBit {
		qh.nextDTDPointer = dtdHead
		qh.dtdToken = 0
		dc.bus.ENDPTPRIME.Set(primeBit)
		primeAttempt := 0
		for !dc.bus.ENDPTSTAT.HasBits(primeBit) {
			if primeAttempt++; primeAttempt >= configDeviceControllerMaxPrimeAttempts {
				return statusError
			}
			if dc.bus.ENDPTCOMPLETE.HasBits(primeBit) {
				break
			}
			dc.bus.ENDPTPRIME.Set(primeBit)
		}
	}

	return statusSuccess
}

func (dc *deviceControl) send(address uint8, buffer []uint8, length uint32) status {
	return dc.transfer((address&specDescriptorEndpointAddressNumberMsk)|
		(specDescriptorEndpointAddressDirectionIn), buffer, length)
}

func (dc *deviceControl) receive(address uint8, buffer []uint8, length uint32) status {
	return dc.transfer((address&specDescriptorEndpointAddressNumberMsk)|
		(specDescriptorEndpointAddressDirectionOut), buffer, length)
}

func (dc *deviceControl) cancel(address uint8) status {
	return statusSuccess
}

func (dc *deviceControl) control(command deviceControlID, param interface{}) (s status) {

	// assume success unless error condition deliberately detected
	s = statusSuccess

	switch command {
	case deviceControlRun:
		dc.bus.USBCMD.SetBits(nxp.USB_USBCMD_RS)

	case deviceControlStop:
		dc.bus.USBCMD.ClearBits(nxp.USB_USBCMD_RS)

	case deviceControlEndpointInit:
		config, ok := param.(*deviceEndpointConfig)
		if !ok {
			return statusInvalidParameter
		}
		s = dc.initEndpoint(config)

	case deviceControlEndpointDeinit:
		address, ok := param.(uint8)
		if !ok {
			return statusInvalidParameter
		}
		s = dc.deinitEndpoint(address)

	case deviceControlEndpointStall:
		address, ok := param.(uint8)
		if !ok {
			return statusInvalidParameter
		}
		s = dc.stallEndpoint(address)

	case deviceControlEndpointUnstall:
		address, ok := param.(uint8)
		if !ok {
			return statusInvalidParameter
		}
		s = dc.unstallEndpoint(address)

	case deviceControlGetDeviceStatus:
		// param should be a pointer to uint16, acting as output parameter.
		stat, ok := param.(*uint16)
		if !ok {
			return statusInvalidParameter
		}
		// configDeviceSelfPowered is a configuration constant on iMXRT1062
		*stat = configDeviceSelfPowered <<
			specRequestStandardGetStatusDeviceSelfPoweredPos

	case deviceControlGetEndpointStatus:
		// param should be pointer to deviceEndpointStatus, acting as output
		// parameter.
		stat, ok := param.(*deviceEndpointStatus)
		if !ok {
			return statusInvalidParameter
		}
		endpoint, direction := unpackEndpoint(stat.address)
		if endpoint >= configDeviceMaxEndpoints {
			return statusInvalidParameter
		}
		mask := uint32(nxp.USB_ENDPTCTRL0_RXS)
		if specOut != direction {
			mask = nxp.USB_ENDPTCTRL0_TXS
		}
		ctrl := dc.endpointControlRegister(endpoint)
		if nil == ctrl {
			return statusInvalidParameter
		}
		if ctrl.HasBits(mask) {
			stat.status = uint16(deviceEndpointStateStalled)
		} else {
			stat.status = uint16(deviceEndpointStateIdle)
		}

	case deviceControlPreSetDeviceAddress:
		address, ok := param.(uint8)
		if !ok {
			return statusInvalidParameter
		}
		dc.bus.DEVICEADDR.Set((uint32(address) << nxp.USB_DEVICEADDR_USBADR_Pos) |
			nxp.USB_DEVICEADDR_USBADRA_Msk)

	case deviceControlSetDeviceAddress:
		// TODO

	case deviceControlGetSynchFrame:
		return statusNotSupported

	case deviceControlSetDefaultStatus:
		for i := uint8(0); i < configDeviceMaxEndpoints; i++ {
			_ = dc.deinitEndpoint(i | specDescriptorEndpointAddressDirectionIn)
			_ = dc.deinitEndpoint(i | specDescriptorEndpointAddressDirectionOut)
		}
		s = dc.resetState()

	case deviceControlGetSpeed:
		// param should be a pointer to uint8, acting as output parameter.
		speed, ok := param.(*uint8)
		if !ok {
			return statusInvalidParameter
		}
		*speed = dc.speed

	case deviceControlGetOTGStatus:
		return statusNotSupported

	case deviceControlSetOTGStatus:
		return statusNotSupported
	}

	return
}

func (dc *deviceControl) critical(enter bool) status {
	if enter {
		// check if critical section already locked
		if dc.criticalState.Get() != 0 {
			return statusRetry
		}
		// lock critical section
		dc.criticalState.Set(1)
		// disable interrupts, storing state in receiver
		dc.interruptMask = arm.DisableInterrupts()
	} else {
		// ensure critical section is locked
		if dc.criticalState.Get() != 0 {
			// re-enable interrupts, using state stored in receiver
			arm.EnableInterrupts(dc.interruptMask)
			// unlock critical section
			dc.criticalState.Set(0)
		}
	}
	return statusSuccess
}

func (dc *deviceControl) resetState() status {

	dc.dtdFree = dc.dtd
	p := dc.dtdFree
	for i := 1; i < configDeviceControllerMaxDTD; i++ {
		p.nextDTDPointer = getDTDBuffer(dc.port, i)
		p = p.nextDTDPointer
	}
	p.nextDTDPointer = nil
	dc.dtdCount = configDeviceControllerMaxDTD

	// no interrupt threshold
	dc.bus.USBCMD.ClearBits(nxp.USB_USBCMD_ITC_Msk)

	// disable setup lockout
	dc.bus.USBMODE.SetBits(nxp.USB_USBMODE_SLOM_Msk)

	// use little-endianness
	dc.bus.USBMODE.ClearBits(nxp.USB_USBMODE_ES_Msk)

	for i := 0; i < 2*configDeviceMaxEndpoints; i++ {
		qh := getQHBuffer(dc.port, 0, i)
		qh.capabilities =
			deviceControllerCapabilities{
				maxPacketSize: configDeviceControllerMaxPacketSize,
			}.pack()
		qh.endpointStatus =
			deviceControllerEndpointStatus{
				isOpened: 0,
			}.pack()
		qh.nextDTDPointer = deviceControllerDTDTerminate
		dc.dtdHead[i] = nil
		dc.dtdTail[i] = nil
	}
	dc.bus.ASYNCLISTADDR.Set(uint32(uintptr(unsafe.Pointer(
		getQHBuffer(dc.port, 0, 0)))))

	dc.bus.DEVICEADDR.Set(0)

	// enable interrupts: bus enable, bus error, port change detect, bus reset
	dc.bus.USBINTR.Set(nxp.USB_USBINTR_UE_Msk | nxp.USB_USBINTR_UEE_Msk |
		nxp.USB_USBINTR_PCE_Msk | nxp.USB_USBINTR_URE_Msk)

	dc.isResetting = false

	return statusSuccess
}

func (dc *deviceControl) reset() {

	println("reset")

	// clear setup flag
	dc.bus.ENDPTSETUPSTAT.Set(dc.bus.ENDPTSETUPSTAT.Get())
	// clear endpoint complete flag
	dc.bus.ENDPTCOMPLETE.Set(dc.bus.ENDPTCOMPLETE.Get())

	// flush any pending transfers
	for dc.bus.ENDPTPRIME.HasBits(nxp.USB_ENDPTPRIME_PERB_Msk | nxp.USB_ENDPTPRIME_PETB_Msk) {
		dc.bus.ENDPTFLUSH.Set(nxp.USB_ENDPTFLUSH_FERB_Msk | nxp.USB_ENDPTFLUSH_FETB_Msk)
	}

	// set receiver flag if port reset bit is set; otherwise, notify device class.
	if dc.bus.PORTSC1.HasBits(nxp.USB_PORTSC1_PR_Msk) {
		dc.isResetting = true
	} else {
		// send reset notification to common device
		dc.device.notify(deviceNotification{code: deviceNotifyBusReset})
	}
}

func (dc *deviceControl) tokenDone() {
	println("token done")
}

func (dc *deviceControl) portChange() {
	// check if port is resetting
	if !dc.bus.PORTSC1.HasBits(nxp.USB_PORTSC1_PR) {
		// not resetting, update bus speed
		if dc.bus.PORTSC1.HasBits(nxp.USB_PORTSC1_HSP) {
			dc.speed = specSpeedHigh
		} else {
			dc.speed = specSpeedFull
		}
		// if reset flag is set, notify device layer reset has finished
		if dc.isResetting {
			dc.device.notify(
				deviceNotification{
					buffer:  nil,
					length:  0,
					code:    deviceNotifyBusReset,
					isSetup: false,
				})
			dc.isResetting = false
		}
	}
}

func (dc *deviceControl) frameStart() {
	println("frame start")
}

func (dc *deviceControl) endpointControlRegister(endpoint uint8) *volatile.Register32 {
	endpoint &= specDescriptorEndpointAddressNumberMsk
	if endpoint < configDeviceMaxEndpoints {
		switch endpoint {
		case 0:
			return &dc.bus.ENDPTCTRL0
		case 1:
			return &dc.bus.ENDPTCTRL1
		case 2:
			return &dc.bus.ENDPTCTRL2
		case 3:
			return &dc.bus.ENDPTCTRL3
		case 4:
			return &dc.bus.ENDPTCTRL4
		case 5:
			return &dc.bus.ENDPTCTRL5
		case 6:
			return &dc.bus.ENDPTCTRL6
		case 7:
			return &dc.bus.ENDPTCTRL7
		}
	}
	return nil
}

func (dc *deviceControl) cancelControlPipe(endpoint, direction uint8) status {
	index := (endpoint << 1) + direction
	message := deviceNotification{
		buffer: nil,
		length: 0,
	}

	// get DTD of control pipe
	currentDTD := (*deviceControllerDTD)(unsafe.Pointer(
		uintptr(unsafe.Pointer(dc.dtdHead[index])) & deviceControllerDTDPointerMsk))

	for nil != currentDTD {
		originalBuffer := currentDTD.originalBuffer.unpack()
		dtdToken := currentDTD.dtdToken.unpack()

		// pass transfer buffer address
		if nil == message.buffer {
			message.buffer = *(*[]uint8)(unsafe.Pointer(
				uintptr((currentDTD.bufferPointerPage[0] & deviceControllerDTDPageMsk) |
					uint32(originalBuffer.originalBufferOffset))))
		}
		if 0 != dtdToken.status&deviceControllerDTDStatusActive {
			message.length = packU32(deviceCDCACMBufferInvalid32)
		} else {
			message.length += originalBuffer.originalBufferLength - uint32(dtdToken.totalBytes)
		}

		if dc.dtdHead[index] == dc.dtdTail[index] {
			dc.dtdHead[index] = nil
			dc.dtdTail[index] = nil
			qh := getQHBuffer(dc.port, 0, int(index))
			qh.nextDTDPointer = deviceControllerDTDTerminate
			qh.dtdToken = 0
		} else {
			dc.dtdHead[index] = dc.dtdHead[index].nextDTDPointer
		}

		if 0 != currentDTD.dtdToken.unpack().ioc ||
			0 == uintptr(unsafe.Pointer(dc.dtdHead[index]))&deviceControllerDTDPointerMsk {
			message.code = deviceNotificationID(endpoint |
				(direction << specDescriptorEndpointAddressDirectionPos))
			message.isSetup = false
			dc.device.notify(message)
			message.buffer = nil
			message.length = 0
		}

		currentDTD.dtdToken = 0
		currentDTD.nextDTDPointer = dc.dtdFree
		dc.dtdFree = currentDTD
		dc.dtdCount++

		// get DTD of control pipe
		currentDTD = (*deviceControllerDTD)(unsafe.Pointer(
			uintptr(unsafe.Pointer(dc.dtdHead[index])) & deviceControllerDTDPointerMsk))

	}
	return statusSuccess
}

func (dc *deviceControl) initEndpoint(config *deviceEndpointConfig) status {
	return statusSuccess
}

func (dc *deviceControl) deinitEndpoint(address uint8) status {
	return statusSuccess
}

func (dc *deviceControl) stallEndpoint(address uint8) status {
	return statusSuccess
}

func (dc *deviceControl) unstallEndpoint(address uint8) status {
	return statusSuccess
}
