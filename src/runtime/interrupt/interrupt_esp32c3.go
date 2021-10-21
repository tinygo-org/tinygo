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

func (i Interrupt) Enable() error {
	mask := riscv.DisableInterrupts()
	defer riscv.EnableInterrupts(mask)

	// map mapRegister to CPU interrupt
	intr, err := HandlerID(i.num).set()
	if err != nil {
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

var (
	interruptMap = map[int][]HandlerID{
		1: {GPIO},
		2: {UART0},
		3: {UART1},
	}
)

// set will update and return module's CPU interrupt
// IMPORTANT: must match interruptMap
func (h HandlerID) set() (int, error) {
	switch h {
	case GPIO:
		esp.INTERRUPT_CORE0.GPIO_INTERRUPT_PRO_MAP.Set(1)
		return 1, nil
	case UART0:
		esp.INTERRUPT_CORE0.UART_INTR_MAP.Set(2)
		return 2, nil
	case UART1:
		esp.INTERRUPT_CORE0.UART1_INTR_MAP.Set(3)
		return 3, nil
	default:
		return -1, ErrInterruptNotImplemented
	}
}

//export handleInterrupt
func handleInterrupt() {
	mcause := riscv.MCAUSE.Get()
	exception := mcause&(1<<31) == 0
	interruptNumber := uint32(mcause & 0x1f)

	if !exception && interruptNumber > 0 {
		// save mepc, which could be overwritten by another CPU interrupt
		mepc := riscv.MEPC.Get()

		// disable interrupt
		interruptBit := uint32(1 << interruptNumber)
		esp.INTERRUPT_CORE0.CPU_INT_ENABLE.ClearBits(interruptBit)

		// reset pending status interrupt
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

		// Call registered interrupt handler(s)
		if vector, ok := interruptMap[int(interruptNumber)]; ok {
			for _, handlerID := range vector {
				callInterruptHandler(int(handlerID))
			}
		} else {
			println("unhandled CPU interrupt", interruptNumber)
		}

		// disable CPU interrupts
		riscv.MSTATUS.ClearBits(0x8)

		// mpie must be set to 1 to resume interrupts after 'MRET'
		riscv.MSTATUS.SetBits(0x80)

		// restore MEPC
		riscv.MEPC.Set(mepc)

		// enable this interrupt
		esp.INTERRUPT_CORE0.CPU_INT_ENABLE.SetBits(interruptBit)

		// do not enable CPU interrupts now
		// the 'MRET' in src/device/riscv/handleinterrupt.S will copies the state of MPIE back into MIE, and subsequently clears MPIE.
		// riscv.MSTATUS.SetBits(0x8)
	} else {
		// Topmost bit is clear, so it is an exception of some sort.
		// We could implement support for unsupported instructions here (such as
		// misaligned loads). However, for now we'll just print a fatal error.
		handleException(mcause)
	}
}

func handleException(mcause uintptr) {
	// TODO need to get location of actual MEPC from the stack stash created in src/device/riscv/handleinterrupt.S
	println("*** Exception:     pc:", riscv.MEPC.Get())
	println("*** Exception:   code:", uint32(mcause&0x1f))
	println("*** Exception: mcause:", mcause)
	for {
		riscv.Asm("wfi")
	}
}

// callInterruptHandler is a compiler-generated function that calls the
// appropriate interrupt handler for the given interrupt ID.
//go:linkname callInterruptHandler runtime.callInterruptHandler
func callInterruptHandler(id int)
