#ifdef __MACH__ // Darwin
.global  _tinygo_startTask
_tinygo_startTask:
#else // Linux etc
.section .text.tinygo_startTask
.global  tinygo_startTask
tinygo_startTask:
#endif
    .cfi_startproc
    // Small assembly stub for starting a goroutine. This is already run on the
    // new stack, with the callee-saved registers already loaded.
    // Most importantly, r12 contain the pc of the to-be-started function and
    // r13 contain the only argument it is given. Multiple arguments are packed
    // into one by storing them in a new allocation.

    // Indicate to the unwinder that there is nothing to unwind, this is the
    // root frame. It avoids bogus extra frames in GDB like here:
    //     #10 0x00000000004277b6 in <goroutine wrapper> () at [...]
    //     #11 0x00000000004278f3 in tinygo_startTask () at [...]
    //     #12 0x0000000000002030 in ?? ()
    //     #13 0x0000000000000071 in ?? ()
    .cfi_undefined rip

    // Set the first argument of the goroutine start wrapper, which contains all
    // the arguments.
    movq %r13, %rdi

    // Branch to the "goroutine start" function.
    callq *%r12

    // After return, exit this goroutine. This is a tail call.
    #ifdef __MACH__
    jmp _tinygo_pause
    #else
    jmp tinygo_pause
    #endif
    .cfi_endproc

#ifdef __MACH__ // Darwin
.global _tinygo_swapTask
_tinygo_swapTask:
#else // Linux etc
.global tinygo_swapTask
.section .text.tinygo_swapTask
tinygo_swapTask:
#endif
    // This function gets the following parameters:
    // %rdi = newStack uintptr
    // %rsi = oldStack *uintptr

    // Save all callee-saved registers:
    pushq %r15
    pushq %r14
    pushq %r13
    pushq %r12
    pushq %rbp
    pushq %rbx

    // Save the current stack pointer in oldStack.
    movq %rsp, (%rsi)

    // Switch to the new stack pointer.
    movq %rdi, %rsp

    // Load saved register from the new stack.
    popq %rbx
    popq %rbp
    popq %r12
    popq %r13
    popq %r14
    popq %r15

    // Return into the new task, as if tinygo_swapTask was a regular call.
    ret

#ifdef __MACH__ // Darwin
.global  _tinygo_handoverTask
_tinygo_handoverTask:
#else // Linux etc
.section .text.tinygo_handoverTask
.global  tinygo_handoverTask
tinygo_handoverTask:
#endif
    .cfi_startproc
    movq %rdi, %rax
    movq %rsi, %rdi
    jmp  *%rax
    .cfi_endproc

#ifdef __MACH__ // Darwin
.global  _tinygo_callMachineEntry
_tinygo_callMachineEntry:
#else // Linux etc
.section .text.tinygo_callMachineEntry
.global  tinygo_callMachineEntry
tinygo_callMachineEntry:
#endif
    .cfi_startproc
    jmp tinygo_machineEntry
    .cfi_endproc

#ifdef __MACH__ // Darwin
// allow these symbols to stripped as dead code
.subsections_via_symbols
#endif
