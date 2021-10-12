// +build esp32c3

package interrupt

import (
	"device/esp"
	"device/riscv"
	"errors"
	"runtime/volatile"
	"unsafe"
)

const (
	// Each module of the ESP32C3 uses different interrupt registers.
	// Below is the list of modules used from machine package to initialize interrupt handler
	GPIO HandlerID = iota
	UART0
	UART1
	USB
	I2C0
	I2S1
	TIMG0
	TIMG1
	UHCI0
	RMT
	SPI1
	SPI2
	TWAI
	RNG
	WIFI
	BT
	WIFI_BT_COMMON
	BT_BASEBAND
	BT_LC
	RSA
	AES
	SHA
	HMAC
	DS
	GDMA
	SYSTIMER
	SARADC
)

var (
	ErrInterruptNotImplemented = errors.New("interrupt number is not implemented")
)

type HandlerID int

func Init() {
	mie := riscv.DisableInterrupts()

	// Reset all interrupt source priorities to zero.
	priReg := &esp.INTERRUPT_CORE0.CPU_INT_PRI_1
	for i := 0; i < 31; i++ {
		priReg.Set(0)
		priReg = (*volatile.Register32)(unsafe.Pointer(uintptr(unsafe.Pointer(priReg)) + uintptr(4)))
	}

	// default threshold for interrupts is 5
	esp.INTERRUPT_CORE0.CPU_INT_THRESH.Set(5)

	// Set the interrupt address.
	// Set MODE field to 1 - a vector base address (only supported by ESP32C3)
	// Note that this address must be aligned to 256 bytes.
	riscv.MTVEC.Set((uintptr(unsafe.Pointer(&_vector_table))) | 1)

	riscv.EnableInterrupts(mie)
}

func (i Interrupt) Enable() error {
	mask := riscv.DisableInterrupts()
	defer riscv.EnableInterrupts(mask)

	intr := getInterruptForHandlerID(HandlerID(i.num))

	// map mapRegister to intr
	if err := HandlerID(i.num).set(intr); err != nil {
		return err
	}

	esp.INTERRUPT_CORE0.CPU_INT_ENABLE.SetBits(1 << intr)
	esp.INTERRUPT_CORE0.CPU_INT_TYPE.SetBits(1 << intr)
	// Set threshold to 5
	reg := (*volatile.Register32)(unsafe.Pointer((uintptr(unsafe.Pointer(&esp.INTERRUPT_CORE0.CPU_INT_PRI_0)) + uintptr(intr)*4)))
	reg.Set(5)

	esp.INTERRUPT_CORE0.CPU_INT_CLEAR.SetBits(1 << intr)
	esp.INTERRUPT_CORE0.CPU_INT_CLEAR.ClearBits(1 << intr)

	riscv.Asm("fence")
	return nil
}

// getHandlerIDsForInterrupt return registered codes for interrupt
func getHandlerIDsForInterrupt(interrupt int) []HandlerID {
	switch interrupt {
	case 1:
		return []HandlerID{GPIO}
	case 2:
		return []HandlerID{UART0}
	case 3:
		return []HandlerID{UART1}
	default:
		return []HandlerID{
			USB,
			I2C0,
			I2S1,
			TIMG0,
			TIMG1,
			UHCI0,
			RMT,
			SPI1,
			SPI2,
			TWAI,
			RNG,
			WIFI,
			BT,
			WIFI_BT_COMMON,
			BT_BASEBAND,
			BT_LC,
			RSA,
			AES,
			SHA,
			HMAC,
			DS,
			GDMA,
			SYSTIMER,
			SARADC,
		}
	}
}

// getInterruptForHandlerID return registered codes for interrupt
func getInterruptForHandlerID(ID HandlerID) int {
	switch ID {
	case GPIO:
		return 1
	case UART0:
		return 2
	case UART1:
		return 3
	default:
		return 4
	}
}

func (h HandlerID) set(interrupt int) error {
	switch h {
	case UART0:
		esp.INTERRUPT_CORE0.UART_INTR_MAP.Set(uint32(interrupt))
	case UART1:
		esp.INTERRUPT_CORE0.UART1_INTR_MAP.Set(uint32(interrupt))
	case GPIO:
		esp.INTERRUPT_CORE0.GPIO_INTERRUPT_PRO_MAP.Set(uint32(interrupt))
	default:
		return ErrInterruptNotImplemented
	}
	return nil
}

//go:extern _vector_table
var _vector_table [0]uintptr

//export handleInterrupt
func handleInterrupt() {
	cause := riscv.MCAUSE.Get()
	interruptNumber := uint32(cause & 0x1f)

	if cause&(1<<31) != 0 && interruptNumber > 0 {
		// disable this interrupt
		interruptBit := uint32(1 << interruptNumber)
		esp.INTERRUPT_CORE0.CPU_INT_ENABLE.ClearBits(interruptBit)

		// reset this interrupt
		if esp.INTERRUPT_CORE0.CPU_INT_TYPE.Get()&(interruptBit) != 0 {
			// this is edge type interrupt
			esp.INTERRUPT_CORE0.CPU_INT_CLEAR.SetBits(interruptBit)
			esp.INTERRUPT_CORE0.CPU_INT_CLEAR.ClearBits(interruptBit)
		} else {
			// this is level type interrupt
			esp.INTERRUPT_CORE0.CPU_INT_CLEAR.ClearBits(interruptBit)
		}

		// enable CPU interrupts
		riscv.MSTATUS.SetBits(0x8)

		// Call the interrupt handler, if any is registered for this code.
		for _, id := range getHandlerIDsForInterrupt(int(interruptNumber)) {
			callInterruptHandler(int(id))
		}

		// disable CPU interrupts
		riscv.MSTATUS.ClearBits(0x8)

		// enable this interrupt
		esp.INTERRUPT_CORE0.CPU_INT_ENABLE.SetBits(interruptBit)

		// enable CPU interrupts
		riscv.MSTATUS.SetBits(0x8)

	} else {
		// Topmost bit is clear, so it is an exception of some sort.
		// We could implement support for unsupported instructions here (such as
		// misaligned loads). However, for now we'll just print a fatal error.
		handleException(interruptNumber)
	}
}

//export handleException
func handleException(code uint32) {
	println("*** Exception: code:", code)
	for {
		riscv.Asm("wfi")
	}
}

// callInterruptHandler is a compiler-generated function that calls the
// appropriate interrupt handler for the given interrupt ID.
//go:linkname callInterruptHandler runtime.callInterruptHandler
func callInterruptHandler(id int)
