.section .init
.global _start
.type _start,@function

_start:
    // Workaround for missing support of the la pseudo-instruction in Clang 8:
    // https://reviews.llvm.org/D55325
    lui sp,      %hi(_stack_top)
    addi sp, sp, %lo(_stack_top)
    // see https://gnu-mcu-eclipse.github.io/arch/riscv/programmer/#the-gp-global-pointer-register
    lui gp,      %hi(__global_pointer$)
    addi gp, gp, %lo(__global_pointer$)
    call main
