//go:build mimxrt1062

package machine

// USB device driver for the i.MX RT1062 (Teensy 4.x). See the device data
// structures section of the USB chapter in IMXRT1060RM.

import (
	"device/arm"
	"device/nxp"
	"machine/usb"
	"runtime/interrupt"
	"runtime/volatile"
	"unsafe"
)

const NumberOfUSBEndpoints = 8

//go:extern _usb_dma_start
var _usb_dma_start [0]byte

// Layout of the non cacheable USB DMA region (4 KiB, see linker script).
// The USB DMA cannot access DTCM, so the region is at the top of OCRAM.
var (
	usbRAMBase    = uintptr(unsafe.Pointer(&_usb_dma_start))
	usbDQHBase    = usbRAMBase + 0x000 // 16 * 64 B, 2 KiB aligned
	usbDTDBase    = usbRAMBase + 0x400 // 16 * 32 B
	usbOutBufBase = usbRAMBase + 0x600 // 8 * 64 B
	usbInBufBase  = usbRAMBase + 0x800 // 8 * 64 B
	usbEP0InBase  = usbRAMBase + 0xA00 // 256 B for EP0 IN control data
)

const usbEP0InLen = 256

// Endpoint queue head (dQH), 64 bytes.
type usbDQH struct {
	config  volatile.Register32
	current volatile.Register32
	next    volatile.Register32
	token   volatile.Register32
	pages   [5]volatile.Register32
	_       uint32
	setup   [2]volatile.Register32
	_       [4]uint32
}

// Endpoint transfer descriptor (dTD), 32 bytes.
type usbDTD struct {
	next  volatile.Register32
	token volatile.Register32
	pages [5]volatile.Register32
	_     uint32
}

const (
	dtdTerminate   = 0x1
	dtdTokenActive = 0x80
	dtdTokenIOC    = 0x8000
	dqhIOS         = 0x8000     // interrupt on setup
	dqhZLTDisable  = 0x20000000 // disable automatic zero-length packet
)

func dqh(ep uint32, in bool) *usbDQH {
	i := 2 * ep
	if in {
		i++
	}
	return (*usbDQH)(unsafe.Pointer(usbDQHBase + uintptr(i)*64))
}

func dtd(ep uint32, in bool) *usbDTD {
	i := 2 * ep
	if in {
		i++
	}
	return (*usbDTD)(unsafe.Pointer(usbDTDBase + uintptr(i)*32))
}

func epOutBuf(ep uint32) []byte {
	return unsafe.Slice((*byte)(unsafe.Pointer(usbOutBufBase+uintptr(ep)*64)), 64)
}

func epInBuf(ep uint32) []byte {
	return unsafe.Slice((*byte)(unsafe.Pointer(usbInBufBase+uintptr(ep)*64)), 64)
}

func ep0InXferBuf() []byte {
	return unsafe.Slice((*byte)(unsafe.Pointer(usbEP0InBase)), usbEP0InLen)
}

// endptCtrl returns ENDPTCTRL[ep] (the registers are consecutive).
func endptCtrl(ep uint32) *volatile.Register32 {
	return (*volatile.Register32)(unsafe.Add(unsafe.Pointer(&nxp.USB1.ENDPTCTRL0), 4*uintptr(ep)))
}

// ENDPTCTRL bits (same layout for every endpoint).
const (
	epctrlRXS    = 0x1     // RX stall
	epctrlRXR    = 0x40    // RX data toggle reset
	epctrlRXE    = 0x80    // RX enable
	epctrlTXS    = 0x10000 // TX stall
	epctrlTXR    = 0x400000
	epctrlTXE    = 0x800000
	epctrlRXTPos = 2
	epctrlTXTPos = 18
)

// Bound for hardware wait loops, far longer than any real operation takes.
const usbSpinLimit = 5_000_000

// Configure the USB peripheral. The config is here for compatibility with the UART interface.
func (dev *USBDevice) Configure(config UARTConfig) {
	if dev.initcomplete {
		return
	}

	// Set a default serial number from the unique chip ID. The device
	// descriptor declares iSerialNumber, so the string must exist.
	if usb.Serial == "" {
		usb.Serial = mimxrtSerialNumber()
	}

	// Ungate the USB clock. PLL3 (480 MHz) is already set up by clock init.
	nxp.CCM.SetCCGR6_CG0(3)

	// Reset and power up the PHY.
	nxp.USBPHY1.CTRL_SET.Set(nxp.USBPHY_CTRL_SFTRST)
	nxp.USBPHY1.CTRL_CLR.Set(nxp.USBPHY_CTRL_SFTRST | nxp.USBPHY_CTRL_CLKGATE)
	nxp.USBPHY1.PWD.Set(0)

	// Reset the controller (the bootloader used it).
	nxp.USB1.USBCMD.SetBits(nxp.USB_USBCMD_RST)
	for nxp.USB1.USBCMD.HasBits(nxp.USB_USBCMD_RST) {
	}

	// Device mode, setup lockout disabled (we use the setup tripwire instead).
	nxp.USB1.USBMODE.Set(2<<nxp.USB_USBMODE_CM_Pos | nxp.USB_USBMODE_SLOM)

	usbRAMClear()
	dqh(0, false).config.Set(64<<16 | dqhIOS)
	dqh(0, true).config.Set(64 << 16)
	// In device mode, ASYNCLISTADDR is ENDPTLISTADDR (same offset 0x158).
	nxp.USB1.ASYNCLISTADDR.Set(uint32(usbDQHBase))

	// The shared machine/usb stack assumes 64-byte (full-speed) endpoints.
	nxp.USB1.PORTSC1.SetBits(nxp.USB_PORTSC1_PFSC)
	nxp.USB1.BURSTSIZE.Set(0x0404)

	// Enable the transfer, error, reset, and suspend interrupts.
	nxp.USB1.USBINTR.Set(nxp.USB_USBSTS_UI | nxp.USB_USBSTS_UEI |
		nxp.USB_USBSTS_URI | nxp.USB_USBSTS_SLI)

	intr := interrupt.New(nxp.IRQ_USB_OTG1, handleUSBIRQ)
	intr.Enable()

	dev.initcomplete = true
	dev.Attach()
}

// mimxrtSerialNumber returns a 16-hex-digit serial from the chip unique ID.
func mimxrtSerialNumber() string {
	id := uint64(nxp.OCOTP.CFG1.Get())<<32 | uint64(nxp.OCOTP.CFG0.Get())
	const hex = "0123456789ABCDEF"
	var b [16]byte
	for i := 0; i < 16; i++ {
		b[15-i] = hex[id&0xF]
		id >>= 4
	}
	return string(b[:])
}

func usbRAMClear() {
	p := unsafe.Slice((*volatile.Register32)(unsafe.Pointer(usbRAMBase)), 0x1000/4)
	for i := range p {
		p[i].Set(0)
	}
}

// Attach connects the device to the USB bus (run the controller).
func (dev *USBDevice) Attach() {
	nxp.USB1.USBCMD.SetBits(nxp.USB_USBCMD_RS)
}

// Detach disconnects the device from the USB bus.
func (dev *USBDevice) Detach() {
	nxp.USB1.USBCMD.ClearBits(nxp.USB_USBCMD_RS)
}

func handleUSBIRQ(intr interrupt.Interrupt) {
	status := nxp.USB1.USBSTS.Get()
	nxp.USB1.USBSTS.Set(status)

	if status&nxp.USB_USBSTS_URI != 0 {
		handleUSBBusReset()
	}

	if status&nxp.USB_USBSTS_UI != 0 {
		// SETUP packets land in the dQH's setup area, flagged in ENDPTSETUPSTAT.
		for {
			ss := nxp.USB1.ENDPTSETUPSTAT.Get()
			if ss == 0 {
				break
			}
			nxp.USB1.ENDPTSETUPSTAT.Set(ss)
			if ss&1 != 0 {
				handleEP0Setup()
			}
		}

		comp := nxp.USB1.ENDPTCOMPLETE.Get()
		if comp != 0 {
			nxp.USB1.ENDPTCOMPLETE.Set(comp)

			// OUT completions, bits 0 to 7
			for ep := uint32(1); ep < NumberOfUSBEndpoints; ep++ {
				if comp&(1<<ep) != 0 {
					n := 64 - int((dtd(ep, false).token.Get()>>16)&0x7FFF)
					buf := epOutBuf(ep)[:n]
					if usbRxHandler[ep] == nil || usbRxHandler[ep](buf) {
						AckUsbOutTransfer(ep)
					}
				}
			}

			// IN completions, bits 16 to 23
			for ep := uint32(1); ep < NumberOfUSBEndpoints; ep++ {
				if comp&(1<<(16+ep)) != 0 {
					if usbTxHandler[ep] != nil {
						usbTxHandler[ep]()
					}
				}
			}
		}
	}
}

// IN endpoints whose in-flight transfer was cancelled by a bus reset flush.
// Only touched from the USB interrupt handler.
var usbTxCancelled uint32

// usbSetupIn is true while the current EP0 setup is a control read
// (device to host). Its status stage is an OUT packet.
var usbSetupIn bool

func handleUSBBusReset() {
	// See the Bus Reset section of the USB chapter in IMXRT1060RM.
	nxp.USB1.ENDPTSETUPSTAT.Set(nxp.USB1.ENDPTSETUPSTAT.Get())
	nxp.USB1.ENDPTCOMPLETE.Set(nxp.USB1.ENDPTCOMPLETE.Get())
	for i := 0; nxp.USB1.ENDPTPRIME.Get() != 0; i++ {
		if i > usbSpinLimit {
			break
		}
	}
	// Record IN transfers the flush below cancels. Their completion callbacks
	// run in initEndpoint once the host reconfigures the device.
	for ep := uint32(1); ep < NumberOfUSBEndpoints; ep++ {
		if dtd(ep, true).token.Get()&dtdTokenActive != 0 {
			usbTxCancelled |= 1 << ep
		}
	}
	nxp.USB1.ENDPTFLUSH.Set(0xFFFFFFFF)
	nxp.USB1.DEVICEADDR.Set(0)
	usbConfiguration = 0
}

// handleEP0Setup reads the setup packet (guarded by the setup tripwire) and
// dispatches it to the shared stack.
func handleEP0Setup() {
	var raw [8]byte
	for i := 0; ; i++ {
		if i > usbSpinLimit {
			return
		}
		nxp.USB1.USBCMD.SetBits(nxp.USB_USBCMD_SUTW)
		s0 := dqh(0, false).setup[0].Get()
		s1 := dqh(0, false).setup[1].Get()
		if nxp.USB1.USBCMD.HasBits(nxp.USB_USBCMD_SUTW) {
			raw[0], raw[1], raw[2], raw[3] = byte(s0), byte(s0>>8), byte(s0>>16), byte(s0>>24)
			raw[4], raw[5], raw[6], raw[7] = byte(s1), byte(s1>>8), byte(s1>>16), byte(s1>>24)
			break
		}
	}
	nxp.USB1.USBCMD.ClearBits(nxp.USB_USBCMD_SUTW)

	// A new setup cancels any transfer still pending on EP0.
	nxp.USB1.ENDPTFLUSH.Set(1<<16 | 1)
	for i := 0; nxp.USB1.ENDPTFLUSH.HasBits(1<<16 | 1); i++ {
		if i > usbSpinLimit {
			break
		}
	}

	setup := usb.NewSetup(raw[:])
	usbSetupIn = setup.BmRequestType&0x80 != 0

	// A control write has an OUT data stage. Prime EP0 OUT before the
	// dispatch so the handler can read it with ReceiveUSBControlPacket.
	if !usbSetupIn && setup.WLength > 0 {
		usbPrime(0, false, usbOutBufBase, 64)
	}

	ok := false
	if (setup.BmRequestType & usb.REQUEST_TYPE) == usb.REQUEST_STANDARD {
		ok = handleStandardSetup(setup)
	} else {
		if setup.WIndex < uint16(len(usbSetupHandler)) && usbSetupHandler[setup.WIndex] != nil {
			ok = usbSetupHandler[setup.WIndex](setup)
		}
	}
	if !ok {
		USBDev.SetStallEPIn(0)
	}
}

// usbPrime arms one dTD for the given buffer and primes the endpoint.
// The hardware splits the transfer into max packet size packets.
func usbPrime(ep uint32, in bool, addr uintptr, size int) {
	d := dtd(ep, in)
	d.next.Set(dtdTerminate)
	d.token.Set(uint32(size)<<16 | dtdTokenIOC | dtdTokenActive)
	p := uint32(addr)
	d.pages[0].Set(p)
	d.pages[1].Set(p&^0xFFF + 0x1000)
	d.pages[2].Set(p&^0xFFF + 0x2000)
	d.pages[3].Set(p&^0xFFF + 0x3000)
	d.pages[4].Set(p&^0xFFF + 0x4000)

	q := dqh(ep, in)
	q.next.Set(uint32(uintptr(unsafe.Pointer(d))))
	q.token.Set(0)

	// Descriptor writes must reach memory before the prime register write.
	arm.Asm("dsb 0xF")

	mask := uint32(1) << ep
	if in {
		mask = 1 << (16 + ep)
	}
	// Clear a stale completion bit from an earlier transfer so it is not
	// read as the completion of this transfer.
	nxp.USB1.ENDPTCOMPLETE.Set(mask)
	// The hardware clears ENDPTPRIME bits on its own. A read modify write
	// can set a cleared bit of another endpoint again. Zero bits are ignored.
	nxp.USB1.ENDPTPRIME.Set(mask)
	for i := 0; nxp.USB1.ENDPTPRIME.HasBits(mask); i++ {
		if i > usbSpinLimit {
			return
		}
	}
}

func initEndpoint(ep, config uint32) {
	// The ENDPTCTRL type value is 2 for bulk and 3 for interrupt.
	var t uint32
	switch config &^ (usb.EndpointIn | usb.EndpointOut) {
	case usb.ENDPOINT_TYPE_BULK:
		t = 2
	case usb.ENDPOINT_TYPE_INTERRUPT:
		t = 3
	case usb.ENDPOINT_TYPE_CONTROL:
		return // EP0 is configured in Configure
	default:
		t = 2
	}

	in := config&usb.EndpointIn != 0

	// Stop the endpoint before the dQH and dTD writes below. Record a
	// cancelled IN transfer so its completion callback runs.
	if in && dtd(ep, true).token.Get()&dtdTokenActive != 0 {
		usbTxCancelled |= 1 << ep
	}
	mask := uint32(1) << ep
	if in {
		mask = 1 << (16 + ep)
	}
	nxp.USB1.ENDPTFLUSH.Set(mask)
	for i := 0; nxp.USB1.ENDPTFLUSH.HasBits(mask); i++ {
		if i > usbSpinLimit {
			break
		}
	}
	dtd(ep, in).token.Set(0)

	if in {
		dqh(ep, true).config.Set(64<<16 | dqhZLTDisable)
		endptCtrl(ep).SetBits(t<<epctrlTXTPos | epctrlTXR | epctrlTXE)
		// Report a transfer cancelled by a bus reset as complete, now that the
		// endpoint works again. Otherwise the class driver waits forever.
		if usbTxCancelled&(1<<ep) != 0 {
			usbTxCancelled &^= 1 << ep
			if usbTxHandler[ep] != nil {
				usbTxHandler[ep]()
			}
		}
	} else {
		dqh(ep, false).config.Set(64<<16 | dqhZLTDisable)
		endptCtrl(ep).SetBits(t<<epctrlRXTPos | epctrlRXR | epctrlRXE)
		usbPrime(ep, false, usbOutBufBase+uintptr(ep)*64, 64)
	}
}

// SendUSBInPacket sends a packet for USB (interrupt in / bulk in). It reports
// false when the data does not fit or the previous transfer is still active.
func SendUSBInPacket(ep uint32, data []byte) bool {
	ep &= 0x7F
	if ep != 0 {
		if len(data) > 64 {
			return false
		}
		// The USB interrupt handler primes the same endpoint. Keep the
		// active check and the prime together.
		mask := interrupt.Disable()
		if dtd(ep, true).token.Get()&dtdTokenActive != 0 {
			interrupt.Restore(mask)
			return false
		}
		sendUSBPacket(ep, data)
		interrupt.Restore(mask)
		return true
	}
	sendUSBPacket(ep, data)
	return true
}

//go:noinline
func sendUSBPacket(ep uint32, data []byte) {
	ep &= 0x7F
	if ep == 0 {
		n := len(data)
		if n > usbEP0InLen {
			n = usbEP0InLen
		}
		copy(ep0InXferBuf(), data[:n])
		usbPrime(0, true, usbEP0InBase, n)
		if usbSetupIn {
			// A control read ends with an OUT status stage, also when the
			// data stage is empty. A control write ends with this IN packet.
			usbPrime(0, false, usbOutBufBase, 64)
		}
	} else {
		n := len(data)
		if n > 64 {
			n = 64
		}
		copy(epInBuf(ep), data[:n])
		usbPrime(ep, true, usbInBufBase+uintptr(ep)*64, n)
	}
}

// ReceiveUSBControlPacket waits for and returns the EP0 OUT data stage that
// was primed when the setup packet was dispatched.
func ReceiveUSBControlPacket() (b [cdcLineInfoSize]byte, err error) {
	for i := 0; nxp.USB1.ENDPTCOMPLETE.Get()&1 == 0; i++ {
		if i > usbSpinLimit {
			return b, ErrUSBReadTimeout
		}
	}
	nxp.USB1.ENDPTCOMPLETE.Set(1)
	arm.Asm("dsb 0xF")

	n := 64 - int((dtd(0, false).token.Get()>>16)&0x7FFF)
	if n > len(b) {
		n = len(b)
	}
	out := epOutBuf(0)
	for i := 0; i < n; i++ {
		b[i] = out[i]
	}
	return b, nil
}

// AckUsbOutTransfer re-arms an OUT endpoint after its data was consumed.
// Thread context callers must not interleave with the USB interrupt handler.
func AckUsbOutTransfer(ep uint32) {
	ep &= 0x7F
	mask := interrupt.Disable()
	usbPrime(ep, false, usbOutBufBase+uintptr(ep)*64, 64)
	interrupt.Restore(mask)
}

func SendZlp() {
	sendUSBPacket(0, nil)
}

func handleUSBSetAddress(setup usb.Setup) bool {
	// USBADRA defers the address change until after the status stage.
	nxp.USB1.DEVICEADDR.Set(uint32(setup.WValueL)<<25 | nxp.USB_DEVICEADDR_USBADRA)
	SendZlp()
	return true
}

// Set ENDPOINT_HALT/stall status on a USB IN endpoint.
func (dev *USBDevice) SetStallEPIn(ep uint32) {
	endptCtrl(ep & 0x7F).SetBits(epctrlTXS)
}

// Set ENDPOINT_HALT/stall status on a USB OUT endpoint.
func (dev *USBDevice) SetStallEPOut(ep uint32) {
	endptCtrl(ep & 0x7F).SetBits(epctrlRXS)
}

// Clear the ENDPOINT_HALT/stall on a USB IN endpoint.
func (dev *USBDevice) ClearStallEPIn(ep uint32) {
	ep &= 0x7F
	endptCtrl(ep).ClearBits(epctrlTXS)
	endptCtrl(ep).SetBits(epctrlTXR) // reset data toggle to DATA0
}

// Clear the ENDPOINT_HALT/stall on a USB OUT endpoint.
func (dev *USBDevice) ClearStallEPOut(ep uint32) {
	ep &= 0x7F
	endptCtrl(ep).ClearBits(epctrlRXS)
	endptCtrl(ep).SetBits(epctrlRXR)
}

// EnterBootloader resets into the HalfKay bootloader. The bootloader chip
// watches for this breakpoint, the same as Teensyduino soft reboot.
func EnterBootloader() {
	arm.DisableInterrupts()
	arm.Asm("bkpt #251")
	for {
	}
}
