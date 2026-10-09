//go:build tinygo

#ifdef __riscv_32e
#define NREG 3
#else
#define NREG 13
#endif

#if __riscv_xlen==64
#define REGSIZE 8
#define SREG sd
#define LREG ld
#else
#define REGSIZE 4
#define SREG sw
#define LREG lw
#endif

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

    // Keep the prior hart state in t0. It is not part of the task context.
    csrrci t0, mstatus, 8

    // Push all callee-saved registers.
    addi sp, sp, -NREG*REGSIZE
    SREG ra,  0*REGSIZE(sp)
    SREG s0,  1*REGSIZE(sp)
    SREG s1,  2*REGSIZE(sp)
#ifndef __riscv_32e
    SREG s2,  3*REGSIZE(sp)
    SREG s3,  4*REGSIZE(sp)
    SREG s4,  5*REGSIZE(sp)
    SREG s5,  6*REGSIZE(sp)
    SREG s6,  7*REGSIZE(sp)
    SREG s7,  8*REGSIZE(sp)
    SREG s8,  9*REGSIZE(sp)
    SREG s9,  10*REGSIZE(sp)
    SREG s10, 11*REGSIZE(sp)
    SREG s11, 12*REGSIZE(sp)
#endif

    // Save the current stack pointer in oldStack.
    sw sp,  0(a1)

    // Switch to the new stack pointer.
    mv sp,  a0

    // Clear the saved system stack only after switching back to it.
    beqz a2, 1f
    sw zero, 0(a2)
1:

    // Pop all saved registers from this new stack.
#ifndef __riscv_32e
    LREG s11, 12*REGSIZE(sp)
    LREG s10, 11*REGSIZE(sp)
    LREG s9,  10*REGSIZE(sp)
    LREG s8,  9*REGSIZE(sp)
    LREG s7,  8*REGSIZE(sp)
    LREG s6,  7*REGSIZE(sp)
    LREG s5,  6*REGSIZE(sp)
    LREG s4,  5*REGSIZE(sp)
    LREG s3,  4*REGSIZE(sp)
    LREG s2,  3*REGSIZE(sp)
#endif
    LREG s1,  2*REGSIZE(sp)
    LREG s0,  1*REGSIZE(sp)
    LREG ra,  0*REGSIZE(sp)
    addi sp, sp, NREG*REGSIZE

    // Restore the interrupt-enable state.
    andi t0, t0, 8
    beqz t0, 2f
    csrsi mstatus, 8
2:

    // Return into the task.
    ret
