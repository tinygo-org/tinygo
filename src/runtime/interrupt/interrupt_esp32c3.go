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
	INT_CODE_INSTR_ACCESS_FAULT = 0x1 // PMP Instruction access fault
	INT_CODE_ILL_INSTR          = 0x2 // Illegal Instruction
	INT_CODE_BRK                = 0x3 // Hardware Breakpoint/Watchpoint or EBREAK
	INT_CODE_LOAD_FAULT         = 0x5 // PMP Load access fault
	INT_CODE_STORE_FAULT        = 0x7 // PMP Store access fault
	INT_CODE_USR_CALL           = 0x8 // ECALL from U mode
	INT_CODE_MACH_CALL          = 0xb // ECALL from M mode
)

var (
	interruptMap              map[int][]int = make(map[int][]int, 32)
	ErrInvalidInterruptNumber               = errors.New("interrupt number is not correct")
	ErrMissingMapRegister                   = errors.New("interrupt map register is missing")
)

func (i Interrupt) Enable(intr uint32, mapRegister *volatile.Register32) error {
	if intr == 0 || intr > 31 {
		return ErrInvalidInterruptNumber
	}
	if mapRegister == nil {
		return ErrMissingMapRegister
	}
	mask := riscv.DisableInterrupts()
	defer riscv.EnableInterrupts(mask)

	RegisterCode(int(intr), i.num)

	esp.INTERRUPT_CORE0.CPU_INT_ENABLE.SetBits(1 << intr)
	esp.INTERRUPT_CORE0.CPU_INT_TYPE.SetBits(1 << intr)
	// Set threshold to 5
	reg := (*volatile.Register32)(unsafe.Pointer((uintptr(unsafe.Pointer(&esp.INTERRUPT_CORE0.CPU_INT_PRI_0)) + uintptr(intr)*4)))
	reg.Set(5)

	esp.INTERRUPT_CORE0.CPU_INT_CLEAR.SetBits(1 << intr)
	esp.INTERRUPT_CORE0.CPU_INT_CLEAR.ClearBits(1 << intr)

	// map mapRegister to intr
	mapRegister.Set(intr)

	riscv.Asm("fence")
	return nil
}

//go:extern _vector_table
var _vector_table [0]uintptr

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

// registerInterruptCode associate interrupt handler id with CPU interrut number
func RegisterCode(interrupt int, id int) {
	vector := interruptMap[interrupt]
	if vector == nil {
		interruptMap[interrupt] = []int{id}
		return
	}
	// make sure we don't have duplicates
	for _, i := range vector {
		if i == id {
			// already registered
			return
		}
	}
	interruptMap[interrupt] = append(interruptMap[interrupt], id)
}

// IDsForCode return registered codes for interrupt
func IDsForCode(interrupt int) []int {
	ret := make([]int, 0)
	if interruptMap[interrupt] != nil {
		for _, id := range interruptMap[interrupt] {
			ret = append(ret, id)
		}
	}
	return ret
}
