.syntax unified
.cfi_sections .debug_frame

.section .text.HardFault_Handler
.global  HardFault_Handler
.type    HardFault_Handler, %function
HardFault_Handler:
    .cfi_startproc
    // Put the old stack pointer in the first argument, for easy debugging. This
    // is especially useful on Cortex-M0, which supports far fewer debug
    // facilities.
    mov r0, sp

    // Load the default stack pointer from address 0 so that we can call normal
    // functions again that expect a working stack. However, it will corrupt the
    // old stack so the function below must not attempt to recover from this
    // fault.
    movs r3, #0
    ldr r3, [r3]
    mov sp, r3

    // Continue handling this error in Go.
    bl handleHardFault
    .cfi_endproc
.size HardFault_Handler, .-HardFault_Handler

// This is a convenience function for semihosting support.
// At some point, this should be replaced by inline assembly.
.section .text.SemihostingCall
.global  SemihostingCall
.type    SemihostingCall, %function
SemihostingCall:
    .cfi_startproc
    bkpt 0xab
    bx   lr
    .cfi_endproc
.size SemihostingCall, .-SemihostingCall
