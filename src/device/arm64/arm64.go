package arm64

// Run the given assembly code. The code will be marked as having side effects,
// as it doesn't produce output and thus would normally be eliminated by the
// optimizer.
func Asm(asm string)

// Run the given inline assembly. The code will be marked as having side
// effects, as it would otherwise be optimized away. The inline assembly string
// recognizes template values in the form {name}, like so:
//
//	arm.AsmFull(
//	    "str {value}, {result}",
//	    map[string]interface{}{
//	        "value":  1
//	        "result": &dest,
//	    })
//
// You can use {} in the asm string (which expands to a register) to set the
// return value.
func AsmFull(asm string, regs map[string]interface{}) uintptr

// Run the following system call (SVCall) with 0 arguments.
func SVCall0(num uintptr) uintptr

// Run the following system call (SVCall) with 1 argument.
func SVCall1(num uintptr, a1 interface{}) uintptr

// Run the following system call (SVCall) with 2 arguments.
func SVCall2(num uintptr, a1, a2 interface{}) uintptr

// Run the following system call (SVCall) with 3 arguments.
func SVCall3(num uintptr, a1, a2, a3 interface{}) uintptr

// Run the following system call (SVCall) with 4 arguments.
func SVCall4(num uintptr, a1, a2, a3, a4 interface{}) uintptr
