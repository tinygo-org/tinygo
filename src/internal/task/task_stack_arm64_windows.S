.section .text.tinygo_startTask,"ax"
.global  tinygo_startTask
tinygo_startTask:
    .cfi_startproc
    // Small assembly stub for starting a goroutine. This is already run on the
    // new stack, with the callee-saved registers already loaded.
    // Most importantly, x19 contains the pc of the to-be-started function and
    // x20 contains the only argument it is given. Multiple arguments are packed
    // into one by storing them in a new allocation.

    // Indicate to the unwinder that there is nothing to unwind, this is the
    // root frame. It avoids the following (bogus) error message in GDB:
    //     Backtrace stopped: previous frame identical to this frame (corrupt stack?)
    .cfi_undefined lr

    // Set the first argument of the goroutine start wrapper, which contains all
    // the arguments.
    mov   x0, x20

    // Branch to the "goroutine start" function. By using blx instead of bx,
    // we'll return here instead of tail calling.
    blr   x19

    // After return, exit this goroutine. This is a tail call.
    b     tinygo_pause
    .cfi_endproc


.global tinygo_swapTask
tinygo_swapTask:
    // This function gets the following parameters:
    // x0 = newStack uintptr
    // x1 = oldStack *uintptr

    // Save all callee-saved registers:
    stp     x19, x20, [sp, #-160]!
    stp     x21, x22, [sp, #16]
    stp     x23, x24, [sp, #32]
    stp     x25, x26, [sp, #48]
    stp     x27, x28, [sp, #64]
    stp     x29, x30, [sp, #80]
    stp     d8,  d9,  [sp, #96]
    stp     d10, d11, [sp, #112]
    stp     d12, d13, [sp, #128]
    stp     d14, d15, [sp, #144]

    // Save the current stack pointer in oldStack.
    mov x8, sp
    str x8, [x1]

    // Switch to the new stack pointer.
    mov sp, x0

    // Restore stack state and return.
    ldp     d14, d15, [sp, #144]
    ldp     d12, d13, [sp, #128]
    ldp     d10, d11, [sp, #112]
    ldp     d8, d9, [sp, #96]
    ldp     x29, x30, [sp, #80]
    ldp     x27, x28, [sp, #64]
    ldp     x25, x26, [sp, #48]
    ldp     x23, x24, [sp, #32]
    ldp     x21, x22, [sp, #16]
    ldp     x19, x20, [sp], #160
    ret
