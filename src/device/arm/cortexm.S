.syntax unified
.cfi_sections .debug_frame

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
