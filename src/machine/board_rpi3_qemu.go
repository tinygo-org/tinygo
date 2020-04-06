// +build rpi3_qemu

package machine

import (
	"device/arm"
	"runtime"
	"runtime/volatile"
	"unsafe"
)

var MiniUART *UART

// XXX Unclear how to do the PIN mapping for a RPI3 because of function select.
// XXX Pins on RPI have many functions and can be remapped in software.
// XXX see GPIO in rpi3.svd.go
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

func NewUART() *UART {
	return &UART{
		rxhead:   0,
		rxtail:   0,
		rxbuffer: make([]uint8, RxBufMax+1),
	}
}

//
// Config options for the MiniUART
//
type UARTConfig struct {
	RXInterrupt bool
	Data7Bits   bool
	DisableTx   bool
	DisableRx   bool
}

// This is the standard setup which is available in many places on the internet.
func (uart UART) Configure(conf *UARTConfig) error {

	Aux.Enable.SetMiniUART()              //tell Aux we want MiniUART on
	Aux.MUCNTL.ClearTransmitterEnable()   //turn off transmitter
	Aux.MUCNTL.ClearReceiverEnable()      //turn off receiver
	Aux.MUIER.ClearReceive()              //no interrupts yet
	Aux.MULCR.SetEightBit()               //8,n,1 are the settings
	Aux.MUMCR.ClearRTS()                  //why is this necessary?
	Aux.MUIIR.SetZeroTransmitAndReceive() //dump the fifos
	Aux.MUBaud.SetBaudrate(270)           //115200 baud
	Aux.MUCNTL.SetReceiverEnable()        //enable receiver
	Aux.MUCNTL.SetTransmitterEnable()     //enable receiver

	if conf.RXInterrupt { //user wants interrupts?
		Aux.MUIER.SetReceive() //recv
		//Aux.AuxMUIER.SetReadErr() //do we want read errors?
	}

	return nil
}

//
// Writing a byte over serial.  Blocking.
//
func (uart UART) WriteByte(c byte) error {
	// wait until we can send
	for {
		//maybe should use Aux.MiniUARTStatus.SpaceAvailableIsSet()?
		if Aux.MULSR.TransmitterEmptyIsSet() { //HasBits(0x20) {
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
		if Aux.MULSR.DataReadyIsSet() { //HasBits(0x01){
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

//export SystemTime
func SystemTime() uint64 {
	return runtime.Semihostingv2ClockMicros()
}

type MiniUARTWriter struct {
}

func (m *MiniUARTWriter) Write(p []byte) (n int, err error) {
	for _, c := range p {
		MiniUART.WriteByte(c)
	}
	return len(p), nil
}

//ActivityLED obviously does not control a real LED on the qemu simulator.
func ActivityLED(on bool) {
	print("Activite LED: ", on, "\n")
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

func Abort() {
	print("Aborting...\n")
	runtime.Semihostingv2Call(uint64(runtime.Semihostingv2OpExit), uint64(runtime.Semihostingv2StopApplicationExit))

}
