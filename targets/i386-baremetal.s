// Based on https://wiki.osdev.org/Bare_Bones

// Declare constants for the multiboot header.
.set ALIGN,    1<<0             // align loaded modules on page boundaries
.set MEMINFO,  1<<1             // provide memory map
.set FLAGS,    ALIGN | MEMINFO  // this is the Multiboot 'flag' field
.set MAGIC,    0x1BADB002       // 'magic number' lets bootloader find the header
.set CHECKSUM, -(MAGIC + FLAGS) // checksum of above, to prove we are multiboot

// Declare the multiboot section, which the bootloader will search for in the
// first 8kB of the kernel image.
.section .multiboot
.align 4
.long MAGIC
.long FLAGS
.long CHECKSUM

// Entry point, jumped to from the bootloader.
.section .text
.global _start
.type _start, @function
_start:
    // Set up a stack: this is not done by the bootloader.
    mov $_stack_top, %esp

    // Jump to the main function.
    call main

    // Lock up the system, as a safety guard. We should never reach this point.
    cli
1:  hlt
    jmp 1b


.section .text.outb
.global outb
.type outb, @function
outb:
    movl    8(%esp), %eax
    movl    4(%esp), %edx
    outb %al, %dx
    ret

.section .text.halt
.global halt
.type halt, @function
halt:
    movw    $0x2000, %ax
    movl    $0x604, %edx
    outw    %ax, %dx
    // Lock up in case shutdown wasn't enough (we're not running in QEMU).
1:  hlt
    jmp 1b
