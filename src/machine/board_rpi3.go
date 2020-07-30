// +build rpi3

package machine

import (
	"device/arm"
	"runtime/volatile"
	"unsafe"
)

var MiniUART *UART

// XXX Unclear how to do the PIN mapping for a RPI3 because of function select.
// XXX Pins on RPI have many functions and can be remapped in software.
// XXX see GPIO in rpi3.sysdec.go
type PinMode struct{}

func (p Pin) Set(_ bool) {}

//
// RPI has many uarts, this is the "miniuart" which is the simplest to configure.
//
const RxBufMax = 0xfff

type UART struct {
	rxhead   int
	rxtail   int
	rxbuffer []uint8
}

//go:noinline
func NewUART() *UART {
	return &UART{
		rxhead:   0,
		rxtail:   0,
		rxbuffer: make([]uint8, RxBufMax+1),
	}
}

//
// Configuration options for the MiniUART.
//
type UARTConfig struct {
	RXInterrupt bool
	Data7Bits   bool
	DisableTx   bool
	DisableRx   bool
}

// This is basicaly the "standard" setup that is available many places
// on the internet.
func (uart UART) Configure(conf *UARTConfig) error {

	Aux.Enable.SetMiniUART()              //tell Aux we want MiniUART on
	Aux.MUCNTL.ClearTransmitterEnable()   //turn off transmitter
	Aux.MUCNTL.ClearReceiverEnable()      //turn off receiver
	Aux.MUIER.ClearReceive()              //no interrupts yet
	Aux.MULCR.SetEightBit()               //8,n,1 are the settings
	Aux.MUMCR.ClearRTS()                  //why is this necessary?
	Aux.MUIIR.SetZeroTransmitAndReceive() //dump the fifos
	Aux.MUCNTL.SetReceiverEnable()        //enable receiver
	Aux.MUCNTL.SetTransmitterEnable()     //enable receiver

	//use mailboxes to get the clock rate
	rate, ok := GetClockRate()
	if !ok {
		panic("unable to read the clock rate from the mailbox")
	}
	//calculate divisor
	baudRate := uint32(115200)
	divisor := (rate / (baudRate * 8)) - 1
	Aux.MUBaud.SetBaudrate(divisor) //115200 baud
	if conf.RXInterrupt {
		Aux.MUIER.SetReceive()
	}

	GPIOSetup(14, FunctionSelectGPIOAltFunc5)
	GPIOSetup(15, FunctionSelectGPIOAltFunc5)
	GPIODisablePullUpDown(0, (1<<14)|(1<<15))

	Aux.MUCNTL.SetReceiverEnable()    //enable rcv
	Aux.MUCNTL.SetTransmitterEnable() //enable tx

	return nil
}

//
// Writing a byte over serial.  Blocking.
//
func (uart UART) WriteByte(c byte) error {
	// wait until we can send
	for {
		//maybe should use extra control SpaceAvailableIsSet()?
		if Aux.MULSR.TransmitterEmptyIsSet() {
			break
		}
		arm.Asm("nop")
	}

	// write the character to the buffer
	c32 := uint32(c)
	Aux.MUData.SetTransmit(c32)
	return nil
}

//
// Reading a byte from serial. Blocking.
//
func (uart UART) ReadByte() uint8 {
	for {
		if Aux.MULSR.DataReadyIsSet() {
			break
		}
		arm.Asm("nop")
	}
	r := Aux.MUData.Receive()
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

// CopyRxBuffer copies entire RX buffer into target and leaves it empty
// Caller needs to insure that all of the buffer can fit into target
func (uart *UART) CopyRxBuffer(target []byte) uint32 {
	moved := uint32(0)
	found := false
	for !uart.EmptyRx() {
		index := uart.rxtail
		target[moved] = uart.rxbuffer[index]
		targ := target[moved]
		if !found {
			moved++
		}
		if targ == 10 {
			found = true
		}
		tail := index + 1
		tail &= RxBufMax
		uart.rxtail = tail
	}
	return moved
}

// DumpRxBuffer pushes the entire RX buffer out to serial and leaves the
// buffer empty.
func (uart *UART) DumpRxBuffer() uint32 {
	moved := uint32(0)
	for !uart.EmptyRx() {
		index := uart.rxtail
		uart.WriteByte(uart.rxbuffer[index])
		tail := index + 1
		tail &= RxBufMax
		uart.rxtail = tail
		moved++
	}
	return moved
}

// LoadRx puts a byte in the RxBuffer as if it came in from
// the other side.  Probably should be called from an exception
// hadler.
func (uart *UART) LoadRx(b uint8) {
	//receiver holds a valid byte
	index := uart.rxhead
	uart.rxbuffer[index] = b
	head := index + 1
	head &= RxBufMax
	uart.rxhead = head
}

// EmptyRx is true if the receiver ring buffer is empty.
func (uart *UART) EmptyRx() bool {
	return uart.rxtail == uart.rxhead
}

// Returns the next element from the read queue.  Note that
// this busy waits on EmptyRx() so you should be sure
// there is data there before you call this or it will block
// and only an interrupt can save that...
func (uart *UART) NextRx() uint8 {
	for {
		if !uart.EmptyRx() {
			break
		}
	}
	result := uart.rxbuffer[uart.rxtail]
	tail := uart.rxtail + 1
	tail &= RxBufMax
	uart.rxtail = tail
	return result
}

//
// GPIO Helpers
//
func (g *GPIODef) Wait150Cycles() {
	//sleep 150 cycles
	r := 150
	for r > 0 {
		r--
		arm.Asm("nop")
	}
}

type FunctionSelectType int

const (
	FunctionSelectGPIOInput    FunctionSelectType = 0
	FunctionSelectGPIOOutput   FunctionSelectType = 1
	FunctionSelectGPIOAltFunc5 FunctionSelectType = 2
	FunctionSelectGPIOAltFunc4 FunctionSelectType = 3
	FunctionSelectGPIOAltFunc0 FunctionSelectType = 4
	FunctionSelectGPIOAltFunc1 FunctionSelectType = 5
	FunctionSelectGPIOAltFunc2 FunctionSelectType = 6
	FunctionSelectGPIOAltFunc3 FunctionSelectType = 7
)

//GPIOSetup sets the preferred function mode f on the pin given.
func GPIOSetup(pin int, fst FunctionSelectType) {
	if pin < 0 || pin > 54 {
		panic("bad pin number trying to initialize GPIO pins")
	}
	f := int(fst)
	if f < int(FunctionSelectGPIOInput) || f > int(FunctionSelectGPIOAltFunc3) {
		panic("bad alternate function selected for GPIO pin")
	}
	shift := ((pin % 10) * 3)                   // Create shift
	mem := GPIO.FSel[pin/10]                    // Read register
	intermediate := (mem.Get() &^ (7 << shift)) // Clear GPIO mode bits for that port
	mem.Set(intermediate | uint32(f<<shift))    // Logical OR GPIO mode bits and write it back

}

// index is 0 for pins less than 32, 1 for pins greater than or equal to 32
// shifted pin numbers assumes that you mod 32 the pin number like this: 1<<(pinNumber%32)
func GPIODisablePullUpDown(index int, shiftedPinNumbers uint32) {
	GPIO.GPPUD.Set(0)
	GPIO.Wait150Cycles()
	GPIO.GPUDClk[index].Set(shiftedPinNumbers)
	GPIO.Wait150Cycles()
	GPIO.GPUDClk[index].Set(0)
}

//Activity LED turns on or off the green LED on the RPI3.
var model *uint32

func ActivityLED(on bool) bool {
	if model == nil {
		m, ok := BoardRevision()
		if !ok {
			return false
		}
		model = &m
		print("model is ", uintptr(*model), "\n")
	}
	var UseExpanderGPIO bool

	if (*model == 0xa02082) || (*model == 0xa020a0) || (*model == 0xa22082) || (*model == 0xa32082) { // These are Pi3B series originals (ARM8)
		UseExpanderGPIO = true // Must use expander GPIO
	} else { // These are Pi3B+ series (ARM8)
		UseExpanderGPIO = false // Don't use expander GPIO
	}
	if UseExpanderGPIO { // Activity LED uses expander GPIO
		gpioExpanderActivityLED(on)
	} else {
		GPIOOutput(29, on)
	}
	return true
}

// GPIOOutput sets the output of a particular pin to be high or low.
func GPIOOutput(gpio uint32, on bool) bool {
	if gpio < 0 || gpio < 54 { // Check GPIO pin number valid, return false if invalid
		return false
	}
	regnum := gpio / 32             // Register number
	bit := uint32(1) << (gpio % 32) // Create mask bit
	if on {
		out := GPIO.GPSet[regnum]
		out.SetBits(bit)
	} else {
		out := GPIO.GPClr[regnum]
		out.SetBits(bit)
	}
	return true
}

//
// MISC Utilities
//

func Abort() {
	MiniUART.WriteString("Aborting...\n")
	for {
		arm.Asm("wfe")
	}
}

//
// Wait MuSec waits for at least n musecs based on the system timer. This is a busy wait.
//
//export WaitMuSec
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

//export SystemTime()
func SystemTime() uint64 {
	h := uint32(0xffffffff)
	var l uint32

	// the reads from system timer are two separate 32 bit reads
	for {
		h = SystemTimer.MostSignificant32.Get()
		l = SystemTimer.LeastSignificant32.Get()
		again := SystemTimer.MostSignificant32.Get() //watch for rollover
		if h == again {
			break
		}
	}
	high := uint64(h << 32)
	return high | uint64(l)
}

///
/// MAILBOX STUFF
///

const MailboxFull = 0x80000000
const MailboxEmpty = 0x40000000
const MailboxResponse = 0x80000000
const MailboxRequest = 0x0

/* channels */
const MailboxChannelProperties = 8

const MailboxTagSerial = 0x100004
const MailboxTagFirmwareVersion = 0x1
const MailboxTagBoardModel = 0x00010001
const MailboxTagBoardRevision = 0x00010002
const MailboxTagMACAddress = 0x00010003
const MailboxTagGetClockRate = 0x00030002
const MailboxTagLast = 0x0
const MailboxTagGetVCMemory = 0x00010006
const MailboxTagGetARMMemory = 0x00010005
const MailboxTagSetGPIOState = 0x00038041

func GetClockRate() (uint32, bool) {
	buffer := message(2, MailboxTagGetClockRate, 2)

	buffer.s[5].Set(0x4)
	buffer.s[6].Set(0)
	if !Call(MailboxChannelProperties, buffer) {
		return 7903, false
	}
	if buffer.s[5].Get() != 4 {
		return 7904, false
	}
	return buffer.s[6].Get(), true
}

func BoardID() (uint64, bool) {
	return MessageNoParams(MailboxTagSerial, 2)
}

func FirmwareVersion() (uint32, bool) {
	firmware, ok := MessageNoParams(MailboxTagFirmwareVersion, 1)
	if !ok {
		return 0x872720, ok
	}
	return uint32(firmware), true
}

func MACAddress() (uint64, bool) {
	addr, ok := MessageNoParams(MailboxTagMACAddress, 2)
	if !ok {
		return 0xab127348, false
	}

	addr &= 0x0000ffffffffffffffff
	return addr, true
}

func BoardModel() (uint32, bool) {
	model, ok := MessageNoParams(MailboxTagBoardModel, 1)
	if !ok {
		return 0x872728, ok
	}
	return uint32(model), true
}

func BoardRevision() (uint32, bool) {
	revision, ok := MessageNoParams(MailboxTagBoardRevision, 1)
	if !ok {
		return 0x872727, ok
	}
	return uint32(revision), true
}

// Note: This can ONLY be used on the Pi3B (not Pi3B+) because Pi3B+ uses the
// Note: the GPIO pins a different way.
func gpioExpanderActivityLED(on bool) bool {
	value := uint32(0)
	if on {
		value = 1
	}
	buffer := message(2, MailboxTagSetGPIOState, 0)
	buffer.s[5].Set(130)
	buffer.s[6].Set(value)
	return Call(MailboxChannelProperties, buffer)
}

type sequenceOfSlots struct {
	s [36]volatile.Register32
}

func message(requestSlots int, tag uint32, responseSlots int) *sequenceOfSlots {

	totalSlots := uint32(1 + 1 + 1 + requestSlots + 1 + 1 + responseSlots + 1)
	larger := responseSlots
	if requestSlots > larger {
		larger = requestSlots
	}
	ptr := sixteenByteAlignedPointer(uintptr(totalSlots << 2)) //32 bit slots
	ptr32 := ((*uint32)(unsafe.Pointer(ptr)))
	seq := (*sequenceOfSlots)(unsafe.Pointer(ptr32))

	seq.s[0].Set(4 * totalSlots) //bytes of total size
	seq.s[1].Set(MailboxRequest)
	seq.s[2].Set(tag)
	seq.s[3].Set(uint32(larger) * 4)
	seq.s[4].Set(0) //request
	//s5...s5+larger-1 will be the outgoing data
	next := 5 + larger
	seq.s[next].Set(MailboxTagLast)
	return seq
}

// Uses of this function are NOT multithread safe. This is uses a single, shared
// mailbox data area.
func Call(ch uint8, mboxBuffer *sequenceOfSlots) bool {
	mask := uintptr(^uint64(0xf))
	rawPtr := uintptr(unsafe.Pointer(mboxBuffer))
	if rawPtr&0xf != 0 {
		Abort()
	}
	addrWithChannel := (uintptr(unsafe.Pointer(rawPtr)) & mask) | uintptr(ch&0xf)
	for {
		if GPUMailbox.Status.FullIsSet() {
			arm.Asm("nop")
		} else {
			break
		}
	}
	GPUMailbox.Write.Set(uint32(addrWithChannel))
	//for i := 0; i < 20; i++ {
	//      happiness.Console.Logf("%x,%x\n ", rawPtr, addrWithChannel)
	//}
	//happiness.Console.Logf("wasting time so the mailbox won't feel in a hurry...")
	for {
		if GPUMailbox.Status.EmptyIsSet() {
			arm.Asm("nop")
		} else {
			read := GPUMailbox.Receive.Get()
			if read == uint32(addrWithChannel) {
				//did we get a confirm?
				return mboxBuffer.s[1].Get() == MailboxResponse
			}
		}
	}
	return false //how would this happen?

}

func sixteenByteAlignedPointer(size uintptr) *uint64 {
	units := (((size / 16) + 1) * 16) / 8
	bigger := make([]uint64, units)
	hackFor16ByteAlignment := ((*uint64)(unsafe.Pointer(&bigger[0])))
	ptr := uintptr(unsafe.Pointer(hackFor16ByteAlignment))
	if ptr&0xf != 0 {
		diff := uintptr(16 - (ptr & 0xf))
		hackFor16ByteAlignment = ((*uint64)(unsafe.Pointer(ptr + diff)))
	}
	return hackFor16ByteAlignment
}

func MessageNoParams(tag uint32, reqRespSlots int) (uint64, bool) {
	seq := message(0, tag, 2)
	if !Call(MailboxChannelProperties, seq) {
		return 77281, false
	}
	if reqRespSlots == 1 {
		return uint64(seq.s[5].Get()), true
	}
	if reqRespSlots == 2 {
		upper := uint64(seq.s[6].Get() << 32)
		lower := uint64(seq.s[5].Get())
		return upper + lower, true
	}
	panic("too many response slots")
}

type MiniUARTWriter struct {
}

func (m *MiniUARTWriter) Write(p []byte) (n int, err error) {
	for _, c := range p {
		MiniUART.WriteByte(c)
	}
	return len(p), nil
}
