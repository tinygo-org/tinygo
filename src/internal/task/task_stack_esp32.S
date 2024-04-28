//go:build tinygo

.section .text.tinygo_startTask,"ax",@progbits
.global  tinygo_startTask
.type    tinygo_startTask, %function
tinygo_startTask:
    // Small assembly stub for starting a goroutine. This already runs on the
    // new stack, control reaches this function after returning from the initial
    // tinygo_swapTask below (the retw.n instruction).
    //
    // The stack was set up in such a way that it looks as if this function was
    // paused using tinygo_swapTask by setting up the parent register window and
    // return pointer as a call4 instruction - except such a call never took
    // place. Instead, the stack pointer is switched to the new stack after all
    // live-but-invisible registers have been flushed to the stack. This means
    // that all registers as present in tinygo_swapTask are moved four up (a2 in
    // tinygo_swapTask is a6 in this function). We don't use any of those
    // registers however. Instead, the retw.n instruction will load them through
    // an underflow exception from the stack which means we get a0-a3 as defined
    // in task_stack_esp32.go.

    // Branch to the "goroutine start" function. The first (and only) parameter
    // is stored in a2, but has to be moved to a6 to make it appear as a2 in the
    // goroutine start function (due to changing the register window by four
    // with callx4).
    mov.n a6, a2
    callx4 a3

    // After return, exit this goroutine. This call never returns.
    call4  tinygo_pause

.section .text.tinygo_swapTask,"ax",@progbits
.global tinygo_swapTask
.type tinygo_swapTask, %function
tinygo_swapTask:
    // This function gets the following parameters:
    // a2 = newStack uintptr
    // a3 = oldStack *uintptr

    // Reserve 32 bytes on the stack. It really needs to be 32 bytes, with 16
    // extra at the bottom to adhere to the ABI.
    entry sp, 32

    // Disable interrupts while flushing registers. This is necessary because
    // interrupts might want to use the stack pointer (at a2) which will be some
    // arbitrary register while registers are flushed.
    rsil a4, 3 // XCHAL_EXCM_LEVEL

    // Flush all unsaved registers to the stack.
    // This trick has been borrowed from the Zephyr project:
    // https://github.com/zephyrproject-rtos/zephyr/blob/d79b003758/arch/xtensa/include/xtensa-asm2-s.h#L17
    and a12, a12, a12
    rotw 3
    and a12, a12, a12
    rotw 3
    and a12, a12, a12
    rotw 3
    and a12, a12, a12
    rotw 3
    and a12, a12, a12
    rotw 4

    // Restore interrupts.
    wsr.ps a4

    // At this point, the following is true:
    //     WindowStart == 1 << WindowBase
    // Therefore, we don't need to do this manually.
    // It also means that the stack pointer can now be safely modified.

    // Save a0, which stores the return address and the parent register window
    // in the upper two bits.
    s32i.n a0, sp, 0

    // Save the current stack pointer in oldStack.
    s32i.n  sp, a3, 0

    // Switch to the new stack pointer (newStack).
    mov.n   sp, a2

    // Load a0, which is the previous return address from before the previous
    // switch or the constructed return address to tinygo_startTask. This
    // register also stores the parent register window.
    l32i.n a0, sp, 0

    // Return into the new stack. This instruction will trigger a window
    // underflow, reloading the saved registers from the stack.
    retw.n
