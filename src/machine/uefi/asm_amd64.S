.section .text

.global uefiCall0
uefiCall0:
    pushq  %rbp
    movq   %rsp, %rbp
    subq   $0x20, %rsp  // Shadow space of UEFI call
    callq  *%rcx
    movq   %rbp, %rsp
    popq   %rbp
    ret

.global uefiCall1
uefiCall1:
    pushq  %rbp
    movq   %rsp, %rbp
    subq   $0x20, %rsp  // Shadow space of UEFI call
    movq   %rcx, %rax   // rax = fn
    movq   %rdx, %rcx
    callq  *%rax
    movq   %rbp, %rsp
    popq   %rbp
    ret

.global uefiCall2
uefiCall2:
    pushq  %rbp
    movq   %rsp, %rbp
    subq   $0x20, %rsp  // Shadow space of UEFI call
    movq   %rcx, %rax   // rax = fn
    movq   %rdx, %rcx
    movq   %r8,  %rdx
    callq  *%rax
    movq   %rbp, %rsp
    popq   %rbp
    ret

.global uefiCall3
uefiCall3:
    pushq  %rbp
    movq   %rsp, %rbp
    subq   $0x20, %rsp  // Shadow space of UEFI call
    movq   %rcx, %rax   // rax = fn
    movq   %rdx, %rcx
    movq   %r8,  %rdx
    movq   %r9,  %r8
    callq  *%rax
    movq   %rbp, %rsp
    popq   %rbp
    ret

.global uefiCall4
uefiCall4:
    pushq  %rbp
    movq   %rsp, %rbp
    subq   $0x20, %rsp  // Shadow space of UEFI call
    movq   %rcx, %rax   // rax = fn
    movq   %rdx, %rcx
    movq   %r8,  %rdx
    movq   %r9,  %r8

    movq   0x30(%rbp),%r9 // 0x08(return_address) + 0x08(pushq rbp) + 0x20

    callq  *%rax
    movq   %rbp, %rsp
    popq   %rbp
    ret

.global uefiCall5
uefiCall5:
    pushq  %rbp
    movq   %rsp, %rbp
    subq   $0x28,%rsp // 0x20 shadow + 1×8 spill
    movq   %rcx, %rax // rax = fn
    movq   %rdx, %rcx // a
    movq   %r8,  %rdx // b
    movq   %r9,  %r8  // c
    movq   0x30(%rbp),%r9 // d
    movq   0x38(%rbp),%r10 // e
    movq   %r10,0x20(%rsp)
    callq  *%rax
    movq   %rbp, %rsp
    popq   %rbp
    ret

.global uefiCall6
uefiCall6:
    pushq  %rbp
    movq   %rsp, %rbp
    subq   $0x28,%rsp // 0x20 shadow + 2×8 spill
    movq   %rcx, %rax // rax = fn
    movq   %rdx, %rcx // a
    movq   %r8,  %rdx // b
    movq   %r9,  %r8  // c
    movq   0x30(%rbp),%r9 // d
    movq   0x38(%rbp),%r10 // e
    movq   %r10,0x20(%rsp)
    movq   0x40(%rbp),%r10 // f
    movq   %r10,0x28(%rsp)
    callq  *%rax
    movq   %rbp, %rsp
    popq   %rbp
    ret

.global uefiCall7
uefiCall7:
    pushq   %rbp
    movq    %rsp, %rbp
    subq    $0x38, %rsp            // 0x20 shadow + 3*8 spill = 0x38

    movq    %rcx, %rax             // RAX ← fn ptr
    movq    %rdx, %rcx             // RCX ← a
    movq    %r8,  %rdx             // RDX ← b
    movq    %r9,  %r8              // R8  ← c
    movq    0x30(%rbp), %r9        // R9  ← d

    movq    0x38(%rbp), %r10       // e
    movq    %r10, 0x20(%rsp)

    movq    0x40(%rbp), %r10       // f
    movq    %r10, 0x28(%rsp)

    movq    0x48(%rbp), %r10       // g
    movq    %r10, 0x30(%rsp)

    callq   *%rax                  // jump into firmware

    movq    %rbp, %rsp             // tear-down
    popq    %rbp
    ret

.global uefiCall8
uefiCall8:
    pushq   %rbp
    movq    %rsp, %rbp
    subq    $0x40, %rsp            // 0x20 shadow + 4*8 spill = 0x40

    movq    %rcx, %rax             // RAX ← fn ptr
    movq    %rdx, %rcx             // RCX ← a
    movq    %r8,  %rdx             // RDX ← b
    movq    %r9,  %r8              // R8  ← c
    movq    0x30(%rbp), %r9        // R9  ← d

    movq    0x38(%rbp), %r10       // e
    movq    %r10, 0x20(%rsp)

    movq    0x40(%rbp), %r10       // f
    movq    %r10, 0x28(%rsp)

    movq    0x48(%rbp), %r10       // g
    movq    %r10, 0x30(%rsp)

    movq    0x50(%rbp), %r10       // h
    movq    %r10, 0x38(%rsp)

    callq   *%rax                  // jump into firmware

    movq    %rbp, %rsp             // tear-down
    popq    %rbp
    ret

.global uefiCall9
uefiCall9:
    pushq   %rbp
    movq    %rsp, %rbp
    subq    $0x48, %rsp            // 0x20 shadow + 5*8 spill = 0x48

    movq    %rcx, %rax             // RAX ← fn ptr
    movq    %rdx, %rcx             // RCX ← a
    movq    %r8,  %rdx             // RDX ← b
    movq    %r9,  %r8              // R8  ← c
    movq    0x30(%rbp), %r9        // R9  ← d

    movq    0x38(%rbp), %r10       // e
    movq    %r10, 0x20(%rsp)

    movq    0x40(%rbp), %r10       // f
    movq    %r10, 0x28(%rsp)

    movq    0x48(%rbp), %r10       // g
    movq    %r10, 0x30(%rsp)

    movq    0x50(%rbp), %r10       // h
    movq    %r10, 0x38(%rsp)

    movq    0x58(%rbp), %r10       // i
    movq    %r10, 0x40(%rsp)

    callq   *%rax                  // jump into firmware

    movq    %rbp, %rsp             // tear-down
    popq    %rbp
    ret

.global uefiCall10
uefiCall10:
    pushq   %rbp
    movq    %rsp, %rbp
    subq    $0x50, %rsp            // 0x20 shadow + 6*8 spill = 0x50

    movq    %rcx, %rax             // RAX ← fn ptr
    movq    %rdx, %rcx             // RCX ← a
    movq    %r8,  %rdx             // RDX ← b
    movq    %r9,  %r8              // R8  ← c
    movq    0x30(%rbp), %r9        // R9  ← d

    movq    0x38(%rbp), %r10       // e
    movq    %r10, 0x20(%rsp)

    movq    0x40(%rbp), %r10       // f
    movq    %r10, 0x28(%rsp)

    movq    0x48(%rbp), %r10       // g
    movq    %r10, 0x30(%rsp)

    movq    0x50(%rbp), %r10       // h
    movq    %r10, 0x38(%rsp)

    movq    0x58(%rbp), %r10       // i
    movq    %r10, 0x40(%rsp)

    movq    0x60(%rbp), %r10       // j
    movq    %r10, 0x48(%rsp)

    callq   *%rax                  // jump into firmware

    movq    %rbp, %rsp             // tear-down
    popq    %rbp
    ret

// MSVC-compatible __chkstk implementation for x86_64
// This touches each page on the stack to ensure it is committed.

.global ___chkstk_ms
___chkstk_ms:
	ret    // Rather than protect the stack, this seems to cause a fault 🤷
    mov    %rcx, %r10              // Save the stack allocation size
    mov    %rsp, %rax              // Current stack pointer
    sub    %r10, %rax              // Target stack pointer
    cmp    %rax, %rsp              // Already below limit?
    jae    .Ldone

.Lloop:
    sub    $0x1000, %rsp           // Page size (4KB)
    movb   $0, (%rsp)              // Touch page
    cmp    %rsp, %rax              // Reached target?
    jne    .Lloop

.Ldone:
    mov    %rax, %rsp              // Update stack pointer
    ret
