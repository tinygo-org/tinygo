package x86

// ReadRegister returns the contents of the specified register. The register
// must be a processor register, reachable with the "mov" instruction.
func ReadRegister(name string) uintptr
