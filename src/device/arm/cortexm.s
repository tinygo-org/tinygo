.syntax unified

// Note: some of the below interrupt handlers are unused on smaller cores (such
// as Cortex-M0) but are defined here anyway for consistency.

// Define fault handlers in a consistent way.
.macro FaultIRQ handler id
    .section .text.\handler
    .global  \handler
    .weak    \handler
    .type    \handler, %function
    .weak    \handler
    \handler:
        movs r0, \id
        b    Fault_Handler
.endm

FaultIRQ NMI_Handler              2
FaultIRQ HardFault_Handler        3
FaultIRQ MemoryManagement_Handler 4
FaultIRQ BusFault_Handler         5
FaultIRQ UsageFault_Handler       6

.section .text.Fault_Handler
.global  Fault_Handler
.type    Fault_Handler, %function
Fault_Handler:
    // Put the old stack pointer in the first argument, for easy debugging. This
    // is especially useful on Cortex-M0, which supports far fewer debug
    // facilities.
    mov r1, sp

    // Load the default stack pointer from address 0 so that we can call normal
    // functions again that expect a working stack. However, it will corrupt the
    // old stack so the function below must not attempt to recover from this
    // fault.
    movs r3, #0
    ldr r3, [r3]
    mov sp, r3

    // Continue handling this error in Go.
    bl handleFault

// This is a convenience function for semihosting support.
// At some point, this should be replaced by inline assembly.
.section .text.SemihostingCall
.global  SemihostingCall
.type    SemihostingCall, %function
SemihostingCall:
    bkpt 0xab
    bx   lr
