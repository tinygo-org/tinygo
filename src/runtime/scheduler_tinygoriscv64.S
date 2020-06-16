.section .text.tinygo_scanCurrentStack
.global  tinygo_scanCurrentStack
.type    tinygo_scanCurrentStack, %function
tinygo_scanCurrentStack:
   // Push callee-saved registers onto the stack.
   addi sp, sp, -104
   sd ra,  96(sp)
   sd s11, 88(sp)
   sd s10, 80(sp)
   sd s9,  72(sp)
   sd s8,  64(sp)
   sd s7,  56(sp)
   sd s6,  48(sp)
   sd s5,  40(sp)
   sd s4,  32(sp)
   sd s3,  24(sp)
   sd s2,  16(sp)
   sd s1,  8(sp)
   sd s0,  0(sp)

   // Scan the stack.
   mv a0, sp
   call tinygo_scanstack

   // Restore return address.
   ld ra, 96(sp)

   // Restore stack state.
   addi sp, sp, 104

   // Return to the caller.
   ret
