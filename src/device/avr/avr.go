package avr

// Magic type recognozed by the compiler to mark this string as inline assembly.
type __asm string

// Run the given assembly code. The code will be marked as having side effects,
// as it doesn't produce output and thus would normally be eliminated by the
// optimizer.
func Asm(asm __asm)
