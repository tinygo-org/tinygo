.section .text.tinygo_scanCurrentStack,"ax"
.global tinygo_scanCurrentStack
tinygo_scanCurrentStack:
    // Save callee-saved registers.
    pushq %rbx
    pushq %rbp
    pushq %rdi
    pushq %rsi
    pushq %r12
    pushq %r13
    pushq %r14
    pushq %r15

    // Scan the stack.
    subq $8, %rsp // adjust the stack before the call to maintain 16-byte alignment
    movq %rsp, %rdi
    callq tinygo_scanstack

    // Restore the stack pointer. Registers do not need to be restored as they
    // were only pushed to be discoverable by the GC.
    addq $72, %rsp
    retq
