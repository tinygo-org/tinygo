//go:build rp2350

package machine

import (
	"device/rp"
	"machine/usb"
	"runtime/interrupt"
)

// Configure the USB peripheral. The config is here for compatibility with the UART interface.
func (dev *USBDevice) Configure(config UARTConfig) {
	// Reset usb controller
	resetBlock(rp.RESETS_RESET_USBCTRL)
	unresetBlockWait(rp.RESETS_RESET_USBCTRL)

	// Clear any previous state in dpram just in case
	_usbDPSRAM.clear()

	// Enable USB interrupt at processor
	rp.USB.INTE.Set(0)
	intr := interrupt.New(rp.IRQ_USBCTRL_IRQ, handleUSBIRQ)
	intr.SetPriority(0x00)
	intr.Enable()
	irqSet(rp.IRQ_USBCTRL_IRQ, true)

	// Mux the controller to the onboard usb phy
	rp.USB.USB_MUXING.Set(rp.USB_USB_MUXING_TO_PHY | rp.USB_USB_MUXING_SOFTCON)

	// Force VBUS detect so the device thinks it is plugged into a host
	rp.USB.USB_PWR.Set(rp.USB_USB_PWR_VBUS_DETECT | rp.USB_USB_PWR_VBUS_DETECT_OVERRIDE_EN)

	// Enable the USB controller in device mode.
	rp.USB.MAIN_CTRL.Set(rp.USB_MAIN_CTRL_CONTROLLER_EN)

	// Enable an interrupt per EP0 transaction
	rp.USB.SIE_CTRL.Set(rp.USB_SIE_CTRL_EP0_INT_1BUF)

	// Enable interrupts for when a buffer is done, when the bus is reset,
	// and when a setup packet is received
	rp.USB.INTE.Set(rp.USB_INTE_BUFF_STATUS |
		rp.USB_INTE_BUS_RESET |
		rp.USB_INTE_SETUP_REQ)

	// Present full speed device by enabling pull up on DP
	rp.USB.SIE_CTRL.SetBits(rp.USB_SIE_CTRL_PULLUP_EN)

	// 12.7.2 Disable phy isolation
	rp.USB.SetMAIN_CTRL_PHY_ISO(0x0)
}

func handleUSBIRQ(intr interrupt.Interrupt) {
	status := rp.USB.INTS.Get()

	// Setup packet received
	if (status & rp.USB_INTS_SETUP_REQ) > 0 {
		rp.USB.SIE_STATUS.Set(rp.USB_SIE_STATUS_SETUP_REC)
		setup := usb.NewSetup(_usbDPSRAM.setupBytes())

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
			// Stall endpoint?
			USBDev.SetStallEPIn(0)
		}

	}

	// Buffer status, one or more buffers have completed
	if (status & rp.USB_INTS_BUFF_STATUS) > 0 {
		if sendOnEP0DATADONE.offset > 0 {
			ep := uint32(0)
			data := sendOnEP0DATADONE.data
			count := len(data) - sendOnEP0DATADONE.offset
			if ep == 0 && count > usb.EndpointPacketSize {
				count = usb.EndpointPacketSize
			}

			sendViaEPIn(ep, data[sendOnEP0DATADONE.offset:], count)
			sendOnEP0DATADONE.offset += count
			if sendOnEP0DATADONE.offset == len(data) {
				sendOnEP0DATADONE.offset = 0
			}
		}

		s2 := rp.USB.BUFF_STATUS.Get()

		// OUT (PC -> rp2350)
		for i := 0; i < 16; i++ {
			if s2&(1<<(i*2+1)) > 0 {
				buf := handleEndpointRx(uint32(i))
				if usbRxHandler[i] == nil || usbRxHandler[i](buf) {
					AckUsbOutTransfer(uint32(i))
				}
			}
		}

		// IN (rp2350 -> PC)
		for i := 0; i < 16; i++ {
			if s2&(1<<(i*2)) > 0 {
				if usbTxHandler[i] != nil {
					usbTxHandler[i]()
				}
			}
		}

		rp.USB.BUFF_STATUS.Set(s2)
	}

	// Bus is reset
	if (status & rp.USB_INTS_BUS_RESET) > 0 {
		rp.USB.SIE_STATUS.Set(rp.USB_SIE_STATUS_BUS_RESET)
		//fixRP2040UsbDeviceEnumeration()

		rp.USB.ADDR_ENDP.Set(0)
		initEndpoint(0, usb.ENDPOINT_TYPE_CONTROL)
	}
}

func handleUSBSetAddress(setup usb.Setup) bool {
	// Using 570μs timeout which is exactly the same as SAMD21.
	const ackTimeout = 570

	rp.USB.SIE_STATUS.Set(rp.USB_SIE_STATUS_ACK_REC)
	sendUSBPacket(0, []byte{}, 0)

	// Wait for transfer to complete with a timeout.
	t := timer.timeElapsed()
	for (rp.USB.SIE_STATUS.Get() & rp.USB_SIE_STATUS_ACK_REC) == 0 {
		if dt := timer.timeElapsed() - t; dt >= ackTimeout {
			return false
		}
	}

	// Set the device address to that requested by host.
	rp.USB.ADDR_ENDP.Set(uint32(setup.WValueL) & rp.USB_ADDR_ENDP_ADDRESS_Msk)
	return true
}

func armEPZeroStall() {
	rp.USB.EP_STALL_ARM.Set(rp.USB_EP_STALL_ARM_EP0_IN)
}
