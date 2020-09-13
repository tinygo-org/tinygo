// Macro for writing less code
.macro FUNC name
    .section .text.\name, "ax", %progbits
    .global \name
    .type \name, %function
    .align 2
\name:
.endm

FUNC armGetSystemTick
    mrs x0, cntpct_el0
    ret

// Horizon System Calls
// https://switchbrew.org/wiki/SVC
FUNC svcSetHeapSize
    str x0, [sp, #-16]!
    svc 0x1
    ldr x2, [sp], #16
    str x1, [x2]
    ret

FUNC svcExitProcess
    svc 0x7
    ret

FUNC svcSleepThread
    svc 0xB
    ret

FUNC svcOutputDebugString
    svc 0x27
    ret

FUNC svcGetInfo
    str x0, [sp, #-16]!
    svc 0x29
    ldr x2, [sp], #16
    str x1, [x2]
    ret
