#ifdef __XTENSA_WINDOWED_ABI__
.section .text.__xtensa_libgcc_window_spill,"ax",@progbits
.global __xtensa_libgcc_window_spill
.type __xtensa_libgcc_window_spill, %function
__xtensa_libgcc_window_spill:
    entry sp, 16
    movi a6, 15
    call4 .Lspill_windows
    retw
.size __xtensa_libgcc_window_spill, .-__xtensa_libgcc_window_spill

.balign 4
.Lspill_windows:
    entry sp, 16
    beqz a2, .Lspill_done
    addi a2, a2, -1
    mov a6, a2
    call4 .Lspill_windows
.Lspill_done:
    retw
#endif

.section .text.tinygo_longjmp,"ax",@progbits
.global tinygo_longjmp
.type tinygo_longjmp, %function
tinygo_longjmp:
#ifdef __XTENSA_WINDOWED_ABI__
    // longjmp's entry rotates a10/a11 into its a2/a3 arguments.
    movi a11, 1
#else
    movi a3, 1
#endif
    j longjmp
.size tinygo_longjmp, .-tinygo_longjmp
