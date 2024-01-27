//go:build tinygo

// Only generate .debug_frame, don't generate .eh_frame.
.cfi_sections .debug_frame

.section .text.tinygo_startTask
.global  tinygo_startTask
.type    tinygo_startTask, %function
tinygo_startTask:
    .cfi_startproc
    // Small assembly stub for starting a goroutine. This is already run on the
    // new stack, with the callee-saved registers already loaded.
    // Most importantly, r4 contains the pc of the to-be-started function and r5
    // contains the only argument it is given. Multiple arguments are packed
    // into one by storing them in a new allocation.

    // Indicate to the unwinder that there is nothing to unwind, this is the
    // root frame. It avoids the following (bogus) error message in GDB:
    //     Backtrace stopped: previous frame identical to this frame (corrupt stack?)
    .cfi_undefined lr

    // Set the first argument of the goroutine start wrapper, which contains all
    // the arguments.
    mov   r0, r5

    // Branch to the "goroutine start" function. By using blx instead of bx,
    // we'll return here instead of tail calling.
    blx   r4

    // After return, exit this goroutine. This is a tail call.
    bl    tinygo_pause
    .cfi_endproc
.size tinygo_startTask, .-tinygo_startTask

.global tinygo_swapTask
.type tinygo_swapTask, %function
tinygo_swapTask:
    // This function gets the following parameters:
    // r0 = newStack uintptr
    // r1 = oldStack *uintptr

    // Save all callee-saved registers:
    push {r4-r11, lr}

    // Save the current stack pointer in oldStack.
    str sp, [r1]

    // Switch to the new stack pointer.
    mov sp, r0

    // Load state from new task and branch to the previous position in the
    // program.
    pop {r4-r11, pc}
