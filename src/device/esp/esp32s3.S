// -----------------------------------------------------------------------
// tinygo_scanCurrentStack — Spill all Xtensa register windows to the
// stack, then call tinygo_scanstack(sp) so the conservative GC can
// discover live heap pointers that are currently in physical registers.
//
// On RISC-V / ARM the equivalent function pushes callee-saved registers
// before the call.  On Xtensa windowed ABI the same effect is achieved
// by forcing hardware window-overflow for every occupied pane: each
// overflow saves the four registers in that pane to the stack frame
// pointed to by the pane's a1 (sp).  After all panes are flushed, a
// scan from the current sp to stackTop covers every live value.
// -----------------------------------------------------------------------
.section .text.tinygo_scanCurrentStack

.global tinygo_scanCurrentStack
tinygo_scanCurrentStack:
    entry a1, 48

    // Disable interrupts while flushing register windows.
    rsr   a4, PS
    s32i  a4, a1, 0         // save PS for later restore
    rsil  a4, 3             // XCHAL_EXCM_LEVEL

    // Flush all register windows using recursive call4.
    // For NAREG=64 (16 panes), 15 recursive levels cover all panes
    // except the current one (which is kept active).
    movi a6, 15
    call4 .Lscan_spill

    // Restore interrupts.
    l32i  a4, a1, 0
    wsr.ps a4
    rsync

    // Pass current sp to tinygo_scanstack.
    // call4 maps caller's a5→callee's a1 (stack ptr for callee's entry)
    // and caller's a6→callee's a2 (first argument = sp).
    mov  a5, a1             // callee's a1 = valid stack pointer
    mov  a6, a1             // callee's a2 = sp argument
    call4 tinygo_scanstack

    retw

    .balign 4
.Lscan_spill:
    entry a1, 16
    beqz  a2, .Lscan_spill_done
    addi  a2, a2, -1
    mov   a6, a2
    call4 .Lscan_spill
.Lscan_spill_done:
    retw
