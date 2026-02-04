// UEFI Call Stub for x86_64
//
// Generic stub for calling any UEFI function via the Microsoft x64 ABI.
//
// Go signature: func callAsm(fn uintptr, args *uintptr, nargs uintptr) Status
// MS x64 entry:
//   RCX = fn (function pointer to call)
//   RDX = args (pointer to uintptr array of arguments)
//   R8  = nargs (number of arguments, 0-N)
//
// MS x64 ABI register mapping for the callee:
//   Args 1-4:  RCX, RDX, R8, R9
//   Args 5+:   stack (after 32-byte shadow space)
//   Return:    RAX
//   Stack:     16-byte aligned before CALL
//

.section .text.uefiCall,"ax"
.global uefiCall
uefiCall:
    pushq %rbp
    movq %rsp, %rbp
    pushq %rsi
    pushq %rdi

    movq %rcx, %r10         // Save fn
    movq %rdx, %rsi         // Save args pointer
    movq %r8, %rdi          // Save nargs

    // Calculate stack space: align16(32 + max(0, nargs-4) * 8)
    // After 3 pushes, RSP is 16-byte aligned. Subtracting an aligned
    // value keeps it aligned for the CALL instruction.
    xorq %rax, %rax
    cmpq $4, %rdi
    jle 1f
    movq %rdi, %rax
    subq $4, %rax
1:
    shlq $3, %rax            // nStackArgs * 8
    addq $47, %rax           // + 32 (shadow) + 15 (round-up)
    andq $-16, %rax          // align to 16
    subq %rax, %rsp

    // Copy stack args: args[4..nargs-1] -> RSP+32+j*8
    movq %rdi, %rcx
    subq $4, %rcx
    jle 2f
    xorq %rax, %rax
3:  movq 32(%rsi,%rax,8), %r11
    movq %r11, 32(%rsp,%rax,8)
    incq %rax
    cmpq %rcx, %rax
    jl 3b
2:
    // Load register args
    xorq %rcx, %rcx
    xorq %rdx, %rdx
    xorq %r8, %r8
    xorq %r9, %r9

    testq %rdi, %rdi
    jz 4f
    movq 0(%rsi), %rcx       // args[0] -> RCX
    cmpq $2, %rdi
    jl 4f
    movq 8(%rsi), %rdx       // args[1] -> RDX
    cmpq $3, %rdi
    jl 4f
    movq 16(%rsi), %r8       // args[2] -> R8
    cmpq $4, %rdi
    jl 4f
    movq 24(%rsi), %r9       // args[3] -> R9
4:
    callq *%r10

    leaq -16(%rbp), %rsp
    popq %rdi
    popq %rsi
    popq %rbp
    retq
