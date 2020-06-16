.section .text.tinygo_scanCurrentStack
.global  tinygo_scanCurrentStack
.type    tinygo_scanCurrentStack, %function
tinygo_scanCurrentStack:
   // Push callee-saved registers onto the stack.
   addi sp, sp, -128
   sd ra,  120(sp)
   sd s11, 112(sp)
   sd s10, 104(sp)
   sd s9,  96(sp)
   sd s8,  88(sp)
   sd s7,  80(sp)
   sd s6,  72(sp)
   sd s5,  64(sp)
   sd s4,  56(sp)
   sd s3,  48(sp)
   sd s2,  40(sp)
   sd s1,  32(sp)
   sd s0,  24(sp)

   // Scan the stack.
   mv a0, sp
   call tinygo_scanstack

   // Restore return address.
   ld ra, 60(sp)

   // Restore stack state.
   addi sp, sp, 128

   // Return to the caller.
   ret
