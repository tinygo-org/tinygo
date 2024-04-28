//go:build tinygo

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
    // instruction available will need to use rcall instead.
    // Note that they will probably not be able to run more than the main
    // goroutine anyway, but this file is compiled for all AVRs so it needs to
    // compile at least.
    rcall  tinygo_pause
#else
    // Other devices can (and must) use the regular call instruction.
    call tinygo_pause
#endif

.global tinygo_swapTask
.type tinygo_swapTask, %function
tinygo_swapTask:
    // This function gets the following parameters:
    // r24:r25 = newStack uintptr
    // r22:r23 = oldStack *uintptr

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

    // Save the current stack pointer in oldStack.
    in  r2, 0x3d; SPL
    in  r3, 0x3e; SPH
    movw Y, r22
    std Y+0, r2
    std Y+1, r3

    // Switch to the new stack pointer.
    in   r0, 0x3f ; SREG
    cli
    out  0x3d, r24; SPL
    out  0x3f, r0 ; SREG, restore interrupts (after the next instruction)
    out  0x3e, r25; SPH

    // Load saved register from the new stack.
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

    // Return into the new task, as if tinygo_swapTask was a regular call.
    ret
