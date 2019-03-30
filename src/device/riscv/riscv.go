package riscv

// Run the given assembly code. The code will be marked as having side effects,
// as it doesn't produce output and thus would normally be eliminated by the
// optimizer.
func Asm(asm string)

// ReadRegister returns the contents of the specified register. The register
// must be a processor register, reachable with the "mov" instruction.
func ReadRegister(name string) uintptr
