.section .init
.global _start
.type _start,@function

_start:
    // Load the stack pointer.
    la sp,      _stack_top

    // Load the globals pointer. The program will load pointers relative to this
    // register, so it must be set to the right value on startup.
    // See: https://gnu-mcu-eclipse.github.io/arch/riscv/programmer/#the-gp-global-pointer-register
    la gp,      __global_pointer$

    // Jump to runtime.main
    call main

.section .text.handleInterruptASM
.global handleInterruptASM
.type handleInterruptASM,@function
handleInterruptASM:
    // Save and restore all registers, because the hardware only saves/restores
    // the pc.
    // Note: we have to do this in assembly because the "interrupt"="machine"
    // attribute is broken in LLVM: https://bugs.llvm.org/show_bug.cgi?id=42984
    addi    sp, sp, -128
    sd      ra, 120(sp)
    sd      t0, 112(sp)
    sd      t1, 104(sp)
    sd      t2, 96(sp)
    sd      a0, 88(sp)
    sd      a1, 80(sp)
    sd      a2, 72(sp)
    sd      a3, 64(sp)
    sd      a4, 56(sp)
    sd      a5, 48(sp)
    sd      a6, 40(sp)
    sd      a7, 32(sp)
    sd      t3, 24(sp)
    sd      t4, 16(sp)
    sd      t5, 8(sp)
    sd      t6, 0(sp)
    call    handleInterrupt
    ld      t6, 0(sp)
    ld      t5, 8(sp)
    ld      t4, 16(sp)
    ld      t3, 24(sp)
    ld      a7, 32(sp)
    ld      a6, 40(sp)
    ld      a5, 48(sp)
    ld      a4, 56(sp)
    ld      a3, 64(sp)
    ld      a2, 72(sp)
    ld      a1, 80(sp)
    ld      a0, 88(sp)
    ld      t2, 96(sp)
    ld      t1, 104(sp)
    ld      t0, 112(sp)
    ld      ra, 120(sp)
    addi    sp, sp, 128
    mret
