//go:build (scheduler.tasks || scheduler.cores) && cortexm
#include <stdint.h>

uintptr_t SystemStack() {
    uintptr_t sp;
    asm volatile(
        "mrs %0, MSP"
        : "=r"(sp)
        :
        : "memory"
    );
    return sp;
}

uintptr_t GoroutineStack() {
    uintptr_t sp;
    asm volatile(
        "mrs %0, PSP"
        : "=r"(sp)
        :
        : "memory"
    );
    return sp;
}
