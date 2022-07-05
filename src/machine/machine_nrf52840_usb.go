//go:build nrf52840
// +build nrf52840

package machine

import (
	"device/arm"
	"device/nrf"
	"runtime/interrupt"
	"runtime/volatile"
	"unsafe"
)

// USBCDC is the USB CDC aka serial over USB interface on the nRF52840
type USBCDC struct {
	Buffer            *RingBuffer
	interrupt         interrupt.Interrupt
	initcomplete      bool
	TxIdx             volatile.Register8
	waitTxc           bool
	waitTxcRetryCount uint8
	sent              bool
}

const (
	usbcdcTxSizeMask          uint8 = 0x3F
	usbcdcTxBankMask          uint8 = ^usbcdcTxSizeMask
	usbcdcTxBank1st           uint8 = 0x00
	usbcdcTxBank2nd           uint8 = usbcdcTxSizeMask + 1
	usbcdcTxMaxRetriesAllowed uint8 = 5
)

// Flush flushes buffered data.
func (usbcdc *USBCDC) Flush() error {
	if usbLineInfo.lineState > 0 {
		idx := usbcdc.TxIdx.Get()
		sz := idx & usbcdcTxSizeMask
		bk := idx & usbcdcTxBankMask
		if 0 < sz {

			if usbcdc.waitTxc {
				// waiting for the next flush(), because the transmission is not complete
				usbcdc.waitTxcRetryCount++
				return nil
			}
			usbcdc.waitTxc = true
			usbcdc.waitTxcRetryCount = 0

			// set the data
			enterCriticalSection()
			sendViaEPIn(
				usb_CDC_ENDPOINT_IN,
				&udd_ep_in_cache_buffer[usb_CDC_ENDPOINT_IN][bk],
				int(sz),
			)
			if bk == usbcdcTxBank1st {
				usbcdc.TxIdx.Set(usbcdcTxBank2nd)
			} else {
				usbcdc.TxIdx.Set(usbcdcTxBank1st)
			}

			usbcdc.sent = true
		}
	}
	return nil
}

// WriteByte writes a byte of data to the USB CDC interface.
func (usbcdc *USBCDC) WriteByte(c byte) error {
	// Supposedly to handle problem with Windows USB serial ports?
	if usbLineInfo.lineState > 0 {
		ok := false
		for {
			mask := interrupt.Disable()

			idx := usbcdc.TxIdx.Get()
			if (idx & usbcdcTxSizeMask) < usbcdcTxSizeMask {
				udd_ep_in_cache_buffer[usb_CDC_ENDPOINT_IN][idx] = c
				usbcdc.TxIdx.Set(idx + 1)
				ok = true
			}

			interrupt.Restore(mask)

			if ok {
				break
			} else if usbcdcTxMaxRetriesAllowed < usbcdc.waitTxcRetryCount {
				mask := interrupt.Disable()
				usbcdc.waitTxc = false
				usbcdc.waitTxcRetryCount = 0
				usbcdc.TxIdx.Set(0)
				usbLineInfo.lineState = 0
				interrupt.Restore(mask)
				break
			} else {
				mask := interrupt.Disable()
				if usbcdc.sent {
					if usbcdc.waitTxc {
						if !easyDMABusy.HasBits(1) {
							usbcdc.waitTxc = false
							usbcdc.Flush()
						}
					} else {
						usbcdc.Flush()
					}
				}
				interrupt.Restore(mask)
			}
		}
	}

	return nil
}

func (usbcdc *USBCDC) DTR() bool {
	return (usbLineInfo.lineState & usb_CDC_LINESTATE_DTR) > 0
}

func (usbcdc *USBCDC) RTS() bool {
	return (usbLineInfo.lineState & usb_CDC_LINESTATE_RTS) > 0
}

var (
	USB        = &_USB
	_USB       = USBCDC{Buffer: NewRingBuffer()}
	waitHidTxc bool

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

	usbConfiguration         uint8
	usbSetInterface          uint8
	usbLineInfo              = cdcLineInfo{115200, 0x00, 0x00, 0x08, 0x00}
	epinen                   uint32
	epouten                  uint32
	easyDMABusy              volatile.Register8
	epout0data_setlinecoding bool
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

// Configure the USB CDC interface. The config is here for compatibility with the UART interface.
func (usbcdc *USBCDC) Configure(config UARTConfig) {
	if usbcdc.initcomplete {
		return
	}

	// Enable IRQ. Make sure this is higher than the SWI2 interrupt handler so
	// that it is possible to print to the console from a BLE interrupt. You
	// shouldn't generally do that but it is useful for debugging and panic
	// logging.
	usbcdc.interrupt = interrupt.New(nrf.IRQ_USBD, _USB.handleInterrupt)
	usbcdc.interrupt.SetPriority(0x40) // interrupt priority 2 (lower number means more important)
	usbcdc.interrupt.Enable()

	// enable USB
	nrf.USBD.ENABLE.Set(1)

	// enable interrupt for end of reset and start of frame
	nrf.USBD.INTENSET.Set(
		nrf.USBD_INTENSET_EPDATA |
			nrf.USBD_INTENSET_EP0DATADONE |
			nrf.USBD_INTENSET_USBEVENT |
			nrf.USBD_INTENSET_SOF |
			nrf.USBD_INTENSET_EP0SETUP,
	)

	nrf.USBD.USBPULLUP.Set(0)

	usbcdc.initcomplete = true
}

func (usbcdc *USBCDC) handleInterrupt(interrupt.Interrupt) {
	if nrf.USBD.EVENTS_SOF.Get() == 1 {
		nrf.USBD.EVENTS_SOF.Set(0)
		usbcdc.Flush()
		if hidCallback != nil && !waitHidTxc {
			hidCallback()
		}
		// if you want to blink LED showing traffic, this would be the place...
	}

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

	if nrf.USBD.EVENTS_EP0DATADONE.Get() == 1 {
		// done sending packet - either need to send another or enter status stage
		nrf.USBD.EVENTS_EP0DATADONE.Set(0)
		if epout0data_setlinecoding {
			nrf.USBD.EPOUT[0].PTR.Set(uint32(uintptr(unsafe.Pointer(&udd_ep_out_cache_buffer[0]))))
			nrf.USBD.EPOUT[0].MAXCNT.Set(64)
			nrf.USBD.TASKS_STARTEPOUT[0].Set(1)
			return
		}
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
			} else if setup.bmRequestType == usb_SET_REPORT_TYPE && setup.bRequest == usb_SET_IDLE {
				sendZlp()
				ok = true
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
		for i = 1; i < uint32(len(endPoints)); i++ {
			// Check if endpoint has a pending interrupt
			inDataDone := epDataStatus&(nrf.USBD_EPDATASTATUS_EPIN1<<(i-1)) > 0
			outDataDone := epDataStatus&(nrf.USBD_EPDATASTATUS_EPOUT1<<(i-1)) > 0
			if inDataDone || outDataDone {
				switch i {
				case usb_CDC_ENDPOINT_OUT:
					// setup buffer to receive from host
					if outDataDone {
						enterCriticalSection()
						nrf.USBD.EPOUT[i].PTR.Set(uint32(uintptr(unsafe.Pointer(&udd_ep_out_cache_buffer[i]))))
						count := nrf.USBD.SIZE.EPOUT[i].Get()
						nrf.USBD.EPOUT[i].MAXCNT.Set(count)
						nrf.USBD.TASKS_STARTEPOUT[i].Set(1)
					}
				case usb_CDC_ENDPOINT_IN: //, usb_CDC_ENDPOINT_ACM:
					if inDataDone {
						usbcdc.waitTxc = false
						exitCriticalSection()
					}
				case usb_HID_ENDPOINT_IN:
					waitHidTxc = false
				}
			}
		}
	}

	// ENDEPOUT[n] events
	for i := 0; i < len(endPoints); i++ {
		if nrf.USBD.EVENTS_ENDEPOUT[i].Get() > 0 {
			nrf.USBD.EVENTS_ENDEPOUT[i].Set(0)
			if i == 0 && epout0data_setlinecoding {
				epout0data_setlinecoding = false
				count := int(nrf.USBD.SIZE.EPOUT[0].Get())
				if count >= 7 {
					parseUSBLineInfo(udd_ep_out_cache_buffer[0][:count])
					if usbLineInfo.dwDTERate == 1200 && usbLineInfo.lineState&usb_CDC_LINESTATE_DTR == 0 {
						EnterBootloader()
					}
				}
				nrf.USBD.TASKS_EP0STATUS.Set(1)
			}
			if i == usb_CDC_ENDPOINT_OUT {
				usbcdc.handleEndpoint(uint32(i))
			}
			exitCriticalSection()
		}
	}
}

func parseUSBLineInfo(b []byte) {
	usbLineInfo.dwDTERate = uint32(b[0]) | uint32(b[1])<<8 | uint32(b[2])<<16 | uint32(b[3])<<24
	usbLineInfo.bCharFormat = b[4]
	usbLineInfo.bParityType = b[5]
	usbLineInfo.bDataBits = b[6]
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
		nrf.USBD.INTENSET.Set(nrf.USBD_INTENSET_ENDEPOUT0 << ep)
		nrf.USBD.SIZE.EPOUT[ep].Set(0)
		enableEPOut(ep)

	case usb_ENDPOINT_TYPE_INTERRUPT | usbEndpointOut:
		nrf.USBD.INTENSET.Set(nrf.USBD_INTENSET_ENDEPOUT0 << ep)
		nrf.USBD.SIZE.EPOUT[ep].Set(0)
		enableEPOut(ep)

	case usb_ENDPOINT_TYPE_BULK | usbEndpointIn:
		enableEPIn(ep)

	case usb_ENDPOINT_TYPE_CONTROL:
		enableEPIn(0)
		enableEPOut(0)
		nrf.USBD.INTENSET.Set(nrf.USBD_INTENSET_ENDEPOUT0)
		nrf.USBD.TASKS_EP0STATUS.Set(1)
	}
}

func handleStandardSetup(setup usbSetup) bool {
	switch setup.bRequest {
	case usb_GET_STATUS:
		buf := []byte{0, 0}

		if setup.bmRequestType != 0 { // endpoint
			if isEndpointHalt {
				buf[0] = 1
			}
		}

		sendUSBPacket(0, buf, setup.wLength)
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
		sendUSBPacket(0, buff, setup.wLength)
		return true

	case usb_SET_CONFIGURATION:
		if setup.bmRequestType&usb_REQUEST_RECIPIENT == usb_REQUEST_DEVICE {
			nrf.USBD.TASKS_EP0STATUS.Set(1)
			for i := 1; i < len(endPoints); i++ {
				initEndpoint(uint32(i), endPoints[i])
			}

			// Enable interrupt for HID messages from host
			if hidCallback != nil {
				nrf.USBD.INTENSET.Set(nrf.USBD_INTENSET_ENDEPOUT0 << usb_HID_ENDPOINT_IN)
			}

			usbConfiguration = setup.wValueL
			return true
		} else {
			return false
		}

	case usb_GET_INTERFACE:
		buff := []byte{usbSetInterface}
		sendUSBPacket(0, buff, setup.wLength)
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
			var b [cdcLineInfoSize]byte
			b[0] = byte(usbLineInfo.dwDTERate)
			b[1] = byte(usbLineInfo.dwDTERate >> 8)
			b[2] = byte(usbLineInfo.dwDTERate >> 16)
			b[3] = byte(usbLineInfo.dwDTERate >> 24)
			b[4] = byte(usbLineInfo.bCharFormat)
			b[5] = byte(usbLineInfo.bParityType)
			b[6] = byte(usbLineInfo.bDataBits)

			sendUSBPacket(0, b[:], setup.wLength)
			return true
		}
	}

	if setup.bmRequestType == usb_REQUEST_HOSTTODEVICE_CLASS_INTERFACE {
		if setup.bRequest == usb_CDC_SET_LINE_CODING {
			epout0data_setlinecoding = true
			nrf.USBD.TASKS_EP0RCVOUT.Set(1)
			return true
		}

		if setup.bRequest == usb_CDC_SET_CONTROL_LINE_STATE {
			usbLineInfo.lineState = setup.wValueL
			if usbLineInfo.dwDTERate == 1200 && usbLineInfo.lineState&usb_CDC_LINESTATE_DTR == 0 {
				EnterBootloader()
			}
			nrf.USBD.TASKS_EP0STATUS.Set(1)
		}

		if setup.bRequest == usb_CDC_SEND_BREAK {
			nrf.USBD.TASKS_EP0STATUS.Set(1)
		}
		return true
	}
	return false
}

// SendUSBHIDPacket sends a packet for USBHID (interrupt / in).
func SendUSBHIDPacket(ep uint32, data []byte) bool {
	if waitHidTxc {
		return false
	}

	sendUSBPacket(ep, data, 0)

	// clear transfer complete flag
	nrf.USBD.INTENCLR.Set(nrf.USBD_INTENCLR_ENDEPOUT0 << 4)

	waitHidTxc = true

	return true
}

//go:noinline
func sendUSBPacket(ep uint32, data []byte, maxsize uint16) {
	count := len(data)
	if 0 < int(maxsize) && int(maxsize) < count {
		count = int(maxsize)
	}
	copy(udd_ep_in_cache_buffer[ep][:], data[:count])
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

func (usbcdc *USBCDC) handleEndpoint(ep uint32) {
	// get data
	count := int(nrf.USBD.EPOUT[ep].AMOUNT.Get())

	// move to ring buffer
	for i := 0; i < count; i++ {
		usbcdc.Receive(byte(udd_ep_out_cache_buffer[ep][i]))
	}

	// set ready for next data
	nrf.USBD.SIZE.EPOUT[ep].Set(0)
}

func sendZlp() {
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
