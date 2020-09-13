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
    .word __text_start - _start
    .word __text_size
    .word __rodata_start - _start
    .word __rodata_size
    .word __data_start - _start
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

.section .text.start, "x"
.global start
start:
    // save lr
    mov  x7, x30

    // get aslr base
    bl   +4
    sub  x6, x30, #0x88

    // context ptr and main thread handle
    mov  x25, x0
    mov  x26, x1

    // Save ASLR Base to use later
    mov x0, x6

    adrp x4, _saved_return_address
    str  x7, [x4, #:lo12:_saved_return_address]

    adrp x4, _context
    str x25, [x4, #:lo12:_context]

    adrp x4, _main_thread
    str x26, [x4, #:lo12:_main_thread]

    // store stack pointer
    mov  x26, sp
    adrp x4, _stack_top
    str  x26, [x4, #:lo12:_stack_top]

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
    // ASLR base on x0
    adrp x1, _DYNAMIC
    add  x1, x1, #:lo12:_DYNAMIC
    bl   __dynamic_loader

    // call entrypoint
    b    main

.global __nx_exit
.type   __nx_exit, %function
__nx_exit:
    // Exit code in x0

    // restore stack pointer
    mov sp, x1

    // jump back to loader
    br   x2

.section .data.horizon
.align 8
.global _saved_return_address // Saved return address.
                              // This might be different than null when coming from launcher
_saved_return_address:
    .dword 0
.global _context // Homebrew Launcher Context
                 // This might be different than null when not coming from launcher
_context:
    .dword 0

.global _main_thread
_main_thread:
    .dword 0

.global _stack_top
_stack_top:
    .dword 0
