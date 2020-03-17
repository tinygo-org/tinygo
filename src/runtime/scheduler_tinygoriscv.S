#if __riscv_xlen==64
#define REGSIZE 8
#define SREG sd
#define LREG ld
#else
#define REGSIZE 4
#define SREG sw
#define LREG lw
#endif

.section .text.tinygo_scanCurrentStack
.global  tinygo_scanCurrentStack
.type    tinygo_scanCurrentStack, %function
tinygo_scanCurrentStack:
   // Push callee-saved registers onto the stack.
   addi sp, sp, -13*REGSIZE
   SREG ra,  0*REGSIZE(sp)
   SREG s11, 1*REGSIZE(sp)
   SREG s10, 2*REGSIZE(sp)
   SREG s9,  3*REGSIZE(sp)
   SREG s8,  4*REGSIZE(sp)
   SREG s7,  5*REGSIZE(sp)
   SREG s6,  6*REGSIZE(sp)
   SREG s5,  7*REGSIZE(sp)
   SREG s4,  8*REGSIZE(sp)
   SREG s3,  9*REGSIZE(sp)
   SREG s2,  10*REGSIZE(sp)
   SREG s1,  11*REGSIZE(sp)
   SREG s0,  12*REGSIZE(sp)

   // Scan the stack.
   mv a0, sp
   call tinygo_scanstack

   // Restore return address.
   LREG ra, 0(sp)

   // Restore stack state.
   addi sp, sp, 13*REGSIZE

   // Return to the caller.
   ret

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

.section .text.tinygo_getSystemStackPointer
.global  tinygo_getSystemStackPointer
.type    tinygo_getSystemStackPointer, %function
tinygo_getSystemStackPointer:
    // The thread pointer (tp) is reserved by the ABI but is otherwise unused in
    // TinyGo. Use it here to store the system stack pointer (equivalent of MSP
    // in ARM). When it's zero we're already on the main stack pointer,
    // otherwise return the main stack pointer from tp.
    // Pseudocode:
    //     if tp:
    //         return tp
    //     return sp
    mv a0, sp
    beq tp, zero, 1f
    mv a0, tp
1:
    ret

.section .text.tinygo_switchToTask
.global  tinygo_switchToTask
.type    tinygo_switchToTask, %function
tinygo_switchToTask:
    // a0 = sp uintptr

    // Push all the registers.
    addi sp, sp, -64
    sw ra,  60(sp)
    sw s11, 56(sp)
    sw s10, 52(sp)
    sw s9,  48(sp)
    sw s8,  44(sp)
    sw s7,  40(sp)
    sw s6,  36(sp)
    sw s5,  32(sp)
    sw s4,  28(sp)
    sw s3,  24(sp)
    sw s2,  20(sp)
    sw s1,  16(sp)
    sw s0,  12(sp)

    // Store the old stack pointer in tp (thread pointer, otherwise unused in
    // TinyGo).
    mv tp,  sp

    // Load the new stack pointer from a0 (newTask) and switch to it.
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
    addi sp, sp, 48

    // Return into the task.
    ret

.section .text.tinygo_switchToScheduler
.global  tinygo_switchToScheduler
.type    tinygo_switchToScheduler, %function
tinygo_switchToScheduler:
    // a0 = sp *uintptr

    // Currently on the task stack. Push all callee-saved registers on this
    // stack.
    addi sp, sp, -48
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

    // Store the current stack pointer into a0 (&task.sp).
    sw sp,    (a0)

    // Switch to the system stack pointer, which was saved in tp.
    // Also, clear the tp to make sure tinygo_getSystemStackPointer will work
    // correctly.
    mv sp, tp
    mv tp, zero

    // Pop all registers from the system stack.
    lw ra,  60(sp)
    lw s11, 56(sp)
    lw s10, 52(sp)
    lw s9,  48(sp)
    lw s8,  44(sp)
    lw s7,  40(sp)
    lw s6,  36(sp)
    lw s5,  32(sp)
    lw s4,  28(sp)
    lw s3,  24(sp)
    lw s2,  20(sp)
    lw s1,  16(sp)
    lw s0,  12(sp)
    addi sp, sp, 64

    // Return into the scheduler.
    ret

