//go:build tinygo

#if __riscv_xlen == 64
#define REGSIZE 8
#define SREG sd
#define LREG ld
#else
#define REGSIZE 4
#define SREG sw
#define LREG lw
#endif

#if defined(__riscv_flen) && __riscv_flen == 32
// fs0-fs11 and fcsr follow the 13 integer registers.
#define SAVE_FREGS 1
#define FRAMESIZE 26
#else
#define FRAMESIZE 13
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
    addi sp, sp, -FRAMESIZE*REGSIZE
    SREG ra,  12*REGSIZE(sp)
    SREG s11, 11*REGSIZE(sp)
    SREG s10, 10*REGSIZE(sp)
    SREG s9,   9*REGSIZE(sp)
    SREG s8,   8*REGSIZE(sp)
    SREG s7,   7*REGSIZE(sp)
    SREG s6,   6*REGSIZE(sp)
    SREG s5,   5*REGSIZE(sp)
    SREG s4,   4*REGSIZE(sp)
    SREG s3,   3*REGSIZE(sp)
    SREG s2,   2*REGSIZE(sp)
    SREG s1,   1*REGSIZE(sp)
    SREG s0,   0*REGSIZE(sp)
#ifdef SAVE_FREGS
    fsw  fs0,  13*REGSIZE(sp)
    fsw  fs1,  14*REGSIZE(sp)
    fsw  fs2,  15*REGSIZE(sp)
    fsw  fs3,  16*REGSIZE(sp)
    fsw  fs4,  17*REGSIZE(sp)
    fsw  fs5,  18*REGSIZE(sp)
    fsw  fs6,  19*REGSIZE(sp)
    fsw  fs7,  20*REGSIZE(sp)
    fsw  fs8,  21*REGSIZE(sp)
    fsw  fs9,  22*REGSIZE(sp)
    fsw  fs10, 23*REGSIZE(sp)
    fsw  fs11, 24*REGSIZE(sp)
    frcsr t1
    SREG t1,   25*REGSIZE(sp)
#endif

    // Save the current stack pointer in oldStack.
    SREG sp,  0(a1)

    // Switch to the new stack pointer.
    mv sp,  a0

    // Clear the saved system stack only after switching back to it.
    beqz a2, 1f
    SREG zero, 0(a2)
1:

    // Pop all saved registers from this new stack.
    LREG ra,  12*REGSIZE(sp)
    LREG s11, 11*REGSIZE(sp)
    LREG s10, 10*REGSIZE(sp)
    LREG s9,   9*REGSIZE(sp)
    LREG s8,   8*REGSIZE(sp)
    LREG s7,   7*REGSIZE(sp)
    LREG s6,   6*REGSIZE(sp)
    LREG s5,   5*REGSIZE(sp)
    LREG s4,   4*REGSIZE(sp)
    LREG s3,   3*REGSIZE(sp)
    LREG s2,   2*REGSIZE(sp)
    LREG s1,   1*REGSIZE(sp)
    LREG s0,   0*REGSIZE(sp)
#ifdef SAVE_FREGS
    flw  fs0,  13*REGSIZE(sp)
    flw  fs1,  14*REGSIZE(sp)
    flw  fs2,  15*REGSIZE(sp)
    flw  fs3,  16*REGSIZE(sp)
    flw  fs4,  17*REGSIZE(sp)
    flw  fs5,  18*REGSIZE(sp)
    flw  fs6,  19*REGSIZE(sp)
    flw  fs7,  20*REGSIZE(sp)
    flw  fs8,  21*REGSIZE(sp)
    flw  fs9,  22*REGSIZE(sp)
    flw  fs10, 23*REGSIZE(sp)
    flw  fs11, 24*REGSIZE(sp)
    LREG t1,   25*REGSIZE(sp)
    fscsr t1
#endif
    addi sp, sp, FRAMESIZE*REGSIZE

    // Restore the interrupt-enable state.
    andi t0, t0, 8
    beqz t0, 2f
    csrsi mstatus, 8
2:

    // Return into the task.
    ret
