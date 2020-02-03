// +build nrf52840

package machine

import (
	"device/arm"
	"device/nrf"
	"runtime/interrupt"
	"runtime/volatile"
	"unicode/utf16"
	"unsafe"
)

// USBCDC is the USB CDC aka serial over USB interface on the nRF52840
type USBCDC struct {
	Buffer    *RingBuffer
	interrupt interrupt.Interrupt
}

// WriteByte writes a byte of data to the USB CDC interface.
func (usbcdc USBCDC) WriteByte(c byte) error {
	// Supposedly to handle problem with Windows USB serial ports?
	if usbLineInfo.lineState > 0 {
		cdcInBusy.Set(1)
		udd_ep_in_cache_buffer[usb_CDC_ENDPOINT_IN][0] = c
		sendViaEPIn(
			usb_CDC_ENDPOINT_IN,
			&udd_ep_in_cache_buffer[usb_CDC_ENDPOINT_IN][0],
			1,
		)

		for cdcInBusy.Get() == 1 {
			arm.Asm("wfi")
		}
	}

	return nil
}

func (usbcdc USBCDC) DTR() bool {
	return (usbLineInfo.lineState & usb_CDC_LINESTATE_DTR) > 0
}

func (usbcdc USBCDC) RTS() bool {
	return (usbLineInfo.lineState & usb_CDC_LINESTATE_RTS) > 0
}

var (
	USB                    = USBCDC{Buffer: NewRingBuffer()}
	usbEndpointDescriptors [8]usbDeviceDescriptor

	udd_ep_in_cache_buffer  [7][128]uint8
	udd_ep_out_cache_buffer [7][128]uint8

	sendOnEP0DATADONE struct {
		ptr   *byte
		count int
	}
	isEndpointHalt        = false
	isRemoteWakeUpEnabled = false
	endPoints             = []uint32{usb_ENDPOINT_TYPE_CONTROL,
		(usb_ENDPOINT_TYPE_INTERRUPT | usbEndpointIn),
		(usb_ENDPOINT_TYPE_BULK | usbEndpointOut),
		(usb_ENDPOINT_TYPE_BULK | usbEndpointIn)}

	usbConfiguration uint8
	usbSetInterface  uint8
	usbLineInfo      = cdcLineInfo{115200, 0x00, 0x00, 0x08, 0x00}
	epinen           uint32
	epouten          uint32
	cdcInBusy        volatile.Register8
)

// Configure the USB CDC interface. The config is here for compatibility with the UART interface.
func (usbcdc *USBCDC) Configure(config UARTConfig) {
	// enable IRQ
	usbcdc.interrupt = interrupt.New(nrf.IRQ_USBD, USB.handleInterrupt)
	usbcdc.interrupt.SetPriority(0xDF)
	usbcdc.interrupt.Enable()

	// enable USB
	nrf.USBD.ENABLE.Set(1)

	// enable interrupt for end of reset and start of frame
	nrf.USBD.INTENSET.Set(
		nrf.USBD_INTENSET_SOF |
			nrf.USBD_INTENSET_EP0DATADONE |
			nrf.USBD_INTENSET_USBEVENT |
			nrf.USBD_INTENSET_EP0SETUP,
	)

	nrf.USBD.USBPULLUP.Set(0)
}

func (usbcdc *USBCDC) handleInterrupt(interrupt.Interrupt) {
	// USBD ready event
	if nrf.USBD.EVENTS_USBEVENT.Get() == 1 {
		nrf.USBD.EVENTS_USBEVENT.Set(0)
		if (nrf.USBD.EVENTCAUSE.Get() & nrf.USBD_EVENTCAUSE_READY) > 0 {

			// Configure control endpoint
			initEndpoint(0, usb_ENDPOINT_TYPE_CONTROL)

			// Enable Setup-Received interrupt
			nrf.USBD.INTENSET.Set(nrf.USBD_INTENSET_EP0SETUP)
			nrf.USBD.USBPULLUP.Set(1)

			usbConfiguration = 0
		}
		nrf.USBD.EVENTCAUSE.Set(0)
	}

	// Start of frame
	if nrf.USBD.EVENTS_SOF.Get() == 1 {
		nrf.USBD.EVENTS_SOF.Set(0)
		// if you want to blink LED showing traffic, this would be the place...
	}

	if nrf.USBD.EVENTS_EP0DATADONE.Get() == 1 {
		// done sending packet - either need to send another or enter status stage
		nrf.USBD.EVENTS_EP0DATADONE.Set(0)
		if sendOnEP0DATADONE.ptr != nil {
			// previous data was too big for one packet, so send a second
			sendViaEPIn(
				0,
				sendOnEP0DATADONE.ptr,
				sendOnEP0DATADONE.count,
			)

			// clear, so we know we're done
			sendOnEP0DATADONE.ptr = nil
		} else {
			// no more data, so set status stage
			nrf.USBD.TASKS_EP0STATUS.Set(1)
		}
		return
	}

	// Endpoint 0 Setup interrupt
	if nrf.USBD.EVENTS_EP0SETUP.Get() == 1 {
		// ack setup received
		nrf.USBD.EVENTS_EP0SETUP.Set(0)

		// parse setup
		setup := parseUSBSetupRegisters()

		ok := false
		if (setup.bmRequestType & usb_REQUEST_TYPE) == usb_REQUEST_STANDARD {
			// Standard Requests
			ok = handleStandardSetup(setup)
		} else {
			if setup.wIndex == usb_CDC_ACM_INTERFACE {
				ok = cdcSetup(setup)
			}
		}

		if !ok {
			// Stall endpoint
			nrf.USBD.TASKS_EP0STALL.Set(1)
		}
	}

	// Now the actual transfer handlers, ignore endpoint number 0 (setup)
	if nrf.USBD.EVENTS_EPDATA.Get() > 0 {
		nrf.USBD.EVENTS_EPDATA.Set(0)
		epDataStatus := nrf.USBD.EPDATASTATUS.Get()
		nrf.USBD.EPDATASTATUS.Set(0)
		var i uint32
		for i = 1; i < uint32(len(endPoints)); i++ {
			// Check if endpoint has a pending interrupt
			inDataDone := epDataStatus&(1<<i) > 0
			outDataDone := epDataStatus&(0x10000<<i) > 0
			if inDataDone || outDataDone {
				switch i {
				case usb_CDC_ENDPOINT_OUT:
					nrf.USBD.EPOUT[i].PTR.Set(uint32(uintptr(unsafe.Pointer(&udd_ep_out_cache_buffer[i]))))
					nrf.USBD.EPOUT[i].MAXCNT.Set(64)
					nrf.USBD.TASKS_STARTEPOUT[i].Set(1)
				case usb_CDC_ENDPOINT_IN, usb_CDC_ENDPOINT_ACM:
					cdcInBusy.Set(0)
				}
			}
		}
	}

	for i := 0; i < len(endPoints); i++ {
		if nrf.USBD.EVENTS_ENDEPOUT[i].Get() > 0 {
			nrf.USBD.EVENTS_ENDEPOUT[i].Set(0)
			handleEndpoint(uint32(i))
		}
	}
}

func parseUSBSetupRegisters() usbSetup {
	return usbSetup{
		bmRequestType: uint8(nrf.USBD.BMREQUESTTYPE.Get()),
		bRequest:      uint8(nrf.USBD.BREQUEST.Get()),
		wValueL:       uint8(nrf.USBD.WVALUEL.Get()),
		wValueH:       uint8(nrf.USBD.WVALUEH.Get()),
		wIndex:        uint16((nrf.USBD.WINDEXH.Get() << 8) | nrf.USBD.WINDEXL.Get()),
		wLength:       uint16(((nrf.USBD.WLENGTHH.Get() & 0xff) << 8) | (nrf.USBD.WLENGTHL.Get() & 0xff)),
	}
}

func initEndpoint(ep, config uint32) {
	switch config {
	case usb_ENDPOINT_TYPE_INTERRUPT | usbEndpointIn:
		enableEPIn(ep)

	case usb_ENDPOINT_TYPE_BULK | usbEndpointOut:
		nrf.USBD.INTENSET.Set(nrf.USBD_INTENSET_EPDATA)
		nrf.USBD.INTENSET.Set(nrf.USBD_INTENSET_ENDEPOUT0 << ep)
		nrf.USBD.SIZE.EPOUT[ep].Set(0)
		enableEPOut(ep)

	case usb_ENDPOINT_TYPE_INTERRUPT | usbEndpointOut:
		nrf.USBD.INTENSET.Set(nrf.USBD_INTENSET_ENDEPOUT0 << ep)
		nrf.USBD.SIZE.EPOUT[ep].Set(0)
		enableEPOut(ep)

	case usb_ENDPOINT_TYPE_BULK | usbEndpointIn:
		nrf.USBD.INTENSET.Set(nrf.USBD_INTENSET_EPDATA)
		enableEPIn(ep)

	case usb_ENDPOINT_TYPE_CONTROL:
		enableEPIn(0)
		enableEPOut(0)
		nrf.USBD.EPOUT[0].PTR.Set(uint32(uintptr(unsafe.Pointer(&udd_ep_out_cache_buffer[0]))))
		nrf.USBD.EPOUT[0].MAXCNT.Set(64)
		nrf.USBD.TASKS_STARTEPOUT[0].Set(1)
		nrf.USBD.TASKS_EP0STATUS.Set(1)
	}
}

func handleStandardSetup(setup usbSetup) bool {
	switch setup.bRequest {
	case usb_GET_STATUS:
		buf := []byte{0, 0}

		if setup.bmRequestType != 0 { // endpoint
			// TODO: actually check if the endpoint in question is currently halted
			if isEndpointHalt {
				buf[0] = 1
			}
		}

		sendUSBPacket(0, buf)
		return true

	case usb_CLEAR_FEATURE:
		if setup.wValueL == 1 { // DEVICEREMOTEWAKEUP
			isRemoteWakeUpEnabled = false
		} else if setup.wValueL == 0 { // ENDPOINTHALT
			isEndpointHalt = false
		}
		nrf.USBD.TASKS_EP0STATUS.Set(1)
		return true

	case usb_SET_FEATURE:
		if setup.wValueL == 1 { // DEVICEREMOTEWAKEUP
			isRemoteWakeUpEnabled = true
		} else if setup.wValueL == 0 { // ENDPOINTHALT
			isEndpointHalt = true
		}
		nrf.USBD.TASKS_EP0STATUS.Set(1)
		return true

	case usb_SET_ADDRESS:
		// nrf USBD handles this
		return true

	case usb_GET_DESCRIPTOR:
		sendDescriptor(setup)
		return true

	case usb_SET_DESCRIPTOR:
		return false

	case usb_GET_CONFIGURATION:
		buff := []byte{usbConfiguration}
		sendUSBPacket(0, buff)
		return true

	case usb_SET_CONFIGURATION:
		if setup.bmRequestType&usb_REQUEST_RECIPIENT == usb_REQUEST_DEVICE {
			nrf.USBD.TASKS_EP0STATUS.Set(1)
			for i := 1; i < len(endPoints); i++ {
				initEndpoint(uint32(i), endPoints[i])
			}

			usbConfiguration = setup.wValueL
			return true
		} else {
			return false
		}

	case usb_GET_INTERFACE:
		buff := []byte{usbSetInterface}
		sendUSBPacket(0, buff)
		return true

	case usb_SET_INTERFACE:
		usbSetInterface = setup.wValueL

		nrf.USBD.TASKS_EP0STATUS.Set(1)
		return true

	default:
		return true
	}
}

func cdcSetup(setup usbSetup) bool {
	if setup.bmRequestType == usb_REQUEST_DEVICETOHOST_CLASS_INTERFACE {
		if setup.bRequest == usb_CDC_GET_LINE_CODING {
			b := make([]byte, 7)
			b[0] = byte(usbLineInfo.dwDTERate)
			b[1] = byte(usbLineInfo.dwDTERate >> 8)
			b[2] = byte(usbLineInfo.dwDTERate >> 16)
			b[3] = byte(usbLineInfo.dwDTERate >> 24)
			b[4] = byte(usbLineInfo.bCharFormat)
			b[5] = byte(usbLineInfo.bParityType)
			b[6] = byte(usbLineInfo.bDataBits)

			sendUSBPacket(0, b)
			return true
		}
	}

	if setup.bmRequestType == usb_REQUEST_HOSTTODEVICE_CLASS_INTERFACE {
		if setup.bRequest == usb_CDC_SET_LINE_CODING {
			b := receiveUSBControlPacket()
			if len(b) >= 7 {
				usbLineInfo.dwDTERate = uint32(b[0]) | uint32(b[1])<<8 | uint32(b[2])<<16 | uint32(b[3])<<24
				usbLineInfo.bCharFormat = b[4]
				usbLineInfo.bParityType = b[5]
				usbLineInfo.bDataBits = b[6]
			}
		}

		if setup.bRequest == usb_CDC_SET_CONTROL_LINE_STATE {
			usbLineInfo.lineState = setup.wValueL
		}

		if setup.bRequest == usb_CDC_SET_LINE_CODING || setup.bRequest == usb_CDC_SET_CONTROL_LINE_STATE {
			// auto-reset into the bootloader
			if usbLineInfo.dwDTERate == 1200 && usbLineInfo.lineState&usb_CDC_LINESTATE_DTR == 0 {
				// TODO: do we want to do this on the nRF52840?
				// ResetProcessor()
			} else {
				// TODO: cancel any reset
			}
			nrf.USBD.TASKS_EP0STATUS.Set(1)
		}

		if setup.bRequest == usb_CDC_SEND_BREAK {
			// TODO: something with this value?
			// breakValue = ((uint16_t)setup.wValueH << 8) | setup.wValueL;
			// return false;
			nrf.USBD.TASKS_EP0STATUS.Set(1)
		}
		return true
	}
	return false
}

func sendUSBPacket(ep uint32, data []byte) {
	count := len(data)
	copy(udd_ep_in_cache_buffer[ep][:], data)
	if ep == 0 && count > usbEndpointPacketSize {
		sendOnEP0DATADONE.ptr = &udd_ep_in_cache_buffer[ep][usbEndpointPacketSize]
		sendOnEP0DATADONE.count = count - usbEndpointPacketSize
		count = usbEndpointPacketSize
	}
	sendViaEPIn(
		ep,
		&udd_ep_in_cache_buffer[ep][0],
		count,
	)
}

func receiveUSBControlPacket() []byte {
	// set byte count to zero
	nrf.USBD.SIZE.EPOUT[0].Set(0)
	nrf.USBD.TASKS_EP0RCVOUT.Set(1)

	// Wait until OUT transfer is ready.
	timeout := 3000
	for nrf.USBD.EVENTS_EP0DATADONE.Get() == 0 {
		timeout--
		if timeout == 0 {
			return []byte{}
		}
	}

	// get data
	bytesread := int(nrf.USBD.SIZE.EPOUT[0].Get())
	data := make([]byte, bytesread)
	copy(data, udd_ep_out_cache_buffer[0][:])
	return data
}

// sendDescriptor creates and sends the various USB descriptor types that
// can be requested by the host.
func sendDescriptor(setup usbSetup) {
	switch setup.wValueH {
	case usb_CONFIGURATION_DESCRIPTOR_TYPE:
		sendConfiguration(setup)
		return
	case usb_DEVICE_DESCRIPTOR_TYPE:
		if setup.wLength == 8 {
			// composite descriptor requested, so only send 8 bytes
			dd := NewDeviceDescriptor(0xEF, 0x02, 0x01, 64, usb_VID, usb_PID, 0x100, usb_IMANUFACTURER, usb_IPRODUCT, usb_ISERIAL, 1)
			sendUSBPacket(0, dd.Bytes()[:8])
		} else {
			// complete descriptor requested so send entire packet
			dd := NewDeviceDescriptor(0x00, 0x00, 0x00, 64, usb_VID, usb_PID, 0x100, usb_IMANUFACTURER, usb_IPRODUCT, usb_ISERIAL, 1)
			sendUSBPacket(0, dd.Bytes())
		}
		return

	case usb_STRING_DESCRIPTOR_TYPE:
		switch setup.wValueL {
		case 0:
			b := make([]byte, 4)
			b[0] = byte(usb_STRING_LANGUAGE[0] >> 8)
			b[1] = byte(usb_STRING_LANGUAGE[0] & 0xff)
			b[2] = byte(usb_STRING_LANGUAGE[1] >> 8)
			b[3] = byte(usb_STRING_LANGUAGE[1] & 0xff)
			sendUSBPacket(0, b)

		case usb_IPRODUCT:
			b := strToUTF16LEDescriptor(usb_STRING_PRODUCT)
			sendUSBPacket(0, b[:setup.wLength])

		case usb_IMANUFACTURER:
			b := strToUTF16LEDescriptor(usb_STRING_MANUFACTURER)
			sendUSBPacket(0, b[:setup.wLength])

		case usb_ISERIAL:
			// TODO: allow returning a product serial number
			nrf.USBD.TASKS_EP0STATUS.Set(1)
		}
		return
	}

	// do not know how to handle this message, so return zero
	nrf.USBD.TASKS_EP0STATUS.Set(1)
	return
}

// sendConfiguration creates and sends the configuration packet to the host.
func sendConfiguration(setup usbSetup) {
	if setup.wLength == 9 {
		sz := uint16(configDescriptorSize + cdcSize)
		config := NewConfigDescriptor(sz, 2)

		sendUSBPacket(0, config.Bytes())
	} else {
		iad := NewIADDescriptor(0, 2, usb_CDC_COMMUNICATION_INTERFACE_CLASS, usb_CDC_ABSTRACT_CONTROL_MODEL, 0)

		cif := NewInterfaceDescriptor(usb_CDC_ACM_INTERFACE, 1, usb_CDC_COMMUNICATION_INTERFACE_CLASS, usb_CDC_ABSTRACT_CONTROL_MODEL, 0)

		header := NewCDCCSInterfaceDescriptor(usb_CDC_HEADER, usb_CDC_V1_10&0xFF, (usb_CDC_V1_10>>8)&0x0FF)

		controlManagement := NewACMFunctionalDescriptor(usb_CDC_ABSTRACT_CONTROL_MANAGEMENT, 6)

		functionalDescriptor := NewCDCCSInterfaceDescriptor(usb_CDC_UNION, usb_CDC_ACM_INTERFACE, usb_CDC_DATA_INTERFACE)

		callManagement := NewCMFunctionalDescriptor(usb_CDC_CALL_MANAGEMENT, 1, 1)

		cifin := NewEndpointDescriptor((usb_CDC_ENDPOINT_ACM | usbEndpointIn), usb_ENDPOINT_TYPE_INTERRUPT, 0x10, 0x10)

		dif := NewInterfaceDescriptor(usb_CDC_DATA_INTERFACE, 2, usb_CDC_DATA_INTERFACE_CLASS, 0, 0)

		out := NewEndpointDescriptor((usb_CDC_ENDPOINT_OUT | usbEndpointOut), usb_ENDPOINT_TYPE_BULK, usbEndpointPacketSize, 0)

		in := NewEndpointDescriptor((usb_CDC_ENDPOINT_IN | usbEndpointIn), usb_ENDPOINT_TYPE_BULK, usbEndpointPacketSize, 0)

		cdc := NewCDCDescriptor(iad,
			cif,
			header,
			controlManagement,
			functionalDescriptor,
			callManagement,
			cifin,
			dif,
			out,
			in)

		sz := uint16(configDescriptorSize + cdcSize)
		config := NewConfigDescriptor(sz, 2)

		buf := make([]byte, 0)
		buf = append(buf, config.Bytes()...)
		buf = append(buf, cdc.Bytes()...)
		sendUSBPacket(0, buf)
	}
}

func handleEndpoint(ep uint32) {
	// get data
	count := int(nrf.USBD.SIZE.EPOUT[ep].Get())

	// move to ring buffer
	for i := 0; i < count; i++ {
		USB.Receive(byte(udd_ep_out_cache_buffer[ep][i]))
	}

	// set ready for next data
	nrf.USBD.SIZE.EPOUT[ep].Set(0)
}

func sendViaEPIn(ep uint32, ptr *byte, count int) {
	nrf.USBD.EPIN[ep].PTR.Set(
		uint32(uintptr(unsafe.Pointer(ptr))),
	)
	nrf.USBD.EPIN[ep].MAXCNT.Set(uint32(count))
	nrf.USBD.EVENTS_ENDEPIN[ep].Set(0)
	nrf.USBD.TASKS_STARTEPIN[ep].Set(1)
}

func enableEPOut(ep uint32) {
	epouten = epouten | (nrf.USBD_EPOUTEN_OUT0 << ep)
	nrf.USBD.EPOUTEN.Set(epouten)
}

func enableEPIn(ep uint32) {
	epinen = epinen | (nrf.USBD_EPINEN_IN0 << ep)
	nrf.USBD.EPINEN.Set(epinen)
}

func strToUTF16LEDescriptor(in string) []byte {
	runes := []rune(in)
	encoded := utf16.Encode(runes)
	size := (len(encoded) << 1) + 2
	out := make([]byte, size)
	out[0] = byte(size)
	out[1] = 0x03
	for i, value := range encoded {
		out[(i<<1)+2] = byte(value & 0xff)
		out[(i<<1)+3] = byte(value >> 8)
	}
	return out
}
