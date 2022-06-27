//go:build rp2040
// +build rp2040

package machine

import (
	"device/arm"
	"device/rp"
	"runtime/interrupt"
	"unsafe"
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
	usbDPSRAM.clear()

	//// Enable USB interrupt at processor
	rp.USBCTRL_REGS.INTE.Set(0)
	intr := interrupt.New(rp.IRQ_USBCTRL_IRQ, handleUSBIRQ)
	intr.SetPriority(0x00)
	intr.Enable()
	irqSet(rp.IRQ_USBCTRL_IRQ, true)

	//// Mux the controller to the onboard usb phy
	rp.USBCTRL_REGS.USB_MUXING.Set(rp.USBCTRL_REGS_USB_MUXING_TO_PHY | rp.USBCTRL_REGS_USB_MUXING_SOFTCON)

	//// Force VBUS detect so the device thinks it is plugged into a host
	rp.USBCTRL_REGS.USB_PWR.Set(rp.USBCTRL_REGS_USB_PWR_VBUS_DETECT | rp.USBCTRL_REGS_USB_PWR_VBUS_DETECT_OVERRIDE_EN)

	//// Enable the USB controller in device mode.
	rp.USBCTRL_REGS.MAIN_CTRL.Set(rp.USBCTRL_REGS_MAIN_CTRL_CONTROLLER_EN)

	//// Enable an interrupt per EP0 transaction
	rp.USBCTRL_REGS.SIE_CTRL.Set(rp.USBCTRL_REGS_SIE_CTRL_EP0_INT_1BUF)

	//// Enable interrupts for when a buffer is done, when the bus is reset,
	//// and when a setup packet is received
	rp.USBCTRL_REGS.INTE.Set(rp.USBCTRL_REGS_INTE_BUFF_STATUS |
		rp.USBCTRL_REGS_INTE_BUS_RESET |
		rp.USBCTRL_REGS_INTE_SETUP_REQ)

	//// Present full speed device by enabling pull up on DP
	rp.USBCTRL_REGS.SIE_CTRL.SetBits(rp.USBCTRL_REGS_SIE_CTRL_PULLUP_EN)
}

func handleUSBIRQ(intr interrupt.Interrupt) {
	status := rp.USBCTRL_REGS.INTS.Get()

	// Setup packet received
	if (status & rp.USBCTRL_REGS_INTS_SETUP_REQ) > 0 {
		rp.USBCTRL_REGS.SIE_STATUS.Set(rp.USBCTRL_REGS_SIE_STATUS_SETUP_REC)
		setup := newUSBSetup(usbDPSRAM.setupBytes())

		ok := false
		if (setup.BmRequestType & usb_REQUEST_TYPE) == usb_REQUEST_STANDARD {
			// Standard Requests
			ok = handleStandardSetup(setup)
		} else {
			// Class Interface Requests
			if setup.WIndex < uint16(len(usbSetupHandler)) && usbSetupHandler[setup.WIndex] != nil {
				ok = usbSetupHandler[setup.WIndex](setup)
			}
		}

		if ok {
			// set Bank1 ready?
		} else {
			// Stall endpoint?
		}

	}

	// Buffer status, one or more buffers have completed
	if (status & rp.USBCTRL_REGS_INTS_BUFF_STATUS) > 0 {
		if sendOnEP0DATADONE.offset > 0 {
			ep := uint32(0)
			data := sendOnEP0DATADONE.data
			count := len(data) - sendOnEP0DATADONE.offset
			if ep == 0 && count > usbEndpointPacketSize {
				count = usbEndpointPacketSize
			}

			sendViaEPIn(ep, data[sendOnEP0DATADONE.offset:], count)
			sendOnEP0DATADONE.offset += count
			if sendOnEP0DATADONE.offset == len(data) {
				sendOnEP0DATADONE.offset = 0
			}
		}

		s2 := rp.USBCTRL_REGS.BUFF_STATUS.Get()

		// OUT (PC -> rp2040)
		for i := 0; i < 16; i++ {
			if s2&(1<<(i*2+1)) > 0 {
				buf := handleEndpointRx(uint32(i))
				if usbRxHandler[i] != nil {
					usbRxHandler[i](buf)
				}
			}
		}

		// IN (rp2040 -> PC)
		for i := 0; i < 16; i++ {
			if s2&(1<<(i*2)) > 0 {
				if usbTxHandler[i] != nil {
					usbTxHandler[i]()
				}
			}
		}

		rp.USBCTRL_REGS.BUFF_STATUS.Set(0xFFFFFFFF)
	}

	// Bus is reset
	if (status & rp.USBCTRL_REGS_INTS_BUS_RESET) > 0 {
		rp.USBCTRL_REGS.SIE_STATUS.Set(rp.USBCTRL_REGS_SIE_STATUS_BUS_RESET)
		initEndpoint(0, usb_ENDPOINT_TYPE_CONTROL)
	}
}

func initEndpoint(ep, config uint32) {
	val := uint32(0x80000000) | uint32(0x20000000)
	offset := ep*2*64 + 0x100
	val |= offset

	switch config {
	case usb_ENDPOINT_TYPE_INTERRUPT | usbEndpointIn:
		val |= 0x0C000000
		usbDPSRAM.EPxControl[ep].In = val

	case usb_ENDPOINT_TYPE_BULK | usbEndpointOut:
		val |= 0x08000000
		usbDPSRAM.EPxControl[ep].Out = val
		usbDPSRAM.EPxBufferControl[ep].Out = 0x00000040
		usbDPSRAM.EPxBufferControl[ep].Out |= 0x00000400

	case usb_ENDPOINT_TYPE_INTERRUPT | usbEndpointOut:
		// TODO: not really anything, seems like...

	case usb_ENDPOINT_TYPE_BULK | usbEndpointIn:
		val |= 0x08000000
		usbDPSRAM.EPxControl[ep].In = val

	case usb_ENDPOINT_TYPE_CONTROL:
		usbDPSRAM.EPxBufferControl[ep].Out = 0x00000400

	}
}

func handleUSBSetAddress(setup USBSetup) bool {
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
		epXdata0[ep] = true
	}

	sendViaEPIn(ep, data, count)
}

func ReceiveUSBControlPacket() ([cdcLineInfoSize]byte, error) {
	var b [cdcLineInfoSize]byte
	return b, nil
}

func handleEndpointRx(ep uint32) []byte {
	ctrl := usbDPSRAM.EPxBufferControl[ep].Out
	usbDPSRAM.EPxBufferControl[ep].Out = 0x00000040
	sz := ctrl & 0x0000003F
	buf := make([]byte, sz)
	copy(buf, usbDPSRAM.EPxBuffer[ep].Buffer0[:sz])

	epXdata0[ep] = !epXdata0[ep]
	if epXdata0[ep] {
		usbDPSRAM.EPxBufferControl[ep].Out |= 0x00002000
	}

	usbDPSRAM.EPxBufferControl[ep].Out |= 0x00000400

	return buf
}

func SendZlp() {
	sendUSBPacket(0, []byte{}, 0)
}

func sendViaEPIn(ep uint32, data []byte, count int) {
	// Prepare buffer control register value
	val := uint32(count) | 0x00000400

	// DATA0 or DATA1
	epXdata0[ep&0x7F] = !epXdata0[ep&0x7F]
	if !epXdata0[ep&0x7F] {
		val |= 0x00002000
	}

	// Mark as full
	val |= 0x00008000

	copy(usbDPSRAM.EPxBuffer[ep&0x7F].Buffer0[:], data[:count])
	usbDPSRAM.EPxBufferControl[ep&0x7F].In = val
}

// EnterBootloader should perform a system reset in preparation
// to switch to the bootloader to flash new firmware.
func EnterBootloader() {
	arm.DisableInterrupts()

	// TODO: Perform magic reset into bootloader

	arm.SystemReset()
}

type USBDPSRAM struct {
	// Note that EPxControl[0] is not EP0Control but 8-byte setup data.
	EPxControl [16]USBEndpointControlRegister

	EPxBufferControl [16]USBBufferControlRegister

	EPxBuffer [16]USBBuffer
}

type USBEndpointControlRegister struct {
	In  uint32
	Out uint32
}
type USBBufferControlRegister struct {
	In  uint32
	Out uint32
}

type USBBuffer struct {
	Buffer0 [64]byte
	Buffer1 [64]byte
}

var (
	usbDPSRAM = (*USBDPSRAM)(unsafe.Pointer(uintptr(0x50100000)))
	epXdata0  [16]bool
)

func (d *USBDPSRAM) setupBytes() []byte {
	var buf [8]byte

	buf[0] = byte(d.EPxControl[usb_CONTROL_ENDPOINT].In)
	buf[1] = byte(d.EPxControl[usb_CONTROL_ENDPOINT].In >> 8)
	buf[2] = byte(d.EPxControl[usb_CONTROL_ENDPOINT].In >> 16)
	buf[3] = byte(d.EPxControl[usb_CONTROL_ENDPOINT].In >> 24)
	buf[4] = byte(d.EPxControl[usb_CONTROL_ENDPOINT].Out)
	buf[5] = byte(d.EPxControl[usb_CONTROL_ENDPOINT].Out >> 8)
	buf[6] = byte(d.EPxControl[usb_CONTROL_ENDPOINT].Out >> 16)
	buf[7] = byte(d.EPxControl[usb_CONTROL_ENDPOINT].Out >> 24)

	return buf[:]
}

func (d *USBDPSRAM) clear() {
	for i := 0; i < len(d.EPxControl); i++ {
		d.EPxControl[i].In = 0
		d.EPxControl[i].Out = 0
		d.EPxBufferControl[i].In = 0
		d.EPxBufferControl[i].Out = 0
	}
}
