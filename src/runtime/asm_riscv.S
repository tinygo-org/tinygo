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

.section .text.tinygo_scanCurrentStack
.global  tinygo_scanCurrentStack
.type    tinygo_scanCurrentStack, %function
tinygo_scanCurrentStack:
   // Push callee-saved registers onto the stack.
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

   // Scan the stack.
   mv a0, sp
   call tinygo_scanstack

   // Restore return address.
   LREG ra, 0(sp)

   // Restore stack state.
   addi sp, sp, NREG*REGSIZE

   // Return to the caller.
   ret


.section .text.tinygo_longjmp
.global tinygo_longjmp
tinygo_longjmp:
    // Note: the code we jump to assumes a0 is non-zero, which is already the
    // case because that's the defer frame pointer.
    LREG sp, 0(a0)       // jumpSP
    LREG a1, REGSIZE(a0) // jumpPC
    jr a1

.section .text.tinygo_checkpointJump
.global tinygo_checkpointJump
tinygo_checkpointJump:
    // Note: the code we jump to assumes a0 is non-zero, which is already the
    // case because that's the stack pointer.
    mv sp, a0          // jumpSP
    csrw mepc, a1      // update MEPC value, so we resume there after the mret
    mret               // jump to jumpPC
