.section .text

.global asmPause
asmPause:
    pause
    ret

.global asmReadRdtsc
asmReadRdtsc:
    rdtsc
    shlq  $0x20, %rdx
    orq   %rdx, %rax
    ret

.global asmCpuid
asmCpuid:
    pushq  %rbx

    mov    %ecx, %eax
    pushq  %rax

    pushq  %rdx
    cpuid

    test   %r9, %r9
    jz     .SkipEcx
    mov    %ecx, (%r9)
.SkipEcx:
    popq   %rcx
    jrcxz  .SkipEax
    mov    %eax, (%rcx)
.SkipEax:
    mov    %r8, %rcx
    jrcxz  .SkipEbx
    mov    %ebx, (%rcx)
.SkipEbx:
    popq   %rax
    popq   %rbx
    ret
