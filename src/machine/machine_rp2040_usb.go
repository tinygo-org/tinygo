//go:build rp2040
// +build rp2040

package machine

import (
	"device/arm"
	"device/rp"
	"runtime/interrupt"
	"unsafe"
)

const (
// these are rp2040 specific.
)

var (
	sendOnEP0DATADONE struct {
		offset int
		data   []byte
		pid    uint32
	}
)

// Configure the USB peripheral. The config is here for compatibility with the UART interface.
func (dev *USBDevice) Configure(config UARTConfig) {
	//// Clear any previous state in dpram just in case
	//memset(usb_dpram, 0, sizeof(*usb_dpram)); // <1>
	usbDPSRAM.clear()

	//// Enable USB interrupt at processor
	//irq_set_enabled(USBCTRL_IRQ, true);
	rp.USBCTRL_REGS.INTE.Set(0)
	intr := interrupt.New(rp.IRQ_USBCTRL_IRQ, handleUSBIRQ)
	intr.SetPriority(0x00)
	intr.Enable()
	irqSet(rp.IRQ_USBCTRL_IRQ, true)

	//// Mux the controller to the onboard usb phy
	//usb_hw->muxing = USB_USB_MUXING_TO_PHY_BITS | USB_USB_MUXING_SOFTCON_BITS;
	rp.USBCTRL_REGS.USB_MUXING.Set(rp.USBCTRL_REGS_USB_MUXING_TO_PHY | rp.USBCTRL_REGS_USB_MUXING_SOFTCON)

	//// Force VBUS detect so the device thinks it is plugged into a host
	//usb_hw->pwr = USB_USB_PWR_VBUS_DETECT_BITS | USB_USB_PWR_VBUS_DETECT_OVERRIDE_EN_BITS;
	rp.USBCTRL_REGS.USB_PWR.Set(rp.USBCTRL_REGS_USB_PWR_VBUS_DETECT | rp.USBCTRL_REGS_USB_PWR_VBUS_DETECT_OVERRIDE_EN)

	//// Enable the USB controller in device mode.
	//usb_hw->main_ctrl = USB_MAIN_CTRL_CONTROLLER_EN_BITS;
	rp.USBCTRL_REGS.MAIN_CTRL.Set(rp.USBCTRL_REGS_MAIN_CTRL_CONTROLLER_EN)

	//// Enable an interrupt per EP0 transaction
	//usb_hw->sie_ctrl = USB_SIE_CTRL_EP0_INT_1BUF_BITS; // <2>
	rp.USBCTRL_REGS.SIE_CTRL.Set(rp.USBCTRL_REGS_SIE_CTRL_EP0_INT_1BUF)

	//// Enable interrupts for when a buffer is done, when the bus is reset,
	//// and when a setup packet is received
	//usb_hw->inte = USB_INTS_BUFF_STATUS_BITS |
	//               USB_INTS_BUS_RESET_BITS |
	//               USB_INTS_SETUP_REQ_BITS;
	rp.USBCTRL_REGS.INTE.Set(rp.USBCTRL_REGS_INTE_BUFF_STATUS |
		rp.USBCTRL_REGS_INTE_BUS_RESET |
		rp.USBCTRL_REGS_INTE_SETUP_REQ)

	//// Set up endpoints (endpoint control registers)
	//// described by device configuration
	//usb_setup_endpoints();
	// void usb_setup_endpoints() {
	//     const struct usb_endpoint_configuration *endpoints = dev_config.endpoints;
	//     for (int i = 0; i < USB_NUM_ENDPOINTS; i++) {
	//         if (endpoints[i].descriptor && endpoints[i].handler) {
	//             usb_setup_endpoint(&endpoints[i]);
	//         }
	//     }
	// }
	//
	// void usb_setup_endpoint(const struct usb_endpoint_configuration *ep) {
	//     printf("Set up endpoint 0x%x with buffer address 0x%p\n", ep->descriptor->bEndpointAddress, ep->data_buffer);
	//
	//     // EP0 doesn't have one so return if that is the case
	//     if (!ep->endpoint_control) {
	//         return;
	//     }
	//
	//     // Get the data buffer as an offset of the USB controller's DPRAM
	//     uint32_t dpram_offset = usb_buffer_offset(ep->data_buffer);
	//     uint32_t reg = EP_CTRL_ENABLE_BITS
	//                    | EP_CTRL_INTERRUPT_PER_BUFFER
	//                    | (ep->descriptor->bmAttributes << EP_CTRL_BUFFER_TYPE_LSB)
	//                    | dpram_offset;
	//     *ep->endpoint_control = reg;
	// }

	//// Present full speed device by enabling pull up on DP
	//usb_hw_set->sie_ctrl = USB_SIE_CTRL_PULLUP_EN_BITS;
	rp.USBCTRL_REGS.SIE_CTRL.SetBits(rp.USBCTRL_REGS_SIE_CTRL_PULLUP_EN)

	// 追加
	val := uint32(0)
	val |= 0x00000400
	usbDPSRAM.EP0OutBufferControl = USBBufferControlRegister(val)
}

func handleUSBIRQ(intr interrupt.Interrupt) {
	status := rp.USBCTRL_REGS.INTS.Get()

	// setup
	//rp.USBCTRL_REGS.SIE_STATUS.Set(0xFFFFFFFF)

	//if false {
	//	// SOF
	//	x := rp.USBCTRL_REGS.SOF_RD.Get()
	//}

	// void isr_usbctrl(void) {
	//     // USB interrupt handler
	//     uint32_t status = usb_hw->ints;
	//     uint32_t handled = 0;

	//     // Setup packet received
	//     if (status & USB_INTS_SETUP_REQ_BITS) {
	//         handled |= USB_INTS_SETUP_REQ_BITS;
	//         usb_hw_clear->sie_status = USB_SIE_STATUS_SETUP_REC_BITS;
	//         usb_handle_setup_packet();
	//     }
	// /// \end::isr_setup_packet[]
	if (status & rp.USBCTRL_REGS_INTS_SETUP_REQ) > 0 {
		rp.USBCTRL_REGS.SIE_STATUS.Set(rp.USBCTRL_REGS_SIE_STATUS_SETUP_REC)
		setup := newUSBSetup(usbDPSRAM.Setup[:])

		ok := false
		if (setup.BmRequestType & usb_REQUEST_TYPE) == usb_REQUEST_STANDARD {
			// Standard Requests
			ok = handleStandardSetup(setup)
		} else {
			// Class Interface Requests
			if setup.WIndex < uint16(len(callbackUSBSetup)) && callbackUSBSetup[setup.WIndex] != nil {
				ok = callbackUSBSetup[setup.WIndex](setup)
			}
		}

		if ok {
			// set Bank1 ready
			//setEPSTATUSSET(0, sam.USB_DEVICE_ENDPOINT_EPSTATUSSET_BK1RDY)
		} else {
			// Stall endpoint
			//setEPSTATUSSET(0, sam.USB_DEVICE_ENDPOINT_EPINTFLAG_STALL1)
		}

		//if getEPINTFLAG(0)&sam.USB_DEVICE_ENDPOINT_EPINTFLAG_STALL1 > 0 {
		//	// ack the stall
		//	setEPINTFLAG(0, sam.USB_DEVICE_ENDPOINT_EPINTFLAG_STALL1)

		//	// clear stall request
		//	setEPINTENCLR(0, sam.USB_DEVICE_ENDPOINT_EPINTENCLR_STALL1)
		//}
	}

	//     // Buffer status, one or more buffers have completed
	//     if (status & USB_INTS_BUFF_STATUS_BITS) {
	//         handled |= USB_INTS_BUFF_STATUS_BITS;
	//         usb_handle_buff_status();
	//     }
	if (status & rp.USBCTRL_REGS_INTS_BUFF_STATUS) > 0 {
		if sendOnEP0DATADONE.offset > 0 {
			ep := uint32(0)
			data := sendOnEP0DATADONE.data
			count := len(data) - sendOnEP0DATADONE.offset
			if ep == 0 && count > usbEndpointPacketSize {
				count = usbEndpointPacketSize
			}
			sendViaEPIn(ep, data[sendOnEP0DATADONE.offset:], count, 0)
			sendOnEP0DATADONE.offset += count
			if sendOnEP0DATADONE.offset == len(data) {
				sendOnEP0DATADONE.offset = 0
			}
		}
		//s2 := rp.USBCTRL_REGS.BUFF_STATUS.Get()
		//if (s2 & 0x00000001) > 0 {
		//	// EP0_IN
		//} else if (s2 & 0x00000002) > 0 {
		//	// EP0_OUT
		//	//sendUSBPacket(0, []byte{}, 0)
		//}
		rp.USBCTRL_REGS.BUFF_STATUS.Set(0xFFFFFFFF)
	}

	//     // Bus is reset
	//     if (status & USB_INTS_BUS_RESET_BITS) {
	//         printf("BUS RESET\n");
	//         handled |= USB_INTS_BUS_RESET_BITS;
	//         usb_hw_clear->sie_status = USB_SIE_STATUS_BUS_RESET_BITS;
	//         usb_bus_reset();
	//     }
	// void usb_bus_reset(void) {
	//     // Set address back to 0
	//     dev_addr = 0;
	//     should_set_address = false;
	//     usb_hw->dev_addr_ctrl = 0;
	//     configured = false;
	// }
	if (status & rp.USBCTRL_REGS_INTS_BUS_RESET) > 0 {
		rp.USBCTRL_REGS.SIE_STATUS.Set(rp.USBCTRL_REGS_SIE_STATUS_BUS_RESET)
	}

	//
	//     if (status ^ handled) {
	//         panic("Unhandled IRQ 0x%x\n", (uint) (status ^ handled));
	//     }
	// }
}

func initEndpoint(ep, config uint32) {
	if ep == 0 {
		return
	}

	val := uint32(0x80000000) | uint32(0x20000000)
	offset := (ep-1)*64 + 0x180
	val |= offset

	switch config {
	case usb_ENDPOINT_TYPE_INTERRUPT | usbEndpointIn:
		val |= 0x0C000000
		// ep1 in
		usbDPSRAM.EP1InControl = USBEndpointControlRegister(val)
	case usb_ENDPOINT_TYPE_BULK | usbEndpointOut:
		val |= 0x08000000
		// ep2 out
		usbDPSRAM.EP2OutControl = USBEndpointControlRegister(val)
	case usb_ENDPOINT_TYPE_INTERRUPT | usbEndpointOut:
		// TODO: not really anything, seems like...

	case usb_ENDPOINT_TYPE_BULK | usbEndpointIn:
		val |= 0x08000000
		usbDPSRAM.EP3InControl = USBEndpointControlRegister(val)
		// ep3 in
	case usb_ENDPOINT_TYPE_CONTROL:
		// skip
	}
}

func handleUSBSetAddress(setup USBSetup) bool {
	//	setup.BmRequestType, setup.BRequest, setup.WValueL, setup.WValueH, setup.WIndex, setup.WLength)
	sendUSBPacket(0, []byte{}, 0)

	// last, set the device address to that requested by host
	// wait for transfer to complete
	timeout := 3000
	rp.USBCTRL_REGS.SIE_STATUS.Set(rp.USBCTRL_REGS_SIE_STATUS_ACK_REC)
	for (rp.USBCTRL_REGS.SIE_STATUS.Get() & rp.USBCTRL_REGS_SIE_STATUS_ACK_REC) == 0 {
		timeout--
		if timeout == 0 {
			return true
		}
	}

	rp.USBCTRL_REGS.ADDR_ENDP.Set(uint32(setup.WValueL) & rp.USBCTRL_REGS_ADDR_ENDP_ADDRESS_Msk)

	return true
}

// SendUSBInPacket sends a packet for USB (interrupt in / bulk in).
func SendUSBInPacket(ep uint32, data []byte) bool {
	sendUSBPacket(ep, data, 0)
	return true
}

//go:noinline
func sendUSBPacket(ep uint32, data []byte, maxsize uint16) {
	count := len(data)
	if 0 < int(maxsize) && int(maxsize) < count {
		count = int(maxsize)
	}

	if ep == 0 {
		if count > usbEndpointPacketSize {
			count = usbEndpointPacketSize

			sendOnEP0DATADONE.offset = count
			sendOnEP0DATADONE.data = data
		} else {
			sendOnEP0DATADONE.offset = 0
		}
	}

	sendViaEPIn(ep, data, count, 1)
}

func ReceiveUSBControlPacket() ([cdcLineInfoSize]byte, error) {
	var b [cdcLineInfoSize]byte
	return b, nil
}

func handleEndpointRx(ep uint32) []byte {
	return nil
}

func SendZlp() {
	sendUSBPacket(0, []byte{}, 0)
}

var ep3data0 = true

func sendViaEPIn(ep uint32, data []byte, count int, pid int) {
	// void usb_start_transfer(struct usb_endpoint_configuration *ep, uint8_t *buf, uint16_t len) {
	//     // We are asserting that the length is <= 64 bytes for simplicity of the example.
	//     // For multi packet transfers see the tinyusb port.
	//     assert(len <= 64);
	//
	//     printf("Start transfer of len %d on ep addr 0x%x\n", len, ep->descriptor->bEndpointAddress);
	//
	//     // Prepare buffer control register value
	//     uint32_t val = len | USB_BUF_CTRL_AVAIL;
	val := uint32(count) | 0x00000400

	//     if (ep_is_tx(ep)) {
	//         // Need to copy the data from the user buffer to the usb memory
	//         memcpy((void *) ep->data_buffer, (void *) buf, len);
	// DATA0 or DATA1
	if ep == 3 {
		if ep3data0 {
			val |= 0x00002000
		}
		ep3data0 = !ep3data0
	} else if pid == 1 {
		val |= 0x00002000
	}

	//         // Mark as full
	//         val |= USB_BUF_CTRL_FULL;
	val |= 0x00008000

	//     }
	//     // Set pid and flip for next transfer
	//     val |= ep->next_pid ? USB_BUF_CTRL_DATA1_PID : USB_BUF_CTRL_DATA0_PID;
	//     ep->next_pid ^= 1u;
	//
	//     *ep->buffer_control = val;
	switch ep & 0x7F {
	case 0:
		copy(usbDPSRAM.EP0Buffer0[:], data[:count])
		usbDPSRAM.EP0InBufferControl = USBBufferControlRegister(val)
	case 1:
		copy(usbDPSRAM.EP1Buffer[:], data[:count])
		usbDPSRAM.EP1InBufferControl = USBBufferControlRegister(val)
	case 2:
		usbDPSRAM.EP2OutBufferControl = USBBufferControlRegister(val)
	case 3:
		copy(usbDPSRAM.EP3Buffer[:], data[:count])
		usbDPSRAM.EP3InBufferControl = USBBufferControlRegister(val)
	default:
	}

	// }
}

// ResetProcessor should perform a system reset in preparation
// to switch to the bootloader to flash new firmware.
func ResetProcessor() {
	arm.DisableInterrupts()

	//// Perform magic reset into bootloader, as mentioned in
	//// https://github.com/arduino/ArduinoCore-samd/issues/197
	//*(*uint32)(unsafe.Pointer(uintptr(0x20000000 + HSRAM_SIZE - 4))) = RESET_MAGIC_VALUE

	arm.SystemReset()
}

type USBDPSRAM struct {
	Setup [8]byte

	EP1InControl   USBEndpointControlRegister // 0x0008
	EP1OutControl  USBEndpointControlRegister // 0x000c
	EP2InControl   USBEndpointControlRegister // 0x0010
	EP2OutControl  USBEndpointControlRegister // 0x0014
	EP3InControl   USBEndpointControlRegister // 0x0018
	EP3OutControl  USBEndpointControlRegister // 0x001c
	EP4InControl   USBEndpointControlRegister // 0x0020
	EP4OutControl  USBEndpointControlRegister // 0x0024
	EP5InControl   USBEndpointControlRegister // 0x0028
	EP5OutControl  USBEndpointControlRegister // 0x002c
	EP6InControl   USBEndpointControlRegister // 0x0030
	EP6OutControl  USBEndpointControlRegister // 0x0034
	EP7InControl   USBEndpointControlRegister // 0x0038
	EP7OutControl  USBEndpointControlRegister // 0x003c
	EP8InControl   USBEndpointControlRegister // 0x0040
	EP8OutControl  USBEndpointControlRegister // 0x0044
	EP9InControl   USBEndpointControlRegister // 0x0048
	EP9OutControl  USBEndpointControlRegister // 0x004c
	EP10InControl  USBEndpointControlRegister // 0x0050
	EP10OutControl USBEndpointControlRegister // 0x0054
	EP11InControl  USBEndpointControlRegister // 0x0058
	EP11OutControl USBEndpointControlRegister // 0x005c
	EP12InControl  USBEndpointControlRegister // 0x0060
	EP12OutControl USBEndpointControlRegister // 0x0064
	EP13InControl  USBEndpointControlRegister // 0x0068
	EP13OutControl USBEndpointControlRegister // 0x006c
	EP14InControl  USBEndpointControlRegister // 0x0070
	EP14OutControl USBEndpointControlRegister // 0x0074
	EP15InControl  USBEndpointControlRegister // 0x0078
	EP15OutControl USBEndpointControlRegister // 0x007c

	EP0InBufferControl   USBBufferControlRegister // 0x0080
	EP0OutBufferControl  USBBufferControlRegister // 0x0084
	EP1InBufferControl   USBBufferControlRegister // 0x0088
	EP1OutBufferControl  USBBufferControlRegister // 0x008c
	EP2InBufferControl   USBBufferControlRegister // 0x0090
	EP2OutBufferControl  USBBufferControlRegister // 0x0094
	EP3InBufferControl   USBBufferControlRegister // 0x0098
	EP3OutBufferControl  USBBufferControlRegister // 0x009c
	EP4InBufferControl   USBBufferControlRegister // 0x00a0
	EP4OutBufferControl  USBBufferControlRegister // 0x00a4
	EP5InBufferControl   USBBufferControlRegister // 0x00a8
	EP5OutBufferControl  USBBufferControlRegister // 0x00ac
	EP6InBufferControl   USBBufferControlRegister // 0x00b0
	EP6OutBufferControl  USBBufferControlRegister // 0x00b4
	EP7InBufferControl   USBBufferControlRegister // 0x00b8
	EP7OutBufferControl  USBBufferControlRegister // 0x00bc
	EP8InBufferControl   USBBufferControlRegister // 0x00c0
	EP8OutBufferControl  USBBufferControlRegister // 0x00c4
	EP9InBufferControl   USBBufferControlRegister // 0x00c8
	EP9OutBufferControl  USBBufferControlRegister // 0x00cc
	EP10InBufferControl  USBBufferControlRegister // 0x00d0
	EP10OutBufferControl USBBufferControlRegister // 0x00d4
	EP11InBufferControl  USBBufferControlRegister // 0x00d8
	EP11OutBufferControl USBBufferControlRegister // 0x00dc
	EP12InBufferControl  USBBufferControlRegister // 0x00e0
	EP12OutBufferControl USBBufferControlRegister // 0x00e4
	EP13InBufferControl  USBBufferControlRegister // 0x00e8
	EP13OutBufferControl USBBufferControlRegister // 0x00ec
	EP14InBufferControl  USBBufferControlRegister // 0x00f0
	EP14OutBufferControl USBBufferControlRegister // 0x00f4
	EP15InBufferControl  USBBufferControlRegister // 0x00f8
	EP15OutBufferControl USBBufferControlRegister // 0x00fc

	// offset : 0x100
	EP0Buffer0 [64]byte
	EP0Buffer1 [64]byte

	// offset : 0x180 ..
	EP1Buffer [64]byte
	EP2Buffer [64]byte
	EP3Buffer [64]byte
	EP4Buffer [64]byte
	EP5Buffer [64]byte
	EP6Buffer [64]byte
	EP7Buffer [64]byte
}

type USBEndpointControlRegister uint32
type USBBufferControlRegister uint32

var usbDPSRAM = (*USBDPSRAM)(unsafe.Pointer(uintptr(0x50100000)))

func (d *USBDPSRAM) clear() {
	for i := range d.Setup {
		d.Setup[i] = 0
	}

	d.EP1InControl = 0
	d.EP1OutControl = 0
	d.EP2InControl = 0
	d.EP2OutControl = 0
	d.EP3InControl = 0
	d.EP3OutControl = 0
	d.EP4InControl = 0
	d.EP4OutControl = 0
	d.EP5InControl = 0
	d.EP5OutControl = 0
	d.EP6InControl = 0
	d.EP6OutControl = 0
	d.EP7InControl = 0
	d.EP7OutControl = 0
	d.EP8InControl = 0
	d.EP8OutControl = 0
	d.EP9InControl = 0
	d.EP9OutControl = 0
	d.EP10InControl = 0
	d.EP10OutControl = 0
	d.EP11InControl = 0
	d.EP11OutControl = 0
	d.EP12InControl = 0
	d.EP12OutControl = 0
	d.EP13InControl = 0
	d.EP13OutControl = 0
	d.EP14InControl = 0
	d.EP14OutControl = 0
	d.EP15InControl = 0
	d.EP15OutControl = 0

	d.EP0InBufferControl = 0
	d.EP0OutBufferControl = 0
	d.EP1InBufferControl = 0
	d.EP1OutBufferControl = 0
	d.EP2InBufferControl = 0
	d.EP2OutBufferControl = 0
	d.EP3InBufferControl = 0
	d.EP3OutBufferControl = 0
	d.EP4InBufferControl = 0
	d.EP4OutBufferControl = 0
	d.EP5InBufferControl = 0
	d.EP5OutBufferControl = 0
	d.EP6InBufferControl = 0
	d.EP6OutBufferControl = 0
	d.EP7InBufferControl = 0
	d.EP7OutBufferControl = 0
	d.EP8InBufferControl = 0
	d.EP8OutBufferControl = 0
	d.EP9InBufferControl = 0
	d.EP9OutBufferControl = 0
	d.EP10InBufferControl = 0
	d.EP10OutBufferControl = 0
	d.EP11InBufferControl = 0
	d.EP11OutBufferControl = 0
	d.EP12InBufferControl = 0
	d.EP12OutBufferControl = 0
	d.EP13InBufferControl = 0
	d.EP13OutBufferControl = 0
	d.EP14InBufferControl = 0
	d.EP14OutBufferControl = 0
	d.EP15InBufferControl = 0
	d.EP15OutBufferControl = 0
}
