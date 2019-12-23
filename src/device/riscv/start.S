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
    call main
