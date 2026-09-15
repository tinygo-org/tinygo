//go:build tinygo

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
    tail  tinygo_task_exit

.section .text.tinygo_swapTaskRISC
.global  tinygo_swapTaskRISC
.type    tinygo_swapTaskRISC, %function
tinygo_swapTaskRISC:
.global  tinygo_swapTask
.type    tinygo_swapTask, %function
.set     tinygo_swapTask, tinygo_swapTaskRISC

    // This function gets the following parameters:
    //   a0 = newStack uintptr
    //   a1 = oldStack *uintptr
    //   a2 = stack pointer to clear after switching, or nil

    // Callers enter with interrupts enabled. Keep the prior hart state in t0;
    // it is not part of the task context.
    csrrci t0, mstatus, 8

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

    // Clear the saved system stack only after switching back to it.
    beqz a2, 1f
    sw zero, 0(a2)
1:

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

    // Restore the interrupt-enable state.
    andi t0, t0, 8
    beqz t0, 2f
    csrsi mstatus, 8
2:

    // Return into the task.
    ret
