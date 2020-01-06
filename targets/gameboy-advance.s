.section    .init
.global     _start
.align
.arm

_start:
    b       start_vector
    .fill   156,1,0                // Nintendo Logo Character Data (8000004h)
    .fill   16,1,0                 // Game Title
    .byte   0x30,0x31              // Maker Code (80000B0h)
    .byte   0x96                   // Fixed Value (80000B2h)
    .byte   0x00                   // Main Unit Code (80000B3h)
    .byte   0x00                   // Device Type (80000B4h)
    .fill   7,1,0                  // unused
    .byte   0x00                   // Software Version No (80000BCh)
    .byte   0xf0                   // Complement Check (80000BDh)
    .byte   0x00,0x00              // Checksum (80000BEh)

start_vector:
    mov     r0, #0x4000000                 // REG_BASE
    str     r0, [r0, #0x208]

    mov     r0, #0x12                      // Switch to IRQ Mode
    msr     cpsr, r0
    ldr     sp, =_stack_top_irq            // Set IRQ stack
    mov     r0, #0x1f                      // Switch to System Mode
    msr     cpsr, r0
    ldr     sp, =_stack_top                // Set user stack

    // Jump to user code (switching to Thumb mode)
    ldr     r3, =main
    bx      r3

