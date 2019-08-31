.section .text.tinygo_startTask
.global  tinygo_startTask
.type    tinygo_startTask, %function
tinygo_startTask:
    // Small assembly stub for starting a goroutine. This is already run on the
    // new stack, with the callee-saved registers already loaded.
    // Most importantly, r4 contains the pc of the to-be-started function and r5
    // contains the only argument it is given. Multiple arguments are packed
    // into one by storing them in a new allocation.

    // Set the first argument of the goroutine start wrapper, which contains all
    // the arguments.
    mov   r0, r5

    // Branch to the "goroutine start" function. By using blx instead of bx,
    // we'll return here instead of tail calling.
    blx   r4

    // After return, exit this goroutine. This is a tail call.
    bl    runtime.yield

.section .text.tinygo_swapTask
.global  tinygo_swapTask
.type    tinygo_swapTask, %function
tinygo_swapTask:
    // r0 = oldTask *task
    // r1 = newTask *task

    // This function stores the current register state to a task struct and
    // loads the state of another task to replace the current state. Apart from
    // saving and restoring all relevant callee-saved registers, it also ends
    // with branching to the last program counter (saved as the lr register, to
    // follow the ARM calling convention).

    // On pre-Thumb2 CPUs (Cortex-M0 in particular), registers r8-r15 cannot be
    // used directly. Only very few operations work on them, such as mov. That's
    // why the higher register values are first stored in the temporary register
    // r3 when loading/storing them.

    // Store state to old task. It saves the lr instead of the pc, because that
    // will be the pc after returning back to the old task (in a different
    // invocation of swapTask).
    str r4, [r0, #0]
    str r5, [r0, #4]
    str r6, [r0, #8]
    str r7, [r0, #12]
    #if defined(__thumb2__)
    str r8, [r0, #16]
    str r9, [r0, #20]
    str r10, [r0, #24]
    str r11, [r0, #28]
    str sp, [r0, #32]
    str lr, [r0, #36]
    #else
    mov r3, r8
    str r3, [r0, #16]
    mov r3, r9
    str r3, [r0, #20]
    mov r3, r10
    str r3, [r0, #24]
    mov r3, r11
    str r3, [r0, #28]
    mov r3, sp
    str r3, [r0, #32]
    mov r3, lr
    str r3, [r0, #36]
    #endif

    // Load state from new task and branch to the previous position in the
    // program.
    ldr r4, [r1, #0]
    ldr r5, [r1, #4]
    ldr r6, [r1, #8]
    ldr r7, [r1, #12]
    #if defined(__thumb2__)
    ldr r8, [r1, #16]
    ldr r9, [r1, #20]
    ldr r10, [r1, #24]
    ldr r11, [r1, #28]
    ldr sp, [r1, #32]
    #else
    ldr r3, [r1, #16]
    mov r8, r3
    ldr r3, [r1, #20]
    mov r9, r3
    ldr r3, [r1, #24]
    mov r10, r3
    ldr r3, [r1, #28]
    mov r11, r3
    ldr r3, [r1, #32]
    mov sp, r3
    #endif
    ldr r3, [r1, #36]
    bx r3
