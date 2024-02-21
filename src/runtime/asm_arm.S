// Only generate .debug_frame, don't generate .eh_frame.
.cfi_sections .debug_frame

.section .text.tinygo_scanCurrentStack
.global  tinygo_scanCurrentStack
.type    tinygo_scanCurrentStack, %function
tinygo_scanCurrentStack:
    .cfi_startproc
    // Save callee-saved registers onto the stack.
    #if defined(__thumb2__)
        #if defined(FPU_ENABLED)
            vpush {s0-s31}              // Push the floating-point registers to the stack
            .cfi_def_cfa_offset 32*4
        #endif
        push {r4-r11, lr}
        #if defined(FPU_ENABLED)
            .cfi_def_cfa_offset 41*4
        #else
            .cfi_def_cfa_offset 9*4
        #endif
    #else
    mov r0, r8
    mov r1, r9
    mov r2, r10
    mov r3, r11
    push {r0-r3, lr}
    .cfi_def_cfa_offset 5*4
    push {r4-r7}
    .cfi_def_cfa_offset 4*4
    #endif

    // Scan the stack.
    mov r0, sp
    bl tinygo_scanstack

    // Restore stack state and return.
    add sp, #32
    .cfi_def_cfa_offset 1*4
    pop {pc}
    .cfi_endproc
.size tinygo_scanCurrentStack, .-tinygo_scanCurrentStack


.section .text.tinygo_longjmp
.global tinygo_longjmp
.type   tinygo_longjmp, %function
tinygo_longjmp:
    .cfi_startproc
    // Note: the code we jump to assumes r0 is set to a non-zero value if we
    // jump from here (which is conveniently already the case).

    ldm r0, {r0, r1}
    mov sp, r0 // jumpSP
    mov pc, r1 // jumpPC
    .cfi_endproc
.size tinygo_longjmp, .-tinygo_longjmp
