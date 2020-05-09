.section .text.jmp, "x"
.global _start
_start:
    b start
    .word _mod_header - _start

.section .data.mod0
    .word 0, 8

.global _mod_header
_mod_header:
    .ascii "MOD0"
    .word __dynamic_start - _mod_header
    .word __bss_start - _mod_header
    .word __bss_end - _mod_header
    .word 0, 0 // eh_frame_hdr start/end
    .word 0 // runtime-generated module object offset

.section .text, "x"
.global start
start:

    // save lr
    mov  x7, x30

    // get aslr base
    bl   +4
    sub  x6, x30, #0x88

    // context ptr and main thread handle
    mov  x5, x0
    mov  x4, x1

    // clear .bss
    adrp x5, __bss_start
    add x5, x5, #:lo12:__bss_start
    adrp x6, __bss_end
    add x6, x6, #:lo12:__bss_end

bssloop:
    cmp x5, x6
    b.eq run
    str xzr, [x5]
    add x5, x5, 8
    b bssloop

run:
    // call entrypoint
    adrp x30, exit
    add  x30, x30, #:lo12:exit
    b    main
