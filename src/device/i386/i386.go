package i386

// Run the given assembly code. The code will be marked as having side effects,
// as it doesn't produce output and thus would normally be eliminated by the
// optimizer.
func Asm(asm string)

// Run the given inline assembly. The code will be marked as having side
// effects, as it would otherwise be optimized away. The inline assembly string
// recognizes template values in the form {name}, like so:
//
//     i386.AsmFull(
//         "str {value}, {result}",
//         map[string]interface{}{
//             "value":  1
//             "result": &dest,
//         })
//
// You can use {} in the asm string (which expands to a register) to set the
// return value.
func AsmFull(asm string, regs map[string]interface{}) uintptr
