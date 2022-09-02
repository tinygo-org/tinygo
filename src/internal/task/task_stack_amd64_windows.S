// Windows on amd64 has a slightly different ABI than other (*nix) systems.
// Therefore, assembly functions need to be tweaked slightly.
//
// The calling convention is described here:
// https://docs.microsoft.com/en-us/cpp/build/x64-calling-convention?view=msvc-170

.section .text.tinygo_startTask,"ax"
.global  tinygo_startTask
tinygo_startTask:
    // Small assembly stub for starting a goroutine. This is already run on the
    // new stack, with the callee-saved registers already loaded.
    // Most importantly, r12 contain the pc of the to-be-started function and
    // r13 contain the only argument it is given. Multiple arguments are packed
    // into one by storing them in a new allocation.

    // Set the first argument of the goroutine start wrapper, which contains all
    // the arguments.
    movq %r13, %rcx

    // Branch to the "goroutine start" function.
    callq *%r12

    // After return, exit this goroutine.
    // This has to be a call, not a jump, to keep the stack correctly aligned.
    callq tinygo_pause

.global tinygo_swapTask
.section .text.tinygo_swapTask,"ax"
tinygo_swapTask:
    // This function gets the following parameters:
    // %rcx = newStack uintptr
    // %rdx = oldStack *uintptr

    // Save all callee-saved registers:
    pushq %r15
    pushq %r14
    pushq %r13
    pushq %r12
    pushq %rsi
    pushq %rdi
    pushq %rbp
    sub   $160, %rsp
    movaps  %xmm6, 144(%rsp)
    movaps  %xmm7, 128(%rsp)
    movaps  %xmm8, 112(%rsp)
    movaps  %xmm9,  96(%rsp)
    movaps  %xmm10, 80(%rsp)
    movaps  %xmm11, 64(%rsp)
    movaps  %xmm12, 48(%rsp)
    movaps  %xmm13, 32(%rsp)
    movaps  %xmm14, 16(%rsp)
    movaps  %xmm15,  0(%rsp)
    pushq %rbx

    // Save the current stack pointer in oldStack.
    movq %rsp, (%rdx)

    // Switch to the new stack pointer.
    movq %rcx, %rsp

    // Load saved register from the new stack.
    popq %rbx
    movaps 0(%rsp),   %xmm15
    movaps 16(%rsp),  %xmm14
    movaps 32(%rsp),  %xmm13
    movaps 48(%rsp),  %xmm12
    movaps 64(%rsp),  %xmm11
    movaps 80(%rsp),  %xmm10
    movaps 96(%rsp),  %xmm9
    movaps 112(%rsp), %xmm8
    movaps 128(%rsp), %xmm7
    movaps 144(%rsp), %xmm6
    add   $160, %rsp
    popq %rbp
    popq %rdi
    popq %rsi
    popq %r12
    popq %r13
    popq %r14
    popq %r15

    // Return into the new task, as if tinygo_swapTask was a regular call.
    ret
