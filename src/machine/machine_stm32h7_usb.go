//go:build stm32 && stm32h7

package machine

import (
	"device/arm"
	"device/stm32"
	"machine/usb"
	"runtime/interrupt"
	"runtime/volatile"
	"unsafe"
)

// Synopsys DesignWare OTG registers
// The SVD-generated Go device file is missing some device-mode registers,
// so we define them here based on the Synopsys OTG IP.
type usbOTGRegs struct {
	// Global registers (0x000)
	GOTGCTL   volatile.Register32 // 0x00
	GOTGINT   volatile.Register32 // 0x04
	GAHBCFG   volatile.Register32 // 0x08
	GUSBCFG   volatile.Register32 // 0xC
	GRSTCTL   volatile.Register32 // 0x10
	GINTSTS   volatile.Register32 // 0x14
	GINTMSK   volatile.Register32 // 0x18
	GRXSTSR   volatile.Register32 // 0x1C
	GRXSTSP   volatile.Register32 // 0x20
	GRXFSIZ   volatile.Register32 // 0x24
	GNPTXFSIZ volatile.Register32 // 0x28
	GNPTXSTS  volatile.Register32 // 0x2C
	_         [8]byte
	GCCFG     volatile.Register32 // 0x38
	CID       volatile.Register32 // 0x3C
	_         [20]byte
	GLPMCFG   volatile.Register32 // 0x54
	GPWRDN    volatile.Register32 // 0x58
	_         [4]byte
	GDFIFO_S  volatile.Register32 // 0x60
	_         [156]byte
	HPTXFSIZ  volatile.Register32     // 0x100
	DIEPTXF   [15]volatile.Register32 // 0x104
	_         [1728]byte

	// Device registers (0x800)
	DCFG       volatile.Register32 // 0x800
	DCTL       volatile.Register32 // 0x804
	DSTS       volatile.Register32 // 0x808
	_          [4]byte
	DIEPMSK    volatile.Register32 // 0x810
	DOEPMSK    volatile.Register32 // 0x814
	DAINT      volatile.Register32 // 0x818
	DAINTMSK   volatile.Register32 // 0x81C
	_          [32]byte
	DIEPEMPMSK volatile.Register32 // 0x840
	_          [188]byte

	// Endpoint registers
	INEP [16]struct {
		CTL    volatile.Register32 // 0x900 + n*0x20
		_      [4]byte
		INT    volatile.Register32 // 0x908 + n*0x20
		_      [4]byte
		TSIZ   volatile.Register32 // 0x910 + n*0x20
		DMA    volatile.Register32 // 0x914 + n*0x20
		TXFSTS volatile.Register32 // 0x918 + n*0x20
		_      [4]byte
	}
	OUTEP [16]struct {
		CTL  volatile.Register32 // 0xB00 + n*0x20
		_    [4]byte
		INT  volatile.Register32 // 0xB08 + n*0x20
		_    [4]byte
		TSIZ volatile.Register32 // 0xB10 + n*0x20
		DMA  volatile.Register32 // 0xB14 + n*0x20
		_    [8]byte
	}

	_ [256]byte

	// Power and clock gating registers (0xE00)
	PCGCCTL volatile.Register32 // 0xE00
}

// USB2 OTG_FS: the FS-only core wired to PA11/PA12 (Nucleo CN13).
// USB1 OTG_HS at 0x40040000 uses ULPI or its own embedded PHY on PB14/PB15,
// which is NOT routed to the user USB connector on this board.
var usbOTG = (*usbOTGRegs)(unsafe.Pointer(uintptr(0x40080000)))

const (
	// GUSBCFG bits
	GUSBCFG_PHYSEL   = 1 << 6
	GUSBCFG_TRDT_Pos = 10
	GUSBCFG_FDMOD    = 1 << 30

	// GAHBCFG bits
	GAHBCFG_GINT = 1 << 0

	// GRSTCTL bits
	GRSTCTL_CSRST      = 1 << 0
	GRSTCTL_RXFFLSH    = 1 << 4
	GRSTCTL_TXFFLSH    = 1 << 5
	GRSTCTL_TXFNUM_ALL = 0x10 << 6
	GRSTCTL_AHBIDL     = 1 << 31

	// GINTSTS / GINTMSK bits
	GINT_RXFLVL   = 1 << 4
	GINT_GINAKEFF = 1 << 6
	GINT_GONAKEFF = 1 << 7
	GINT_USBSUSP  = 1 << 11
	GINT_USBRST   = 1 << 12
	GINT_ENUMDNE  = 1 << 13
	GINT_IEPINT   = 1 << 18
	GINT_OEPINT   = 1 << 19

	// DCFG bits
	DCFG_DSPD_FS = 0x3 << 0

	// DCTL bits
	DCTL_RWUSIG = 1 << 0
	DCTL_SDIS   = 1 << 1
	DCTL_GINSTS = 1 << 2
	DCTL_GONSTS = 1 << 3

	// DIEPCTL / DOEPCTL bits
	DEPCTL_MPSIZ_Pos  = 0
	DEPCTL_USBAEP     = 1 << 15
	DEPCTL_EPTYP_Pos  = 18
	DEPCTL_STALL      = 1 << 21
	DEPCTL_CNAK       = 1 << 26
	DEPCTL_SNAK       = 1 << 27
	DEPCTL_TXFNUM_Pos = 22
	DEPCTL_EPDIS      = 1 << 30
	DEPCTL_EPENA      = 1 << 31

	// DIEPINT / DOEPINT bits
	DEPINT_XFERC  = 1 << 0
	DEPINT_EPDISD = 1 << 1
	DEPINT_SETUP  = 1 << 3

	NumberOfUSBEndpoints = 9

	// FIFO layout in 32-bit words: shared RX FIFO plus one 64-word TX FIFO
	// for EP0 and each of the 8 IN endpoints (256 + 9*64 = 832 ≤ 1024).
	rxFIFOWords = 256
	txFIFOWords = 64
)

var (
	// ep0OutReceived signals that an OUT packet was received on EP0.
	// Volatile: written from the USB interrupt, busy-waited on from thread mode.
	ep0OutReceived volatile.Register8
)

// Configure the USB peripheral.
func (dev *USBDevice) Configure(config UARTConfig) {
	if dev.initcomplete {
		return
	}

	// 1. Enable clocks
	stm32.RCC.AHB1ENR.SetBits(stm32.RCC_AHB1ENR_USB2OTGHSEN)
	// The FS core has no ULPI clock, but AHB1LPENR resets with
	// USB2OTGULPILPEN set, so in CPU Sleep mode (the scheduler's WFE) the RCC
	// waits on a ULPI clock that never comes and the core's AHB interface
	// stalls, killing USB whenever the CPU sleeps.
	// Keep the OTG bus clock running in Sleep, drop the ULPI one.
	stm32.RCC.AHB1LPENR.SetBits(stm32.RCC_AHB1LPENR_USB2OTGLPEN)
	stm32.RCC.AHB1LPENR.ClearBits(stm32.RCC_AHB1LPENR_USB2OTGULPILPEN)
	// Enable USB regulator (USB33DEN) for internal PHY.
	// Already done in initCLK, but setting here as well for safety.
	stm32.PWR.CR3.SetBits(stm32.PWR_CR3_USB33DEN)

	// Pulse RCC reset to clear any stale state from a warm reset.
	stm32.RCC.AHB1RSTR.SetBits(stm32.RCC_AHB1RSTR_USB2OTGRST)
	stm32.RCC.AHB1RSTR.ClearBits(stm32.RCC_AHB1RSTR_USB2OTGRST)

	// 2. Setup pins (PA11=DM, PA12=DP) for USB2 OTG_FS — AF10.
	PA11.ConfigureAltFunc(PinConfig{Mode: PinModeUSB}, AF10_OTG_HS_FS_SAI2_QUADSPI_SDMMC2)
	PA12.ConfigureAltFunc(PinConfig{Mode: PinModeUSB}, AF10_OTG_HS_FS_SAI2_QUADSPI_SDMMC2)

	// 3. Select internal FS PHY BEFORE the core reset below — the reset FSM
	// samples the PHY clock, which only runs once PHYSEL is set. Give the
	// clock a few cycles to start or CSRST can hang / self-clear too early.
	usbOTG.GUSBCFG.SetBits(GUSBCFG_PHYSEL)
	for j := 0; j < 10_000; j++ {
		arm.Asm("nop")
	}

	// 4. Core Reset — wait for AHB idle then pulse CSRST.
	for usbOTG.GRSTCTL.Get()&GRSTCTL_AHBIDL == 0 {
	}
	usbOTG.GRSTCTL.SetBits(GRSTCTL_CSRST)
	for usbOTG.GRSTCTL.Get()&GRSTCTL_CSRST != 0 {
	}

	// Power up the FS transceiver AFTER the core reset: CSRST wipes GCCFG,
	// so setting PWRDWN earlier leaves the transceiver off and the DP
	// pull-up never appears (host sees no cable).
	// No HW VBUS sensing: CN13's VBUS pin is not wired to the MCU VBUS-sense
	// input on this board, so leave GCCFG.VBDEN (bit 21) clear and force
	// session/VBUS valid via GOTGCTL instead (below).
	usbOTG.GCCFG.Set(1 << 16) // PWRDWN

	// Make sure the PHY clock is not gated (e.g. by a bootloader).
	usbOTG.PCGCCTL.Set(0)

	// Stay soft-disconnected until configuration is complete; CSRST left
	// DCTL at its default "connected" state.
	dev.Detach()

	// 5. Force device mode now that the core is out of reset. The mode
	// change takes effect only after up to 25 ms (RM0433); poll GINTSTS.CMOD
	// (bit 0: 0 = device) with a generous busy-wait bound.
	usbOTG.GUSBCFG.SetBits(GUSBCFG_FDMOD)
	for j := 0; j < 20_000_000 && usbOTG.GINTSTS.Get()&0x1 != 0; j++ {
		arm.Asm("nop")
	}

	// Override all session/VBUS valid bits regardless of hardware pin state.
	// GOTGCTL[2]=VBVALOEN, [3]=VBVALOVAL, [6]=BVALOEN, [7]=BVALOVAL.
	usbOTG.GOTGCTL.SetBits(0x4 | 0x8 | 0x40 | 0x80)

	// Set turnaround time: HCLK=200MHz → TRDT=6 per RM0433 Table 362.
	usbOTG.GUSBCFG.ReplaceBits(0x6<<GUSBCFG_TRDT_Pos, 0xF<<GUSBCFG_TRDT_Pos, 0)

	// 6. FIFO Configuration (total 1024 words shared by RX + all TX FIFOs).
	// initEndpoint assigns TX FIFO n to IN endpoint n, so every IN endpoint
	// 1..8 needs a configured FIFO even if the current class uses only a few.
	usbOTG.GRXFSIZ.Set(rxFIFOWords)
	usbOTG.GNPTXFSIZ.Set(txFIFOWords<<16 | rxFIFOWords)
	for i, offset := 0, uint32(rxFIFOWords+txFIFOWords); i < 8; i++ {
		usbOTG.DIEPTXF[i].Set(txFIFOWords<<16 | offset) // DIEPTXF[i] = FIFO i+1
		offset += txFIFOWords
	}

	// Flush all FIFOs after (re)sizing them.
	usbOTG.GRSTCTL.SetBits(GRSTCTL_RXFFLSH)
	for usbOTG.GRSTCTL.Get()&GRSTCTL_RXFFLSH != 0 {
	}
	usbOTG.GRSTCTL.SetBits(GRSTCTL_TXFFLSH | GRSTCTL_TXFNUM_ALL)
	for usbOTG.GRSTCTL.Get()&GRSTCTL_TXFFLSH != 0 {
	}

	// 7. Device Configuration
	// Device Speed (FS)
	usbOTG.DCFG.ReplaceBits(DCFG_DSPD_FS, 0x3, 0)

	// 8. Per-endpoint interrupt masks
	usbOTG.DIEPMSK.Set(DEPINT_XFERC)
	usbOTG.DOEPMSK.Set(DEPINT_XFERC | DEPINT_SETUP)

	// 9. Interrupts
	// Clear anything pending, then unmask Reset, Enumeration Done,
	// RX FIFO Non-Empty, Setup Done (via OEPINT).
	usbOTG.GINTSTS.Set(0xFFFFFFFF)
	usbOTG.GINTMSK.SetBits(GINT_USBSUSP | GINT_USBRST | GINT_ENUMDNE | GINT_RXFLVL | GINT_IEPINT | GINT_OEPINT)
	// Global Interrupt Enable
	usbOTG.GAHBCFG.SetBits(GAHBCFG_GINT)

	// 10. Enable IRQ
	i := interrupt.New(stm32.IRQ_OTG_FS, handleUSBIRQ)
	i.SetPriority(0)
	i.Enable()

	dev.initcomplete = true

	// Release soft-disconnect: pulls D+ high, making device visible to host.
	dev.Attach()
}

// Attach connects the device to the USB bus by releasing soft disconnect,
// allowing the host to detect and enumerate it. It can be used together with
// Detach to delay enumeration until the USB configuration (device
// identifiers, classes, ...) is complete.
func (dev *USBDevice) Attach() {
	usbOTG.DCTL.ClearBits(DCTL_SDIS)
}

// Detach disconnects the device from the USB bus by asserting soft
// disconnect. To the host this appears as if the device was unplugged. A
// subsequent Attach makes the host enumerate the device again.
func (dev *USBDevice) Detach() {
	usbOTG.DCTL.SetBits(DCTL_SDIS)
}

func initEndpoint(ep, config uint32) {
	if ep == 0 {
		// Control endpoint
		// IN
		usbOTG.INEP[0].CTL.ReplaceBits(0, 0x3, DEPCTL_MPSIZ_Pos) // Max packet size 64 (00)
		usbOTG.INEP[0].INT.Set(0xFF)                             // Clear interrupts
		// OUT
		usbOTG.OUTEP[0].CTL.ReplaceBits(0, 0x3, DEPCTL_MPSIZ_Pos) // Max packet size 64 (00)
		usbOTG.OUTEP[0].INT.Set(0xFF)                             // Clear interrupts

		// Unmask interrupts for EP0
		usbOTG.DAINTMSK.SetBits(0x10001) // EP0 IN and OUT
	} else {
		isIn := (config & uint32(usb.EndpointIn)) != 0
		typ := config & 0x03

		if isIn {
			// Configure IN endpoint — do NOT set EPENA; set it only when queuing a transfer.
			ctl := uint32(DEPCTL_USBAEP)
			ctl |= (typ << DEPCTL_EPTYP_Pos)
			ctl |= (ep << DEPCTL_TXFNUM_Pos)
			ctl |= (64 << DEPCTL_MPSIZ_Pos) // MPS = 64 bytes
			ctl |= DEPCTL_SNAK              // Start NAKing until data is ready
			usbOTG.INEP[ep].CTL.Set(ctl)
			usbOTG.INEP[ep].INT.Set(0xFF) // Clear any stale interrupts
			usbOTG.DAINTMSK.SetBits(1 << ep)
		} else {
			// Configure OUT endpoint — do NOT set EPENA here; AckUsbOutTransfer arms it.
			ctl := uint32(DEPCTL_USBAEP)
			ctl |= (typ << DEPCTL_EPTYP_Pos)
			ctl |= (64 << DEPCTL_MPSIZ_Pos) // MPS = 64 bytes
			ctl |= DEPCTL_SNAK
			usbOTG.OUTEP[ep].CTL.Set(ctl)
			usbOTG.OUTEP[ep].INT.Set(0xFF) // Clear any stale interrupts
			usbOTG.DAINTMSK.SetBits(1 << (ep + 16))
			// Arm immediately so host can send data.
			AckUsbOutTransfer(ep)
		}
	}
}

func handleUSBSetAddress(setup usb.Setup) bool {
	addr := uint32(setup.WValueL)
	usbOTG.DCFG.ReplaceBits(addr<<4, 0x7F<<4, 0)
	SendZlp()
	return true
}

func SendZlp() {
	sendUSBPacket(0, nil)
}

func sendUSBPacket(ep uint32, data []byte) {
	// 1. Wait until the TX FIFO has room for the whole transfer, so a packet
	// queued while the previous one is still draining cannot corrupt the FIFO.
	// DTXFSTS reports free space in words; bounded wait in case the endpoint
	// is stuck (e.g. host stopped polling).
	words := uint32((len(data) + 3) / 4)
	for i := 0; i < 1_000_000 && usbOTG.INEP[ep].TXFSTS.Get()&0xFFFF < words; i++ {
	}

	// 2. Setup transfer size
	pktCnt := uint32((len(data) + 63) / 64)
	if len(data) == 0 {
		pktCnt = 1
	}
	usbOTG.INEP[ep].TSIZ.Set(uint32(len(data)) | (pktCnt << 19))

	// 3. Enable endpoint and clear NAK
	usbOTG.INEP[ep].CTL.SetBits(DEPCTL_EPENA | DEPCTL_CNAK)

	// 4. Write data to FIFO
	// FIFOs are at 0x1000, 0x2000, ... from base
	fifo := (*volatile.Register32)(unsafe.Pointer(uintptr(unsafe.Pointer(usbOTG)) + 0x1000 + uintptr(ep)*0x1000))
	for i := 0; i < len(data); i += 4 {
		var word uint32
		for j := 0; j < 4 && i+j < len(data); j++ {
			word |= uint32(data[i+j]) << (8 * j)
		}
		fifo.Set(word)
	}
}

func AckUsbOutTransfer(ep uint32) {
	// Prepare for next OUT transfer
	if ep == 0 {
		// EP0 OUT: 1 packet, 64 bytes, 3 SETUP packets
		usbOTG.OUTEP[0].TSIZ.Set(64 | (1 << 19) | (3 << 29))
	} else {
		usbOTG.OUTEP[ep].TSIZ.Set(64 | (1 << 19))
	}
	usbOTG.OUTEP[ep].CTL.SetBits(DEPCTL_EPENA | DEPCTL_CNAK)
}

func (dev *USBDevice) SetStallEPIn(ep uint32) {
	usbOTG.INEP[ep].CTL.SetBits(DEPCTL_STALL)
}

func (dev *USBDevice) SetStallEPOut(ep uint32) {
	usbOTG.OUTEP[ep].CTL.SetBits(DEPCTL_STALL)
}

func (dev *USBDevice) ClearStallEPIn(ep uint32) {
	usbOTG.INEP[ep].CTL.ClearBits(DEPCTL_STALL)
	usbOTG.INEP[ep].CTL.SetBits(1 << 28) // SD0PID
}

func (dev *USBDevice) ClearStallEPOut(ep uint32) {
	usbOTG.OUTEP[ep].CTL.ClearBits(DEPCTL_STALL)
	usbOTG.OUTEP[ep].CTL.SetBits(1 << 28) // SD0PID
}

// SendUSBInPacket sends a packet for USB (interrupt in / bulk in).
func SendUSBInPacket(ep uint32, data []byte) bool {
	sendUSBPacket(ep, data)
	return true
}

// ReceiveUSBControlPacket receives a control packet (used for CDC line coding).
//
// This runs inside the setup handler, which itself runs inside handleUSBIRQ.
// The interrupt cannot re-enter to deliver the data stage, so the RX FIFO is
// drained manually here until the EP0 OUT packet arrives.
func ReceiveUSBControlPacket() ([cdcLineInfoSize]byte, error) {
	var b [cdcLineInfoSize]byte
	ep0OutReceived.Set(0)
	for i := 0; i < 1_000_000; i++ {
		if usbOTG.GINTSTS.Get()&GINT_RXFLVL != 0 {
			handleRxFIFO()
		}
		if ep0OutReceived.Get() != 0 {
			copy(b[:], udd_ep_out_cache_buffer[0][:])
			ep0OutReceived.Set(0)
			return b, nil
		}
	}
	return b, ErrUSBReadTimeout
}

// handleRxFIFO pops one status entry from the shared RX FIFO and processes it.
// Called from the USB interrupt, and re-entrantly from
// ReceiveUSBControlPacket while a setup handler is waiting for the data stage.
func handleRxFIFO() {
	pop := usbOTG.GRXSTSP.Get()
	ep := pop & 0xF
	byteCnt := (pop >> 4) & 0x7FF
	pktSts := (pop >> 17) & 0xF

	// All OUT/SETUP data is read from the shared RX FIFO (DFIFO[0]).
	fifo := (*volatile.Register32)(unsafe.Pointer(uintptr(unsafe.Pointer(usbOTG)) + 0x1000))

	switch pktSts {
	case 0x2: // OUT data packet received
		// Guard against out-of-range endpoint or oversized packet: both come
		// straight from hardware and would panic if used to slice the 64-byte
		// cache buffers. Drain and discard instead.
		if ep >= NumberOfUSBEndpoints || byteCnt > uint32(len(udd_ep_out_cache_buffer[0])) {
			for i := uint32(0); i < byteCnt; i += 4 {
				fifo.Get()
			}
			return
		}

		buf := udd_ep_out_cache_buffer[ep][:byteCnt]
		for i := uint32(0); i < byteCnt; i += 4 {
			word := fifo.Get()
			for j := uint32(0); j < 4 && i+j < byteCnt; j++ {
				buf[i+j] = byte(word >> (8 * j))
			}
		}

		if ep == 0 {
			ep0OutReceived.Set(1)
			AckUsbOutTransfer(0)
		} else if usbRxHandler[ep] != nil {
			if usbRxHandler[ep](buf) {
				AckUsbOutTransfer(ep)
			}
		}

	case 0x6: // SETUP data packet received (always 8 bytes)
		setupBuf := udd_ep_out_cache_buffer[0][:8]
		for i := uint32(0); i < 8; i += 4 {
			word := fifo.Get()
			setupBuf[i] = byte(word)
			setupBuf[i+1] = byte(word >> 8)
			setupBuf[i+2] = byte(word >> 16)
			setupBuf[i+3] = byte(word >> 24)
		}

		setup := usb.NewSetup(setupBuf)

		ok := false
		if (setup.BmRequestType & 0x60) == 0 { // Standard request
			ok = handleStandardSetup(setup)
		} else {
			if setup.WIndex < uint16(len(usbSetupHandler)) && usbSetupHandler[setup.WIndex] != nil {
				ok = usbSetupHandler[setup.WIndex](setup)
			}
		}

		if !ok {
			// Stall EP0 — host will retry.
			usbOTG.INEP[0].CTL.SetBits(DEPCTL_STALL)
			usbOTG.OUTEP[0].CTL.SetBits(DEPCTL_STALL)
		}
		// Do NOT re-arm EP0 here. The FIFO will deliver a pktSts=4
		// (SETUP complete) entry next; we re-arm there.

	case 0x3: // OUT transfer complete (host sent ACK) — no payload.
		// Nothing to do; EP already re-armed in case 0x2.

	case 0x4: // SETUP transaction complete — re-arm EP0 for next SETUP/OUT.
		AckUsbOutTransfer(0)
	}
}

func handleUSBIRQ(intr interrupt.Interrupt) {
	status := usbOTG.GINTSTS.Get()

	// Suppress suspend interrupts — suspend fires before enumeration completes.
	if status&GINT_USBSUSP != 0 {
		usbOTG.GINTSTS.Set(GINT_USBSUSP)
	}

	if status&GINT_USBRST != 0 {
		usbOTG.GINTSTS.Set(GINT_USBRST)

		// Flush all FIFOs.
		usbOTG.GRSTCTL.SetBits(GRSTCTL_RXFFLSH)
		for usbOTG.GRSTCTL.Get()&GRSTCTL_RXFFLSH != 0 {
		}
		usbOTG.GRSTCTL.SetBits(GRSTCTL_TXFFLSH | GRSTCTL_TXFNUM_ALL)
		for usbOTG.GRSTCTL.Get()&GRSTCTL_TXFFLSH != 0 {
		}

		// Reset device address.
		usbOTG.DCFG.ClearBits(0x7F << 4)

		// Init EP0.
		initEndpoint(0, 0)
		usbConfiguration = 0

		// TRDT for HCLK ≥ 30 MHz → 6.
		usbOTG.GUSBCFG.ReplaceBits(0x6<<GUSBCFG_TRDT_Pos, 0xF<<GUSBCFG_TRDT_Pos, 0)

		// Arm EP0 OUT to receive first SETUP/OUT.
		AckUsbOutTransfer(0)
	}

	if status&GINT_ENUMDNE != 0 {
		usbOTG.GINTSTS.Set(GINT_ENUMDNE)
		// Enumeration done: activate EP0 at negotiated speed.
		usbOTG.INEP[0].CTL.SetBits(DEPCTL_CNAK)
	}

	if status&GINT_RXFLVL != 0 {
		// RXFLVL is level-triggered: mask it while processing, not W1C.
		usbOTG.GINTMSK.ClearBits(GINT_RXFLVL)
		for usbOTG.GINTSTS.Get()&GINT_RXFLVL != 0 {
			handleRxFIFO()
		}
		usbOTG.GINTMSK.SetBits(GINT_RXFLVL)
	}

	if status&GINT_IEPINT != 0 {
		daint := usbOTG.DAINT.Get() & 0xFFFF
		for ep := uint32(0); ep < NumberOfUSBEndpoints; ep++ {
			if daint&(1<<ep) != 0 {
				epInt := usbOTG.INEP[ep].INT.Get()
				usbOTG.INEP[ep].INT.Set(epInt) // W1C
				if epInt&DEPINT_XFERC != 0 {
					if ep != 0 && usbTxHandler[ep] != nil {
						usbTxHandler[ep]()
					}
				}
			}
		}
	}

	if status&GINT_OEPINT != 0 {
		daint := (usbOTG.DAINT.Get() >> 16) & 0xFFFF
		for ep := uint32(0); ep < NumberOfUSBEndpoints; ep++ {
			if daint&(1<<ep) != 0 {
				epInt := usbOTG.OUTEP[ep].INT.Get()
				usbOTG.OUTEP[ep].INT.Set(epInt) // W1C
			}
		}
	}
}
