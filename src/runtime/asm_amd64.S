#ifdef __ELF__
.section .text.tinygo_scanCurrentStack
.global tinygo_scanCurrentStack
tinygo_scanCurrentStack:
#else // Darwin
.global _tinygo_scanCurrentStack
_tinygo_scanCurrentStack:
#endif
    // Save callee-saved registers.
    pushq %rbx
    pushq %rbp
    pushq %r12
    pushq %r13
    pushq %r14
    pushq %r15

    // Scan the stack.
    subq $8, %rsp // adjust the stack before the call to maintain 16-byte alignment
    movq %rsp, %rdi
#ifdef __ELF__
    callq tinygo_scanstack
#else
    callq _tinygo_scanstack // Darwin
#endif

    // Restore the stack pointer. Registers do not need to be restored as they
    // were only pushed to be discoverable by the GC.
    addq $56, %rsp
    retq


#ifdef __ELF__
.section .text.tinygo_longjmp
.global tinygo_longjmp
tinygo_longjmp:
#else // Darwin
.global _tinygo_longjmp
_tinygo_longjmp:
#endif
    // Note: the code we jump to assumes rax is set to a non-zero value if we
    // jump from here, so we use rax as the temporary value for jumpPC.
    movq 0(%rdi), %rsp // jumpSP
    movq 8(%rdi), %rax // jumpPC
    jmpq *%rax


#ifdef __MACH__ // Darwin
// allow these symbols to stripped as dead code
.subsections_via_symbols
#endif
