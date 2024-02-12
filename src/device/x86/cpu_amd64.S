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

// @param  Index         The 32-bit value to load into EAX prior to invoking the CPUID
//                       instruction.
// @param  RegisterEax   A pointer to the 32-bit EAX value returned by the CPUID
//                       instruction. This is an optional parameter that may be NULL.
// @param  RegisterEbx   A pointer to the 32-bit EBX value returned by the CPUID
//                       instruction. This is an optional parameter that may be NULL.
// @param  RegisterEcx   A pointer to the 32-bit ECX value returned by the CPUID
//                       instruction. This is an optional parameter that may be NULL.
// @param  RegisterEdx   A pointer to the 32-bit EDX value returned by the CPUID
//                       instruction. This is an optional parameter that may be NULL.
// @return Index.
.global asmCpuid
asmCpuid:
    // rcx = Index
    // rdx = RegisterEax
    // r8  = RegisterEbx
    // r9  = RegisterEcx
    // rsp + 0x28 = RegisterEdx
    pushq  %rbx

    mov    %ecx, %eax // eax <- index
    pushq  %rax       // save Index on stack

    pushq  %rdx // RegisterEax
    cpuid

    test    %r9, %r9 // RegisterEcx
    jz      .SkipEcx
    mov     %ecx, (%r9)
.SkipEcx:
    popq     %rcx // RegisterEax
    jrcxz   .SkipEax
    mov     %eax, (%rcx)
.SkipEax:
    mov     %r8, %rcx // RegisterEbx
    jrcxz   .SkipEbx
    mov     %ebx, (%rcx)
.SkipEbx:
    mov     0x38(%rsp), %rcx // 0x28 + 0x10
    jrcxz   .SkipEdx
    mov     %edx, (%rcx)
.SkipEdx:
    popq     %rax // restore Index to rax as return value
    popq     %rbx
    ret
