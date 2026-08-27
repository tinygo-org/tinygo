//go:build stm32f4 || stm32f7

package machine

import (
	"device/stm32"
	"machine/usb"
	"runtime/interrupt"
	"runtime/volatile"
	"unsafe"
)

// NumberOfUSBEndpoints is sized to cover TinyGo's full endpoint index space
// (0=control, 1=CDC ACM, 2=CDC OUT, 3=CDC IN, 4=HID IN, 5=HID OUT, 6=MIDI IN, 7=MIDI OUT).
// Physical OTG FS hardware has 4 IN + 4 OUT endpoints (0–3).
const NumberOfUSBEndpoints = 8

// Default USB identifiers; board files with USB support should override these.
const (
	usb_VID                 = uint16(0x239A)
	usb_PID                 = uint16(0x0001)
	usb_STRING_MANUFACTURER = "TinyGo"
	usb_STRING_PRODUCT      = "STM32 USB Device"
)

// OTG FS register blocks.
var (
	otgDevice = (*usbDeviceRegs)(unsafe.Pointer(stm32.OTG_FS_DEVICE))
	otgPower  = (*usbPowerRegs)(unsafe.Pointer(stm32.OTG_FS_PWRCLK))
)

// usbDeviceRegs represents the USB device-mode control block at base+0x800.
type usbDeviceRegs struct {
	DCFG       volatile.Register32 // 0x800
	DCTL       volatile.Register32 // 0x804
	DSTS       volatile.Register32 // 0x808
	_          uint32              // 0x80C
	DIEPMSK    volatile.Register32 // 0x810
	DOEPMSK    volatile.Register32 // 0x814
	DAINT      volatile.Register32 // 0x818
	DAINTMSK   volatile.Register32 // 0x81C
	_          [5]uint32           // 0x820 - 0x830
	DIEPEMPMSK volatile.Register32 // 0x834
}

// usbInEndpointRegs represents the registers for a single IN endpoint at 0x900 + ep*0x20.
type usbInEndpointRegs struct {
	CTL   volatile.Register32 // 0x00
	_     uint32              // 0x04
	INT   volatile.Register32 // 0x08
	_     uint32              // 0x0C
	TSIZ  volatile.Register32 // 0x10
	_     uint32              // 0x14
	TXFST volatile.Register32 // 0x18
	_     uint32              // 0x1C
}

// usbOutEndpointRegs represents the registers for a single OUT endpoint at 0xB00 + ep*0x20.
type usbOutEndpointRegs struct {
	CTL  volatile.Register32 // 0x00
	_    uint32              // 0x04
	INT  volatile.Register32 // 0x08
	_    uint32              // 0x0C
	TSIZ volatile.Register32 // 0x10
	_    [3]uint32           // 0x14 - 0x1C
}

// usbPowerRegs represents the power and clock gating block at base+0xE00.
type usbPowerRegs struct {
	PCGCCTL volatile.Register32 // 0xE00
}

// otgInEP returns the IN endpoint registers for physical endpoint ep.
func otgInEP(ep uint32) *usbInEndpointRegs {
	return (*usbInEndpointRegs)(unsafe.Pointer(uintptr(unsafe.Pointer(stm32.OTG_FS_GLOBAL)) + 0x900 + uintptr(ep)*0x20))
}

// otgOutEP returns the OUT endpoint registers for physical endpoint ep.
func otgOutEP(ep uint32) *usbOutEndpointRegs {
	return (*usbOutEndpointRegs)(unsafe.Pointer(uintptr(unsafe.Pointer(stm32.OTG_FS_GLOBAL)) + 0xB00 + uintptr(ep)*0x20))
}

// otgDFIFO returns a volatile pointer to the data FIFO for physical endpoint ep.
// DFIFO[ep] is located at 0x50001000 + ep*0x1000.
func otgDFIFO(ep uint32) *volatile.Register32 {
	return (*volatile.Register32)(unsafe.Pointer(uintptr(unsafe.Pointer(stm32.OTG_FS_GLOBAL)) + 0x1000 + uintptr(ep)*0x1000))
}

// DCFG bit positions and masks.
const (
	dcfgDSPD    = uint32(0x3) // FS PHY speed (bits [1:0] = 0b11)
	dcfgDAD_Pos = uint32(4)   // device address field start bit
	dcfgDAD_Msk = uint32(0x7F << 4)
)

// DCTL bits.
const (
	dctlSDIS   = uint32(1 << 1)  // soft disconnect
	dctlSGINAK = uint32(1 << 7)  // set global IN NAK
	dctlCGINAK = uint32(1 << 8)  // clear global IN NAK
	dctlSGONAK = uint32(1 << 9)  // set global OUT NAK
	dctlCGONAK = uint32(1 << 10) // clear global OUT NAK
)

// DSTS bits.
const (
	dstsSUSPSTS     = uint32(1 << 0)
	dstsENUMSPD_Pos = uint32(1)
	dstsENUMSPD_Msk = uint32(0x3 << 1)
)

// DIEPCTLn / DOEPCTLn bits (shared between IN and OUT endpoint control registers).
const (
	depctlEPENA      = uint32(1 << 31) // endpoint enable
	depctlEPDIS      = uint32(1 << 30) // endpoint disable
	depctlSD0PID     = uint32(1 << 28) // set DATA0 PID
	depctlSNAK       = uint32(1 << 27) // set NAK
	depctlCNAK       = uint32(1 << 26) // clear NAK
	depctlSTALL      = uint32(1 << 21) // STALL handshake
	depctlETYP_Pos   = uint32(18)      // endpoint type field start
	depctlUSBAEP     = uint32(1 << 15) // USB active endpoint
	depctlTXFNUM_Pos = uint32(22)      // TX FIFO number field start (IN eps only)
	depctlMPSIZ_Pos  = uint32(0)       // max packet size field start
)

// EP0 max packet size encoding in DIEPCTL0 / DOEPCTL0 bits [1:0].
const ep0Mps64 = uint32(0x0) // 64 bytes (FS default)

// DIEPINTn / DOEPINTn bits.
const (
	depintXFRC = uint32(1 << 0) // transfer complete
	depintSTUP = uint32(1 << 3) // SETUP phase done (DOEPINTn)
)

// GRXSTSP_Device PKTSTS field values.
const (
	rxPktstsGNAK      = uint32(0x1) // global OUT NAK
	rxPktstsOUTData   = uint32(0x2) // OUT data packet received
	rxPktstsOUTDone   = uint32(0x3) // OUT transfer complete
	rxPktstsSetupDone = uint32(0x4) // SETUP transaction complete
	rxPktstsSetupData = uint32(0x6) // SETUP data packet received (always 8 bytes)
)

// GINTSTS / GINTMSK bits (values taken from stm32f405 SVD constants).
const (
	gintRXFLVL  = uint32(stm32.USB_OTG_FS_GINTSTS_RXFLVL)  // 0x10
	gintUSBRST  = uint32(stm32.USB_OTG_FS_GINTSTS_USBRST)  // 0x1000
	gintENUMDNE = uint32(stm32.USB_OTG_FS_GINTSTS_ENUMDNE) // 0x2000
	gintUSBSUSP = uint32(stm32.USB_OTG_FS_GINTSTS_USBSUSP) // 0x800
	gintWKUPINT = uint32(stm32.USB_OTG_FS_GINTSTS_WKUPINT) // 0x80000000
	gintIEPINT  = uint32(stm32.USB_OTG_FS_GINTSTS_IEPINT)  // 0x40000
	gintOEPINT  = uint32(stm32.USB_OTG_FS_GINTSTS_OEPINT)  // 0x80000
)

// GRSTCTL bits.
const (
	grstCSRST      = uint32(stm32.USB_OTG_FS_GRSTCTL_CSRST)   // core soft reset
	grstRXFFLSH    = uint32(stm32.USB_OTG_FS_GRSTCTL_RXFFLSH) // RX FIFO flush
	grstTXFFLSH    = uint32(stm32.USB_OTG_FS_GRSTCTL_TXFFLSH) // TX FIFO flush
	grstTXFNUM_Pos = uint32(stm32.USB_OTG_FS_GRSTCTL_TXFNUM_Pos)
	grstAHBIDL     = uint32(stm32.USB_OTG_FS_GRSTCTL_AHBIDL) // AHB master idle
)

// FIFO size layout in 32-bit words (total budget = 320 words).
const (
	rxFIFODepth    = uint32(128) // shared RX FIFO
	ep0TxFIFODepth = uint32(16)  // EP0 TX FIFO
	ep1TxFIFODepth = uint32(64)  // EP1 TX FIFO
	ep2TxFIFODepth = uint32(64)  // EP2 TX FIFO
	ep3TxFIFODepth = uint32(48)  // EP3 TX FIFO

	ep0TxFIFOStart = rxFIFODepth
	ep1TxFIFOStart = ep0TxFIFOStart + ep0TxFIFODepth
	ep2TxFIFOStart = ep1TxFIFOStart + ep1TxFIFODepth
	ep3TxFIFOStart = ep2TxFIFOStart + ep2TxFIFODepth
)

// Driver state.
var (
	// sendOnEP0DATADONE tracks multi-chunk EP0 IN transfers.
	sendOnEP0DATADONE struct {
		ptr    *byte
		count  int
		offset int
	}

	// usbSetupBuf holds the 8-byte SETUP packet from the RX FIFO.
	usbSetupBuf [8]byte

	// usbRxBufLen tracks the byte count of the most recently received OUT packet
	// per physical endpoint (index 0–3).
	usbRxBufLen [4]uint32
)

// Configure initialises the OTG FS USB peripheral in device mode.
// The config parameter is unused (present for interface compatibility).
func (dev *USBDevice) Configure(config UARTConfig) {
	if dev.initcomplete {
		return
	}

	// ---- 1. Enable peripheral clocks ----------------------------------------

	// GPIOA clock (PA11 = D-, PA12 = D+)
	stm32.RCC.AHB1ENR.SetBits(stm32.RCC_AHB1ENR_GPIOAEN)
	// OTG FS peripheral clock
	stm32.RCC.AHB2ENR.SetBits(stm32.RCC_AHB2ENR_OTGFSEN)

	// ---- 2. Configure GPIO pins (PA11 D-, PA12 D+) as AF, very high speed ----

	for _, pin := range [2]Pin{PA11, PA12} {
		pos := uint8(pin%16) * 2
		port := pin.getPort()
		port.MODER.ReplaceBits(gpioModeAlternate, gpioModeMask, pos)
		port.OSPEEDR.ReplaceBits(gpioOutputSpeedVeryHigh, gpioOutputSpeedMask, pos)
		port.PUPDR.ReplaceBits(gpioPullFloating, gpioPullMask, pos)
		// OTYPER remains 0 (push-pull)
		pin.SetAltFunc(10) // AF10 = OTG FS on both F4 and F7
	}

	// ---- 3. OTG core reset --------------------------------------------------

	// Wait for AHB master idle before core reset.
	for stm32.OTG_FS_GLOBAL.GRSTCTL.Get()&grstAHBIDL == 0 {
	}

	// Core soft reset
	stm32.OTG_FS_GLOBAL.GRSTCTL.SetBits(grstCSRST)
	for stm32.OTG_FS_GLOBAL.GRSTCTL.HasBits(grstCSRST) {
	}

	// Wait for AHB idle again after reset
	for stm32.OTG_FS_GLOBAL.GRSTCTL.Get()&grstAHBIDL == 0 {
	}

	// ---- 4. Force device mode, set turnaround time --------------------------

	gusbcfg := stm32.OTG_FS_GLOBAL.GUSBCFG.Get()
	gusbcfg &^= stm32.USB_OTG_FS_GUSBCFG_FHMOD |
		stm32.USB_OTG_FS_GUSBCFG_FDMOD |
		stm32.USB_OTG_FS_GUSBCFG_TRDT_Msk
	gusbcfg |= stm32.USB_OTG_FS_GUSBCFG_FDMOD |
		(9 << stm32.USB_OTG_FS_GUSBCFG_TRDT_Pos) // turnaround time = 9 for 216MHz HCLK
	stm32.OTG_FS_GLOBAL.GUSBCFG.Set(gusbcfg)

	// ---- 5. PHY / VBUS configuration (platform-specific) --------------------

	initOTGFSPHY()

	// ---- 6. Enable PHY clock, soft-disconnect before further init -----------

	// Clear stop-clock / stop-phy-clock bits so the PHY clock runs.
	// If a bootloader left these set, USB would hang silently.
	otgPower.PCGCCTL.Set(0)

	// Soft-disconnect now (after CSRST reset DCTL to its default connected state).
	dev.Detach()

	// ---- 7. Configure data FIFOs --------------------------------------------

	// RX FIFO (shared for all OUT + SETUP packets)
	stm32.OTG_FS_GLOBAL.GRXFSIZ.Set(rxFIFODepth)

	// EP0 TX FIFO: start = rxFIFODepth, depth = ep0TxFIFODepth
	stm32.OTG_FS_GLOBAL.DIEPTXF0.Set(
		(ep0TxFIFODepth << 16) | ep0TxFIFOStart,
	)

	// EP1–3 TX FIFOs
	stm32.OTG_FS_GLOBAL.DIEPTXF1.Set(
		(ep1TxFIFODepth << 16) | ep1TxFIFOStart,
	)
	stm32.OTG_FS_GLOBAL.DIEPTXF2.Set(
		(ep2TxFIFODepth << 16) | ep2TxFIFOStart,
	)
	stm32.OTG_FS_GLOBAL.DIEPTXF3.Set(
		(ep3TxFIFODepth << 16) | ep3TxFIFOStart,
	)

	// ---- 8. Flush FIFOs ----------------------------------------------------

	flushRxFIFO()
	flushTxFIFO(0x10) // flush all TX FIFOs (TXFNUM = 0x10 = all)

	// ---- 9. Configure device: full-speed, no SOF output --------------------

	otgDevice.DCFG.Set(dcfgDSPD) // FS PHY speed

	// Clear any stale interrupts
	stm32.OTG_FS_GLOBAL.GINTSTS.Set(0xFFFFFFFF)

	// ---- 10. Enable interrupts ----------------------------------------------

	stm32.OTG_FS_GLOBAL.GINTMSK.Set(
		gintUSBRST | gintENUMDNE | gintRXFLVL | gintIEPINT | gintOEPINT |
			gintUSBSUSP | gintWKUPINT,
	)

	// Enable device-level IN and OUT endpoint interrupt masks
	otgDevice.DIEPMSK.Set(depintXFRC)
	otgDevice.DOEPMSK.Set(depintXFRC | depintSTUP)

	// Enable global interrupt
	stm32.OTG_FS_GLOBAL.GAHBCFG.SetBits(stm32.USB_OTG_FS_GAHBCFG_GINT)

	// ---- 11. Register and enable NVIC interrupt -----------------------------

	intr := interrupt.New(stm32.IRQ_OTG_FS, handleUSBIRQ)
	intr.SetPriority(0) // Highest priority
	intr.Enable()

	// ---- 12. Connect to host (clear soft-disconnect) -----------------------

	dev.Attach()

	dev.initcomplete = true
}

// Attach connects the device to the USB bus by releasing soft disconnect,
// allowing the host to detect and enumerate it. It can be used together with
// Detach to delay enumeration until the USB configuration (device
// identifiers, classes, ...) is complete.
func (dev *USBDevice) Attach() {
	otgDevice.DCTL.ClearBits(dctlSDIS)
}

// Detach disconnects the device from the USB bus by asserting soft
// disconnect. To the host this appears as if the device was unplugged. A
// subsequent Attach makes the host enumerate the device again.
func (dev *USBDevice) Detach() {
	otgDevice.DCTL.SetBits(dctlSDIS)
}

// handleUSBIRQ is the OTG FS interrupt handler, dispatching on GINTSTS bits.
func handleUSBIRQ(intr interrupt.Interrupt) {
	status := stm32.OTG_FS_GLOBAL.GINTSTS.Get() &
		stm32.OTG_FS_GLOBAL.GINTMSK.Get()

	if status&gintUSBSUSP != 0 {
		stm32.OTG_FS_GLOBAL.GINTSTS.Set(gintUSBSUSP)
		// Stop PHY clock during suspend. Only STPPCLK (bit 0); do NOT set
		// GATEHCLK (bit 1) — that gates the AHB bus, which prevents the ISR
		// from reading GINTSTS when WKUPINT fires.
		otgPower.PCGCCTL.SetBits(1) // STPPCLK
	}

	if status&gintWKUPINT != 0 {
		stm32.OTG_FS_GLOBAL.GINTSTS.Set(gintWKUPINT)
		// Restart clocks before any endpoint activity can resume.
		otgPower.PCGCCTL.ClearBits(1 | 2)
	}

	if status&gintUSBRST != 0 {
		stm32.OTG_FS_GLOBAL.GINTSTS.Set(gintUSBRST)
		otgPower.PCGCCTL.ClearBits(1 | 2) // ensure clocks running after reset-from-suspend
		handleUSBReset()
	}

	if status&gintRXFLVL != 0 {
		// RXFLVL is level-triggered; drain entire FIFO in a loop.
		stm32.OTG_FS_GLOBAL.GINTMSK.ClearBits(gintRXFLVL)
		handleRxFIFO()
		stm32.OTG_FS_GLOBAL.GINTMSK.SetBits(gintRXFLVL)
	}

	if status&gintENUMDNE != 0 {
		stm32.OTG_FS_GLOBAL.GINTSTS.Set(gintENUMDNE)
		handleEnumDone()
	}

	if status&gintIEPINT != 0 {
		handleInEndpoints()
	}

	if status&gintOEPINT != 0 {
		handleOutEndpoints()
	}
}

// handleUSBReset is called on USB bus reset (USBRST interrupt).
func handleUSBReset() {
	// Set NAK on all OUT endpoints.
	for ep := uint32(0); ep < 4; ep++ {
		otgOutEP(ep).CTL.SetBits(depctlSNAK)
	}

	// Flush RX and all TX FIFOs.
	flushRxFIFO()
	flushTxFIFO(0x10)

	// Clear all endpoint interrupts.
	otgDevice.DAINT.Set(0xFFFFFFFF)
	otgDevice.DAINTMSK.Set(0)

	// Enable EP0 IN and OUT interrupt sources.
	otgDevice.DAINTMSK.Set((1 << 0) | (1 << 16)) // DIEP0 + DOEP0

	// Re-arm EP0 OUT for up to 3 back-to-back SETUP packets.
	armEP0Out()

	// Signal upper layer: device is no longer configured.
	usbConfiguration = 0
	USBDev.InitEndpointComplete = false
}

// handleEnumDone is called after USB enumeration speed is detected (ENUMDNE).
func handleEnumDone() {
	// Activate EP0 (max packet 64, type control, TX FIFO 0).
	ep0Ctl := ep0Mps64 | depctlUSBAEP | (0 << depctlETYP_Pos) // control type
	otgInEP(0).CTL.SetBits(ep0Ctl)
	otgOutEP(0).CTL.SetBits(ep0Mps64 | depctlUSBAEP)

	// Clear global IN NAK so EP0 IN can send.
	otgDevice.DCTL.SetBits(dctlCGINAK)
}

// handleRxFIFO drains the RX FIFO completely, processing each pop via GRXSTSP.
// RXFLVL is level-triggered, so this must loop until the FIFO is empty.
func handleRxFIFO() {
	for stm32.OTG_FS_GLOBAL.GINTSTS.HasBits(gintRXFLVL) {
		status := stm32.OTG_FS_GLOBAL.GRXSTSP_Device.Get()

		ep := status & stm32.USB_OTG_FS_GRXSTSP_Device_EPNUM_Msk
		bcnt := (status & stm32.USB_OTG_FS_GRXSTSP_Device_BCNT_Msk) >>
			stm32.USB_OTG_FS_GRXSTSP_Device_BCNT_Pos
		pktsts := (status >> stm32.USB_OTG_FS_GRXSTSP_Device_PKTSTS_Pos) & 0xF

		pep := ep // GRXSTSP.EPNUM is already a physical endpoint (0–3)

		switch pktsts {
		case rxPktstsSetupData:
			// 8-byte SETUP packet: read exactly 2 words from DFIFO[0].
			w0 := otgDFIFO(0).Get()
			w1 := otgDFIFO(0).Get()
			usbSetupBuf[0] = byte(w0)
			usbSetupBuf[1] = byte(w0 >> 8)
			usbSetupBuf[2] = byte(w0 >> 16)
			usbSetupBuf[3] = byte(w0 >> 24)
			usbSetupBuf[4] = byte(w1)
			usbSetupBuf[5] = byte(w1 >> 8)
			usbSetupBuf[6] = byte(w1 >> 16)
			usbSetupBuf[7] = byte(w1 >> 24)

		case rxPktstsSetupDone:
			// SETUP transaction complete: process the buffered SETUP packet.
			setup := usb.Setup{
				BmRequestType: usbSetupBuf[0],
				BRequest:      usbSetupBuf[1],
				WValueL:       usbSetupBuf[2],
				WValueH:       usbSetupBuf[3],
				WIndex:        uint16(usbSetupBuf[4]) | (uint16(usbSetupBuf[5]) << 8),
				WLength:       uint16(usbSetupBuf[6]) | (uint16(usbSetupBuf[7]) << 8),
			}

			ok := false
			if setup.BmRequestType&usb.REQUEST_TYPE == usb.REQUEST_STANDARD {
				ok = handleStandardSetup(setup)
			} else {
				if setup.WIndex < uint16(len(usbSetupHandler)) &&
					usbSetupHandler[setup.WIndex] != nil {
					ok = usbSetupHandler[setup.WIndex](setup)
				}
			}
			if !ok {
				// Stall EP0 IN and OUT on unrecognised requests.
				otgInEP(0).CTL.SetBits(depctlSTALL)
				otgOutEP(0).CTL.SetBits(depctlSTALL)
			}
			// Re-arm EP0 OUT for the next SETUP.
			armEP0Out()

		case rxPktstsOUTData:
			// OUT data: read bcnt bytes from DFIFO[pep] into cache buffer.
			if bcnt > 0 && pep < 4 {
				readFIFO(pep, bcnt)
				usbRxBufLen[pep] = bcnt
			}

		case rxPktstsOUTDone:
			// OUT transfer complete: nothing to do here; handled in handleOutEndpoints.
		}
	}
}

// readFIFO reads bcnt bytes from the shared RX FIFO (DFIFO 0) into udd_ep_out_cache_buffer[ep].
func readFIFO(ep, bcnt uint32) {
	buf := udd_ep_out_cache_buffer[ep][:]
	words := (bcnt + 3) / 4
	for i := uint32(0); i < words; i++ {
		w := otgDFIFO(0).Get() // Always read from FIFO 0
		b := i * 4
		buf[b] = byte(w)
		if b+1 < bcnt {
			buf[b+1] = byte(w >> 8)
		}
		if b+2 < bcnt {
			buf[b+2] = byte(w >> 16)
		}
		if b+3 < bcnt {
			buf[b+3] = byte(w >> 24)
		}
	}
}

// handleInEndpoints handles IEPINT: checks each active IN endpoint for XFRC.
func handleInEndpoints() {
	daint := otgDevice.DAINT.Get() & 0x0000FFFF // lower 16 bits = IN EPs
	daintmsk := otgDevice.DAINTMSK.Get() & 0x0000FFFF
	active := daint & daintmsk

	for ep := uint32(0); ep < 4; ep++ {
		if active&(1<<ep) == 0 {
			continue
		}
		diep := otgInEP(ep)
		diepint := diep.INT.Get()
		diepintmsk := otgDevice.DIEPMSK.Get()
		fired := diepint & diepintmsk

		if fired&depintXFRC != 0 {
			// Clear XFRC.
			diep.INT.Set(depintXFRC)

			if ep == 0 {
				// EP0 IN transfer complete.
				if sendOnEP0DATADONE.ptr != nil {
					// More data to send.
					ptr := sendOnEP0DATADONE.ptr
					count := sendOnEP0DATADONE.count
					if count > usb.EndpointPacketSize {
						sendOnEP0DATADONE.offset += usb.EndpointPacketSize
						sendOnEP0DATADONE.ptr = &udd_ep_control_cache_buffer[sendOnEP0DATADONE.offset]
						count = usb.EndpointPacketSize
					}
					sendOnEP0DATADONE.count -= count
					sendViaEPIn(0, ptr, count)
					if sendOnEP0DATADONE.count == 0 {
						sendOnEP0DATADONE.ptr = nil
						sendOnEP0DATADONE.offset = 0
					}
				} else {
					// All EP0 IN data sent; arm EP0 OUT for the status ZLP from host.
					armEP0Out()
				}
			} else {
				// Non-EP0 IN: find the virtual endpoint(s) mapped to this physical EP
				// and call the registered TX handler. Multiple virtual EPs may share
				// a physical EP (e.g., HID_IN=4 and CDC_IN=3 both → physical 1 or 3).
				for vep := uint32(0); vep < NumberOfUSBEndpoints; vep++ {
					if vep == ep && usbTxHandler[vep] != nil {
						usbTxHandler[vep]()
					}
				}
			}
		}

	}
}

// handleOutEndpoints handles OEPINT: checks each active OUT endpoint for STUP / XFRC.
func handleOutEndpoints() {
	daint := otgDevice.DAINT.Get() >> 16 // upper 16 bits = OUT EPs
	daintmsk := otgDevice.DAINTMSK.Get() >> 16
	active := daint & daintmsk

	for ep := uint32(0); ep < 4; ep++ {
		if active&(1<<ep) == 0 {
			continue
		}
		doep := otgOutEP(ep)
		doepint := doep.INT.Get()
		doepintmsk := otgDevice.DOEPMSK.Get()
		fired := doepint & doepintmsk

		if fired&depintSTUP != 0 {
			// EP0 SETUP phase done (already processed in handleRxFIFO).
			doep.INT.Set(depintSTUP)
		}

		if fired&depintXFRC != 0 {
			doep.INT.Set(depintXFRC)
			if ep > 0 {
				buf := handleEndpointRx(ep)
				// Find the virtual endpoint(s) mapped to this physical EP and call the RX handler.
				for vep := uint32(0); vep < NumberOfUSBEndpoints; vep++ {
					if vep == ep && usbRxHandler[vep] != nil {
						if usbRxHandler[vep](buf) {
							AckUsbOutTransfer(ep)
						}
						break
					}
				}
			}
		}
	}
}

// initEndpoint configures a USB endpoint for the given type and direction.
// MPS is hardcoded to 64 bytes; the caller (usb.go) does not pass a descriptor.
func initEndpoint(ep, config uint32) {
	pep := ep
	if pep == 0 {
		return // EP0 is always active; configured in handleEnumDone
	}

	txFIFONum := pep // TX FIFO number matches physical EP

	switch config {
	case usb.ENDPOINT_TYPE_INTERRUPT | usb.EndpointIn:
		ctl := (64 << depctlMPSIZ_Pos) | depctlUSBAEP |
			(txFIFONum << depctlTXFNUM_Pos) |
			(3 << depctlETYP_Pos) | // interrupt type
			depctlSD0PID
		otgInEP(pep).CTL.Set(ctl)
		otgDevice.DAINTMSK.SetBits(1 << pep)

	case usb.ENDPOINT_TYPE_BULK | usb.EndpointIn:
		ctl := (64 << depctlMPSIZ_Pos) | depctlUSBAEP |
			(txFIFONum << depctlTXFNUM_Pos) |
			(2 << depctlETYP_Pos) | // bulk type
			depctlSD0PID
		otgInEP(pep).CTL.Set(ctl)
		otgDevice.DAINTMSK.SetBits(1 << pep)

	case usb.ENDPOINT_TYPE_INTERRUPT | usb.EndpointOut:
		ctl := uint32(64) | depctlUSBAEP | depctlSD0PID |
			(3 << depctlETYP_Pos) // interrupt type
		otgOutEP(pep).CTL.Set(ctl)
		otgOutEP(pep).TSIZ.Set((1 << 19) | 64)
		otgOutEP(pep).CTL.SetBits(depctlEPENA | depctlCNAK)
		otgDevice.DAINTMSK.SetBits(1 << (pep + 16))

	case usb.ENDPOINT_TYPE_BULK | usb.EndpointOut:
		ctl := uint32(64) | depctlUSBAEP | depctlSD0PID |
			(2 << depctlETYP_Pos) // bulk type
		otgOutEP(pep).CTL.Set(ctl)
		otgOutEP(pep).TSIZ.Set((1 << 19) | 64)
		otgOutEP(pep).CTL.SetBits(depctlEPENA | depctlCNAK)
		otgDevice.DAINTMSK.SetBits(1 << (pep + 16))

	case usb.ENDPOINT_TYPE_CONTROL:
		// EP0 activated in handleEnumDone.
	}
}

// SendUSBInPacket sends data on a USB IN endpoint (interrupt or bulk).
func SendUSBInPacket(ep uint32, data []byte) bool {
	sendUSBPacket(ep, data)
	return true
}

// sendUSBPacket copies data into the endpoint cache buffer then initiates the transfer.
//
//go:noinline
func sendUSBPacket(ep uint32, data []byte) {
	count := len(data)
	var buf []byte
	if ep == 0 {
		buf = udd_ep_control_cache_buffer[:]
		if count > usb.EndpointPacketSize {
			// Large response: queue continuation via sendOnEP0DATADONE.
			sendOnEP0DATADONE.offset = usb.EndpointPacketSize
			sendOnEP0DATADONE.ptr = &udd_ep_control_cache_buffer[usb.EndpointPacketSize]
			sendOnEP0DATADONE.count = count - usb.EndpointPacketSize
			count = usb.EndpointPacketSize
		}
	} else {
		pep := ep
		buf = udd_ep_in_cache_buffer[pep][:]
	}
	copy(buf[:len(data)], data)
	sendViaEPIn(ep, &buf[0], count)
}

// sendViaEPIn arms the IN endpoint and writes count bytes from ptr into the TX FIFO.
func sendViaEPIn(ep uint32, ptr *byte, count int) {
	pep := ep
	diep := otgInEP(pep)

	// Verify TX FIFO has enough space before writing.
	// DTXFSTS[15:0] = INEPTFSAV: available words. Stall if insufficient.
	if count > 0 {
		need := uint32((count + 3) / 4)
		avail := diep.TXFST.Get() & 0xFFFF
		if avail < need {
			return
		}
	}

	// Program transfer size: 1 packet, count bytes.
	diep.TSIZ.Set(
		uint32(count) | (1 << 19), // XFRSIZ = count, PKTCNT = 1
	)

	// Enable endpoint and clear NAK (starts transfer).
	diep.CTL.SetBits(depctlEPENA | depctlCNAK)

	// Write bytes to FIFO in 32-bit words (last word padded if needed).
	fifo := otgDFIFO(pep)
	data := unsafe.Slice(ptr, count)
	words := (count + 3) / 4
	for i := 0; i < words; i++ {
		b := i * 4
		var w uint32
		w = uint32(data[b])
		if b+1 < count {
			w |= uint32(data[b+1]) << 8
		}
		if b+2 < count {
			w |= uint32(data[b+2]) << 16
		}
		if b+3 < count {
			w |= uint32(data[b+3]) << 24
		}
		fifo.Set(w)
	}
}

// SendZlp sends a zero-length packet on EP0 IN (status stage for OUT control transfers).
func SendZlp() {
	// PKTCNT=1, XFRSIZ=0
	otgInEP(0).TSIZ.Set(1 << 19)
	otgInEP(0).CTL.SetBits(depctlEPENA | depctlCNAK)
}

// handleEndpointRx returns the bytes received on the given physical endpoint.
func handleEndpointRx(ep uint32) []byte {
	pep := ep
	return udd_ep_out_cache_buffer[pep][:usbRxBufLen[pep]]
}

// AckUsbOutTransfer re-arms the OUT endpoint to receive the next packet.
func AckUsbOutTransfer(ep uint32) {
	pep := ep
	usbRxBufLen[pep] = 0
	doep := otgOutEP(pep)
	doep.TSIZ.Set(
		(1 << 19) | 64, // PKTCNT=1, XFRSIZ=64
	)
	doep.CTL.SetBits(depctlEPENA | depctlCNAK)
}

// handleUSBSetAddress applies the new device address from a SET_ADDRESS request
// and sends the status ZLP. The address is written to DCFG before the ZLP is
// enqueued: the OTG FS core has already committed the current IN token to
// address 0, so the ZLP goes out at the old address while the new address is
// already in DCFG and ready for the host's next transaction.
func handleUSBSetAddress(setup usb.Setup) bool {
	addr := uint8(setup.WValueL) & 0x7F

	dcfg := otgDevice.DCFG.Get()
	dcfg &^= dcfgDAD_Msk
	dcfg |= uint32(addr) << dcfgDAD_Pos
	otgDevice.DCFG.Set(dcfg)

	SendZlp()
	return true
}

// ReceiveUSBControlPacket synchronously receives a CDC control OUT packet on EP0.
func ReceiveUSBControlPacket() ([cdcLineInfoSize]byte, error) {
	var b [cdcLineInfoSize]byte

	// Arm EP0 OUT for up to 64 bytes.
	armEP0Out()

	// Busy-wait for data to arrive. We call handleRxFIFO() manually to drain
	// the shared RX FIFO because we are currently in an interrupt context
	// (this is called from the setup handler) and the hardware-triggered
	// handleRxFIFO loop is blocked waiting for us to return.
	const timeout = 300000
	for i := 0; i < timeout; i++ {
		if stm32.OTG_FS_GLOBAL.GINTSTS.HasBits(gintRXFLVL) {
			handleRxFIFO()
		}
		if usbRxBufLen[0] > 0 {
			n := usbRxBufLen[0]
			if n > cdcLineInfoSize {
				n = cdcLineInfoSize
			}
			copy(b[:n], udd_ep_out_cache_buffer[0][:n])
			usbRxBufLen[0] = 0
			SendZlp()
			return b, nil
		}
	}
	return b, ErrUSBReadTimeout
}

// SetStallEPIn stalls an IN endpoint.
func (dev *USBDevice) SetStallEPIn(ep uint32) {
	pep := ep
	otgInEP(pep).CTL.SetBits(depctlSTALL)
}

// ClearStallEPIn clears the stall condition on an IN endpoint.
func (dev *USBDevice) ClearStallEPIn(ep uint32) {
	pep := ep
	// Clear STALL and reset DATA0 PID.
	ctl := &otgInEP(pep).CTL
	ctl.ClearBits(depctlSTALL)
	ctl.SetBits(depctlSD0PID)
}

// SetStallEPOut stalls an OUT endpoint.
func (dev *USBDevice) SetStallEPOut(ep uint32) {
	pep := ep
	otgOutEP(pep).CTL.SetBits(depctlSTALL)
}

// ClearStallEPOut clears the stall condition on an OUT endpoint.
func (dev *USBDevice) ClearStallEPOut(ep uint32) {
	pep := ep
	ctl := &otgOutEP(pep).CTL
	ctl.ClearBits(depctlSTALL)
	ctl.SetBits(depctlSD0PID)
}

// armEP0Out re-arms EP0 OUT to receive the next SETUP or status ZLP from the host.
func armEP0Out() {
	// STUPCNT=3 (bits[30:29]=11): accept up to 3 back-to-back SETUPs.
	// PKTCNT=1 (bit[19]):         one packet.
	// XFRSIZ=64 (bits[6:0]):      max 64 bytes.
	otgOutEP(0).TSIZ.Set((3 << 29) | (1 << 19) | 64)
	otgOutEP(0).CTL.SetBits(depctlEPENA | depctlCNAK)
}

// flushTxFIFO flushes the selected TX FIFO(s).
// txfnum: 0–3 for a specific FIFO, 0x10 to flush all TX FIFOs.
func flushTxFIFO(txfnum uint32) {
	stm32.OTG_FS_GLOBAL.GRSTCTL.Set(
		grstTXFFLSH | (txfnum << grstTXFNUM_Pos),
	)
	for stm32.OTG_FS_GLOBAL.GRSTCTL.HasBits(grstTXFFLSH) {
	}
}

// flushRxFIFO flushes the shared RX FIFO.
func flushRxFIFO() {
	stm32.OTG_FS_GLOBAL.GRSTCTL.Set(grstRXFFLSH)
	for stm32.OTG_FS_GLOBAL.GRSTCTL.HasBits(grstRXFFLSH) {
	}
}
