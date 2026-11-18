//go:build tinygo && esp32p4

// Task context switch for ESP32-P4 (RV32IMAFC, ilp32f ABI).
//
// The ilp32f ABI designates fs0-fs11 (f8-f9, f18-f27) as callee-saved.
// The generic task_stack_tinygoriscv.S only saves integer regs; goroutines
// that use float arithmetic across a yield point would see corrupted FPU state.
// This file replaces the generic version for esp32p4 and saves fs0-fs11+fcsr.
//
// Stack frame layout (sp points to bottom after push):
//   sp+ 0: s0    sp+ 4: s1    sp+ 8: s2    sp+12: s3
//   sp+16: s4    sp+20: s5    sp+24: s6    sp+28: s7
//   sp+32: s8    sp+36: s9    sp+40: s10   sp+44: s11
//   sp+48: ra
//   sp+52: fcsr
//   sp+56: fs0   sp+60: fs1   sp+64: fs2   sp+68: fs3
//   sp+72: fs4   sp+76: fs5   sp+80: fs6   sp+84: fs7
//   sp+88: fs8   sp+92: fs9   sp+96: fs10  sp+100: fs11
// Total: 104 bytes  (matches sizeof(calleeSavedRegs) in task_stack_esp32p4.go)

.section .text.tinygo_startTask
.global  tinygo_startTask
.type    tinygo_startTask, %function
tinygo_startTask:
    // Entry point for a newly created goroutine.
    // s0 = function to call, s1 = argument pointer.
    mv    a0, s1
    jalr  s0
    tail  tinygo_task_exit

.section .text.tinygo_swapTask
.global  tinygo_swapTask
.type    tinygo_swapTask, %function
tinygo_swapTask:
    // Save the current goroutine and switch to another.
    // a0 = newStack (uintptr)
    // a1 = oldStack (*uintptr)

    // ── Allocate frame (104 bytes) ──────────────────────────────────────────
    addi  sp, sp, -104

    // ── Save integer callee-saved registers ─────────────────────────────────
    sw    s0,    0(sp)
    sw    s1,    4(sp)
    sw    s2,    8(sp)
    sw    s3,   12(sp)
    sw    s4,   16(sp)
    sw    s5,   20(sp)
    sw    s6,   24(sp)
    sw    s7,   28(sp)
    sw    s8,   32(sp)
    sw    s9,   36(sp)
    sw    s10,  40(sp)
    sw    s11,  44(sp)
    sw    ra,   48(sp)

    // ── Save FPU callee-saved registers (ilp32f ABI) ─────────────────────────
    csrr  t0, fcsr
    sw    t0,   52(sp)
    fsw   fs0,  56(sp)
    fsw   fs1,  60(sp)
    fsw   fs2,  64(sp)
    fsw   fs3,  68(sp)
    fsw   fs4,  72(sp)
    fsw   fs5,  76(sp)
    fsw   fs6,  80(sp)
    fsw   fs7,  84(sp)
    fsw   fs8,  88(sp)
    fsw   fs9,  92(sp)
    fsw   fs10, 96(sp)
    fsw   fs11, 100(sp)

    // ── Switch stacks ────────────────────────────────────────────────────────
    sw    sp,  0(a1)      // *oldStack = sp
    mv    sp,  a0         // sp = newStack

    // ── Restore integer callee-saved registers ───────────────────────────────
    lw    s0,    0(sp)
    lw    s1,    4(sp)
    lw    s2,    8(sp)
    lw    s3,   12(sp)
    lw    s4,   16(sp)
    lw    s5,   20(sp)
    lw    s6,   24(sp)
    lw    s7,   28(sp)
    lw    s8,   32(sp)
    lw    s9,   36(sp)
    lw    s10,  40(sp)
    lw    s11,  44(sp)
    lw    ra,   48(sp)

    // ── Restore FPU callee-saved registers ───────────────────────────────────
    lw    t0,   52(sp)
    csrw  fcsr, t0
    flw   fs0,  56(sp)
    flw   fs1,  60(sp)
    flw   fs2,  64(sp)
    flw   fs3,  68(sp)
    flw   fs4,  72(sp)
    flw   fs5,  76(sp)
    flw   fs6,  80(sp)
    flw   fs7,  84(sp)
    flw   fs8,  88(sp)
    flw   fs9,  92(sp)
    flw   fs10, 96(sp)
    flw   fs11, 100(sp)

    // ── Deallocate frame and return ──────────────────────────────────────────
    addi  sp, sp, 104
    ret
