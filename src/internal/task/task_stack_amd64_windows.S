// Windows on amd64 has a slightly different ABI than other (*nix) systems.
// Therefore, assembly functions need to be tweaked slightly.

.section .text.tinygo_startTask,"ax"
.global  tinygo_startTask
tinygo_startTask:
    // Small assembly stub for starting a goroutine. This is already run on the
    // new stack, with the callee-saved registers already loaded.
    // Most importantly, r12 contain the pc of the to-be-started function and
    // r13 contain the only argument it is given. Multiple arguments are packed
    // into one by storing them in a new allocation.

    // Set the first argument of the goroutine start wrapper, which contains all
    // the arguments.
    movq %r13, %rcx

    // Branch to the "goroutine start" function.
    callq *%r12

    // After return, exit this goroutine.
    callq tinygo_pause

.global tinygo_swapTask
.section .text.tinygo_swapTask,"ax"
tinygo_swapTask:
    // This function gets the following parameters:
    // %rcx = newStack uintptr
    // %rdx = oldStack *uintptr

    // Save all callee-saved registers:
    pushq %r15
    pushq %r14
    pushq %r13
    pushq %r12
    pushq %rsi
    pushq %rdi
    pushq %rbp
    pushq %rbx

    // Save the current stack pointer in oldStack.
    movq %rsp, (%rdx)

    // Switch to the new stack pointer.
    movq %rcx, %rsp

    // Load saved register from the new stack.
    popq %rbx
    popq %rbp
    popq %rdi
    popq %rsi
    popq %r12
    popq %r13
    popq %r14
    popq %r15

    // Return into the new task, as if tinygo_swapTask was a regular call.
    ret
