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


.section .text.tinygo_longjmp
.global tinygo_longjmp
tinygo_longjmp:
    // Note: the code we jump to assumes a0 is non-zero, which is already the
    // case because that's jumpSP (the stack pointer).
    mv sp, a0 // jumpSP
    jr a1     // jumpPC
