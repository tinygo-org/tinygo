.section .text.tinygo_scanCurrentStack
.global tinygo_scanCurrentStack
.type tinygo_scanCurrentStack, %function
tinygo_scanCurrentStack:
    // Sources:
    //   * https://developer.arm.com/architectures/learn-the-architecture/armv8-a-instruction-set-architecture/procedure-call-standard
    //   * https://godbolt.org/z/qrvrEh

    // Save callee-saved registers.
    stp     x29, x30, [sp, #-96]!
    stp     x28, x27, [sp, #16]
    stp     x26, x25, [sp, #32]
    stp     x24, x23, [sp, #48]
    stp     x22, x21, [sp, #64]
    stp     x20, x19, [sp, #80]

    // Scan the stack.
    mov     x0, sp
    bl      tinygo_scanstack

    // Restore stack state and return.
    ldp     x29, x30, [sp], #96
    ret
