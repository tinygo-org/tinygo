.section .text.tinygo_scanCurrentStack
.global tinygo_scanCurrentStack
.type tinygo_scanCurrentStack, %function
tinygo_scanCurrentStack:
    // Save callee-saved registers.
    push r29 // Y
    push r28 // Y
    push r17
    push r16
    push r15
    push r14
    push r13
    push r12
    push r11
    push r10
    push r9
    push r8
    push r7
    push r6
    push r5
    push r4
    push r3
    push r2

    // Scan the stack.
    in  r24, 0x3d; SPL
    in  r25, 0x3e; SPH
#if __AVR_ARCH__ == 2 || __AVR_ARCH__ == 25
    rcall tinygo_scanstack
#else
    call tinygo_scanstack
#endif

    // Restore callee-saved registers.
    pop r2
    pop r3
    pop r4
    pop r5
    pop r6
    pop r7
    pop r8
    pop r9
    pop r10
    pop r11
    pop r12
    pop r13
    pop r14
    pop r15
    pop r16
    pop r17
    pop r28 // Y
    pop r29 // Y


.section .text.tinygo_longjmp
.global tinygo_longjmp
tinygo_longjmp:
    ; Move the *DeferFrame pointer to the X register.
    mov r26, r24
    mov r27, r25

    ; Load the stack pointer
    ld r24, X+
    ld r25, X+

    ; Switch to the given stack pointer.
    in r0, 0x3f   ; SREG
    cli           ; disable interrupts
    out 0x3d, r24 ; SPL
    out 0x3f, r0  ; re-enable interrupts (with one instruction delay)
    out 0x3e, r25 ; SPH

    ; Load the new PC
    ld r30, X+
    ld r31, X+

    ; Load the new Y register
    ld r28, X+
    ld r29, X+

    ; Jump to the PC (stored in the Z register)
    icall
