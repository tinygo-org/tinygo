.section .text.tinygo_scanCurrentStack
.global  tinygo_scanCurrentStack
.type    tinygo_scanCurrentStack, %function
tinygo_scanCurrentStack:
   // Push callee-saved registers onto the stack.
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

   // Scan the stack.
   mv a0, sp
   call tinygo_scanstack

   // Restore stack state.
   addi sp, sp, 64

   // Return to the caller.
   ret
