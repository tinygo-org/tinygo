.section .init
.global _start
.type _start,@function

_start:
    // Load the stack pointer.
    la sp,      _stack_top

    // Load the globals pointer. The program will load pointers relative to this
    // register, so it must be set to the right value on startup.
    // See: https://gnu-mcu-eclipse.github.io/arch/riscv/programmer/#the-gp-global-pointer-register
    // Linker relaxations must be disabled to avoid the initialization beign
    // relaxed with an uninitialized global pointer: mv gp, gp
    .option push
    .option norelax
    la gp,      __global_pointer$
    .option pop

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
    addi    sp, sp, -64
    sw      ra, 60(sp)
    sw      t0, 56(sp)
    sw      t1, 52(sp)
    sw      t2, 48(sp)
    sw      a0, 44(sp)
    sw      a1, 40(sp)
    sw      a2, 36(sp)
    sw      a3, 32(sp)
    sw      a4, 28(sp)
    sw      a5, 24(sp)
    sw      a6, 20(sp)
    sw      a7, 16(sp)
    sw      t3, 12(sp)
    sw      t4, 8(sp)
    sw      t5, 4(sp)
    sw      t6, 0(sp)
    call    handleInterrupt
    lw      t6, 0(sp)
    lw      t5, 4(sp)
    lw      t4, 8(sp)
    lw      t3, 12(sp)
    lw      a7, 16(sp)
    lw      a6, 20(sp)
    lw      a5, 24(sp)
    lw      a4, 28(sp)
    lw      a3, 32(sp)
    lw      a2, 36(sp)
    lw      a1, 40(sp)
    lw      a0, 44(sp)
    lw      t2, 48(sp)
    lw      t1, 52(sp)
    lw      t0, 56(sp)
    lw      ra, 60(sp)
    addi    sp, sp, 64
    mret
