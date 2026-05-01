.section .text

.global uefiCall0
uefiCall0:
    pushq  %rbp
    movq   %rsp, %rbp
    subq   $0x20, %rsp
    callq  *%rcx
    movq   %rbp, %rsp
    popq   %rbp
    ret

.global uefiCall1
uefiCall1:
    pushq  %rbp
    movq   %rsp, %rbp
    subq   $0x20, %rsp
    movq   %rcx, %rax
    movq   %rdx, %rcx
    callq  *%rax
    movq   %rbp, %rsp
    popq   %rbp
    ret

.global uefiCall2
uefiCall2:
    pushq  %rbp
    movq   %rsp, %rbp
    subq   $0x20, %rsp
    movq   %rcx, %rax
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
    subq   $0x20, %rsp
    movq   %rcx, %rax
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
    subq   $0x20, %rsp
    movq   %rcx, %rax
    movq   %rdx, %rcx
    movq   %r8,  %rdx
    movq   %r9,  %r8
    movq   0x30(%rbp), %r9
    callq  *%rax
    movq   %rbp, %rsp
    popq   %rbp
    ret

.global uefiCall5
uefiCall5:
    pushq  %rbp
    movq   %rsp, %rbp
    subq   $0x30, %rsp
    movq   %rcx, %rax
    movq   %rdx, %rcx
    movq   %r8,  %rdx
    movq   %r9,  %r8
    movq   0x30(%rbp), %r9
    movq   0x38(%rbp), %r10
    movq   %r10, 0x20(%rsp)
    callq  *%rax
    movq   %rbp, %rsp
    popq   %rbp
    ret

.global ___chkstk_ms
___chkstk_ms:
    ret
