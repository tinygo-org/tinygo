.section .text.tinygo_scanCurrentStack
.global  tinygo_scanCurrentStack
.type    tinygo_scanCurrentStack, %function
tinygo_scanCurrentStack:
    // Push callee-saved registers onto the stack.
    addiu $sp, $sp, -40
    sw $ra, 36($sp)
    sw $s8, 32($sp)
    sw $s7, 28($sp)
    sw $s6, 24($sp)
    sw $s5, 20($sp)
    sw $s4, 16($sp)
    sw $s3, 12($sp)
    sw $s2,  8($sp)
    sw $s1,  4($sp)
    sw $s0,   ($sp)

    // Scan the stack.
    jal tinygo_scanstack
    move $a0, $sp // in the branch delay slot

    // Restore return address.
    lw $ra, 36($sp)

    // Restore stack state.
    addiu $sp, $sp, 40

    // Return to the caller.
    jalr $ra
    nop

.section .text.tinygo_longjmp
.global tinygo_longjmp
tinygo_longjmp:
    // Note: the code we jump to assumes a0 is non-zero, which is already the
    // case because that's the defer frame pointer.
    lw $sp, 0($a0) // jumpSP
    lw $a1, 4($a0) // jumpPC
    jr $a1
    nop
