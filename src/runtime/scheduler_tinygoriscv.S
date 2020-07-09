#ifdef __riscv_flen
#define NREG 25
#else
#define NREG 13
#endif

#if __riscv_xlen==64
#define REGSIZE 8
#define SREG sd
#define LREG ld
#ifdef __riscv_flen
#define LFREG fld
#define SFREG fsd
#endif
#else
#define REGSIZE 4
#define SREG sw
#define LREG lw
#ifdef __riscv_flen
#define LFREG flw
#define SFREG fsw
#endif
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
    addi sp, sp, -NREG*REGSIZE
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
#ifdef __riscv_flen
    SFREG   fs11, 13*REGSIZE(sp)
    SFREG   fs10, 14*REGSIZE(sp)
    SFREG   fs9,  15*REGSIZE(sp)
    SFREG   fs8,  16*REGSIZE(sp)
    SFREG   fs7,  17*REGSIZE(sp)
    SFREG   fs6,  18*REGSIZE(sp)
    SFREG   fs5,  19*REGSIZE(sp)
    SFREG   fs4,  20*REGSIZE(sp)
    SFREG   fs3,  21*REGSIZE(sp)
    SFREG   fs2,  22*REGSIZE(sp)
    SFREG   fs1,  23*REGSIZE(sp)
    SFREG   fs0,  24*REGSIZE(sp)
#endif

    // Store the old stack pointer in tp (thread pointer, otherwise unused in
    // TinyGo).
    mv tp,  sp

    // Load the new stack pointer from a0 (newTask) and switch to it.
    mv sp,  a0

    // Pop all saved registers from this new stack.
    LREG ra,  (NREG - 1)*REGSIZE(sp)

    LREG s11, ((NREG - 1) - 1)*REGSIZE(sp)
    LREG s10, ((NREG - 1) - 2)*REGSIZE(sp)
    LREG s9,  ((NREG - 1) - 3)*REGSIZE(sp)
    LREG s8,  ((NREG - 1) - 4)*REGSIZE(sp)
    LREG s7,  ((NREG - 1) - 5)*REGSIZE(sp)
    LREG s6,  ((NREG - 1) - 6)*REGSIZE(sp)
    LREG s5,  ((NREG - 1) - 7)*REGSIZE(sp)
    LREG s4,  ((NREG - 1) - 8)*REGSIZE(sp)
    LREG s3,  ((NREG - 1) - 9)*REGSIZE(sp)
    LREG s2,  ((NREG - 1) - 10)*REGSIZE(sp)
    LREG s1,  ((NREG - 1) - 11)*REGSIZE(sp)
    LREG s0,  ((NREG - 1) - 12)*REGSIZE(sp)

#ifdef __riscv_flen
    LFREG   fs11, ((NREG - 1) - 13)*REGSIZE(sp)
    LFREG   fs10, ((NREG - 1) - 14)*REGSIZE(sp)
    LFREG   fs9,  ((NREG - 1) - 15)*REGSIZE(sp)
    LFREG   fs8,  ((NREG - 1) - 16)*REGSIZE(sp)
    LFREG   fs7,  ((NREG - 1) - 17)*REGSIZE(sp)
    LFREG   fs6,  ((NREG - 1) - 18)*REGSIZE(sp)
    LFREG   fs5,  ((NREG - 1) - 19)*REGSIZE(sp)
    LFREG   fs4,  ((NREG - 1) - 20)*REGSIZE(sp)
    LFREG   fs3,  ((NREG - 1) - 21)*REGSIZE(sp)
    LFREG   fs2,  ((NREG - 1) - 22)*REGSIZE(sp)
    LFREG   fs1,  ((NREG - 1) - 23)*REGSIZE(sp)
    LFREG   fs0,  ((NREG - 1) - 24)*REGSIZE(sp)
#endif

    addi sp, sp, (NREG - 1)*REGSIZE

    // Return into the task.
    ret

.section .text.tinygo_switchToScheduler
.global  tinygo_switchToScheduler
.type    tinygo_switchToScheduler, %function
tinygo_switchToScheduler:
    // a0 = sp *uintptr

    // Currently on the task stack. Push all callee-saved registers on this
    // stack.
    addi sp, sp, -(NREG - 1)*REGSIZE
    SREG ra, (NREG - 1)*REGSIZE(sp)
    
    SREG s11, ((NREG - 1) - 1)*REGSIZE(sp)
    SREG s10, ((NREG - 1) - 2)*REGSIZE(sp)
    SREG s9,  ((NREG - 1) - 3)*REGSIZE(sp)
    SREG s8,  ((NREG - 1) - 4)*REGSIZE(sp)
    SREG s7,  ((NREG - 1) - 5)*REGSIZE(sp)
    SREG s6,  ((NREG - 1) - 6)*REGSIZE(sp)
    SREG s5,  ((NREG - 1) - 7)*REGSIZE(sp)
    SREG s4,  ((NREG - 1) - 8)*REGSIZE(sp)
    SREG s3,  ((NREG - 1) - 9)*REGSIZE(sp)
    SREG s2,  ((NREG - 1) - 10)*REGSIZE(sp)
    SREG s1,  ((NREG - 1) - 11)*REGSIZE(sp)
    SREG s0,  ((NREG - 1) - 12)*REGSIZE(sp)


#ifdef __riscv_flen
    SFREG fs11, ((NREG - 1) - 13)*REGSIZE(sp)
    SFREG fs10, ((NREG - 1) - 14)*REGSIZE(sp)
    SFREG fs9,  ((NREG - 1) - 15)*REGSIZE(sp)
    SFREG fs8,  ((NREG - 1) - 16)*REGSIZE(sp)
    SFREG fs7,  ((NREG - 1) - 17)*REGSIZE(sp)
    SFREG fs6,  ((NREG - 1) - 18)*REGSIZE(sp)
    SFREG fs5,  ((NREG - 1) - 19)*REGSIZE(sp)
    SFREG fs4,  ((NREG - 1) - 20)*REGSIZE(sp)
    SFREG fs3,  ((NREG - 1) - 21)*REGSIZE(sp)
    SFREG fs2,  ((NREG - 1) - 22)*REGSIZE(sp)
    SFREG fs1,  ((NREG - 1) - 23)*REGSIZE(sp)
    SFREG fs0,  ((NREG - 1) - 24)*REGSIZE(sp)
#endif

    // Store the current stack pointer into a0 (&task.sp).
    SREG sp,    (a0)

    // Switch to the system stack pointer, which was saved in tp.
    // Also, clear the tp to make sure tinygo_getSystemStackPointer will work
    // correctly.
    mv sp, tp
    mv tp, zero

    // Pop all registers from the system stack.
    LREG ra,  0*REGSIZE(sp)
    LREG s11, 1*REGSIZE(sp)
    LREG s10, 2*REGSIZE(sp)
    LREG s9,  3*REGSIZE(sp)
    LREG s8,  4*REGSIZE(sp)
    LREG s7,  5*REGSIZE(sp)
    LREG s6,  6*REGSIZE(sp)
    LREG s5,  7*REGSIZE(sp)
    LREG s4,  8*REGSIZE(sp)
    LREG s3,  9*REGSIZE(sp)
    LREG s2,  10*REGSIZE(sp)
    LREG s1,  11*REGSIZE(sp)
    LREG s0,  12*REGSIZE(sp)
#ifdef __riscv_flen
    LFREG fs11, 13*REGSIZE(sp)
    LFREG fs10, 14*REGSIZE(sp)
    LFREG fs9,  15*REGSIZE(sp)
    LFREG fs8,  16*REGSIZE(sp)
    LFREG fs7,  17*REGSIZE(sp)
    LFREG fs6,  18*REGSIZE(sp)
    LFREG fs5,  19*REGSIZE(sp)
    LFREG fs4,  20*REGSIZE(sp)
    LFREG fs3,  21*REGSIZE(sp)
    LFREG fs2,  22*REGSIZE(sp)
    LFREG fs1,  23*REGSIZE(sp)
    LFREG fs0,  24*REGSIZE(sp)
#endif
    addi sp, sp, NREG*REGSIZE

    // Return into the scheduler.
    ret

