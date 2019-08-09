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
    ldr     sp, =__sp_irq                  // Set IRQ stack
    mov     r0, #0x1f                      // Switch to System Mode
    msr     cpsr, r0
    ldr     sp, =__sp_usr                  // Set user stack

    // Register the interrupt handler
    mov     r0, #0x3000000
    add     r0, #0x0008000
    ldr     r3, =_gba_asm_interrupt_handler
    str     r3, [r0, #-4]

    // Jump to user code (switching to Thumb mode)
    ldr     r3, =main
    bx      r3

.section .data._gba_asm_interrupt_handler
.global  _gba_asm_interrupt_handler
.type    _gba_asm_interrupt_handler, %function
.align
.arm

_gba_asm_interrupt_handler:
    // Registers:
    //   r0 = IOREG
    // NOTE:
    //   BIOS has saved r0-r3
    //   Must save/restore r4-r11 if used

    // Disable IME.
    mov     r0, #0x04000000
    str     r0, [r0, #0x208] // only uses lower bits, aka 0

    // Save registers to be nuked by context switching.
    mrs     r2, spsr
    stmfd   sp!, {r2, lr} // Stack: spsr, lr_irq

    // Switch to SYS mode.
    mrs     r3, cpsr
    bic     r3, r3, #0xdf
    orr     r3, r3, #0x1f
    msr     cpsr, r3

    // Save registers before call.
    //   We save r0 because we need it as soon as we return.
    //   We save lr again because they're different between modes.
    stmfd   sp!, {r0,lr} // Stack: [IOREG, lr_sys, spsr, lr_irq]

    // Call the user-space handler.
    ldr     r3, =runtime_isr_trampoline
    mov     lr,pc
    bx      r3

    // Restore our registers.
    ldmfd   sp!, {r0,lr} // Stack: [spsr, lr_irq]

    // Disable IME again, for safety.
    str     r0, [r0, #0x208] // only uses lower bits, aka 0

    // Switch to INT mode.
    mrs     r3, cpsr
    bic     r3, r3, #0xdf  // Clear all mode bits (except the ARM/THUMB state bit 6)
    orr     r3, r3, #0x92  // Set M1, THUMB, reserved (bits 2, 5, and 8)
    msr     cpsr, r3

    // Restore registers that were nuked by context switch.
    ldmfd   sp!, {r2, lr} // Stack: []
    msr     spsr, r2

    // Enable IME.
    mov     r3, #1
    str     r3, [r0, #0x208]

    // We're done!  Back to the BIOS.
    bx      lr
