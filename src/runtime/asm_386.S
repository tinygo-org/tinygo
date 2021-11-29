.section .text.tinygo_scanCurrentStack
.global tinygo_scanCurrentStack
.type tinygo_scanCurrentStack, %function
tinygo_scanCurrentStack:
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
    calll tinygo_scanstack

    // Restore the stack pointer. Registers do not need to be restored as they
    // were only pushed to be discoverable by the GC.
    addl $28, %esp
    retl


.section .text.tinygo_longjmp
.global tinygo_longjmp
tinygo_longjmp:
    // Note: the code we jump to assumes eax is set to a non-zero value if we
    // jump from here.
    movl 8(%esp), %eax // jumpPC (stash in volatile register)
    movl 4(%esp), %esp // jumpSP
    jmpl *%eax
