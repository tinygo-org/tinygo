.section .init
.global _start
.type _start,@function

_start:
    // If we're on a multicore system, we need to wait for hart 0 to wake us up.
#if TINYGO_CORES > 1
    csrr a0, mhartid

    // Hart 0 stack
    bnez a0, 1f
    la sp,      _stack_top

1:
    // Hart 1 stack
    li a1, 1
    bne a0, a1, 2f
    la sp,      _stack1_top

2:
    // Hart 2 stack
    #if TINYGO_CORES >= 3
    li a1, 2
    bne a0, a1, 3f
    la sp,      _stack2_top
    #endif

3:
    // Hart 3 stack
    #if TINYGO_CORES >= 4
    li a1, 3
    bne a0, a1, 4f
    la sp,      _stack3_top
    #endif

4:
    // done

#if TINYGO_CORES > 4
#error only up to 4 cores are supported at the moment!
#endif

#else
    // Load the stack pointer.
    la sp,      _stack_top
#endif

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
