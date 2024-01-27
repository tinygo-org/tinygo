//go:build tinygo

.section .text.tinygo_startTask,"ax",@progbits
.global  tinygo_startTask
.type    tinygo_startTask, %function
tinygo_startTask:
    // Small assembly stub for starting a goroutine. This is already run on the
    // new stack, with the callee-saved registers already loaded.
    // Most importantly, r4 contains the pc of the to-be-started function and r5
    // contains the only argument it is given. Multiple arguments are packed
    // into one by storing them in a new allocation.

    // Set the first argument of the goroutine start wrapper, which contains all
    // the arguments.
    mov.n  a2, a13

    // Branch to the "goroutine start" function.
    callx0 a12

    // After return, exit this goroutine. This is a tail call.
    call0  tinygo_pause
.size tinygo_startTask, .-tinygo_startTask

.global tinygo_swapTask
.type tinygo_swapTask, %function
tinygo_swapTask:
    // This function gets the following parameters:
    //   a2 = newStack uintptr
    //   a3 = oldStack *uintptr
    // Note:
    //   a0 is the return address
    //   a1 is the stack pointer (sp)

    // Save all callee-saved registers:
    addi    sp,  sp, -20
    s32i.n  a12, sp, 0
    s32i.n  a13, sp, 4
    s32i.n  a14, sp, 8
    s32i.n  a15, sp, 12
    s32i.n  a0,  sp, 16

    // Save the current stack pointer in oldStack.
    s32i.n  sp, a3, 0

    // Switch to the new stack pointer.
    mov.n   sp, a2

    // Load state from new task and branch to the previous position in the
    // program.
    l32i.n  a12, sp, 0
    l32i.n  a13, sp, 4
    l32i.n  a14, sp, 8
    l32i.n  a15, sp, 12
    l32i.n  a0,  sp, 16
    addi    sp,  sp, 20
    ret.n
