.section .text.tinygo_scanCurrentStack
.global  tinygo_scanCurrentStack
.type    tinygo_scanCurrentStack, %function
tinygo_scanCurrentStack:
    // Save callee-saved registers onto the stack.
    mov r0, r8
    mov r1, r9
    mov r2, r10
    mov r3, r11
    push {r0-r3, lr}
    push {r4-r7}

    // Scan the stack.
    mov r0, sp
    bl tinygo_scanstack

    // Restore stack state and return.
    add sp, #32
    pop {pc}
