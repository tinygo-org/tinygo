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

//export handleInterrupt
func handleInterrupt() {
	cause := riscv.MCAUSE.Get()
	code := uint32(cause & 0xf)
	println("INTR: cause:", cause, "code:", code)
	// device.Asm("ebreak")
	if cause&(1<<31) == 0 {
		handleException(code)
		return
	}

	println("INT_RAW:", esp.UART1.INT_RAW.Get(), "", esp.UART1.INT_ST.Get())

	// Topmost bit is set, which means that it is an interrupt.
	// switch code {
	// case 7: // Machine timer interrupt
	// 	// Signal timeout.
	// 	timerWakeup.Set(1)
	// 	// Disable the timer, to avoid triggering the interrupt right after
	// 	// this interrupt returns.
	// 	riscv.MIE.ClearBits(1 << 7) // MTIE bit
	// case 11: // Machine external interrupt
	// 	hartId := riscv.MHARTID.Get()

	// 	// Claim this interrupt.
	// 	id := kendryte.PLIC.TARGETS[hartId].CLAIM.Get()
	// 	// Call the interrupt handler, if any is registered for this ID.
	// 	callInterruptHandler(int(id))
	// 	// Complete this interrupt.
	// 	kendryte.PLIC.TARGETS[hartId].CLAIM.Set(id)
	// }
}

//export handleException
func handleException(code uint32) {
	println("*** Exception: code:", code)
	// device.Asm("ebreak")
	for {
		riscv.Asm("wfi")
	}
}

type interruptFunc func()

var interruptMap = [32]interruptFunc{}

func AddHandler(intr int, mapRegister *volatile.Register32, h interruptFunc) (err error) {
	println("AddHandler(", intr, ")")
	if intr == 0 {
		return errors.New("Interrupt 0 not allowed")
	}

	mask := riscv.DisableInterrupts()
	defer func() {
		riscv.EnableInterrupts(mask)
		println("AddHandler(", intr, ") done:", err)
	}()

	if interruptMap[intr] != nil {
		err = errors.New("Interrupt already in use")
		return err
	}
	interruptMap[intr] = h

	esp.INTERRUPT_CORE0.CPU_INT_ENABLE.SetBits(1 << intr)
	esp.INTERRUPT_CORE0.CPU_INT_TYPE.SetBits(1 << intr)
	// Set threshold to 5
	priReg := &esp.INTERRUPT_CORE0.CPU_INT_PRI_0
	addr := uintptr(unsafe.Pointer(priReg)) + uintptr(4*intr)
	priReg = (*volatile.Register32)(unsafe.Pointer(addr))
	priReg.Set(5)
	// map mapRegister to intr
	mapRegister.Set(uint32(intr))

	esp.INTERRUPT_CORE0.CPU_INT_CLEAR.SetBits(1 << intr)
	esp.INTERRUPT_CORE0.CPU_INT_CLEAR.ClearBits(1 << intr)

	riscv.Asm("fence")

	return nil
}
