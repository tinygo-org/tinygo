// UEFI Entry Point for x86_64
//
// efi_main receives (per MS x64 ABI):
//   RCX = EFI_HANDLE ImageHandle
//   RDX = EFI_SYSTEM_TABLE *SystemTable
//
// We store these for Go code access, then call main.

.section .text.efi_main,"ax"
.global efi_main
efi_main:
    // Save UEFI parameters to global storage
    // These will be accessible from Go code via extern vars
    leaq efi_image_handle(%rip), %rax
    movq %rcx, (%rax)

    leaq efi_system_table(%rip), %rax
    movq %rdx, (%rax)

    // Save original stack pointer to __stack_top variable
    leaq __stack_top(%rip), %rax
    movq %rsp, (%rax)

    // Align stack to 16 bytes (MS ABI requirement)
    // This is critical for XMM register operations in task switching
    andq $-16, %rsp

    // Reserve shadow space (32 bytes) required by MS x64 ABI
    subq $32, %rsp

    // Call TinyGo main (exported as 'main' from Go and never returns)
    callq main

// =============================================================================
// Data Section
// =============================================================================

.section .data,"aw"
.align 8

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

// Stack top - written by efi_main with the initial RSP value.
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