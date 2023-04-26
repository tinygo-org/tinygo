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
    bl    	tinygo_pause
    .cfi_endproc
.size tinygo_startTask, .-tinygo_startTask

.section .text.tinygo_switchToScheduler
.global  tinygo_switchToScheduler
.type    tinygo_switchToScheduler, %function
tinygo_switchToScheduler:
    .cfi_startproc
    // r0 = sp *uintptr

    // Currently on the task stack (SP=PSP). We need to store the position on
    // the stack where the in-use registers will be stored.
    mov r1, sp
    subs r1, #36
    str r1, [r0]

    b tinygo_swapTask
    .cfi_endproc
.size tinygo_switchToScheduler, .-tinygo_switchToScheduler

.section .text.tinygo_switchToTask
.global  tinygo_switchToTask
.type    tinygo_switchToTask, %function
tinygo_switchToTask:
    .cfi_startproc
    // r0 = sp uintptr

    // Currently on the scheduler stack (SP=MSP). We'll have to update the PSP,
    // and then we can invoke swapTask.
    msr PSP, r0

    b.n tinygo_swapTask
    .cfi_endproc
.size tinygo_switchToTask, .-tinygo_switchToTask

.section .text.tinygo_swapTask
.global  tinygo_swapTask
.type    tinygo_swapTask, %function
tinygo_swapTask:
    .cfi_startproc
    // This function stores the current register state to the stack, switches to
    // the other stack (MSP/PSP), and loads the register state from the other
    // stack. Apart from saving and restoring all relevant callee-saved
    // registers, it also ends with branching to the last program counter (saved
    // as the lr register, to follow the ARM calling convention).

    // On pre-Thumb2 CPUs (Cortex-M0 in particular), registers r8-r15 cannot be
    // used directly. Only very few operations work on them, such as mov. That's
    // why the higher register values are first stored in the temporary register
    // r3 when loading/storing them.
    // It is possible to reduce the swapTask by two instructions (~2 cycles) on
    // Cortex-M0 by reordering the layout of the pushed registers from {r4-r11,
    // lr} to {r8-r11, r4-r8, lr}. However, that also requires a change on the
    // Go side (depending on thumb1/thumb2!) and so is not really worth the
    // complexity.

    // Store state to old task. It saves the lr instead of the pc, because that
    // will be the pc after returning back to the old task (in a different
    // invocation of swapTask).
    #if defined(__thumb2__)
    push {r4-r11, lr}
    .cfi_def_cfa_offset 9*4
    #else
    mov r0, r8
    mov r1, r9
    mov r2, r10
    mov r3, r11
    push {r0-r3, lr}
    .cfi_def_cfa_offset 5*4
    push {r4-r7}
    .cfi_def_cfa_offset 9*4
    #endif

    // Switch the stack. This could either switch from PSP to MSP, or from MSP
    // to PSP. By using an XOR (eor), it will just switch to the other stack.
    mrs  r0, CONTROL // load CONTROL register
    movs r3, #2
    eors r0, r0, r3  // flip the SPSEL (active stack pointer) bit
    msr  CONTROL, r0 // store CONTROL register
    isb              // required to flush the pipeline

    // Load state from new task and branch to the previous position in the
    // program.
    #if defined(__thumb2__)
    pop {r4-r11, pc}
    #else
    pop {r4-r7}
    .cfi_def_cfa_offset 5*4
    pop {r0-r3}
    .cfi_def_cfa_offset 1*4
    mov r8, r0
    mov r9, r1
    mov r10, r2
    mov r11, r3
    pop {pc}
    #endif
    .cfi_endproc
.size tinygo_swapTask, .-tinygo_swapTask
