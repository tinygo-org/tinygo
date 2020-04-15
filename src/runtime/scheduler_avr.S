.section .bss.tinygo_systemStack
.global tinygo_systemStack
.type   tinygo_systemStack, %object
tinygo_systemStack:
    .short 0

.section .text.tinygo_startTask
.global  tinygo_startTask
.type    tinygo_startTask, %function
tinygo_startTask:
    // Small assembly stub for starting a goroutine. This is already run on the
    // new stack, with the callee-saved registers already loaded.
    // Most importantly, r2r3 contain the pc of the to-be-started function and
    // r4r5 contain the only argument it is given. Multiple arguments are packed
    // into one by storing them in a new allocation.

    // Set the first argument of the goroutine start wrapper, which contains all
    // the arguments.
    movw  r24, r4

    // Branch to the "goroutine start" function. Note that the Z register is
    // call-clobbered, so does not need to be restored after use.
    movw  Z, r2
    icall

    // After return, exit this goroutine. This is a tail call.
#if __AVR_ARCH__ == 2 || __AVR_ARCH__ == 25
    // Small memory devices (≤8kB flash) that do not have the long call
    // instruction availble will need to use rcall instead.
    // Note that they will probably not be able to run more than the main
    // goroutine anyway, but this file is compiled for all AVRs so it needs to
    // compile at least.
    rcall  tinygo_pause
#else
    // Other devices can (and must) use the regular call instruction.
    call tinygo_pause
#endif

// Get the system stack pointer, independent of whether we're currently on the
// system stack or a task stack.
.global tinygo_getSystemStackPointer
.type tinygo_getSystemStackPointer, %function
tinygo_getSystemStackPointer:
    // Load system stack pointer.
    lds r24, tinygo_systemStack
    lds r25, tinygo_systemStack+1

    // Compare against 0.
    cp  r24, r1
    cpc r25, r1

    // Branch (and then return) if tinygo_systemStack has a non-zero value.
    brne 1f

    // tinygo_systemStack is zero, so return the current stack pointer.
    in  r24, 0x3d; SPL
    in  r25, 0x3e; SPH

1:
    ret

.global tinygo_switchToTask
.type tinygo_switchToTask, %function
tinygo_switchToTask:
    // The sp parameter is the only parameter, so it will take up r24:r25.
    // r24:r25 = sp uintptr

    // Save all call-saved registers:
    // https://gcc.gnu.org/wiki/avr-gcc#Call-Saved_Registers
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

    // Save the system stack pointer in a global.
    in  r2, 0x3d; SPL
    in  r3, 0x3e; SPH
    sts tinygo_systemStack+0, r2
    sts tinygo_systemStack+1, r3

    // Switch to the task stack pointer.
    out  0x3d, r24; SPL
    out  0x3e, r25; SPH

    // Load saved register from the task stack.
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

    // Return into the new task, as if tinygo_switchToScheduler was a regular
    // call.
    ret

.global tinygo_switchToScheduler
.type tinygo_switchToScheduler, %function
tinygo_switchToScheduler:
    // The sp parameter is the only parameter, so it will take up r24:r25.
    // r24:r25 = sp *uintptr

    // Save all call-saved registers on the task stack:
    // https://gcc.gnu.org/wiki/avr-gcc#Call-Saved_Registers
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

    // Save the task stack.
    in  r2, 0x3d; SPL
    in  r3, 0x3e; SPH
    movw Y, r24
    std Y+0, r2
    std Y+1, r3

    // Switch to the system stack.
    lds r2, tinygo_systemStack
    lds r3, tinygo_systemStack+1
    out  0x3d, r2; SPL
    out  0x3e, r3; SPH

    // Clear tinygo_systemStack to make sure tinygo_getSystemStackPointer knows
    // which pointer to return.
    sts tinygo_systemStack+0, r1
    sts tinygo_systemStack+1, r1

    // Load saved register from the system stack.
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

    // Return into the scheduler, as if tinygo_switchToTask was a regular call.
    ret

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
