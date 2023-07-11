package riscv

// Run the given assembly code. The code will be marked as having side effects,
// as it doesn't produce output and thus would normally be eliminated by the
// optimizer.
func Asm(asm string)

// Run the given inline assembly. The code will be marked as having side
// effects, as it would otherwise be optimized away. The inline assembly string
// recognizes template values in the form {name}, like so:
//
//	arm.AsmFull(
//	    "st {value}, {result}",
//	    map[string]interface{}{
//	        "value":  1
//	        "result": &dest,
//	    })
//
// You can use {} in the asm string (which expands to a register) to set the
// return value.
func AsmFull(asm string, regs map[string]interface{}) uintptr

// DisableInterrupts disables all interrupts, and returns the old interrupt
// state.
func DisableInterrupts() uintptr {
	// Note: this can be optimized with a CSRRW instruction, which atomically
	// swaps the value and returns the old value.
	mask := MSTATUS.Get()
	MSTATUS.ClearBits(1 << 3) // clear the MIE bit
	return mask
}

// EnableInterrupts enables all interrupts again. The value passed in must be
// the mask returned by DisableInterrupts.
func EnableInterrupts(mask uintptr) {
	mask &= 1 << 3        // clear all bits except for the MIE bit
	MSTATUS.SetBits(mask) // set the MIE bit, if it was previously cleared
}
