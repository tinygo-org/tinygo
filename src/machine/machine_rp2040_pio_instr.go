package machine

//
// This file contains the primitives for creating instructions dynamically
//
const (
	PIO_INSTR_BITS_JMP  = 0x0000
	PIO_INSTR_BITS_WAIT = 0x2000
	PIO_INSTR_BITS_IN   = 0x4000
	PIO_INSTR_BITS_OUT  = 0x6000
	PIO_INSTR_BITS_PUSH = 0x8000
	PIO_INSTR_BITS_PULL = 0x8080
	PIO_INSTR_BITS_MOV  = 0xa000
	PIO_INSTR_BITS_IRQ  = 0xc000
	PIO_INSTR_BITS_SET  = 0xe000

	// Bit mask for instruction code
	PIO_INSTR_BITS_Msk = 0xe000
)

type PIOSrcDest uint16

const (
	PIOSrcDestPins    PIOSrcDest = 0
	PIOSrcDestX                  = 1
	PIOSrcDestY                  = 2
	PIOSrcDestNull               = 3
	PIOSrcDestPinDirs            = 4
	PIOSrcDestExecMov            = 4
	PIOSrcDestStatus             = 5
	PIOSrcDestPC                 = 5
	PIOSrcDestISR                = 6
	PIOSrcDestOSR                = 7
	PIOSrcExecOut                = 7
)

func PIOMajorInstrBits(instr uint16) uint16 {
	return instr & PIO_INSTR_BITS_Msk
}

func PIOEncodeInstrAndArgs(instr uint16, arg1 uint16, arg2 uint16) uint16 {
	return instr | (arg1 << 5) | (arg2 & 0x1f)
}

func PIOEncodeInstrAndSrcDest(instr uint16, dest PIOSrcDest, value uint16) uint16 {
	return PIOEncodeInstrAndArgs(instr, uint16(dest)&7, value)
}

func PIOEncodeDelay(cycles uint16) uint16 {
	return cycles << 8
}

func PIOEncodeSideSet(bitCount uint16, value uint16) uint16 {
	return value << (13 - bitCount)
}

func PIOEncodeSetSetOpt(bitCount uint16, value uint16) uint16 {
	return 0x1000 | value<<(12-bitCount)
}

func PIOEncodeJmp(addr uint16) uint16 {
	return PIOEncodeInstrAndArgs(PIO_INSTR_BITS_JMP, 0, addr)
}

func PIOEncodeIRQ(relative bool, irq uint16) uint16 {
	instr := irq

	if relative {
		instr |= 0x10
	}

	return instr
}

func PIOEncodeWaitGPIO(polarity bool, pin uint16) uint16 {
	flag := uint16(0)
	if polarity {
		flag = 0x4
	}

	return PIOEncodeInstrAndArgs(PIO_INSTR_BITS_WAIT, 0|flag, pin)
}

func PIOEncodeWaitPin(polarity bool, pin uint16) uint16 {
	flag := uint16(0)
	if polarity {
		flag = 0x4
	}

	return PIOEncodeInstrAndArgs(PIO_INSTR_BITS_WAIT, 1|flag, pin)
}

func PIOEncodeWaitIRQ(polarity bool, relative bool, irq uint16) uint16 {
	flag := uint16(0)
	if polarity {
		flag = 0x4
	}

	return PIOEncodeInstrAndArgs(PIO_INSTR_BITS_WAIT, 2|flag, PIOEncodeIRQ(relative, irq))
}

func PIOEncodeIn(src PIOSrcDest, value uint16) uint16 {
	return PIOEncodeInstrAndSrcDest(PIO_INSTR_BITS_IN, src, value)
}

func PIOEncodeOut(dest PIOSrcDest, value uint16) uint16 {
	return PIOEncodeInstrAndSrcDest(PIO_INSTR_BITS_OUT, dest, value)
}

func PIOEncodePush(ifFull bool, block bool) uint16 {
	arg := uint16(0)
	if ifFull {
		arg |= 2
	}
	if block {
		arg |= 1
	}

	return PIOEncodeInstrAndArgs(PIO_INSTR_BITS_PUSH, arg, 0)
}

func PIOEncodePull(ifEmpty bool, block bool) uint16 {
	arg := uint16(0)
	if ifEmpty {
		arg |= 2
	}
	if block {
		arg |= 1
	}

	return PIOEncodeInstrAndArgs(PIO_INSTR_BITS_PULL, arg, 0)
}

func PIOEncodeMov(dest PIOSrcDest, src PIOSrcDest) uint16 {
	return PIOEncodeInstrAndSrcDest(PIO_INSTR_BITS_MOV, dest, uint16(src)&7)
}

func PIOEncodeMovNot(dest PIOSrcDest, src PIOSrcDest) uint16 {
	return PIOEncodeInstrAndSrcDest(PIO_INSTR_BITS_MOV, dest, (1<<3)|(uint16(src)&7))
}

func PIOEncodeMovReverse(dest PIOSrcDest, src PIOSrcDest) uint16 {
	return PIOEncodeInstrAndSrcDest(PIO_INSTR_BITS_MOV, dest, (2<<3)|(uint16(src)&7))
}

func PIOEncodeIRQSet(relative bool, irq uint16) uint16 {
	return PIOEncodeInstrAndArgs(PIO_INSTR_BITS_IRQ, 0, PIOEncodeIRQ(relative, irq))
}

func PIOEncodeIRQClear(relative bool, irq uint16) uint16 {
	return PIOEncodeInstrAndArgs(PIO_INSTR_BITS_IRQ, 2, PIOEncodeIRQ(relative, irq))
}

func PIOEncodeSet(dest PIOSrcDest, value uint16) uint16 {
	return PIOEncodeInstrAndSrcDest(PIO_INSTR_BITS_SET, dest, value)
}

func PIOEncodeNOP() uint16 {
	return PIOEncodeMov(PIOSrcDestY, PIOSrcDestY)
}
