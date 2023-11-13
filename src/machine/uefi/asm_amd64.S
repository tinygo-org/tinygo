.section .text

.global uefiCall0
uefiCall0:
    pushq  %rbp
    movq   %rsp, %rbp
    callq  *%rcx
    movq   %rbp, %rsp
    popq   %rbp
    ret

.global uefiCall1
uefiCall1:
    pushq  %rbp
    movq   %rsp, %rbp
    movq   %rcx, %rax // rax = fn
    movq   %rdx, %rcx
    callq  *%rax
    movq   %rbp, %rsp
    popq   %rbp
    ret

.global uefiCall2
uefiCall2:
    pushq  %rbp
    movq   %rsp, %rbp
    movq   %rcx, %rax // rax = fn
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
    movq   %rcx, %rax // rax = fn
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
    movq   %rcx, %rax // rax = fn
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
    subq   $0x30,%rsp
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
    subq   $0x30,%rsp
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
