// For more information on the .nro file format, see:
// https://switchbrew.org/wiki/NRO

.section .text.jmp, "x"
.global _start
_start:
    b start
    .word _mod_header - _start
    .ascii "HOMEBREW"

    .ascii "NRO0"              // magic
    .word 0                    // version (always 0)
    .word __bss_start - _start // total NRO file size
    .word 0                    // flags (unused)

    // segment headers
    .word 0 // __text_start
    .word __text_size
    .word 0 //__rodata_start
    .word __rodata_size
    .word 0 //__data_start
    .word __data_size
    .word __bss_size
    .word 0

    // ModuleId (not supported)
    . = 0x50; // skip 32 bytes

    .word 0 // DSO Module Offset (unused)
    .word 0 // reserved (unused)

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
    // Get ASLR Base
    adrp x6, _start

    // context ptr and main thread handle
    mov  x5, x0
    mov  x4, x1

    // Save lr, context pointer, main thread handler
    adrp x0, _aslr_base
    str  x6, [x0, #:lo12:_aslr_base]

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
    // process .dynamic section
    adrp x0, _aslr_base
    ldr  x0, [x0, #:lo12:_aslr_base]
    adrp x1, _DYNAMIC
    add  x1, x1, #:lo12:_DYNAMIC
    bl   __dynamic_loader

    // set LR to svcExitProcess if it's null
    adrp x3, exit
    add x3, x3, #:lo12:exit
    cmp x30, xzr
    csel x30, x3, x30, eq

    // call entrypoint
    mov x3, sp
    sub sp, sp, 0x10
    stp x29, x30, [sp]
    b    main

.section .data.horizon
.align 8
.global _aslr_base // Placeholder for ASLR Base Address
_aslr_base:
    .dword 0
