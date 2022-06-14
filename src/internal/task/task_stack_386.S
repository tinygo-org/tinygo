.section .text.tinygo_startTask
.global  tinygo_startTask
.type    tinygo_startTask, %function
tinygo_startTask:
    .cfi_startproc
    // Small assembly stub for starting a goroutine. This is already run on the
    // new stack, with the callee-saved registers already loaded.
    // Most importantly, EBX contain the pc of the to-be-started function and
    // ESI contain the only argument it is given. Multiple arguments are packed
    // into one by storing them in a new allocation.

    // Indicate to the unwinder that there is nothing to unwind, this is the
    // root frame. It avoids bogus extra frames in GDB.
    .cfi_undefined eip

    // Set the first argument of the goroutine start wrapper, which contains all
    // the arguments.
    pushl %esi

    // Branch to the "goroutine start" function.
    calll *%ebx

    // After return, exit this goroutine.
    calll tinygo_pause
    .cfi_endproc

.global tinygo_swapTask
.type tinygo_swapTask, %function
tinygo_swapTask:
    // This function gets the following parameters:
    movl 4(%esp), %eax // newStack uintptr
    movl 8(%esp), %ecx // oldStack *uintptr
    // More information on the calling convention:
    // https://wiki.osdev.org/System_V_ABI#i386

    // Save all callee-saved registers:
    pushl %ebp
    pushl %edi
    pushl %esi
    pushl %ebx

    // Save the current stack pointer in oldStack.
    movl %esp, (%ecx)

    // Switch to the new stack pointer.
    movl %eax, %esp

    // Load saved register from the new stack.
    popl %ebx
    popl %esi
    popl %edi
    popl %ebp

    // Return into the new task, as if tinygo_swapTask was a regular call.
    ret
