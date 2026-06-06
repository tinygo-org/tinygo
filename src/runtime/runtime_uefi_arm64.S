// UEFI Entry Point for ARM64
//
// efi_main receives (per AAPCS64):
//   x0 = EFI_HANDLE ImageHandle
//   x1 = EFI_SYSTEM_TABLE *SystemTable
//
// We store these for Go code access, then call main.

.section .text.efi_main,"ax"
.global efi_main
efi_main:
    // Save UEFI parameters to global storage.
    // ARM64 uses adrp+add for PC-relative addressing.
    adrp    x2, efi_image_handle
    add     x2, x2, :lo12:efi_image_handle
    str     x0, [x2]

    adrp    x2, efi_system_table
    add     x2, x2, :lo12:efi_system_table
    str     x1, [x2]

    // Save original stack pointer to __stack_top variable.
    adrp    x2, __stack_top
    add     x2, x2, :lo12:__stack_top
    mov     x3, sp
    str     x3, [x2]

    // Ensure stack is 16-byte aligned (AAPCS64 requirement).
    // SP should already be aligned by UEFI firmware, but be safe.
    mov     x3, sp
    and     x3, x3, #~0xF
    mov     sp, x3

    // Call TinyGo main (exported as 'main' from Go and never returns).
    bl      main

// =============================================================================
// Data Section
// =============================================================================

.section .data,"aw"
.align 3

// UEFI handles
.global efi_image_handle
efi_image_handle:
    .quad 0

.global efi_system_table
efi_system_table:
    .quad 0

// _fltused is required by the PE/COFF linker when floating-point code is present.
.global _fltused
_fltused:
    .long 0

// Dummy symbols required by baremetal.go.
// _globals_start/_globals_end: determined at runtime by parsing PE/COFF headers.
// _heap_start/_heap_end: heap is allocated via UEFI AllocatePages.
.global _globals_start
.global _globals_end
.global _heap_start
.global _heap_end
_globals_start:
_globals_end:
_heap_start:
_heap_end:

// Stack top - written by efi_main with the initial SP value.
// preinit() reads the value stored here to fix runtime.stackTop.
.global __stack_top
.global _stack_top
__stack_top:
_stack_top:
    .quad 0

// SBAT (Secure Boot Advanced Targeting) metadata section.
// Required by modern shim for second-stage binary verification.

.section .sbat,"dr"
.ascii "sbat,1,SBAT Version,sbat,1,https://github.com/rhboot/shim/blob/main/SBAT.md\n"
.ascii "node,1,TinyGo,app.efi,1.0,https://tinygo.org\n"