// +build rpi3

package machine

import (
	"device/arm"
	"runtime/volatile"
	"unsafe"
)

//specific to raspberry pi 3
const MemoryMappedIO = uintptr(0x3F000000)

var Aux *AuxPeripheralsRegisterMap = (*AuxPeripheralsRegisterMap)(unsafe.Pointer(MemoryMappedIO + 0x00215000))
var GPIO *GPIORegisterMap = (*GPIORegisterMap)(unsafe.Pointer(MemoryMappedIO + 0x00200000))
var SysTimer *SysTimerRegisterMap = (*SysTimerRegisterMap)(unsafe.Pointer(MemoryMappedIO + 0x3000))
var InterruptController *IRQRegisterMap = (*IRQRegisterMap)(unsafe.Pointer(MemoryMappedIO + 0xB200))

//decls
var MiniUART UART

// for the interrupt numbers for use with interrupt controller
const AuxInterrupt = 1 << 29

type AuxPeripheralsRegisterMap struct {
	InterruptStatus           volatile.Register32 //0x00
	Enables                   volatile.Register32 //0x04
	reserved00                [14]uint32
	MiniUARTData              volatile.Register32 //0x40, 8 bits wide
	MiniUARTInterruptEnable   volatile.Register32 //0x44
	MiniUARTInterruptIdentify volatile.Register32 //0x48
	MiniUARTLineControl       volatile.Register32 //0x4C
	MiniUARTModemControl      volatile.Register32 //0x50
	MiniUARTLineStatus        volatile.Register32 //0x54, readonly
	MiniUARTModemStatus       volatile.Register32 //0x58, readonly
	MiniUARTScratch           volatile.Register32 //0x5C
	MiniUARTExtraControl      volatile.Register32 //0x60
	MiniUARTExtraStatus       volatile.Register32 //0x64
	MiniUARTBAUD              volatile.Register32 //0x68
	reserved01                [5]uint32
	SPI1ControlRegister0      volatile.Register32 //0x80
	SPI1ControlRegister1      volatile.Register32 //0x84
	SPI1Status                volatile.Register32 //0x88
	reserved02                volatile.Register32 //0x8C
	SPI1Data                  volatile.Register32 //0x90
	SPI1Peek                  volatile.Register32 //0x94
	reserved03                [10]uint32
	SPI2ControlRegister0      volatile.Register32 //0xC0
	SPI2ControlRegister1      volatile.Register32 //0xC4
	SPI2Status                volatile.Register32 //0xC8
	reserved04                volatile.Register32
	SPI2Data                  volatile.Register32 //0xD0
	SPI2Peek                  volatile.Register32 //0xD4
}

// mini uart: peripheral enable
const PeripheralMiniUART = 1 << 0

// mini uart: extra control bitfields
const ReceiveEnable = 1 << 0
const TransmitEnable = 1 << 1
const EnableRTS = 1 << 2
const EnableCTS = 1 << 3
const RTSFlowLevelMask = 0x1F //use with register32.ReplaceBits
const RTSFlowLevelFIFO3Spaces = 0 << 4
const RTSFlowLevelFIFO2Spaces = 1 << 4
const RTSFlowLevelFIFO1Space = 2 << 4
const RTSFlowLevelFIFO4Spaces = 3 << 4
const RTSAssertLevel = 1 << 6
const CTSAssertLevel = 1 << 7

// mini uart: line control register bitfields
//https://elinux.org/BCM2835_datasheet_errata
const DataLength8Bits = 3 << 0
const Break = 1 << 6
const DLab = 1 << 7

// mini uart: line control register bitfields
const ReadyToSend = 1 << 1

// mini uart: interrupt identify register bitfields
const Pending = 1 << 0
const TransmitInterruptsPending = 1 << 1 //Read
const ReceiveInterruptsPending = 2 << 1 //Read

const ClearFIFOsMask = 0x6//use with register32.ReplaceBits
const ClearReceiveFIFO = 1 << 1 //Write
const ClearTransmitFIFO = 1 << 2 //Write

// mini uart: line status register bitfields
const ReceivedDataAvailable = 1 << 0
const ReceivedDataOverrun = 1 << 1
const TransmitFIFOSpaceAvailable = 1 << 5
const TransmitterIdle = 1 << 6

// mini uart: interrupt enable register bitfields
//https://elinux.org/BCM2835_datasheet_errata#p12 (does not explain two magic bits 3:2)
//https://github.com/LdB-ECM/Raspberry-Pi/blob/bc38ce183f731891d52a31df87df24904e466d0c/PlayGround/rpi-SmartStart.c#L255
const ReceiveFIFOReady = 1 << 0
const TransmitFIFOEmpty = 1 << 1
const LineStatusError = 1 << 2 //overrun error, parity error, framing error
const ModemStatusChange = 1 << 3 //changes to DSR/CTS

type GPIORegisterMap struct {
	FuncSelect               [6]volatile.Register32 //0x00,04,08,0C,10, and 14
	reserved00               volatile.Register32 //0x18
	OutputSet0               volatile.Register32 //0x1C
	OutputSet1               volatile.Register32 //0x20
	reserved01               volatile.Register32 //0x24
	OutputClear0             volatile.Register32 //0x28
	OutputClear1             volatile.Register32 //0x2C
	reserved03               volatile.Register32 //0x30
	Level0                   volatile.Register32 //0x34
	Level1                   volatile.Register32 //0x38
	reserved04               volatile.Register32 //0x3C
	EventDetectStatus0       volatile.Register32 //0x40
	EventDetectStatus1       volatile.Register32 //0x44
	reserved05               volatile.Register32 //0x48
	RisingEdgeDetectEnable0  volatile.Register32 //0x4C
	RisingEdgeDetectEnable1  volatile.Register32 //0x50
	reserved06               volatile.Register32 //0x54
	FallingEdgeDetectEnable0 volatile.Register32 //0x58
	FallingEdgeDetectEnable1 volatile.Register32 //0x5C
	reserved07               volatile.Register32 //0x60
	HighDetectEnable9        volatile.Register32 //0x64
	HighDetectEnable1        volatile.Register32 //0x68
	reserved08               volatile.Register32 //0x6C
	LowDetectEnable0         volatile.Register32 //0x70
	LowDetectEnable1         volatile.Register32 //0x74
	reserved09               volatile.Register32 //0x78
	AsyncRisingEdgeDetect0   volatile.Register32 //0x7C
	AsyncRisingEdgeDetect1   volatile.Register32 //0x80
	reserved0A               volatile.Register32 //0x84
	AsyncFallingEdgeDetect0  volatile.Register32 //0x88
	AsyncFallingEdgeDetect1  volatile.Register32 //0x8C
	reserved0B               volatile.Register32 //0x90
	PullUpDownEnable         volatile.Register32 // 0x94
	PullUpDownEnableClock0   volatile.Register32 //0x98
	PullUpDownEnableClock1   volatile.Register32 //0x9C
	reserved0C               volatile.Register32 //0xA0
	test                     volatile.Register32 //0xA4
}

type GPIOMode uint32 //3 bits wide
const GPIOInput GPIOMode = 0
const GPIOOutput GPIOMode = 1
const GPIOAltFunc5 GPIOMode = 2
const GPIOAltFunc4 GPIOMode = 3
const GPIOAltFunc0 GPIOMode = 4
const GPIOAltFunc1 GPIOMode = 5
const GPIOAltFunc2 GPIOMode = 6
const GPIOAltFunc3 GPIOMode = 7

type SysTimerRegisterMap struct {
	ControlStatus   volatile.Register32 //0x00
	CounterLower32  volatile.Register32 //0x04
	CounterHigher32 volatile.Register32 //0x08
	reservedGPU0    volatile.Register32 //0x0C
	Compare1        volatile.Register32 //0x10
	reservedGPU2    volatile.Register32 //0x14
	Compare3        volatile.Register32 //0x18
}

type IRQRegisterMap struct {
	IRQBasicPending  volatile.Register32 //0x00
	IRQPending1      volatile.Register32 //0x04
	IRQPending2      volatile.Register32 //0x08
	FIQControl       volatile.Register32 //0x0C
	EnableIRQs1      volatile.Register32 //0x10
	EnableIRQs2      volatile.Register32 //0x14
	EnableBasicIRQs  volatile.Register32 //0x18
	DisableIRQs1     volatile.Register32 //0x1C
	DisableIRQs2     volatile.Register32 //0x20
	DisableBasicIRQs volatile.Register32 //0x24
}

// ***************************************
// SCTLR_EL1, System Control Register (EL1), Page 2654 of AArch64-Reference-Manual.
// ***************************************

const SystemControlRegisterEL1Reserved = (3 << 28) | (3 << 22) | (1 << 20) | (1 << 11)
const SystemControlRegisterEELittleEndian = (0 << 25)
const SystemControlRegisterEOELittleEndian = (0 << 24)
const SystemControlRegisterICacheDisabled = (0 << 12)
const SystemControlRegisterDCacheDisabled = (0 << 2)
const SystemControlRegisterMMUDisabled = (0 << 0)
const SystemControlRegisterMMUEnabled = (1 << 0)

const SystemControlRegisterValueMMUDisabled = (SystemControlRegisterEL1Reserved | // 0x30000000 | 0xC00000 | 0x100000 | 0x800 => 0x30D00800
	SystemControlRegisterEELittleEndian | //0x0
	SystemControlRegisterICacheDisabled | //0x0
	SystemControlRegisterDCacheDisabled | //0x0
	SystemControlRegisterMMUDisabled) //0x0

// ***************************************
// HCR_EL2, Hypervisor Configuration Register (EL2), Page 2487 of AArch64-Reference-Manual.
// ***************************************

const HypervisorConfigurationRegisterRW = (1 << 31)
const HypervisorConfigurationRegisterValue = HypervisorConfigurationRegisterRW //0x80000000

// ***************************************
// SCR_EL3, Secure Configuration Register (EL3), Page 2648 of AArch64-Reference-Manual.
// ***************************************

const SecureConfigurationRegisterReserved = (3 << 4)
const SecureConfigurationRegisterRW = (1 << 10)
const SecureConfigurationRegisterNS = (1 << 0)
const SecureConfigurationRegisterValue = //0x30 | 0x400 | 0x1 => 0x431
SecureConfigurationRegisterReserved |
	SecureConfigurationRegisterRW |
	SecureConfigurationRegisterNS

// ***************************************
// SPSR_EL3, Saved Program Status Register (EL3) Page 389 of AArch64-Reference-Manual.
// ***************************************

const SavedProgramStatusRegisterMaskAll = (7 << 6)
const SavedProgramStatusRegisterEl1h = (5 << 0)                                                            //EL1 has own stack
const SavedProgramStatusRegisterValue =
	SavedProgramStatusRegisterMaskAll |
		SavedProgramStatusRegisterEl1h //0x1C0

func Abort() {
	MiniUART.WriteString("Aborting...")
	arm.Asm("wfe")
}


//
// Wait MuSec waits for at least n musecs based on the system timer. This is a busy wait.
//
//go:export WaitMuSec
func WaitMuSec(n uint64) {
	var f, t, r uint64
	arm.AsmFull(`mrs x28, cntfrq_el0
		str x28,{f}
		mrs x27, cntpct_el0
		str x27,{t}`, map[string]interface{}{"f": &f, "t": &t})
	//expires at t
	t += ((f / 1000) * n) / 1000
	for r < t {
		arm.AsmFull(`mrs x27, cntpct_el0
			str x27,{r}`, map[string]interface{}{"r": &r})
	}
}

//
// SysTimer gets the 64 bit timer's value.
//
//go:export SystemTime()
func SystemTime() uint64 {
	h := uint32(0xffffffff)
	var l uint32

	// the reads from system timer are two separate 32 bit reads
	h = SysTimer.CounterHigher32.Get()
	l = SysTimer.CounterLower32.Get()
	//the read of hi can fail
	again := SysTimer.CounterHigher32.Get()
	if h != again {
		h = SysTimer.CounterHigher32.Get()
		l = SysTimer.CounterLower32.Get()
	}
	high := uint64(h << 32)
	return high | uint64(l)
}

// XXX Unclear how to do the PIN mapping for a RPI3 because of function select
type PinMode struct{}

func (p Pin) Set(_ bool) {}

//
// RPI has many uarts, this is the "miniuart" which is the simplest to configure.
//
const RxBufMax = 0xfff

type UART struct {
	rxhead   *volatile.Register32
	rxtail   *volatile.Register32
	rxbuffer []uint8
}

func NewUART() UART {
	return UART{
		rxhead:   &volatile.Register32{},
		rxtail:   &volatile.Register32{},
		rxbuffer: make([]uint8, RxBufMax+1),
	}
}

//
// For now, zero value is a non-interrupt UART at 8bits, bidirectional.
// If you set EnableRXInterrupt, you'll need to actually "turn on" the
// interrupts when you are ready:
// InterruptController.EnableIRQs1.SetBits(1<<AuxInterrupt)
type UARTConfig struct {
	RXInterrupt bool
	Data7Bits         bool
	DisableTx         bool
	DisableRx         bool
}

func GPIOSetup(pinNumber uint8, mode GPIOMode) bool {
	if (pinNumber > 54) { //54 pins on RPI
		return false
	}									// Check GPIO pin number valid, return false if invalid
	var shift uint8
	shift = ((pinNumber % 10) * 3);							// Create shift amount

	value := uint32(mode)
	mask:=uint32(7)

	GPIO.FuncSelect[pinNumber % 10].ReplaceBits(value,mask,shift)
	return true;													// Return true
}


// Configure accepts a config object to set some simple
// properties of the UART.  It is not a fully featured 16550 UART,
// rather it is the "mini" UART.   The zero value of conf
// gives you 8 bits, no interrupts, and both tx and rx enabled.
func (uart UART) Configure(conf UARTConfig) error {
	var r uint32

	Aux.Enables.SetBits(PeripheralMiniUART) //enable AUX Mini uart

	//turn off the transmitter and receiver
	Aux.MiniUARTExtraControl.ClearBits(ReceiveEnable|TransmitEnable)
	Aux.MiniUARTExtraControl.Set(0)

	//configure data bits
	if conf.Data7Bits {
		Aux.MiniUARTLineControl.ClearBits(DataLength8Bits) //7 bits
	} else {
		//see errata for why (bad docs!) uses excuse of compat with 16550
		// https://elinux.org/BCM2835_datasheet_errata#p14
		Aux.MiniUARTLineControl.SetBits(DataLength8Bits)
	}

	Aux.MiniUARTModemControl.ClearBits(ReadyToSend) // this asserts the line
	Aux.MiniUARTInterruptIdentify.ReplaceBits(ClearTransmitFIFO|ClearReceiveFIFO, ClearFIFOsMask,0/*no shift*/)

	// derived from clock speed: BCM2835 ARM Peripheral manual page 11
	Aux.MiniUARTBAUD.Set(270)               // 115200 baud

	//set the bits
	if conf.RXInterrupt {
		Aux.MiniUARTInterruptEnable.SetBits(ReceiveFIFOReady)
	} else {
		Aux.MiniUARTInterruptEnable.ClearBits(ReceiveFIFOReady|TransmitFIFOEmpty|LineStatusError|ModemStatusChange)
	}

	// map UART1 to GPIO pins
	GPIOSetup(14,GPIOAltFunc5)
	GPIOSetup(15,GPIOAltFunc5)

	//sleep 150 cycles
	r = 150
	for r > 0 {
		r--
		arm.Asm("nop")
	}

	GPIO.PullUpDownEnableClock0.SetBits((1 << 14) | (1 << 15))

	//sleep 150 cycles
	r = 150
	for r > 0 {
		r--
		arm.Asm("nop")
	}

	GPIO.PullUpDownEnableClock0.Set(0) //flush gpio setup

	if !conf.DisableRx {
		Aux.MiniUARTExtraControl.SetBits(ReceiveEnable)
	} else {
		Aux.MiniUARTExtraControl.ClearBits(ReceiveEnable)
	}
	if !conf.DisableTx {
		Aux.MiniUARTExtraControl.SetBits(TransmitEnable)
	} else {
		Aux.MiniUARTExtraControl.ClearBits(TransmitEnable)
	}

	return nil
}

//
// Writing a byte over serial.  Blocking.
//
func (uart UART) WriteByte(c byte) error {
	// wait until we can send
	for {
		if Aux.MiniUARTLineStatus.HasBits(TransmitFIFOSpaceAvailable) {
			break
		}
		arm.Asm("nop")
	}

	// write the character to the buffer
	c32 := uint32(c)
	Aux.MiniUARTData.Set(c32) //really 8 bit write
	return nil
}

//
// Write a CR (and secretly an LF) to serial.
//
func (uart UART) WriteCR() error {
	if err:=uart.WriteByte(10); err!=nil {
		return err
	}
	if err:=uart.WriteByte(13); err!=nil {
		return err
	}
	return nil
}
//
// Reading a byte from serial. Blocking.
//
func (uart UART) ReadByte() uint8 {
	for {
		if Aux.MiniUARTLineStatus.HasBits(ReceiveFIFOReady) {
			break
		}
		arm.Asm("nop")
	}
	r := Aux.MiniUARTData.Get() //8 bit read
	return uint8(r)
}

//
// Put a whole string out to serial. Blocking.
//
func (uart UART) WriteString(s string) error {
	for i := 0; i < len(s); i++ {
		uart.WriteByte(s[i])
	}
	return nil
}

func (uart UART) Hex32string(d uint32) {
	var rb uint32
	var rc uint32

	rb = 32
	for {
		rb -= 4
		rc = (d >> rb) & 0xF
		if rc > 9 {
			rc += 0x37
		} else {
			rc += 0x30
		}
		uart.WriteByte(uint8(rc))
		if rb == 0 {
			break
		}
	}
	uart.WriteByte(0x20)
}

func (uart UART) Hex64string(d uint64) {
	var rb uint64
	var rc uint64

	rb = 64
	for {
		rb -= 4
		rc = (d >> rb) & 0xF
		if rc > 9 {
			rc += 0x37
		} else {
			rc += 0x30
		}
		uart.WriteByte(uint8(rc))
		if rb == 0 {
			break
		}
	}
	uart.WriteByte(0x20)
}

// DumpRxBuffer is for debugging. Returns the size of the buffer dumped.
func (uart UART) DumpRxBuffer() uint32 {
	moved := uint32(0)
	for !uart.EmptyRx() {
		index := uart.rxtail.Get()
		uart.WriteByte(uart.rxbuffer[index])
		tail := index + 1
		tail &= RxBufMax
		uart.rxtail.Set(tail)
		moved++
	}
	return moved
}

// LoadRx puts a byte in the RxBuffer as if it came in from
// the other side.  This is probably only interesting for callers
// if they are doing testing.
func (uart UART) LoadRx(b uint8) {
	//receiver holds a valid byte
	index := uart.rxhead.Get()
	uart.rxbuffer[index] = b
	head := index + 1
	head &= RxBufMax
	uart.rxhead.Set(head)
}

// EmptyRx is true if the receiver ring buffer is empty.
func (uart UART) EmptyRx() bool {
	return uart.rxtail.Get() == uart.rxhead.Get()
}

// Returns the next element from the read queue.  Note that
// this busy waits on EmptyRx() so you should be sure
// there is data there before you call this or it will block
// and only an interrupt can save that...
func (uart UART) NextRx() uint8 {
	for {
		if !uart.EmptyRx() {
			break
		}
	}
	result := uart.rxbuffer[uart.rxtail.Get()]
	tail := uart.rxtail.Get() + 1
	tail &= RxBufMax
	uart.rxtail.Set(tail)
	return result
}

// pullRXData is only called if you have enabled RX interrupts *and* you have
// inserted functions address into the proper interrupt vector.
func (uart UART) pullRXData() {
	//an interrupt has occurred, find out why
	for { //resolve all interrupts to uart
		//weird, clear means we DO have interrupt
		if Aux.MiniUARTInterruptIdentify.HasBits(1) {
			break //no more interrupts
		}
		//this ignores the possibility that HasBits(6) because docs (!)
		//say that bits 2 and 1 cannot both be set, so we just check bit 2
		if Aux.MiniUARTInterruptIdentify.HasBits(4) {
			//receiver holds a valid byte
			rc := Aux.MiniUARTData.Get() //read byte from rx fifo
			uart.LoadRx(uint8(rc))
		}
	}
}
//////////////////////////////////////////////////////////////////
// ARM64 Exception Handlers
//////////////////////////////////////////////////////////////////
type interruptHandler func(uint64,uint64, uint64)

//go:export excptrs
var excptrs [16]interruptHandler

func handleIRQ(t uint64, esr uint64, addr uint64) {
	MiniUART.WriteString("woot! ")
	MiniUART.WriteString(", ESR 0x")
	MiniUART.Hex64string(esr)
	MiniUART.WriteString(", ADDR 0x")
	MiniUART.Hex64string(addr)
	MiniUART.WriteCR()

	for { //resolve all interrupts to uart
		//weird, clear means we DO have interrupt
		if Aux.MiniUARTInterruptIdentify.HasBits(1) {
			break //no more interrupts
		}
		//this ignores the possibility that HasBits(6) because docs (!)
		//say that bits 2 and 1 cannot both be set, so we just check bit 2
		if Aux.MiniUARTInterruptIdentify.HasBits(4) {
			MiniUART.pullRXData()
		}
	}
}


// MaskDAIF sets the value of the four D-A-I-F interupt masking on the ARM
func MaskDAIF() {
	arm.Asm("msr    daifset, #0xf")

}
// UnmaskDAIF sets the value of the four D-A-I-F interupt masking on the ARM
func UnmaskDAIF() {
	arm.Asm("msr    daifclr, #0xf")
}

// see rpi3.ld
const exceptionBase = 0x2000

//go:extern single_exception_code_ptr
var singleExceptionCodePtr uint64

// InitInterrupts is called by postinit (before main) to make sure all the interrupt
// machinery is in the right startup state.
func InitInterrupts() {
	arm.Asm("adr	x0, #0x76000")		// load VBAR_EL1 with exc vector
	arm.Asm("msr	vbar_el1, x0")
	MaskDAIF()
	InterruptController.DisableIRQs1.SetBits(AuxInterrupt)
}

// when an interrupt falls in the woods and nobody is around to hear it
func showInvalidEntryMessage(t uint64, esr uint64, addr uint64 ) {
	MiniUART.WriteString(entryErrorMessages[t])
	MiniUART.WriteString(", ESR 0x")
	MiniUART.Hex64string(esr)
	MiniUART.WriteString(", ADDR 0x")
	MiniUART.Hex64string(addr)
	MiniUART.WriteCR()
}

var entryErrorMessages = []string {
	"SYNC_INVALID_EL1t",
	"IRQ_INVALID_EL1t",
	"FIQ_INVALID_EL1t",
	"ERROR_INVALID_EL1T",

	"SYNC_INVALID_EL1h",
	"IRQ_INVALID_EL1h",
	"FIQ_INVALID_EL1h",
	"ERROR_INVALID_EL1h",

	"SYNC_INVALID_EL0_64",
	"IRQ_INVALID_EL0_64",
	"FIQ_INVALID_EL0_64",
	"ERROR_INVALID_EL0_64",

	"SYNC_INVALID_EL0_32",
	"IRQ_INVALID_EL0_32",
	"FIQ_INVALID_EL0_32",
	"ERROR_INVALID_EL0_32",
}
