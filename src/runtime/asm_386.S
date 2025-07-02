#ifdef _WIN32
.global _tinygo_scanCurrentStack
_tinygo_scanCurrentStack:
#else
.section .text.tinygo_scanCurrentStack
.global tinygo_scanCurrentStack
.type tinygo_scanCurrentStack, %function
tinygo_scanCurrentStack:
#endif
    // Sources:
    //   * https://stackoverflow.com/questions/18024672/what-registers-are-preserved-through-a-linux-x86-64-function-call
    //   * https://godbolt.org/z/q7e8dn

    // Save callee-saved registers.
    pushl %ebx
    pushl %esi
    pushl %edi
    pushl %ebp

    // Scan the stack.
    subl $8, %esp // adjust the stack before the call to maintain 16-byte alignment
    pushl %esp
    #ifdef _WIN32
    calll _tinygo_scanstack
    #else
    calll tinygo_scanstack
    #endif

    // Restore the stack pointer. Registers do not need to be restored as they
    // were only pushed to be discoverable by the GC.
    addl $28, %esp
    retl


#ifdef _WIN32
.global _tinygo_longjmp
_tinygo_longjmp:
#else
.section .text.tinygo_longjmp
.global tinygo_longjmp
tinygo_longjmp:
#endif
    // Note: the code we jump to assumes eax is set to a non-zero value if we
    // jump from here.
    movl 4(%esp), %eax
    movl 0(%eax), %esp // jumpSP
    movl 4(%eax), %eax // jumpPC (stash in volatile register)
    jmpl *%eax
