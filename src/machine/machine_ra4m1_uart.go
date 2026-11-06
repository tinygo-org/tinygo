//go:build ra4m1

package machine

import (
	"device/renesas"
	"runtime/interrupt"
	"runtime/volatile"
	"unsafe"
)

// SCI modules run from PCLKA, set to 48 MHz in runtime_ra4m1.go.
const sciClockFrequency = 48_000_000

// PSEL values for the SCI pin function (R01UH0887 table 19.7).
const (
	pselSCIEven uint32 = 0x04 // SCI0, SCI2, SCI4, SCI6, SCI8
	pselSCIOdd  uint32 = 0x05 // SCI1, SCI3, SCI5, SCI7, SCI9
)

// ELC event offsets relative to the RXI event of a SCI channel.
const (
	eventRXI = 0
	eventTXI = 1
	eventTEI = 2
	eventERI = 3
)

// UART on the RA4M1 is a SCI module in asynchronous mode.
type UART struct {
	Buffer   *RingBuffer
	Bus      *renesas.SCI2_Type
	rxIRQ    interrupt.Interrupt
	errIRQ   interrupt.Interrupt
	rxIELSR  *volatile.Register32
	errIELSR *volatile.Register32
	event    uint32 // ELC event number of RXI
	mstpMask uint32
	psel     uint32
	txPin    Pin
	rxPin    Pin
}

var (
	// SCI1 has extra FIFO registers but shares the SCI2 layout for the
	// registers used here. The FIFO stays disabled.
	sci1UART = UART{
		Buffer:   NewRingBuffer(),
		Bus:      (*renesas.SCI2_Type)(unsafe.Pointer(renesas.SCI1)),
		rxIELSR:  &renesas.ICU.IELSR4,
		errIELSR: &renesas.ICU.IELSR5,
		event:    0x9E, // ELC_EVENT_SCI1_RXI
		mstpMask: renesas.MSTP_MSTPCRB_MSTPB30_Msk,
		psel:     pselSCIOdd,
	}
	sci2UART = UART{
		Buffer:   NewRingBuffer(),
		Bus:      renesas.SCI2,
		rxIELSR:  &renesas.ICU.IELSR2,
		errIELSR: &renesas.ICU.IELSR3,
		event:    0xA3, // ELC_EVENT_SCI2_RXI
		mstpMask: renesas.MSTP_MSTPCRB_MSTPB29_Msk,
		psel:     pselSCIEven,
	}
	sci9UART = UART{
		Buffer:   NewRingBuffer(),
		Bus:      renesas.SCI9,
		rxIELSR:  &renesas.ICU.IELSR0,
		errIELSR: &renesas.ICU.IELSR1,
		event:    0xA8, // ELC_EVENT_SCI9_RXI
		mstpMask: renesas.MSTP_MSTPCRB_MSTPB22_Msk,
		psel:     pselSCIOdd,
	}
)

func init() {
	sci1UART.rxIRQ = interrupt.New(renesas.IRQ_IEL4, sci1UART.handleRXI)
	sci1UART.errIRQ = interrupt.New(renesas.IRQ_IEL5, sci1UART.handleERI)
	sci2UART.rxIRQ = interrupt.New(renesas.IRQ_IEL2, sci2UART.handleRXI)
	sci2UART.errIRQ = interrupt.New(renesas.IRQ_IEL3, sci2UART.handleERI)
	sci9UART.rxIRQ = interrupt.New(renesas.IRQ_IEL0, sci9UART.handleRXI)
	sci9UART.errIRQ = interrupt.New(renesas.IRQ_IEL1, sci9UART.handleERI)
}

// Configure the UART.
func (uart *UART) Configure(config UARTConfig) {
	// Default baud rate to 115200.
	if config.BaudRate == 0 {
		config.BaudRate = 115200
	}

	// Use the default pins of this SCI instance if not set.
	if config.TX == 0 && config.RX == 0 {
		config.TX = uart.txPin
		config.RX = uart.rxPin
	}

	// Release module stop.
	renesas.MSTP.MSTPCRB.ClearBits(uart.mstpMask)

	// Init sequence from R01UH0887 figure 28.9: 8N1, async, 16x sampling.
	uart.Bus.SCR.Set(0)
	uart.Bus.SCMR.Set(0xF2)
	uart.Bus.SMR.Set(0)
	uart.Bus.SEMR.Set(renesas.SCI2_SEMR_ABCS_Msk | renesas.SCI2_SEMR_BGDM_Msk)
	uart.SetBaudRate(config.BaudRate)

	uart.configurePin(config.TX)
	uart.configurePin(config.RX)

	// The bootloader leaves SCI events linked in the ICU. An event linked
	// twice is only delivered to one slot, so unlink them first.
	slots := (*[32]volatile.Register32)(unsafe.Pointer(&renesas.ICU.IELSR0))
	for i := range slots {
		ev := slots[i].Get() & renesas.ICU_IELSR_IELS_Msk
		if ev >= uart.event+eventRXI && ev <= uart.event+eventERI {
			slots[i].Set(0)
		}
	}
	uart.rxIELSR.Set(uart.event + eventRXI)
	uart.errIELSR.Set(uart.event + eventERI)
	uart.rxIRQ.Enable()
	uart.errIRQ.Enable()

	uart.Bus.SCR.Set(renesas.SCI2_SCR_TE_Msk | renesas.SCI2_SCR_RE_Msk | renesas.SCI2_SCR_RIE_Msk)
}

// SetBaudRate sets the communication speed for the UART.
func (uart *UART) SetBaudRate(br uint32) {
	// TE and RE must be 0 while SMR, BRR and MDDR are written.
	scr := uart.Bus.SCR.Get()
	uart.Bus.SCR.Set(0)

	// With ABCS=1 and BGDM=1 the divisor is 8*4^CKS (R01UH0887 table 28.10).
	// N is rounded down so the raw rate is at or above the target. The bit
	// rate modulator then scales it by MDDR/256 (R01UH0887 section 28.2.14).
	cks := uint32(0)
	divisor := uint32(8)
	n := sciClockFrequency / (divisor * br)
	for n > 256 && cks < 3 {
		cks++
		divisor *= 4
		n = sciClockFrequency / (divisor * br)
	}
	if n > 256 {
		n = 256
	}
	if n < 1 {
		n = 1
	}
	scaled := uint64(br) * 256 * uint64(divisor*n)
	mddr := (scaled*2 + sciClockFrequency) / (2 * sciClockFrequency)
	if mddr >= 256 {
		uart.Bus.SEMR.ClearBits(renesas.SCI2_SEMR_BRME_Msk)
	} else {
		if mddr < 128 {
			mddr = 128
		}
		uart.Bus.MDDR.Set(uint8(mddr))
		uart.Bus.SEMR.SetBits(renesas.SCI2_SEMR_BRME_Msk)
	}
	uart.Bus.SMR.ReplaceBits(uint8(cks), renesas.SCI2_SMR_CKS_Msk, 0)
	uart.Bus.BRR.Set(uint8(n - 1))

	uart.Bus.SCR.Set(scr)
}

func (uart *UART) configurePin(p Pin) {
	if p == NoPin {
		return
	}
	enableWritingPmnPFS()
	port := getPortNumber(p)
	bit := getPinNumber(p)
	reg := (*volatile.Register32)(unsafe.Add(unsafe.Pointer(renesas.PFS), 0x40*port+4*int(bit)))
	mode := uart.psel<<renesas.PFS_P000PFS_PSEL_Pos | renesas.PFS_P000PFS_PMR_Msk
	mask := uint32(renesas.PFS_P000PFS_PSEL_Msk |
		renesas.PFS_P000PFS_PMR_Msk |
		renesas.PFS_P000PFS_ASEL_Msk |
		renesas.PFS_P000PFS_ISEL_Msk |
		renesas.PFS_P000PFS_PCR_Msk |
		renesas.PFS_P000PFS_PDR_Msk)
	reg.ReplaceBits(mode, mask, 0)
	disableWritingPmnPFS()
}

func (uart *UART) handleRXI(interrupt.Interrupt) {
	uart.Receive(uart.Bus.RDR.Get())
	uart.rxIELSR.ClearBits(renesas.ICU_IELSR_IR_Msk)
	uart.rxIELSR.Get() // make sure the IR clear completes before exception return
}

func (uart *UART) handleERI(interrupt.Interrupt) {
	ssr := uart.Bus.SSR.Get()
	if ssr&renesas.SCI2_SSR_RDRF_Msk != 0 {
		uart.Receive(uart.Bus.RDR.Get())
	}
	// Error flags are cleared by writing 0.
	uart.Bus.SSR.Set(ssr &^ (renesas.SCI2_SSR_ORER_Msk | renesas.SCI2_SSR_FER_Msk | renesas.SCI2_SSR_PER_Msk))
	uart.errIELSR.ClearBits(renesas.ICU_IELSR_IR_Msk)
	uart.errIELSR.Get()
}

func (uart *UART) writeByte(c byte) error {
	for !uart.Bus.SSR.HasBits(renesas.SCI2_SSR_TDRE_Msk) {
	}
	uart.Bus.TDR.Set(c)
	return nil
}

func (uart *UART) flush() {
	for !uart.Bus.SSR.HasBits(renesas.SCI2_SSR_TEND_Msk) {
	}
}

// DTR is not supported by the SCI peripheral.
func (uart *UART) DTR() bool {
	return true
}

// RTS is not supported by the SCI peripheral.
func (uart *UART) RTS() bool {
	return false
}
