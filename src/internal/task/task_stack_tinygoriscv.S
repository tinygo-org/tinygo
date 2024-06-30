//go:build tinygo

// We don't support RV32E at the moment.
#if !defined(__riscv_32e)

.section .text.tinygo_startTask
.global  tinygo_startTask
.type    tinygo_startTask, %function
tinygo_startTask:
    // Small assembly stub for starting a goroutine. This is already run on the
    // new stack, with the callee-saved registers already loaded.
    // Most importantly, s0 contains the pc of the to-be-started function and s1
    // contains the only argument it is given. Multiple arguments are packed
    // into one by storing them in a new allocation.

    // Set the first argument of the goroutine start wrapper, which contains all
    // the arguments.
    mv    a0, s1

    // Branch to the "goroutine start" function. Use jalr to write the return
    // address to ra so we'll return here after the goroutine exits.
    jalr  s0

    // After return, exit this goroutine. This is a tail call.
    tail  tinygo_pause

.section .text.tinygo_swapTask
.global  tinygo_swapTask
.type    tinygo_swapTask, %function
tinygo_swapTask:
    // This function gets the following parameters:
    //   a0 = newStack uintptr
    //   a1 = oldStack *uintptr

    // Push all callee-saved registers.
    addi sp, sp, -52
    sw ra,  48(sp)
    sw s11, 44(sp)
    sw s10, 40(sp)
    sw s9,  36(sp)
    sw s8,  32(sp)
    sw s7,  28(sp)
    sw s6,  24(sp)
    sw s5,  20(sp)
    sw s4,  16(sp)
    sw s3,  12(sp)
    sw s2,   8(sp)
    sw s1,   4(sp)
    sw s0,    (sp)

    // Save the current stack pointer in oldStack.
    sw sp,  0(a1)

    // Switch to the new stack pointer.
    mv sp,  a0

    // Pop all saved registers from this new stack.
    lw ra,  48(sp)
    lw s11, 44(sp)
    lw s10, 40(sp)
    lw s9,  36(sp)
    lw s8,  32(sp)
    lw s7,  28(sp)
    lw s6,  24(sp)
    lw s5,  20(sp)
    lw s4,  16(sp)
    lw s3,  12(sp)
    lw s2,   8(sp)
    lw s1,   4(sp)
    lw s0,    (sp)
    addi sp, sp, 52

    // Return into the task.
    ret

#endif
