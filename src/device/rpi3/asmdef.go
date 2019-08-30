// +build rpi3

package rpi3

// Try to keep go vet from complaining about empty function bodies
import _ "unsafe"

// Asm runs the given assembly code. The code will be marked as having side effects,
// as it doesn't produce output and thus would normally be eliminated by the
// optimizer.
func Asm(asm string)

// AsmFull runs the given inline assembly. The code will be marked as having side
// effects, as it would otherwise be optimized away. The inline assembly string
// recognizes template values in the form {name}, like so:
//
//     avr.AsmFull(
//         "str {value}, {result}",
//         map[string]interface{}{
//             "value":  1
//             "result": &dest,
//         })
func AsmFull(asm string, regs map[string]interface{})
