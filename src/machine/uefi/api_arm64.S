// UEFI Call Stub for ARM64
//
// Generic stub for calling any UEFI function via the AAPCS64 calling convention.
//
// Go signature: func callAsm(fn uintptr, args *uintptr, nargs uintptr) Status
// AAPCS64 entry:
//   x0 = fn (function pointer to call)
//   x1 = args (pointer to uintptr array of arguments)
//   x2 = nargs (number of arguments, 0-N)
//
// AAPCS64 register mapping for the callee:
//   Args 1-8:  x0-x7
//   Args 9+:   stack (no shadow space)
//   Return:    x0
//   Stack:     16-byte aligned at all times

.section .text.uefiCall,"ax"
.global uefiCall
uefiCall:
    // Save callee-saved registers and link register.
    stp     x29, x30, [sp, #-64]!
    mov     x29, sp
    stp     x19, x20, [sp, #16]
    stp     x21, x22, [sp, #32]

    // Move parameters to callee-saved registers so they survive
    // the stack setup and register loading below.
    mov     x19, x0             // x19 = fn
    mov     x20, x1             // x20 = args pointer
    mov     x21, x2             // x21 = nargs

    // Calculate and reserve stack space for args 9+.
    // nStackArgs = max(0, nargs - 8)
    // Stack space = align16(nStackArgs * 8)
    subs    x22, x21, #8
    b.le    .Lno_stack_args

    // Round up to 16-byte alignment: ((nStackArgs * 8) + 15) & ~15
    lsl     x3, x22, #3        // nStackArgs * 8
    add     x3, x3, #15
    and     x3, x3, #~0xF
    sub     sp, sp, x3

    // Copy stack args: args[8..nargs-1] -> sp[0..]
    add     x6, x20, #64       // x6 = &args[8]
    mov     x4, #0              // index = 0
.Lcopy_stack:
    ldr     x5, [x6, x4, lsl #3]
    str     x5, [sp, x4, lsl #3]
    add     x4, x4, #1
    cmp     x4, x22
    b.lt    .Lcopy_stack

.Lno_stack_args:
    // Load register arguments x0-x7 from the args array.
    // Zero all registers first, then load as many as nargs specifies.
    mov     x0, #0
    mov     x1, #0
    mov     x2, #0
    mov     x3, #0
    mov     x4, #0
    mov     x5, #0
    mov     x6, #0
    mov     x7, #0

    cbz     x21, .Lcall         // nargs == 0, skip loading
    ldr     x0, [x20, #0]       // args[0]
    cmp     x21, #2
    b.lt    .Lcall
    ldr     x1, [x20, #8]       // args[1]
    cmp     x21, #3
    b.lt    .Lcall
    ldr     x2, [x20, #16]      // args[2]
    cmp     x21, #4
    b.lt    .Lcall
    ldr     x3, [x20, #24]      // args[3]
    cmp     x21, #5
    b.lt    .Lcall
    ldr     x4, [x20, #32]      // args[4]
    cmp     x21, #6
    b.lt    .Lcall
    ldr     x5, [x20, #40]      // args[5]
    cmp     x21, #7
    b.lt    .Lcall
    ldr     x6, [x20, #48]      // args[6]
    cmp     x21, #8
    b.lt    .Lcall
    ldr     x7, [x20, #56]      // args[7]

.Lcall:
    blr     x19                  // Call UEFI function, result in x0

    // Restore stack and callee-saved registers.
    mov     sp, x29
    ldp     x19, x20, [sp, #16]
    ldp     x21, x22, [sp, #32]
    ldp     x29, x30, [sp], #64
    ret
