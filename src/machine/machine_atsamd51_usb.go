//go:build (sam && atsamd51) || (sam && atsame5x)
// +build sam,atsamd51 sam,atsame5x

package machine

import (
	"device/sam"
	"runtime/interrupt"
	"unsafe"
)

const (
	// these are SAMD51 specific.
	usb_DEVICE_PCKSIZE_BYTE_COUNT_Pos  = 0
	usb_DEVICE_PCKSIZE_BYTE_COUNT_Mask = 0x3FFF

	usb_DEVICE_PCKSIZE_SIZE_Pos  = 28
	usb_DEVICE_PCKSIZE_SIZE_Mask = 0x7

	usb_DEVICE_PCKSIZE_MULTI_PACKET_SIZE_Pos  = 14
	usb_DEVICE_PCKSIZE_MULTI_PACKET_SIZE_Mask = 0x3FFF
)

// Configure the USB peripheral. The config is here for compatibility with the UART interface.
func (dev *USBDevice) Configure(config UARTConfig) {
	if dev.initcomplete {
		return
	}

	// reset USB interface
	sam.USB_DEVICE.CTRLA.SetBits(sam.USB_DEVICE_CTRLA_SWRST)
	for sam.USB_DEVICE.SYNCBUSY.HasBits(sam.USB_DEVICE_SYNCBUSY_SWRST) ||
		sam.USB_DEVICE.SYNCBUSY.HasBits(sam.USB_DEVICE_SYNCBUSY_ENABLE) {
	}

	sam.USB_DEVICE.DESCADD.Set(uint32(uintptr(unsafe.Pointer(&usbEndpointDescriptors))))

	// configure pins
	USBCDC_DM_PIN.Configure(PinConfig{Mode: PinCom})
	USBCDC_DP_PIN.Configure(PinConfig{Mode: PinCom})

	// performs pad calibration from store fuses
	handlePadCalibration()

	// run in standby
	sam.USB_DEVICE.CTRLA.SetBits(sam.USB_DEVICE_CTRLA_RUNSTDBY)

	// set full speed
	sam.USB_DEVICE.CTRLB.SetBits(sam.USB_DEVICE_CTRLB_SPDCONF_FS << sam.USB_DEVICE_CTRLB_SPDCONF_Pos)

	// attach
	sam.USB_DEVICE.CTRLB.ClearBits(sam.USB_DEVICE_CTRLB_DETACH)

	// enable interrupt for end of reset
	sam.USB_DEVICE.INTENSET.SetBits(sam.USB_DEVICE_INTENSET_EORST)

	// enable interrupt for start of frame
	sam.USB_DEVICE.INTENSET.SetBits(sam.USB_DEVICE_INTENSET_SOF)

	// enable USB
	sam.USB_DEVICE.CTRLA.SetBits(sam.USB_DEVICE_CTRLA_ENABLE)

	// enable IRQ
	interrupt.New(sam.IRQ_USB_OTHER, handleUSBIRQ).Enable()
	interrupt.New(sam.IRQ_USB_SOF_HSOF, handleUSBIRQ).Enable()
	interrupt.New(sam.IRQ_USB_TRCPT0, handleUSBIRQ).Enable()
	interrupt.New(sam.IRQ_USB_TRCPT1, handleUSBIRQ).Enable()

	dev.initcomplete = true
}

func handlePadCalibration() {
	// Load Pad Calibration data from non-volatile memory
	// This requires registers that are not included in the SVD file.
	// Modeled after defines from samd21g18a.h and nvmctrl.h:
	//
	// #define NVMCTRL_OTP4 0x00806020
	//
	// #define USB_FUSES_TRANSN_ADDR       (NVMCTRL_OTP4 + 4)
	// #define USB_FUSES_TRANSN_Pos        13           /**< \brief (NVMCTRL_OTP4) USB pad Transn calibration */
	// #define USB_FUSES_TRANSN_Msk        (0x1Fu << USB_FUSES_TRANSN_Pos)
	// #define USB_FUSES_TRANSN(value)     ((USB_FUSES_TRANSN_Msk & ((value) << USB_FUSES_TRANSN_Pos)))

	// #define USB_FUSES_TRANSP_ADDR       (NVMCTRL_OTP4 + 4)
	// #define USB_FUSES_TRANSP_Pos        18           /**< \brief (NVMCTRL_OTP4) USB pad Transp calibration */
	// #define USB_FUSES_TRANSP_Msk        (0x1Fu << USB_FUSES_TRANSP_Pos)
	// #define USB_FUSES_TRANSP(value)     ((USB_FUSES_TRANSP_Msk & ((value) << USB_FUSES_TRANSP_Pos)))

	// #define USB_FUSES_TRIM_ADDR         (NVMCTRL_OTP4 + 4)
	// #define USB_FUSES_TRIM_Pos          23           /**< \brief (NVMCTRL_OTP4) USB pad Trim calibration */
	// #define USB_FUSES_TRIM_Msk          (0x7u << USB_FUSES_TRIM_Pos)
	// #define USB_FUSES_TRIM(value)       ((USB_FUSES_TRIM_Msk & ((value) << USB_FUSES_TRIM_Pos)))
	//
	fuse := *(*uint32)(unsafe.Pointer(uintptr(0x00806020) + 4))
	calibTransN := uint16(fuse>>13) & uint16(0x1f)
	calibTransP := uint16(fuse>>18) & uint16(0x1f)
	calibTrim := uint16(fuse>>23) & uint16(0x7)

	if calibTransN == 0x1f {
		calibTransN = 5
	}
	sam.USB_DEVICE.PADCAL.SetBits(calibTransN << sam.USB_DEVICE_PADCAL_TRANSN_Pos)

	if calibTransP == 0x1f {
		calibTransP = 29
	}
	sam.USB_DEVICE.PADCAL.SetBits(calibTransP << sam.USB_DEVICE_PADCAL_TRANSP_Pos)

	if calibTrim == 0x7 {
		calibTrim = 3
	}
	sam.USB_DEVICE.PADCAL.SetBits(calibTrim << sam.USB_DEVICE_PADCAL_TRIM_Pos)
}

func handleUSBIRQ(intr interrupt.Interrupt) {
	// reset all interrupt flags
	flags := sam.USB_DEVICE.INTFLAG.Get()
	sam.USB_DEVICE.INTFLAG.Set(flags)

	// End of reset
	if (flags & sam.USB_DEVICE_INTFLAG_EORST) > 0 {
		// Configure control endpoint
		initEndpoint(0, usb_ENDPOINT_TYPE_CONTROL)

		usbConfiguration = 0

		// ack the End-Of-Reset interrupt
		sam.USB_DEVICE.INTFLAG.Set(sam.USB_DEVICE_INTFLAG_EORST)
	}

	// Start of frame
	if (flags & sam.USB_DEVICE_INTFLAG_SOF) > 0 {
		// if you want to blink LED showing traffic, this would be the place...
	}

	// Endpoint 0 Setup interrupt
	if getEPINTFLAG(0)&sam.USB_DEVICE_ENDPOINT_EPINTFLAG_RXSTP > 0 {
		// ack setup received
		setEPINTFLAG(0, sam.USB_DEVICE_ENDPOINT_EPINTFLAG_RXSTP)

		// parse setup
		setup := newUSBSetup(udd_ep_out_cache_buffer[0][:])

		// Clear the Bank 0 ready flag on Control OUT
		setEPSTATUSCLR(0, sam.USB_DEVICE_ENDPOINT_EPSTATUSCLR_BK0RDY)
		usbEndpointDescriptors[0].DeviceDescBank[0].PCKSIZE.ClearBits(usb_DEVICE_PCKSIZE_BYTE_COUNT_Mask << usb_DEVICE_PCKSIZE_BYTE_COUNT_Pos)

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
			setEPSTATUSSET(0, sam.USB_DEVICE_ENDPOINT_EPSTATUSSET_BK1RDY)
		} else {
			// Stall endpoint
			setEPSTATUSSET(0, sam.USB_DEVICE_ENDPOINT_EPINTFLAG_STALL1)
		}

		if getEPINTFLAG(0)&sam.USB_DEVICE_ENDPOINT_EPINTFLAG_STALL1 > 0 {
			// ack the stall
			setEPINTFLAG(0, sam.USB_DEVICE_ENDPOINT_EPINTFLAG_STALL1)

			// clear stall request
			setEPINTENCLR(0, sam.USB_DEVICE_ENDPOINT_EPINTENCLR_STALL1)
		}
	}

	// Now the actual transfer handlers, ignore endpoint number 0 (setup)
	var i uint32
	for i = 1; i < uint32(len(endPoints)); i++ {
		// Check if endpoint has a pending interrupt
		epFlags := getEPINTFLAG(i)
		setEPINTFLAG(i, epFlags)
		if (epFlags & sam.USB_DEVICE_ENDPOINT_EPINTFLAG_TRCPT0) > 0 {
			buf := handleEndpointRx(i)
			if callbackUSBRx[i] != nil {
				callbackUSBRx[i](buf)
			}
			handleEndpointRxComplete(i)
		} else if (epFlags & sam.USB_DEVICE_ENDPOINT_EPINTFLAG_TRCPT1) > 0 {
			if callbackUSBTx[i] != nil {
				callbackUSBTx[i]()
			}
		}
	}
}

func initEndpoint(ep, config uint32) {
	switch config {
	case usb_ENDPOINT_TYPE_INTERRUPT | usbEndpointIn:
		// set packet size
		usbEndpointDescriptors[ep].DeviceDescBank[1].PCKSIZE.SetBits(epPacketSize(64) << usb_DEVICE_PCKSIZE_SIZE_Pos)

		// set data buffer address
		usbEndpointDescriptors[ep].DeviceDescBank[1].ADDR.Set(uint32(uintptr(unsafe.Pointer(&udd_ep_in_cache_buffer[ep]))))

		// set endpoint type
		setEPCFG(ep, ((usb_ENDPOINT_TYPE_INTERRUPT + 1) << sam.USB_DEVICE_ENDPOINT_EPCFG_EPTYPE1_Pos))

		setEPINTENSET(ep, sam.USB_DEVICE_ENDPOINT_EPINTENSET_TRCPT1)

	case usb_ENDPOINT_TYPE_BULK | usbEndpointOut:
		// set packet size
		usbEndpointDescriptors[ep].DeviceDescBank[0].PCKSIZE.SetBits(epPacketSize(64) << usb_DEVICE_PCKSIZE_SIZE_Pos)

		// set data buffer address
		usbEndpointDescriptors[ep].DeviceDescBank[0].ADDR.Set(uint32(uintptr(unsafe.Pointer(&udd_ep_out_cache_buffer[ep]))))

		// set endpoint type
		setEPCFG(ep, ((usb_ENDPOINT_TYPE_BULK + 1) << sam.USB_DEVICE_ENDPOINT_EPCFG_EPTYPE0_Pos))

		// receive interrupts when current transfer complete
		setEPINTENSET(ep, sam.USB_DEVICE_ENDPOINT_EPINTENSET_TRCPT0)

		// set byte count to zero, we have not received anything yet
		usbEndpointDescriptors[ep].DeviceDescBank[0].PCKSIZE.ClearBits(usb_DEVICE_PCKSIZE_BYTE_COUNT_Mask << usb_DEVICE_PCKSIZE_BYTE_COUNT_Pos)

		// ready for next transfer
		setEPSTATUSCLR(ep, sam.USB_DEVICE_ENDPOINT_EPSTATUSCLR_BK0RDY)

	case usb_ENDPOINT_TYPE_INTERRUPT | usbEndpointOut:
		// TODO: not really anything, seems like...

	case usb_ENDPOINT_TYPE_BULK | usbEndpointIn:
		// set packet size
		usbEndpointDescriptors[ep].DeviceDescBank[1].PCKSIZE.SetBits(epPacketSize(64) << usb_DEVICE_PCKSIZE_SIZE_Pos)

		// set data buffer address
		usbEndpointDescriptors[ep].DeviceDescBank[1].ADDR.Set(uint32(uintptr(unsafe.Pointer(&udd_ep_in_cache_buffer[ep]))))

		// set endpoint type
		setEPCFG(ep, ((usb_ENDPOINT_TYPE_BULK + 1) << sam.USB_DEVICE_ENDPOINT_EPCFG_EPTYPE1_Pos))

		// NAK on endpoint IN, the bank is not yet filled in.
		setEPSTATUSCLR(ep, sam.USB_DEVICE_ENDPOINT_EPSTATUSCLR_BK1RDY)

		setEPINTENSET(ep, sam.USB_DEVICE_ENDPOINT_EPINTENSET_TRCPT1)

	case usb_ENDPOINT_TYPE_CONTROL:
		// Control OUT
		// set packet size
		usbEndpointDescriptors[ep].DeviceDescBank[0].PCKSIZE.SetBits(epPacketSize(64) << usb_DEVICE_PCKSIZE_SIZE_Pos)

		// set data buffer address
		usbEndpointDescriptors[ep].DeviceDescBank[0].ADDR.Set(uint32(uintptr(unsafe.Pointer(&udd_ep_out_cache_buffer[ep]))))

		// set endpoint type
		setEPCFG(ep, getEPCFG(ep)|((usb_ENDPOINT_TYPE_CONTROL+1)<<sam.USB_DEVICE_ENDPOINT_EPCFG_EPTYPE0_Pos))

		// Control IN
		// set packet size
		usbEndpointDescriptors[ep].DeviceDescBank[1].PCKSIZE.SetBits(epPacketSize(64) << usb_DEVICE_PCKSIZE_SIZE_Pos)

		// set data buffer address
		usbEndpointDescriptors[ep].DeviceDescBank[1].ADDR.Set(uint32(uintptr(unsafe.Pointer(&udd_ep_in_cache_buffer[ep]))))

		// set endpoint type
		setEPCFG(ep, getEPCFG(ep)|((usb_ENDPOINT_TYPE_CONTROL+1)<<sam.USB_DEVICE_ENDPOINT_EPCFG_EPTYPE1_Pos))

		// Prepare OUT endpoint for receive
		// set multi packet size for expected number of receive bytes on control OUT
		usbEndpointDescriptors[ep].DeviceDescBank[0].PCKSIZE.SetBits(64 << usb_DEVICE_PCKSIZE_MULTI_PACKET_SIZE_Pos)

		// set byte count to zero, we have not received anything yet
		usbEndpointDescriptors[ep].DeviceDescBank[0].PCKSIZE.ClearBits(usb_DEVICE_PCKSIZE_BYTE_COUNT_Mask << usb_DEVICE_PCKSIZE_BYTE_COUNT_Pos)

		// NAK on endpoint OUT to show we are ready to receive control data
		setEPSTATUSSET(ep, sam.USB_DEVICE_ENDPOINT_EPSTATUSSET_BK0RDY)

		// Enable Setup-Received interrupt
		setEPINTENSET(0, sam.USB_DEVICE_ENDPOINT_EPINTENSET_RXSTP)
	}
}

func handleUSBSetAddress(setup USBSetup) bool {
	// set packet size 64 with auto Zlp after transfer
	usbEndpointDescriptors[0].DeviceDescBank[1].PCKSIZE.Set((epPacketSize(64) << usb_DEVICE_PCKSIZE_SIZE_Pos) |
		uint32(1<<31)) // autozlp

	// ack the transfer is complete from the request
	setEPINTFLAG(0, sam.USB_DEVICE_ENDPOINT_EPINTFLAG_TRCPT1)

	// set bank ready for data
	setEPSTATUSSET(0, sam.USB_DEVICE_ENDPOINT_EPSTATUSSET_BK1RDY)

	// wait for transfer to complete
	timeout := 3000
	for (getEPINTFLAG(0) & sam.USB_DEVICE_ENDPOINT_EPINTFLAG_TRCPT1) == 0 {
		timeout--
		if timeout == 0 {
			return true
		}
	}

	// last, set the device address to that requested by host
	sam.USB_DEVICE.DADD.SetBits(setup.WValueL)
	sam.USB_DEVICE.DADD.SetBits(sam.USB_DEVICE_DADD_ADDEN)

	return true
}

// SendUSBInPacket sends a packet for USB (interrupt in / bulk in).
func SendUSBInPacket(ep uint32, data []byte) bool {
	sendUSBPacket(ep, data, 0)

	// clear transfer complete flag
	setEPINTFLAG(ep, sam.USB_DEVICE_ENDPOINT_EPINTFLAG_TRCPT1)

	// send data by setting bank ready
	setEPSTATUSSET(ep, sam.USB_DEVICE_ENDPOINT_EPSTATUSSET_BK1RDY)

	return true
}

//go:noinline
func sendUSBPacket(ep uint32, data []byte, maxsize uint16) {
	l := uint16(len(data))
	if 0 < maxsize && maxsize < l {
		l = maxsize
	}

	// Set endpoint address for sending data
	if ep == 0 {
		copy(udd_ep_control_cache_buffer[:], data[:l])
		usbEndpointDescriptors[ep].DeviceDescBank[1].ADDR.Set(uint32(uintptr(unsafe.Pointer(&udd_ep_control_cache_buffer))))
	} else {
		copy(udd_ep_in_cache_buffer[ep][:], data[:l])
		usbEndpointDescriptors[ep].DeviceDescBank[1].ADDR.Set(uint32(uintptr(unsafe.Pointer(&udd_ep_in_cache_buffer[ep]))))
	}

	// clear multi-packet size which is total bytes already sent
	usbEndpointDescriptors[ep].DeviceDescBank[1].PCKSIZE.ClearBits(usb_DEVICE_PCKSIZE_MULTI_PACKET_SIZE_Mask << usb_DEVICE_PCKSIZE_MULTI_PACKET_SIZE_Pos)

	// set byte count, which is total number of bytes to be sent
	usbEndpointDescriptors[ep].DeviceDescBank[1].PCKSIZE.ClearBits(usb_DEVICE_PCKSIZE_BYTE_COUNT_Mask << usb_DEVICE_PCKSIZE_BYTE_COUNT_Pos)
	usbEndpointDescriptors[ep].DeviceDescBank[1].PCKSIZE.SetBits((uint32(l) & usb_DEVICE_PCKSIZE_BYTE_COUNT_Mask) << usb_DEVICE_PCKSIZE_BYTE_COUNT_Pos)
}

func ReceiveUSBControlPacket() ([cdcLineInfoSize]byte, error) {
	var b [cdcLineInfoSize]byte

	// address
	usbEndpointDescriptors[0].DeviceDescBank[0].ADDR.Set(uint32(uintptr(unsafe.Pointer(&udd_ep_out_cache_buffer[0]))))

	// set byte count to zero
	usbEndpointDescriptors[0].DeviceDescBank[0].PCKSIZE.ClearBits(usb_DEVICE_PCKSIZE_BYTE_COUNT_Mask << usb_DEVICE_PCKSIZE_BYTE_COUNT_Pos)

	// set ready for next data
	setEPSTATUSCLR(0, sam.USB_DEVICE_ENDPOINT_EPSTATUSCLR_BK0RDY)

	// Wait until OUT transfer is ready.
	timeout := 300000
	for (getEPSTATUS(0) & sam.USB_DEVICE_ENDPOINT_EPSTATUS_BK0RDY) == 0 {
		timeout--
		if timeout == 0 {
			return b, errUSBCDCReadTimeout
		}
	}

	// Wait until OUT transfer is completed.
	timeout = 300000
	for (getEPINTFLAG(0) & sam.USB_DEVICE_ENDPOINT_EPINTFLAG_TRCPT0) == 0 {
		timeout--
		if timeout == 0 {
			return b, errUSBCDCReadTimeout
		}
	}

	// get data
	bytesread := uint32((usbEndpointDescriptors[0].DeviceDescBank[0].PCKSIZE.Get() >>
		usb_DEVICE_PCKSIZE_BYTE_COUNT_Pos) & usb_DEVICE_PCKSIZE_BYTE_COUNT_Mask)

	if bytesread != cdcLineInfoSize {
		return b, errUSBCDCBytesRead
	}

	copy(b[:7], udd_ep_out_cache_buffer[0][:7])

	return b, nil
}

func handleEndpointRx(ep uint32) []byte {
	// get data
	count := int((usbEndpointDescriptors[ep].DeviceDescBank[0].PCKSIZE.Get() >>
		usb_DEVICE_PCKSIZE_BYTE_COUNT_Pos) & usb_DEVICE_PCKSIZE_BYTE_COUNT_Mask)

	return udd_ep_out_cache_buffer[ep][:count]
}

func handleEndpointRxComplete(ep uint32) {
	// set byte count to zero
	usbEndpointDescriptors[ep].DeviceDescBank[0].PCKSIZE.ClearBits(usb_DEVICE_PCKSIZE_BYTE_COUNT_Mask << usb_DEVICE_PCKSIZE_BYTE_COUNT_Pos)

	// set multi packet size to 64
	usbEndpointDescriptors[ep].DeviceDescBank[0].PCKSIZE.SetBits(64 << usb_DEVICE_PCKSIZE_MULTI_PACKET_SIZE_Pos)

	// set ready for next data
	setEPSTATUSCLR(ep, sam.USB_DEVICE_ENDPOINT_EPSTATUSCLR_BK0RDY)
}

func SendZlp() {
	usbEndpointDescriptors[0].DeviceDescBank[1].PCKSIZE.ClearBits(usb_DEVICE_PCKSIZE_BYTE_COUNT_Mask << usb_DEVICE_PCKSIZE_BYTE_COUNT_Pos)
}

func epPacketSize(size uint16) uint32 {
	switch size {
	case 8:
		return 0
	case 16:
		return 1
	case 32:
		return 2
	case 64:
		return 3
	case 128:
		return 4
	case 256:
		return 5
	case 512:
		return 6
	case 1023:
		return 7
	default:
		return 0
	}
}

func getEPCFG(ep uint32) uint8 {
	return sam.USB_DEVICE.DEVICE_ENDPOINT[ep].EPCFG.Get()
}

func setEPCFG(ep uint32, val uint8) {
	sam.USB_DEVICE.DEVICE_ENDPOINT[ep].EPCFG.Set(val)
}

func setEPSTATUSCLR(ep uint32, val uint8) {
	sam.USB_DEVICE.DEVICE_ENDPOINT[ep].EPSTATUSCLR.Set(val)
}

func setEPSTATUSSET(ep uint32, val uint8) {
	sam.USB_DEVICE.DEVICE_ENDPOINT[ep].EPSTATUSSET.Set(val)
}

func getEPSTATUS(ep uint32) uint8 {
	return sam.USB_DEVICE.DEVICE_ENDPOINT[ep].EPSTATUS.Get()
}

func getEPINTFLAG(ep uint32) uint8 {
	return sam.USB_DEVICE.DEVICE_ENDPOINT[ep].EPINTFLAG.Get()
}

func setEPINTFLAG(ep uint32, val uint8) {
	sam.USB_DEVICE.DEVICE_ENDPOINT[ep].EPINTFLAG.Set(val)
}

func setEPINTENCLR(ep uint32, val uint8) {
	sam.USB_DEVICE.DEVICE_ENDPOINT[ep].EPINTENCLR.Set(val)
}

func setEPINTENSET(ep uint32, val uint8) {
	sam.USB_DEVICE.DEVICE_ENDPOINT[ep].EPINTENSET.Set(val)
}
