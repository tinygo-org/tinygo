//go:build tinygo

.section .text.tinygo_startTask,"ax",@progbits
.global  tinygo_startTask
.type    tinygo_startTask, %function
tinygo_startTask:
    // Small assembly stub for starting a goroutine. This already runs on the
    // new stack, control reaches this function after returning from the initial
    // tinygo_swapTask below (the retw.n instruction).
    //
    // The stack was set up in such a way that it looks as if this function was
    // paused using tinygo_swapTask by setting up the parent register window and
    // return pointer as a call4 instruction - except such a call never took
    // place. Instead, the stack pointer is switched to the new stack after all
    // live-but-invisible registers have been flushed to the stack. This means
    // that all registers as present in tinygo_swapTask are moved four up (a2 in
    // tinygo_swapTask is a6 in this function). We don't use any of those
    // registers however. Instead, the retw.n instruction will load them through
    // an underflow exception from the stack which means we get a0-a3 as defined
    // in task_stack_esp32.go.

    // Branch to the "goroutine start" function. The first (and only) parameter
    // is stored in a2, but has to be moved to a6 to make it appear as a2 in the
    // goroutine start function (due to changing the register window by four
    // with callx4).
    mov.n a6, a2
    callx4 a3

    // After return, exit this goroutine. This call never returns.
    call4  tinygo_task_exit

.section .text.tinygo_swapTask,"ax",@progbits
.global tinygo_swapTask
.type tinygo_swapTask, %function
tinygo_swapTask:
    // This function gets the following parameters:
    // a2 = newStack uintptr
    // a3 = oldStack *uintptr

    // Reserve 32 bytes on the stack. It really needs to be 32 bytes, with 16
    // extra at the bottom to adhere to the ABI.
    entry sp, 32

    // Disable interrupts while flushing registers. This is necessary because
    // interrupts might want to use the stack pointer (at a2) which will be some
    // arbitrary register while registers are flushed.
    rsil a4, 3 // XCHAL_EXCM_LEVEL

    // Save the old PS (in a4) to the stack because call4 clobbers a4-a7.
    s32i a4, sp, 4

    // Flush all register windows to the stack using recursive call4.
    //
    // The previous ROTW-based approach (borrowed from Zephyr) does NOT
    // actually trigger window overflow exceptions on Xtensa LX7 — ROTW
    // simply modifies WindowBase without saving any registers to the stack.
    // This leaves stale window data in the physical register file after a
    // goroutine switch, corrupting the overflow save chain when the new
    // goroutine's deeper calls trigger overflow of old goroutine windows.
    //
    // Instead, we recursively call a small function via call4.  Each call4 +
    // entry triggers the hardware window-overflow mechanism for any occupied
    // pane being reused, correctly saving registers to the stack.
    // For NAREG=64 (16 panes), 15 recursive levels cover all panes except
    // the current one (tinygo_swapTask), which must stay active.
    movi a6, 15
    call4 .Lspill_windows

    // After the recursive spill returns, the physical register file still
    // has WindowStart bits set for the spill helper frames.
    // We will clear WindowStart completely (to 0) right before the retw.n
    // below, after the stack switch is done.  This prevents stale overflow
    // when the new goroutine's calls rotate back into these panes.

    // Restore interrupts.
    l32i a4, sp, 4          // reload saved PS
    wsr.ps a4
    rsync

    // At this point, the following is true:
    //     WindowStart == 1 << WindowBase
    // All other windows have been properly flushed to their stacks.
    // It also means that the stack pointer can now be safely modified.

    // Save a0, which stores the return address and the parent register window
    // in the upper two bits.
    s32i.n a0, sp, 0

    // Save the current stack pointer in oldStack.
    s32i.n  sp, a3, 0

    // Switch to the new stack pointer (newStack).
    mov.n   sp, a2

    // Load a0, which is the previous return address from before the previous
    // switch or the constructed return address to tinygo_startTask. This
    // register also stores the parent register window.
    l32i.n a0, sp, 0

    // Clear ALL WindowStart bits.  With all windows spilled to the stack,
    // we must ensure no stale WS bits remain: the retw.n below will trigger
    // underflow4 to load the new goroutine's registers from the new stack
    // (which sets the appropriate WS bit via rfwu).  Any stale WS bits
    // (from spill helpers or the old goroutine) would cause spurious
    // overflows of garbage register values into memory.
    movi a5, 0
    wsr  a5, WINDOWSTART
    rsync

    // Return into the new stack. This instruction will trigger a window
    // underflow, reloading the saved registers from the stack.
    retw.n

// Recursive helper for flushing all register windows.
// Parameter: a2 = remaining recursion depth (passed via caller's a6).
// Each call4 + entry triggers the hardware overflow mechanism for any
// occupied pane at the new WindowBase position.
    .balign 4
.Lspill_windows:
    entry a1, 16
    beqz  a2, .Lspill_done
    addi  a2, a2, -1
    mov   a6, a2
    call4 .Lspill_windows
.Lspill_done:
    retw.n
