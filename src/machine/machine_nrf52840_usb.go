//go:build nrf52840

package machine

import (
	"device/arm"
	"device/nrf"
	"machine/usb"
	"runtime/interrupt"
	"runtime/volatile"
	"unsafe"
)

const NumberOfUSBEndpoints = 8

var (
	sendOnEP0DATADONE struct {
		ptr    *byte
		count  int
		offset int
	}
	epinen      uint32
	epouten     uint32
	easyDMABusy volatile.Register8
	// epOutFlowControl contains the flow control state of the USB OUT endpoints.
	epOutFlowControl [NumberOfUSBEndpoints]struct {
		// nak indicates that we are NAKing any further OUT packets because the rxHandler isn't ready yet.
		// When this is true, we do not restart the DMA for the endpoint, effectively pausing it.
		nak bool
		// dataPending indicates that we have data in the hardware buffer that hasn't been handled yet.
		// Having one in the buffer is what generates the NAK responses, this is a signal to handle it.
		dataPending bool
	}
)

// enterCriticalSection is used to protect access to easyDMA - only one thing
// can be done with it at a time
func enterCriticalSection() {
	waitForEasyDMA()
	easyDMABusy.SetBits(1)
}

func waitForEasyDMA() {
	for easyDMABusy.HasBits(1) {
		arm.Asm("wfi")
	}
}

func exitCriticalSection() {
	easyDMABusy.ClearBits(1)
}

// Configure the USB peripheral. The config is here for compatibility with the UART interface.
func (dev *USBDevice) Configure(config UARTConfig) {
	if dev.initcomplete {
		return
	}

	state := interrupt.Disable()
	defer interrupt.Restore(state)

	nrf.USBD.USBPULLUP.Set(0)

	// Enable IRQ. Make sure this is higher than the SWI2 interrupt handler so
	// that it is possible to print to the console from a BLE interrupt. You
	// shouldn't generally do that but it is useful for debugging and panic
	// logging.
	intr := interrupt.New(nrf.IRQ_USBD, handleUSBIRQ)
	intr.SetPriority(0x40) // interrupt priority 2 (lower number means more important)
	intr.Enable()

	// enable interrupt for end of reset and start of frame
	nrf.USBD.INTEN.Set(nrf.USBD_INTENSET_USBEVENT)

	// errata 187
	// https://infocenter.nordicsemi.com/topic/errata_nRF52840_EngB/ERR/nRF52840/EngineeringB/latest/anomaly_840_187.html
	(*volatile.Register32)(unsafe.Pointer(uintptr(0x4006EC00))).Set(0x00009375)
	(*volatile.Register32)(unsafe.Pointer(uintptr(0x4006ED14))).Set(0x00000003)
	(*volatile.Register32)(unsafe.Pointer(uintptr(0x4006EC00))).Set(0x00009375)

	// enable USB
	nrf.USBD.ENABLE.Set(1)

	timeout := 300000
	for !nrf.USBD.EVENTCAUSE.HasBits(nrf.USBD_EVENTCAUSE_READY) {
		timeout--
		if timeout == 0 {
			return
		}
	}
	nrf.USBD.EVENTCAUSE.ClearBits(nrf.USBD_EVENTCAUSE_READY)

	// errata 187
	(*volatile.Register32)(unsafe.Pointer(uintptr(0x4006EC00))).Set(0x00009375)
	(*volatile.Register32)(unsafe.Pointer(uintptr(0x4006ED14))).Set(0x00000000)
	(*volatile.Register32)(unsafe.Pointer(uintptr(0x4006EC00))).Set(0x00009375)

	dev.initcomplete = true
}

func handleUSBIRQ(interrupt.Interrupt) {
	if nrf.USBD.EVENTS_SOF.Get() == 1 {
		nrf.USBD.EVENTS_SOF.Set(0)

		// if you want to blink LED showing traffic, this would be the place...
	}

	// USBD ready event
	if nrf.USBD.EVENTS_USBEVENT.Get() == 1 {
		nrf.USBD.EVENTS_USBEVENT.Set(0)
		if (nrf.USBD.EVENTCAUSE.Get() & nrf.USBD_EVENTCAUSE_READY) > 0 {

			// Configure control endpoint
			initEndpoint(0, usb.ENDPOINT_TYPE_CONTROL)
			nrf.USBD.USBPULLUP.Set(1)

			usbConfiguration = 0
		}
		nrf.USBD.EVENTCAUSE.Set(0)
	}

	if nrf.USBD.EVENTS_EP0DATADONE.Get() == 1 {
		// done sending packet - either need to send another or enter status stage
		nrf.USBD.EVENTS_EP0DATADONE.Set(0)
		if sendOnEP0DATADONE.ptr != nil {
			// previous data was too big for one packet, so send a second
			ptr := sendOnEP0DATADONE.ptr
			count := sendOnEP0DATADONE.count
			if count > usb.EndpointPacketSize {
				sendOnEP0DATADONE.offset += usb.EndpointPacketSize
				sendOnEP0DATADONE.ptr = &udd_ep_control_cache_buffer[sendOnEP0DATADONE.offset]
				count = usb.EndpointPacketSize
			}
			sendOnEP0DATADONE.count -= count
			sendViaEPIn(
				0,
				ptr,
				count,
			)

			// clear, so we know we're done
			if sendOnEP0DATADONE.count == 0 {
				sendOnEP0DATADONE.ptr = nil
				sendOnEP0DATADONE.offset = 0
			}
		} else {
			// no more data, so set status stage
			SendZlp() // nrf.USBD.TASKS_EP0STATUS.Set(1)
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
		if (setup.BmRequestType & usb.REQUEST_TYPE) == usb.REQUEST_STANDARD {
			// Standard Requests
			ok = handleStandardSetup(setup)
		} else {
			// Class Interface Requests
			if setup.WIndex < uint16(len(usbSetupHandler)) && usbSetupHandler[setup.WIndex] != nil {
				ok = usbSetupHandler[setup.WIndex](setup)
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
		nrf.USBD.EPDATASTATUS.Set(epDataStatus)
		var i uint32
		for i = 1; i < NumberOfUSBEndpoints; i++ {
			// Check if endpoint has a pending interrupt
			inDataDone := epDataStatus&(nrf.USBD_EPDATASTATUS_EPIN1<<(i-1)) > 0
			outDataDone := epDataStatus&(nrf.USBD_EPDATASTATUS_EPOUT1<<(i-1)) > 0
			if inDataDone {
				if usbTxHandler[i] != nil {
					usbTxHandler[i]()
				}
			} else if outDataDone {
				enterCriticalSection()
				nrf.USBD.EPOUT[i].PTR.Set(uint32(uintptr(unsafe.Pointer(&udd_ep_out_cache_buffer[i]))))
				count := nrf.USBD.SIZE.EPOUT[i].Get()
				nrf.USBD.EPOUT[i].MAXCNT.Set(count)
				if !epOutFlowControl[i].nak {
					// Normal case: We want data, so start DMA immediately
					nrf.USBD.TASKS_STARTEPOUT[i].Set(1)
					epOutFlowControl[i].dataPending = false
				} else {
					// NAK case: We want to NAK, so DO NOT start DMA.
					// The data stays in HW buffer. Host receives NAKs.
					// Mark that we have data waiting so we can fetch it later.
					epOutFlowControl[i].dataPending = true
				}
			}
		}
	}

	// ENDEPOUT[n] events
	for i := 0; i < NumberOfUSBEndpoints; i++ {
		if nrf.USBD.EVENTS_ENDEPOUT[i].Get() > 0 {
			nrf.USBD.EVENTS_ENDEPOUT[i].Set(0)
			buf := handleEndpointRx(uint32(i))
			if usbRxHandler[i] == nil || usbRxHandler[i](buf) {
				AckUsbOutTransfer(uint32(i))
			} else {
				// usbRxHandler returned false, so NAK further OUT packets until we're ready
				epOutFlowControl[i].nak = true
				nrf.USBD.SIZE.EPOUT[i].Set(0)
			}
			exitCriticalSection()
		}
	}
}

func parseUSBSetupRegisters() usb.Setup {
	return usb.Setup{
		BmRequestType: uint8(nrf.USBD.BMREQUESTTYPE.Get()),
		BRequest:      uint8(nrf.USBD.BREQUEST.Get()),
		WValueL:       uint8(nrf.USBD.WVALUEL.Get()),
		WValueH:       uint8(nrf.USBD.WVALUEH.Get()),
		WIndex:        uint16((nrf.USBD.WINDEXH.Get() << 8) | nrf.USBD.WINDEXL.Get()),
		WLength:       uint16(((nrf.USBD.WLENGTHH.Get() & 0xff) << 8) | (nrf.USBD.WLENGTHL.Get() & 0xff)),
	}
}

func initEndpoint(ep, config uint32) {
	switch config {
	case usb.ENDPOINT_TYPE_INTERRUPT | usb.EndpointIn:
		enableEPIn(ep)
		setEPDataPID(ep|usb.EndpointIn, false)

	case usb.ENDPOINT_TYPE_BULK | usb.EndpointOut:
		nrf.USBD.INTENSET.Set(nrf.USBD_INTENSET_ENDEPOUT0 << ep)
		nrf.USBD.SIZE.EPOUT[ep].Set(0)
		enableEPOut(ep)
		setEPDataPID(ep, false)

	case usb.ENDPOINT_TYPE_INTERRUPT | usb.EndpointOut:
		nrf.USBD.INTENSET.Set(nrf.USBD_INTENSET_ENDEPOUT0 << ep)
		nrf.USBD.SIZE.EPOUT[ep].Set(0)
		enableEPOut(ep)
		setEPDataPID(ep, false)

	case usb.ENDPOINT_TYPE_BULK | usb.EndpointIn:
		enableEPIn(ep)
		setEPDataPID(ep|usb.EndpointIn, false)

	case usb.ENDPOINT_TYPE_CONTROL:
		enableEPIn(0)
		enableEPOut(0)
		nrf.USBD.INTENSET.Set(nrf.USBD_INTENSET_ENDEPOUT0 |
			nrf.USBD_INTENSET_EP0SETUP |
			nrf.USBD_INTENSET_EPDATA |
			nrf.USBD_INTENSET_EP0DATADONE)
		SendZlp() // nrf.USBD.TASKS_EP0STATUS.Set(1)
	}
}

// SendUSBInPacket sends a packet for USBHID (interrupt in / bulk in).
func SendUSBInPacket(ep uint32, data []byte) bool {
	sendUSBPacket(ep, data)

	// clear transfer complete flag
	nrf.USBD.INTENCLR.Set(nrf.USBD_INTENCLR_ENDEPOUT0 << ep)

	return true
}

// Prevent file size increases: https://github.com/tinygo-org/tinygo/pull/998
//
//go:noinline
func sendUSBPacket(ep uint32, data []byte) {
	// Select the corresponding buffer.
	count := len(data)
	var buffer []byte
	if ep == 0 {
		buffer = udd_ep_control_cache_buffer[:]
		if count > usb.EndpointPacketSize {
			// The packet must be sent in chunks.
			sendOnEP0DATADONE.offset = usb.EndpointPacketSize
			sendOnEP0DATADONE.ptr = &udd_ep_control_cache_buffer[usb.EndpointPacketSize]
			sendOnEP0DATADONE.count = count - usb.EndpointPacketSize
			count = usb.EndpointPacketSize
		}
	} else {
		buffer = udd_ep_in_cache_buffer[ep][:]
	}

	// Copy the packet to the buffer.
	copy(buffer[:len(data)], data)

	// Send the first chunk of the packet.
	sendViaEPIn(
		ep,
		&buffer[0],
		count,
	)
}

func handleEndpointRx(ep uint32) []byte {
	// get data
	count := int(nrf.USBD.EPOUT[ep].AMOUNT.Get())

	return udd_ep_out_cache_buffer[ep][:count]
}

// AckUsbOutTransfer is called to acknowledge the completion of a USB OUT transfer.
// It also clears the NAK state and resumes data flow if it was paused.
func AckUsbOutTransfer(ep uint32) {
	epOutFlowControl[ep].nak = false

	// If we ignored a packet earlier (Buffer Full strategy), we must manually
	// trigger the DMA now to pull it from the HW buffer.
	if epOutFlowControl[ep].dataPending {
		epOutFlowControl[ep].dataPending = false

		// Prepare DMA to move data from HW Buffer -> RAM
		nrf.USBD.EPOUT[ep].PTR.Set(uint32(uintptr(unsafe.Pointer(&udd_ep_out_cache_buffer[ep]))))
		count := nrf.USBD.SIZE.EPOUT[ep].Get()
		nrf.USBD.EPOUT[ep].MAXCNT.Set(count)

		// Kick the DMA
		nrf.USBD.TASKS_STARTEPOUT[ep].Set(1)
		return
	}

	// Otherwise, just re-arm the endpoint to accept the NEXT packet
	nrf.USBD.SIZE.EPOUT[ep].Set(0)
}

func SendZlp() {
	nrf.USBD.TASKS_EP0STATUS.Set(1)
}

func sendViaEPIn(ep uint32, ptr *byte, count int) {
	nrf.USBD.EPIN[ep].PTR.Set(
		uint32(uintptr(unsafe.Pointer(ptr))),
	)
	nrf.USBD.EPIN[ep].MAXCNT.Set(uint32(count))
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

func handleUSBSetAddress(setup usb.Setup) bool {
	// nrf USBD handles this
	return true
}

func ReceiveUSBControlPacket() ([cdcLineInfoSize]byte, error) {
	var b [cdcLineInfoSize]byte

	nrf.USBD.TASKS_EP0RCVOUT.Set(1)

	nrf.USBD.EPOUT[0].PTR.Set(uint32(uintptr(unsafe.Pointer(&udd_ep_out_cache_buffer[0]))))
	nrf.USBD.EPOUT[0].MAXCNT.Set(64)

	timeout := 300000
	count := 0
	for {
		if nrf.USBD.EVENTS_EP0DATADONE.Get() == 1 {
			nrf.USBD.EVENTS_EP0DATADONE.Set(0)
			count = int(nrf.USBD.SIZE.EPOUT[0].Get())
			nrf.USBD.TASKS_STARTEPOUT[0].Set(1)
			break
		}
		timeout--
		if timeout == 0 {
			return b, ErrUSBReadTimeout
		}
	}

	timeout = 300000
	for {
		if nrf.USBD.EVENTS_ENDEPOUT[0].Get() == 1 {
			nrf.USBD.EVENTS_ENDEPOUT[0].Set(0)
			break
		}

		timeout--
		if timeout == 0 {
			return b, ErrUSBReadTimeout
		}
	}

	nrf.USBD.TASKS_EP0STATUS.Set(1)
	nrf.USBD.TASKS_EP0RCVOUT.Set(0)

	copy(b[:7], udd_ep_out_cache_buffer[0][:count])

	return b, nil
}

// Set the USB endpoint Packet ID to DATA0 or DATA1.
// In endpoints must have bit 7 (0x80) set.
func setEPDataPID(ep uint32, dataOne bool) {
	val := ep
	if dataOne {
		val |= nrf.USBD_DTOGGLE_VALUE_Data1 << nrf.USBD_DTOGGLE_VALUE_Pos
	} else {
		val |= nrf.USBD_DTOGGLE_VALUE_Data0 << nrf.USBD_DTOGGLE_VALUE_Pos
	}
	nrf.USBD.DTOGGLE.Set(val)
}

// Set ENDPOINT_HALT/stall status on a USB IN endpoint.
func (dev *USBDevice) SetStallEPIn(ep uint32) {
	if ep&0x7F == 0 {
		nrf.USBD.TASKS_EP0STALL.Set(1)
	} else if ep&0x7F < NumberOfUSBEndpoints {
		//     Stall   In     Endpoint
		val := 0x100 | 0x80 | ep
		nrf.USBD.EPSTALL.Set(val)
	}
}

// Set ENDPOINT_HALT/stall status on a USB OUT endpoint.
func (dev *USBDevice) SetStallEPOut(ep uint32) {
	if ep == 0 {
		nrf.USBD.TASKS_EP0STALL.Set(1)
	} else if ep < NumberOfUSBEndpoints {
		//     Stall   Out    Endpoint
		val := 0x100 | 0x00 | ep
		nrf.USBD.EPSTALL.Set(val)
	}
}

// Clear the ENDPOINT_HALT/stall on a USB IN endpoint.
func (dev *USBDevice) ClearStallEPIn(ep uint32) {
	if ep&0x7F == 0 {
		nrf.USBD.TASKS_EP0STALL.Set(0)
	} else if ep&0x7F < NumberOfUSBEndpoints {
		// Reset the endpoint data PID to DATA0
		ep |= 0x80 // Set endpoint direction bit
		setEPDataPID(ep, false)

		//  No-stall   In     Endpoint
		val := 0x000 | 0x80 | ep
		nrf.USBD.EPSTALL.Set(val)
	}
}

// Clear the ENDPOINT_HALT/stall on a USB OUT endpoint.
func (dev *USBDevice) ClearStallEPOut(ep uint32) {
	if ep == 0 {
		nrf.USBD.TASKS_EP0STALL.Set(0)
	} else if ep < NumberOfUSBEndpoints {
		// Reset the endpoint data PID to DATA0
		setEPDataPID(ep, false)

		//  No-stall   Out    Endpoint
		val := 0x000 | 0x00 | ep
		nrf.USBD.EPSTALL.Set(val)

		// Write a value to the SIZE register to allow nRF to ACK/accept data
		nrf.USBD.SIZE.EPOUT[ep].Set(0)
	}
}
