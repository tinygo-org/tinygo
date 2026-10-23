//go:build esp32s3 && usbotg

package machine

import (
	"device/esp"
	"machine/usb"
	"runtime/interrupt"
	"runtime/volatile"
	"sync/atomic"
	"unsafe"
)

// USB-OTG device driver for the ESP32-S3, enabled with -tags=usbotg. The DWC2
// core runs in slave (FIFO) mode, see the ESP32-S3 TRM USB On-The-Go chapter.

const NumberOfUSBEndpoints = 8

// Physical endpoints in use. TinyGo classes only use endpoints 0 to 3.
const otgNumEndpoints = 4

const cpuInterruptFromUSB = 8

// Espressif VID with the CDC PID of the 0x4000 | class scheme that
// Espressif's esp_tinyusb component uses.
const (
	usb_VID                 = uint16(0x303A)
	usb_PID                 = uint16(0x4001)
	usb_STRING_MANUFACTURER = "TinyGo"
	usb_STRING_PRODUCT      = "ESP32-S3 USB Device"
)

// GPIO matrix inputs of the OTG core and the constant levels. See esp-idf
// soc/esp32s3/include/soc/gpio_sig_map.h and soc/gpio_pins.h.
const (
	usbOTGIDDIGSignal     = 58
	usbOTGAVALIDSignal    = 59
	usbOTGBVALIDSignal    = 60
	usbOTGVBUSVALIDSignal = 61

	gpioMatrixConstHigh = 0x38
	gpioMatrixConstLow  = 0x3C
)

type otgInEndpoint struct {
	CTL    volatile.Register32
	_      uint32
	INT    volatile.Register32
	_      uint32
	TSIZ   volatile.Register32
	_      uint32
	TXFSTS volatile.Register32
	_      uint32
}

type otgOutEndpoint struct {
	CTL  volatile.Register32
	_    uint32
	INT  volatile.Register32
	_    uint32
	TSIZ volatile.Register32
	_    [3]uint32
}

func otgInEP(ep uint32) *otgInEndpoint {
	return (*otgInEndpoint)(unsafe.Add(unsafe.Pointer(&esp.USB0.DIEPCTL0), uintptr(ep)*0x20))
}

func otgOutEP(ep uint32) *otgOutEndpoint {
	return (*otgOutEndpoint)(unsafe.Add(unsafe.Pointer(&esp.USB0.DOEPCTL0), uintptr(ep)*0x20))
}

// otgFIFO returns the data FIFO of an endpoint. Reads from any of them pop
// the shared RX FIFO.
func otgFIFO(ep uint32) *volatile.Register32 {
	return (*volatile.Register32)(unsafe.Add(unsafe.Pointer(esp.USB0), 0x1000+uintptr(ep)*0x1000))
}

// GUSBCFG bits.
const (
	gusbcfgPHYSEL   = 1 << 6
	gusbcfgTRDT_Pos = 10
	gusbcfgTRDT_Msk = 0xF << 10
	gusbcfgFHMOD    = 1 << 29
	gusbcfgFDMOD    = 1 << 30
)

// GRSTCTL bits.
const (
	grstCSRST      = 1 << 0
	grstRXFFLSH    = 1 << 4
	grstTXFFLSH    = 1 << 5
	grstTXFNUM_Pos = 6
	grstAHBIDL     = 1 << 31
)

// GINTSTS and GINTMSK bits.
const (
	gintCMOD      = 1 << 0
	gintRXFLVL    = 1 << 4
	gintGONAKEFF  = 1 << 7
	gintUSBSUSP   = 1 << 11
	gintUSBRST    = 1 << 12
	gintENUMDNE   = 1 << 13
	gintIEPINT    = 1 << 18
	gintOEPINT    = 1 << 19
	gintWKUPINT   = 1 << 31
	gahbcfgGINT   = 1 << 0
	dcfgDSPDFS48  = 3
	dcfgDAD_Pos   = 4
	dcfgDAD_Msk   = 0x7F << 4
	dctlSDIS      = 1 << 1
	dctlCGINAK    = 1 << 8
	dctlSGONAK    = 1 << 9
	dctlCGONAK    = 1 << 10
	pcgcctlStopCK = 1 << 0
)

// DIEPCTLn and DOEPCTLn bits.
const (
	depctlUSBAEP     = 1 << 15
	depctlETYP_Pos   = 18
	depctlSTALL      = 1 << 21
	depctlTXFNUM_Pos = 22
	depctlCNAK       = 1 << 26
	depctlSNAK       = 1 << 27
	depctlSD0PID     = 1 << 28
	depctlEPDIS      = 1 << 30
	depctlEPENA      = 1 << 31
)

// DIEPINTn and DOEPINTn bits.
const (
	depintXFRC   = 1 << 0
	depintEPDISD = 1 << 1
	depintSTUP   = 1 << 3
	depintINEPNE = 1 << 6
)

// GRXSTSP packet status values.
const (
	rxPktstsOUTData   = 2
	rxPktstsSetupData = 6
)

// FIFO sizes in words. The S3 has 256 words of FIFO RAM, see esp-hal
// esp-metadata/devices/esp32s3/soc.toml usb_otg.fifo_depth_words.
const (
	rxFIFOWords  = 112
	tx0FIFOWords = 16
	tx1FIFOWords = 16
	tx2FIFOWords = 48
	tx3FIFOWords = 48
)

var (
	sendOnEP0DATADONE struct {
		offset int
		data   []byte
	}

	usbSetupBuf [8]byte
	usbRxBufLen [NumberOfUSBEndpoints]uint32
	usbRxReady  [NumberOfUSBEndpoints]bool
)

// usbRegLock serializes FIFO and endpoint register access between the USB
// interrupt and senders on the other core with scheduler=cores.
var usbRegLock atomic.Uint32

func lockUSBRegs() interrupt.State {
	for {
		state := interrupt.Disable()
		if usbRegLock.CompareAndSwap(0, 1) {
			return state
		}
		interrupt.Restore(state)
	}
}

func unlockUSBRegs(state interrupt.State) {
	usbRegLock.Store(0)
	interrupt.Restore(state)
}

// Configure starts the USB-OTG peripheral in device mode. The config is here
// for compatibility with the UART interface.
func (dev *USBDevice) Configure(config UARTConfig) {
	if dev.initcomplete {
		return
	}

	// Drop the D+ pull-up of USB Serial/JTAG so the host sees a disconnect,
	// USB 2.0 section 7.1.7.3. Otherwise it keeps talking to the ROM device.
	esp.USB_DEVICE.SetCONF0_DP_PULLUP(0)
	esp.USB_DEVICE.SetCONF0_PAD_PULL_OVERRIDE(1)
	usbDelayMs(100)

	esp.SYSTEM.SetPERIP_CLK_EN0_USB_CLK_EN(1)
	esp.SYSTEM.SetPERIP_RST_EN0_USB_RST(1)
	esp.SYSTEM.SetPERIP_RST_EN0_USB_RST(0)

	// Route the internal PHY to USB-OTG instead of USB Serial/JTAG, as esp-hal
	// does in esp-hal/src/usb/otg/mod.rs common_init.
	esp.RTC_CNTL.SetUSB_CONF_SW_HW_USB_PHY_SEL(1)
	esp.RTC_CNTL.SetUSB_CONF_SW_USB_PHY_SEL(1)
	esp.USB_WRAP.SetOTG_CONF_PHY_SEL(0)
	esp.USB_WRAP.SetOTG_CONF_USB_PAD_ENABLE(1)
	esp.USB_WRAP.SetOTG_CONF_PAD_PULL_OVERRIDE(0)
	esp.USB_WRAP.SetOTG_CONF_CLK_EN(1)
	esp.USB_WRAP.SetOTG_CONF_AHB_CLK_FORCE_ON(1)
	esp.USB_WRAP.SetOTG_CONF_PHY_CLK_FORCE_ON(1)

	// There is no VBUS or ID pin, so tie the session signals to device mode.
	setUSBOTGInput(usbOTGIDDIGSignal, gpioMatrixConstHigh)
	setUSBOTGInput(usbOTGBVALIDSignal, gpioMatrixConstHigh)
	setUSBOTGInput(usbOTGVBUSVALIDSignal, gpioMatrixConstHigh)
	setUSBOTGInput(usbOTGAVALIDSignal, gpioMatrixConstLow)

	// D- is GPIO19 and D+ is GPIO20.
	GPIO19.ioMuxReg().ReplaceBits(3, 3, esp.IO_MUX_GPIO_FUN_DRV_Pos)
	GPIO20.ioMuxReg().ReplaceBits(3, 3, esp.IO_MUX_GPIO_FUN_DRV_Pos)

	waitUSBBits(&esp.USB0.GRSTCTL, grstAHBIDL, grstAHBIDL)
	esp.USB0.GRSTCTL.SetBits(grstCSRST)
	waitUSBBits(&esp.USB0.GRSTCTL, grstCSRST, 0)
	waitUSBBits(&esp.USB0.GRSTCTL, grstAHBIDL, grstAHBIDL)

	// TRDT 5 matches esp-hal calculate_trdt_fn for the S3.
	gusbcfg := esp.USB0.GUSBCFG.Get()
	gusbcfg &^= gusbcfgFHMOD | gusbcfgTRDT_Msk
	gusbcfg |= gusbcfgFDMOD | gusbcfgPHYSEL | 5<<gusbcfgTRDT_Pos
	esp.USB0.GUSBCFG.Set(gusbcfg)
	waitUSBBits(&esp.USB0.GINTSTS, gintCMOD, 0)

	esp.USB0.PCGCCTL.Set(0)
	dev.Detach()

	esp.USB0.GRXFSIZ.Set(rxFIFOWords)
	start := uint32(rxFIFOWords)
	esp.USB0.GNPTXFSIZ.Set(tx0FIFOWords<<16 | start)
	start += tx0FIFOWords
	esp.USB0.DIEPTXF1.Set(tx1FIFOWords<<16 | start)
	start += tx1FIFOWords
	esp.USB0.DIEPTXF2.Set(tx2FIFOWords<<16 | start)
	start += tx2FIFOWords
	esp.USB0.DIEPTXF3.Set(tx3FIFOWords<<16 | start)
	flushRxFIFO()
	flushTxFIFO(0x10)

	esp.USB0.DCFG.Set(dcfgDSPDFS48)

	esp.USB0.GINTSTS.Set(0xFFFFFFFF)
	esp.USB0.GINTMSK.Set(gintUSBRST | gintENUMDNE | gintRXFLVL | gintIEPINT | gintOEPINT |
		gintUSBSUSP | gintWKUPINT)
	esp.USB0.DIEPMSK.Set(depintXFRC)
	esp.USB0.DOEPMSK.Set(depintXFRC | depintSTUP)
	esp.USB0.GAHBCFG.SetBits(gahbcfgGINT)

	esp.INTERRUPT_CORE0.SetUSB_INTR_MAP(cpuInterruptFromUSB)
	interrupt.New(cpuInterruptFromUSB, handleUSBIRQ).Enable()

	dev.Attach()
	dev.initcomplete = true
}

func setUSBOTGInput(signal, level uint32) {
	inFunc(signal).Set(esp.GPIO_FUNC_IN_SEL_CFG_SEL | level<<esp.GPIO_FUNC_IN_SEL_CFG_IN_SEL_Pos)
}

// waitUSBBits waits until reg&mask equals value, with a limit so that a
// missing PHY clock does not hang the program.
func waitUSBBits(reg *volatile.Register32, mask, value uint32) bool {
	for i := 0; i < 1000000; i++ {
		if reg.Get()&mask == value {
			return true
		}
	}
	return false
}

// Attach connects the device to the USB bus by releasing soft disconnect.
func (dev *USBDevice) Attach() {
	esp.USB0.DCTL.ClearBits(dctlSDIS)
}

// Detach disconnects the device from the USB bus with soft disconnect.
func (dev *USBDevice) Detach() {
	esp.USB0.DCTL.SetBits(dctlSDIS)
}

func handleUSBIRQ(interrupt.Interrupt) {
	status := esp.USB0.GINTSTS.Get() & esp.USB0.GINTMSK.Get()

	if status&gintUSBSUSP != 0 {
		esp.USB0.GINTSTS.Set(gintUSBSUSP)
	}
	if status&gintWKUPINT != 0 {
		esp.USB0.GINTSTS.Set(gintWKUPINT)
		esp.USB0.PCGCCTL.ClearBits(pcgcctlStopCK)
	}
	if status&gintUSBRST != 0 {
		esp.USB0.GINTSTS.Set(gintUSBRST)
		handleUSBReset()
	}
	if status&gintENUMDNE != 0 {
		esp.USB0.GINTSTS.Set(gintENUMDNE)
		// EP0 max packet size field 0 is 64 bytes.
		otgInEP(0).CTL.ClearBits(3)
		esp.USB0.DCTL.SetBits(dctlCGINAK)
	}
	if status&gintRXFLVL != 0 {
		handleRxFIFO()
	}
	if status&gintOEPINT != 0 {
		handleOutEndpoints()
	}
	if status&gintIEPINT != 0 {
		handleInEndpoints()
	}
}

func handleUSBReset() {
	for ep := uint32(0); ep < otgNumEndpoints; ep++ {
		otgOutEP(ep).CTL.SetBits(depctlSNAK)
		otgInEP(ep).CTL.SetBits(depctlSNAK)
		usbRxReady[ep] = false
	}
	esp.USB0.DCFG.ClearBits(dcfgDAD_Msk)
	flushRxFIFO()
	flushTxFIFO(0x10)

	esp.USB0.DAINT.Set(0xFFFFFFFF)
	esp.USB0.DAINTMSK.Set(1<<0 | 1<<16)
	sendOnEP0DATADONE.data = nil
	armEP0Out()

	usbConfiguration = 0
	USBDev.InitEndpointComplete = false
}

// handleRxFIFO drains the shared RX FIFO. The data of each OUT packet goes to
// the cache buffer of its endpoint.
func handleRxFIFO() {
	state := lockUSBRegs()
	for esp.USB0.GINTSTS.Get()&gintRXFLVL != 0 {
		status := esp.USB0.GRXSTSP.Get()
		ep := status & 0xF
		bcnt := (status >> 4) & 0x7FF

		switch (status >> 17) & 0xF {
		case rxPktstsSetupData:
			w0 := otgFIFO(0).Get()
			w1 := otgFIFO(0).Get()
			for i := 0; i < 4; i++ {
				usbSetupBuf[i] = byte(w0 >> (8 * i))
				usbSetupBuf[i+4] = byte(w1 >> (8 * i))
			}
		case rxPktstsOUTData:
			readUSBFIFO(ep, bcnt)
		}
	}
	unlockUSBRegs(state)
}

func readUSBFIFO(ep, bcnt uint32) {
	words := (bcnt + 3) / 4
	if ep >= NumberOfUSBEndpoints || bcnt > uint32(len(udd_ep_out_cache_buffer[0])) {
		for i := uint32(0); i < words; i++ {
			otgFIFO(0).Get()
		}
		return
	}
	buf := udd_ep_out_cache_buffer[ep][:]
	for i := uint32(0); i < words; i++ {
		w := otgFIFO(0).Get()
		for j := uint32(0); j < 4 && i*4+j < bcnt; j++ {
			buf[i*4+j] = byte(w >> (8 * j))
		}
	}
	usbRxBufLen[ep] = bcnt
	usbRxReady[ep] = true
}

func handleOutEndpoints() {
	active := (esp.USB0.DAINT.Get() & esp.USB0.DAINTMSK.Get()) >> 16
	for ep := uint32(0); ep < otgNumEndpoints; ep++ {
		if active&(1<<ep) == 0 {
			continue
		}
		out := otgOutEP(ep)
		flags := out.INT.Get()
		out.INT.Set(flags)

		if ep == 0 {
			if flags&depintSTUP != 0 {
				handleUSBSetup()
			} else if flags&depintXFRC != 0 {
				usbRxReady[0] = false
				armEP0Out()
			}
			continue
		}
		if flags&depintXFRC != 0 {
			usbRxReady[ep] = false
			buf := udd_ep_out_cache_buffer[ep][:usbRxBufLen[ep]]
			if usbRxHandler[ep] == nil || usbRxHandler[ep](buf) {
				AckUsbOutTransfer(ep)
			}
		}
	}
}

func handleUSBSetup() {
	setup := usb.NewSetup(usbSetupBuf[:])
	sendOnEP0DATADONE.data = nil

	ok := false
	if setup.BmRequestType&usb.REQUEST_TYPE == usb.REQUEST_STANDARD {
		ok = handleStandardSetup(setup)
	} else if setup.WIndex < uint16(len(usbSetupHandler)) && usbSetupHandler[setup.WIndex] != nil {
		ok = usbSetupHandler[setup.WIndex](setup)
	}
	if !ok {
		// The core clears an EP0 stall on the next SETUP.
		otgInEP(0).CTL.SetBits(depctlSTALL)
		otgOutEP(0).CTL.SetBits(depctlSTALL)
	}
	armEP0Out()
}

func handleInEndpoints() {
	active := esp.USB0.DAINT.Get() & esp.USB0.DAINTMSK.Get() & 0xFFFF
	for ep := uint32(0); ep < otgNumEndpoints; ep++ {
		if active&(1<<ep) == 0 {
			continue
		}
		in := otgInEP(ep)
		flags := in.INT.Get()
		in.INT.Set(flags)
		if flags&depintXFRC == 0 {
			continue
		}

		if ep != 0 {
			if usbTxHandler[ep] != nil {
				usbTxHandler[ep]()
			}
			continue
		}
		if data := sendOnEP0DATADONE.data; data != nil {
			offset := sendOnEP0DATADONE.offset
			count := min(len(data)-offset, usb.EndpointPacketSize)
			sendOnEP0DATADONE.offset += count
			if sendOnEP0DATADONE.offset >= len(data) {
				sendOnEP0DATADONE.data = nil
			}
			sendViaEPIn(0, data[offset:offset+count])
		} else {
			armEP0Out()
		}
	}
}

func initEndpoint(ep, config uint32) {
	if ep == 0 || ep >= otgNumEndpoints {
		return
	}
	epType := config &^ usb.EndpointIn
	ctl := uint32(usb.EndpointPacketSize) | depctlUSBAEP | depctlSD0PID | epType<<depctlETYP_Pos

	if config&usb.EndpointIn != 0 {
		otgInEP(ep).CTL.Set(ctl | ep<<depctlTXFNUM_Pos | depctlSNAK)
		esp.USB0.DAINTMSK.SetBits(1 << ep)
		return
	}
	otgOutEP(ep).CTL.Set(ctl)
	esp.USB0.DAINTMSK.SetBits(1 << (ep + 16))
	AckUsbOutTransfer(ep)
}

// SendUSBInPacket sends a packet for USB (interrupt in / bulk in).
func SendUSBInPacket(ep uint32, data []byte) bool {
	sendUSBPacket(ep, data)
	return true
}

//go:noinline
func sendUSBPacket(ep uint32, data []byte) {
	count := len(data)
	if ep == 0 {
		sendOnEP0DATADONE.data = nil
		if count > usb.EndpointPacketSize {
			count = usb.EndpointPacketSize
			sendOnEP0DATADONE.offset = count
			sendOnEP0DATADONE.data = data
		}
	}
	sendViaEPIn(ep, data[:count])
}

// sendViaEPIn starts a one packet IN transfer and pushes the data into the
// TX FIFO of the endpoint.
func sendViaEPIn(ep uint32, data []byte) {
	state := lockUSBRegs()
	in := otgInEP(ep)
	in.TSIZ.Set(1<<19 | uint32(len(data)))
	in.CTL.SetBits(depctlEPENA | depctlCNAK)

	fifo := otgFIFO(ep)
	for i := 0; i < len(data); i += 4 {
		var w uint32
		for j := 0; j < 4 && i+j < len(data); j++ {
			w |= uint32(data[i+j]) << (8 * j)
		}
		fifo.Set(w)
	}
	unlockUSBRegs(state)
}

// SendZlp sends a zero length packet on EP0 IN.
func SendZlp() {
	sendUSBPacket(0, nil)
}

// AckUsbOutTransfer re-arms an OUT endpoint for the next packet.
func AckUsbOutTransfer(ep uint32) {
	ep &= 0x7F
	state := lockUSBRegs()
	usbRxBufLen[ep] = 0
	out := otgOutEP(ep)
	out.TSIZ.Set(1<<19 | usb.EndpointPacketSize)
	out.CTL.SetBits(depctlEPENA | depctlCNAK)
	unlockUSBRegs(state)
}

// armEP0Out enables EP0 OUT for a data or status packet and up to 3
// back to back SETUP packets.
func armEP0Out() {
	state := lockUSBRegs()
	otgOutEP(0).TSIZ.Set(3<<29 | 1<<19 | usb.EndpointPacketSize)
	otgOutEP(0).CTL.SetBits(depctlEPENA | depctlCNAK)
	unlockUSBRegs(state)
}

// handleUSBSetAddress stores the new address before the status ZLP goes out.
// The core still answers the status stage on address 0.
func handleUSBSetAddress(setup usb.Setup) bool {
	dcfg := esp.USB0.DCFG.Get() &^ dcfgDAD_Msk
	esp.USB0.DCFG.Set(dcfg | uint32(setup.WValueL&0x7F)<<dcfgDAD_Pos)
	SendZlp()
	return true
}

// ReceiveUSBControlPacket receives the data stage of a control OUT request.
// It runs inside the SETUP handler, so it drains the RX FIFO itself.
func ReceiveUSBControlPacket() ([cdcLineInfoSize]byte, error) {
	var b [cdcLineInfoSize]byte
	usbRxReady[0] = false
	armEP0Out()

	for i := 0; i < 2000000; i++ {
		if esp.USB0.GINTSTS.Get()&gintRXFLVL != 0 {
			handleRxFIFO()
		}
		if usbRxReady[0] {
			usbRxReady[0] = false
			copy(b[:], udd_ep_out_cache_buffer[0][:usbRxBufLen[0]])
			return b, nil
		}
	}
	return b, ErrUSBReadTimeout
}

// SetStallEPIn stalls an IN endpoint.
func (dev *USBDevice) SetStallEPIn(ep uint32) {
	ep &= 0x7F
	state := lockUSBRegs()
	in := otgInEP(ep)
	if ep == 0 || in.CTL.Get()&depctlEPENA == 0 {
		in.CTL.SetBits(depctlSNAK | depctlSTALL)
	} else {
		in.CTL.SetBits(depctlSNAK)
		waitUSBBits(&in.INT, depintINEPNE, depintINEPNE)
		in.CTL.SetBits(depctlEPDIS | depctlSTALL)
		waitUSBBits(&in.INT, depintEPDISD, depintEPDISD)
		in.INT.Set(depintEPDISD | depintINEPNE)
	}
	flushTxFIFO(ep)
	unlockUSBRegs(state)
}

// SetStallEPOut stalls an OUT endpoint.
func (dev *USBDevice) SetStallEPOut(ep uint32) {
	ep &= 0x7F
	state := lockUSBRegs()
	out := otgOutEP(ep)
	if ep == 0 || out.CTL.Get()&depctlEPENA == 0 {
		out.CTL.SetBits(depctlSTALL)
		unlockUSBRegs(state)
		return
	}
	// Disabling an enabled OUT endpoint needs a global OUT NAK, as in
	// TinyUSB src/portable/synopsys/dwc2/dcd_dwc2.c edpt_disable.
	esp.USB0.DCTL.SetBits(dctlSGONAK)
	waitUSBBits(&esp.USB0.GINTSTS, gintGONAKEFF, gintGONAKEFF)
	out.CTL.SetBits(depctlEPDIS | depctlSTALL)
	waitUSBBits(&out.INT, depintEPDISD, depintEPDISD)
	out.INT.Set(depintEPDISD)
	esp.USB0.DCTL.SetBits(dctlCGONAK)
	unlockUSBRegs(state)
}

// ClearStallEPIn clears the stall on an IN endpoint and resets it to DATA0.
func (dev *USBDevice) ClearStallEPIn(ep uint32) {
	ep &= 0x7F
	state := lockUSBRegs()
	ctl := &otgInEP(ep).CTL
	ctl.ClearBits(depctlSTALL)
	ctl.SetBits(depctlSD0PID)
	unlockUSBRegs(state)
}

// ClearStallEPOut clears the stall on an OUT endpoint, resets it to DATA0
// and re-arms it.
func (dev *USBDevice) ClearStallEPOut(ep uint32) {
	ep &= 0x7F
	state := lockUSBRegs()
	ctl := &otgOutEP(ep).CTL
	ctl.ClearBits(depctlSTALL)
	ctl.SetBits(depctlSD0PID)
	unlockUSBRegs(state)
	AckUsbOutTransfer(ep)
}

func flushTxFIFO(txfnum uint32) {
	esp.USB0.GRSTCTL.Set(grstTXFFLSH | txfnum<<grstTXFNUM_Pos)
	waitUSBBits(&esp.USB0.GRSTCTL, grstTXFFLSH, 0)
}

func flushRxFIFO() {
	esp.USB0.GRSTCTL.Set(grstRXFFLSH)
	waitUSBBits(&esp.USB0.GRSTCTL, grstRXFFLSH, 0)
}

// FlushSerial is a no-op, the CDC class sends from its own ring buffer.
func FlushSerial() {}

// EnterBootloader restarts into the ROM download mode on USB Serial/JTAG.
func EnterBootloader() {
	interrupt.Disable()

	// Drop off the bus first so the host enumerates USB Serial/JTAG anew.
	esp.USB0.DCTL.SetBits(dctlSDIS)
	usbDelayMs(500)
	usbPHYToSerialJTAG()

	// Reset both cores, as esp_restart does. A core reset keeps the RTC
	// registers, so the ROM sees FORCE_DOWNLOAD_BOOT and the PHY routing.
	esp.RTC_CNTL.SetOPTION1_FORCE_DOWNLOAD_BOOT(1)
	esp.RTC_CNTL.OPTIONS0.SetBits(esp.RTC_CNTL_OPTIONS0_SW_APPCPU_RST | esp.RTC_CNTL_OPTIONS0_SW_PROCPU_RST)
	for {
	}
}

// usbDelayMs spins without timers or interrupts. One RTC register read takes
// about 365 ns, measured on a XIAO ESP32-S3 at 240 MHz.
func usbDelayMs(ms int) {
	for i := 0; i < ms*2740; i++ {
		_ = esp.RTC_CNTL.OPTIONS0.Get()
	}
}

// usbPHYToSerialJTAG gives the internal PHY back to USB Serial/JTAG, as
// usb_phy_ll_int_jtag_enable in esp-idf hal/esp32s3/include/hal/usb_phy_ll.h.
func usbPHYToSerialJTAG() {
	esp.USB_WRAP.SetOTG_CONF_USB_PAD_ENABLE(0)
	esp.SYSTEM.SetPERIP_RST_EN0_USB_RST(1)
	esp.SYSTEM.SetPERIP_CLK_EN0_USB_CLK_EN(0)

	esp.SYSTEM.SetPERIP_CLK_EN1_USB_DEVICE_CLK_EN(1)
	esp.SYSTEM.SetPERIP_RST_EN1_USB_DEVICE_RST(0)
	esp.USB_DEVICE.SetCONF0_PHY_SEL(0)
	esp.USB_DEVICE.SetCONF0_PAD_PULL_OVERRIDE(0)
	esp.USB_DEVICE.SetCONF0_DP_PULLUP(1)
	esp.USB_DEVICE.SetCONF0_USB_PAD_ENABLE(1)
	esp.RTC_CNTL.SetUSB_CONF_SW_HW_USB_PHY_SEL(1)
	esp.RTC_CNTL.SetUSB_CONF_SW_USB_PHY_SEL(0)
}
