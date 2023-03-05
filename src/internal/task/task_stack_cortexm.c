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