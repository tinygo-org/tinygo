#ifdef __MACH__
.global _tinygo_scanCurrentStack
_tinygo_scanCurrentStack:
#else
.section .text.tinygo_scanCurrentStack
.global tinygo_scanCurrentStack
tinygo_scanCurrentStack:
#endif
    // Sources:
    //   * https://developer.arm.com/architectures/learn-the-architecture/armv8-a-instruction-set-architecture/procedure-call-standard
    //   * https://godbolt.org/z/qrvrEh

    // Save callee-saved registers.
    stp     x29, x30, [sp, #-160]!
    stp     x28, x27, [sp, #16]
    stp     x26, x25, [sp, #32]
    stp     x24, x23, [sp, #48]
    stp     x22, x21, [sp, #64]
    stp     x20, x19, [sp, #80]
    stp     d8,  d9,  [sp, #96]
    stp     d10, d11, [sp, #112]
    stp     d12, d13, [sp, #128]
    stp     d14, d15, [sp, #144]

    // Scan the stack.
    mov     x0, sp
#ifdef __MACH__
    bl      _tinygo_scanstack
#else
    bl      tinygo_scanstack
#endif

    // Restore stack state and return.
    ldp     x29, x30, [sp], #160
    ret


#ifdef __MACH__
.global _tinygo_longjmp
_tinygo_longjmp:
#else
.section .text.tinygo_longjmp
.global tinygo_longjmp
tinygo_longjmp:
#endif
    // Note: the code we jump to assumes x0 is set to a non-zero value if we
    // jump from here (which is conveniently already the case).
    ldp x1, x2, [x0] // jumpSP, jumpPC
    mov sp, x1
    br  x2
