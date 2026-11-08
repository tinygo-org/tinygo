//go:build ra4m1 && serial.usb

package machine

import (
	"device/arm"
	"device/renesas"
	"machine/usb"
	"runtime/interrupt"
	"runtime/volatile"
	"unsafe"
)

// USB full speed device driver for the RA4M1 USBFS module.
// Register and bit definitions are from R01UH0887 chapter 27.

const NumberOfUSBEndpoints = 8

// ELC event number of USBFS_INT, see bsp_elc.h in ArduinoCore-renesas.
const usbfsIntEvent = 51

// INTSTS0 and INTENB0 bits.
const (
	usbVBINT = 1 << 15
	usbRESM  = 1 << 14
	usbSOFR  = 1 << 13
	usbDVST  = 1 << 12
	usbCTRT  = 1 << 11
	usbBEMP  = 1 << 10
	usbNRDY  = 1 << 9
	usbBRDY  = 1 << 8
	usbVBSTS = 1 << 7
	usbVALID = 1 << 3

	usbDVSQMask = 0x70
	usbDVSQDef  = 1 << 4
	usbCTSQMask = 0x7
)

// SYSCFG bits.
const (
	usbSYSCFGUSBE  = 1 << 0
	usbSYSCFGDPRPU = 1 << 4
	usbSYSCFGDRPD  = 1 << 5
	usbSYSCFGDCFM  = 1 << 6
	usbSYSCFGSCKE  = 1 << 10
)

// FIFO select and control bits.
const (
	usbFIFOSELMBW16 = 1 << 10
	usbCFIFOSELISEL = 1 << 5
	usbFIFOCTRBVAL  = 1 << 15
	usbFIFOCTRBCLR  = 1 << 14
	usbFIFOCTRFRDY  = 1 << 13
	usbFIFOCTRDTLN  = 0x1FF
	usbCURPIPEMask  = 0xF
)

// DCPCTR and PIPEnCTR bits.
const (
	usbCTRPIDBUF   = 1
	usbCTRPIDSTALL = 2
	usbCTRPIDMask  = 3
	usbCTRCCPL     = 1 << 2
	usbCTRSQCLR    = 1 << 8
	usbCTRACLRM    = 1 << 9
)

// PIPECFG bits.
const (
	usbPIPECFGTypeBulk = 1 << 14
	usbPIPECFGTypeInt  = 2 << 14
	usbPIPECFGDBLB     = 1 << 9
	usbPIPECFGDirIn    = 1 << 4
)

const (
	usbPacketSize         = 64
	usbSpinLimit          = 100000
	usbFirstBulk          = 3
	usbFirstInt           = 6
	usbPipeCount          = 10
	usbBackupMagic uint32 = 0x07738135
)

var (
	usbPipeOf      [2][16]uint8
	usbPipeEP      [usbPipeCount]uint8
	usbPipeIn      [usbPipeCount]bool
	usbPipeUsed    [usbPipeCount]bool
	usbInPending   [usbPipeCount]bool
	usbRxBuf       [usbPacketSize]byte
	usbSetupBuf    [8]byte
	usbIELSR       = &renesas.ICU.IELSR6
	usbEP0         usbControlIn
	usbConfiguring bool
)

type usbControlIn struct {
	data    []byte
	off     int
	wLength uint16
	active  bool
}

func pipeCtr(num uint32) *volatile.Register16 {
	if num == 0 {
		return &renesas.USBFS.DCPCTR
	}
	return (*volatile.Register16)(unsafe.Pointer(uintptr(unsafe.Pointer(&renesas.USBFS.PIPE1CTR)) + uintptr(num-1)*2))
}

func fifoByte(r *volatile.Register16) *volatile.Register8 {
	return (*volatile.Register8)(unsafe.Pointer(r))
}

func fifoWrite(fifo *volatile.Register16, data []byte) {
	i := 0
	for ; i+1 < len(data); i += 2 {
		fifo.Set(uint16(data[i]) | uint16(data[i+1])<<8)
	}
	if i < len(data) {
		fifoByte(fifo).Set(data[i])
	}
}

func fifoRead(fifo *volatile.Register16, data []byte) {
	r := fifoByte(fifo)
	for i := range data {
		data[i] = r.Get()
	}
}

// Configure the USB peripheral. The config is here for compatibility with the UART interface.
func (dev *USBDevice) Configure(config UARTConfig) {
	u := renesas.USBFS

	renesas.SYSTEM.PRCR.Set(0xA5<<8 | renesas.SYSTEM_PRCR_PRC0_Msk)
	renesas.SYSTEM.USBCKCR.Set(1)
	renesas.SYSTEM.PRCR.Set(0xA5 << 8)

	renesas.MSTP.MSTPCRB.ClearBits(renesas.MSTP_MSTPCRB_MSTPB11_Msk)

	// Enable the internal VBUS detection, as the Arduino variant does.
	u.USBMC.SetBits(1 << 7)

	u.SYSCFG.SetBits(usbSYSCFGSCKE)
	for !u.SYSCFG.HasBits(usbSYSCFGSCKE) {
	}
	u.SYSCFG.ClearBits(usbSYSCFGDRPD | usbSYSCFGDCFM)
	u.SYSCFG.SetBits(usbSYSCFGUSBE)

	u.DCPMAXP.Set(usbPacketSize)

	// An event linked twice is only delivered to one slot, so unlink it first.
	slots := (*[32]volatile.Register32)(unsafe.Pointer(&renesas.ICU.IELSR0))
	for i := range slots {
		if slots[i].Get()&renesas.ICU_IELSR_IELS_Msk == usbfsIntEvent {
			slots[i].Set(0)
		}
	}
	usbIELSR.Set(usbfsIntEvent)
	intr := interrupt.New(renesas.IRQ_IEL6, handleUSBIRQ)
	// Below the UART, which has no FIFO. ARMv7-M ARM B1.5.4.
	intr.SetPriority(0x40)
	intr.Enable()

	u.INTSTS0.Set(0)
	u.BEMPENB.Set(1)
	u.BRDYENB.Set(0)
	u.INTENB0.Set(usbVBINT | usbBRDY | usbBEMP | usbDVST | usbCTRT | usbRESM)

	dev.Attach()
}

// Attach connects the device to the USB bus by enabling the DP pull-up.
func (dev *USBDevice) Attach() {
	renesas.USBFS.SYSCFG.SetBits(usbSYSCFGDPRPU)
}

// Detach disconnects the device from the USB bus by disabling the DP pull-up.
func (dev *USBDevice) Detach() {
	renesas.USBFS.SYSCFG.ClearBits(usbSYSCFGDPRPU)
}

const usbEvents = usbVBINT | usbRESM | usbDVST | usbCTRT | usbBEMP | usbBRDY

// The ICU only raises the interrupt on a new edge of the USB signal. Handle
// events until none are pending, then clear the flag and check again.
func handleUSBIRQ(interrupt.Interrupt) {
	u := renesas.USBFS
	for {
		is0 := u.INTSTS0.Get()
		if is0&usbEvents == 0 {
			usbIELSR.ClearBits(renesas.ICU_IELSR_IR_Msk)
			usbIELSR.Get() // make sure the IR clear completes before rechecking
			if u.INTSTS0.Get()&usbEvents == 0 {
				return
			}
			continue
		}
		usbHandleEvents(is0)
	}
}

func usbHandleEvents(is0 uint16) {
	u := renesas.USBFS

	// Write 0 to clear the sticky bits. VALID is cleared after the setup packet is read.
	u.INTSTS0.Set(^(is0 & (usbCTRT | usbDVST | usbSOFR | usbRESM | usbVBINT)) | usbVALID)

	if is0&usbVBINT != 0 {
		if u.INTSTS0.HasBits(usbVBSTS) {
			USBDev.Attach()
		} else {
			USBDev.Detach()
		}
	}

	if is0&usbDVST != 0 && is0&usbDVSQMask == usbDVSQDef {
		usbBusReset()
	}

	if is0&usbCTRT != 0 && is0&1 != 0 {
		usbProcessSetup()
	}

	if is0&usbBEMP != 0 {
		s := u.BEMPSTS.Get()
		u.BEMPSTS.Set(0)
		if s&1 != 0 && usbEP0.active {
			if usbEP0Write() {
				usbEP0.active = false
				u.DCPCTR.Set(usbCTRCCPL | usbCTRPIDBUF)
			}
		}
	}

	if is0&usbBRDY != 0 {
		s := u.BRDYSTS.Get() & u.BRDYENB.Get()
		u.BRDYSTS.Set(^s)
		for num := uint32(1); num < usbPipeCount; num++ {
			if s&(1<<num) == 0 {
				continue
			}
			ep := uint32(usbPipeEP[num])
			if usbPipeIn[num] {
				if usbInPending[num] {
					usbInPending[num] = false
					if usbTxHandler[ep] != nil {
						usbTxHandler[ep]()
					}
				}
			} else {
				usbReceive(num, ep)
			}
		}
	}
}

func usbBusReset() {
	u := renesas.USBFS

	u.BEMPENB.Set(1)
	u.BRDYENB.Set(0)
	u.CFIFOCTR.Set(usbFIFOCTRBCLR)

	u.D0FIFOSEL.Set(0)
	for u.D0FIFOSEL.Get()&usbCURPIPEMask != 0 {
	}
	u.D1FIFOSEL.Set(0)
	for u.D1FIFOSEL.Get()&usbCURPIPEMask != 0 {
	}

	for num := uint32(1); num < usbPipeCount; num++ {
		u.PIPESEL.Set(uint16(num))
		u.PIPECFG.Set(0)
		ctr := pipeCtr(num)
		ctr.Set(usbCTRACLRM)
		ctr.Set(0)
	}

	usbPipeOf = [2][16]uint8{}
	usbPipeUsed = [usbPipeCount]bool{}
	usbInPending = [usbPipeCount]bool{}
	usbEP0.active = false
	usbConfiguration = 0
	USBDev.InitEndpointComplete = false
}

func usbProcessSetup() {
	u := renesas.USBFS
	if !u.INTSTS0.HasBits(usbVALID) {
		return
	}

	u.CFIFOCTR.Set(usbFIFOCTRBCLR)
	words := [4]uint16{u.USBREQ.Get(), u.USBVAL.Get(), u.USBINDX.Get(), u.USBLENG.Get()}
	for i, w := range words {
		usbSetupBuf[2*i] = byte(w)
		usbSetupBuf[2*i+1] = byte(w >> 8)
	}
	u.INTSTS0.Set(^uint16(usbVALID))

	setup := usb.NewSetup(usbSetupBuf[:])
	usbEP0.wLength = setup.WLength

	ok := false
	if setup.BmRequestType&usb.REQUEST_TYPE == usb.REQUEST_STANDARD {
		ok = handleStandardSetup(setup)
	} else if setup.WIndex < uint16(len(usbSetupHandler)) && usbSetupHandler[setup.WIndex] != nil {
		ok = usbSetupHandler[setup.WIndex](setup)
	}
	if !ok {
		USBDev.SetStallEPIn(0)
	}
}

// The hardware acknowledges SET_ADDRESS by itself.
func handleUSBSetAddress(setup usb.Setup) bool {
	return true
}

func usbEP0Write() bool {
	u := renesas.USBFS
	rem := len(usbEP0.data) - usbEP0.off
	if rem == 0 {
		return true
	}
	n := min(rem, usbPacketSize)
	fifoWrite(&u.CFIFO, usbEP0.data[usbEP0.off:usbEP0.off+n])
	if n < usbPacketSize {
		u.CFIFOCTR.Set(usbFIFOCTRBVAL)
	}
	usbEP0.off += n
	return false
}

func usbSelectEP0(write bool) {
	u := renesas.USBFS
	if write {
		u.CFIFOSEL.Set(usbCFIFOSELISEL | usbFIFOSELMBW16)
		for !u.CFIFOSEL.HasBits(usbCFIFOSELISEL) {
		}
		for i := 0; i < usbSpinLimit && !u.CFIFOCTR.HasBits(usbFIFOCTRFRDY); i++ {
		}
	} else {
		u.CFIFOSEL.Set(0)
		for u.CFIFOSEL.HasBits(usbCFIFOSELISEL) {
		}
	}
}

func sendUSBPacket(ep uint32, data []byte) {
	u := renesas.USBFS
	if ep != 0 {
		SendUSBInPacket(ep, data)
		return
	}
	if len(data) == 0 {
		SendZlp()
		return
	}
	usbSelectEP0(true)
	usbEP0.data = data
	usbEP0.off = 0
	usbEP0.active = true
	usbEP0Write()
	u.DCPCTR.Set(usbCTRPIDBUF)
}

// SendZlp sends a zero length packet on the control endpoint.
func SendZlp() {
	u := renesas.USBFS
	if u.INTSTS0.Get()&usbCTSQMask == 1 {
		// Zero length data stage of a control read.
		usbSelectEP0(true)
		usbEP0.data = nil
		usbEP0.off = 0
		usbEP0.active = true
		u.CFIFOCTR.Set(usbFIFOCTRBVAL)
		u.DCPCTR.Set(usbCTRPIDBUF)
		return
	}
	u.DCPCTR.Set(usbCTRCCPL | usbCTRPIDBUF)
}

// SendUSBInPacket sends a packet for USB (interrupt in / bulk in).
func SendUSBInPacket(ep uint32, data []byte) bool {
	if ep == 0 {
		sendUSBPacket(0, data)
		return true
	}
	num := uint32(usbPipeOf[1][ep&0xF])
	if num == 0 || len(data) > usbPacketSize {
		return false
	}

	u := renesas.USBFS
	state := interrupt.Disable()

	usbInPending[num] = true
	u.D1FIFOSEL.Set(uint16(num) | usbFIFOSELMBW16)
	ready := false
	for i := 0; i < usbSpinLimit; i++ {
		if u.D1FIFOSEL.Get()&usbCURPIPEMask == uint16(num) && u.D1FIFOCTR.HasBits(usbFIFOCTRFRDY) {
			ready = true
			break
		}
	}
	if ready {
		fifoWrite(&u.D1FIFO, data)
		if len(data) < usbPacketSize {
			u.D1FIFOCTR.Set(usbFIFOCTRBVAL)
		}
	} else {
		usbInPending[num] = false
	}
	u.D1FIFOSEL.Set(0)
	for i := 0; i < usbSpinLimit && u.D1FIFOSEL.Get()&usbCURPIPEMask != 0; i++ {
	}

	interrupt.Restore(state)
	return ready
}

func usbReceive(num, ep uint32) {
	u := renesas.USBFS

	u.D0FIFOSEL.Set(uint16(num))
	for i := 0; i < usbSpinLimit; i++ {
		if u.D0FIFOSEL.Get()&usbCURPIPEMask == uint16(num) && u.D0FIFOCTR.HasBits(usbFIFOCTRFRDY) {
			break
		}
	}
	n := int(u.D0FIFOCTR.Get() & usbFIFOCTRDTLN)
	n = min(n, usbPacketSize)
	fifoRead(&u.D0FIFO, usbRxBuf[:n])
	if n < usbPacketSize {
		u.D0FIFOCTR.Set(usbFIFOCTRBCLR)
	}
	u.D0FIFOSEL.Set(0)
	for i := 0; i < usbSpinLimit && u.D0FIFOSEL.Get()&usbCURPIPEMask != 0; i++ {
	}

	if n > 0 && usbRxHandler[ep] != nil {
		usbRxHandler[ep](usbRxBuf[:n])
	}
}

// ReceiveUSBControlPacket reads the data stage of a control write.
func ReceiveUSBControlPacket() ([cdcLineInfoSize]byte, error) {
	var b [cdcLineInfoSize]byte
	u := renesas.USBFS

	usbSelectEP0(false)
	u.BRDYSTS.Set(^uint16(1))
	u.DCPCTR.Set(usbCTRPIDBUF)
	for i := 0; ; i++ {
		if u.BRDYSTS.HasBits(1) && u.CFIFOCTR.Get()&usbFIFOCTRDTLN != 0 {
			break
		}
		if i >= usbSpinLimit {
			return b, ErrUSBReadTimeout
		}
	}

	n := min(int(u.CFIFOCTR.Get()&usbFIFOCTRDTLN), len(b))
	fifoRead(&u.CFIFO, b[:n])
	if n < usbPacketSize {
		u.CFIFOCTR.Set(usbFIFOCTRBCLR)
	}
	u.BRDYSTS.Set(^uint16(1))
	return b, nil
}

func initEndpoint(ep, config uint32) {
	u := renesas.USBFS
	isIn := config&usb.EndpointIn != 0
	typ := config &^ usb.EndpointIn

	var first, last uint32
	var cfg uint16
	switch typ {
	case usb.ENDPOINT_TYPE_BULK:
		first, last = usbFirstBulk, usbFirstInt-1
		cfg = usbPIPECFGTypeBulk
		if !isIn {
			cfg |= usbPIPECFGDBLB
		}
	case usb.ENDPOINT_TYPE_INTERRUPT:
		first, last = usbFirstInt, usbPipeCount-1
		cfg = usbPIPECFGTypeInt
	default:
		return
	}

	num := uint32(0)
	for i := first; i <= last; i++ {
		if !usbPipeUsed[i] {
			num = i
			break
		}
	}
	if num == 0 {
		return
	}

	dir := 0
	if isIn {
		dir = 1
		cfg |= usbPIPECFGDirIn
	}
	usbPipeUsed[num] = true
	usbPipeOf[dir][ep&0xF] = uint8(num)
	usbPipeEP[num] = uint8(ep)
	usbPipeIn[num] = isIn
	usbInPending[num] = false

	u.PIPESEL.Set(uint16(num))
	u.PIPEMAXP.Set(usbPacketSize)
	ctr := pipeCtr(num)
	ctr.Set(usbCTRACLRM | usbCTRSQCLR)
	ctr.Set(0)
	u.PIPECFG.Set(cfg | uint16(ep&0xF))
	u.BRDYSTS.Set(^uint16(1 << num))
	u.BRDYENB.SetBits(1 << num)
	ctr.Set(usbCTRPIDBUF)
}

func usbStall(ep uint32, in bool) {
	var ctr *volatile.Register16
	if ep == 0 {
		ctr = pipeCtr(0)
	} else {
		dir := 0
		if in {
			dir = 1
		}
		num := uint32(usbPipeOf[dir][ep&0xF])
		if num == 0 {
			return
		}
		ctr = pipeCtr(num)
	}
	pid := ctr.Get() & usbCTRPIDMask
	ctr.Set(pid | usbCTRPIDSTALL)
	ctr.Set(usbCTRPIDSTALL)
}

func usbClearStall(ep uint32, in bool) {
	dir := 0
	if in {
		dir = 1
	}
	num := uint32(usbPipeOf[dir][ep&0xF])
	if num == 0 {
		return
	}
	ctr := pipeCtr(num)
	ctr.Set(usbCTRSQCLR)
	ctr.Set(usbCTRPIDBUF)
}

// SetStallEPIn sets ENDPOINT_HALT on a USB IN endpoint.
func (dev *USBDevice) SetStallEPIn(ep uint32) {
	usbStall(ep&0x7F, true)
}

// SetStallEPOut sets ENDPOINT_HALT on a USB OUT endpoint.
func (dev *USBDevice) SetStallEPOut(ep uint32) {
	usbStall(ep&0x7F, false)
}

// ClearStallEPIn clears ENDPOINT_HALT on a USB IN endpoint.
func (dev *USBDevice) ClearStallEPIn(ep uint32) {
	usbClearStall(ep&0x7F, true)
}

// ClearStallEPOut clears ENDPOINT_HALT on a USB OUT endpoint.
func (dev *USBDevice) ClearStallEPOut(ep uint32) {
	usbClearStall(ep&0x7F, false)
}

// EnterBootloader resets into the Arduino bootloader.
// Source: ArduinoCore-renesas 1.6.0 cores/arduino/boot.cpp.
func EnterBootloader() {
	arm.DisableInterrupts()

	// Stay detached long enough for the host to notice the disconnect.
	renesas.USBFS.SYSCFG.ClearBits(usbSYSCFGDPRPU)
	for i := 0; i < 8000000; i++ {
		arm.Asm("nop")
	}

	renesas.SYSTEM.PRCR.Set(0xA5<<8 | renesas.SYSTEM_PRCR_PRC1_Msk)
	magic := (*volatile.Register32)(unsafe.Pointer(&renesas.SYSTEM.VBTBKR[0]))
	magic.Set(usbBackupMagic)
	for i := 0; i < usbSpinLimit && magic.Get() != usbBackupMagic; i++ {
	}
	renesas.SYSTEM.PRCR.Set(0xA5 << 8)
	arm.SystemReset()
}
