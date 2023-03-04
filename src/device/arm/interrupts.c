#include <stdint.h>

void EnableInterrupts(uintptr_t mask) {
    asm volatile(
        "msr PRIMASK, %0"
        :
        : "r"(mask)
        : "memory"
    );
}

uintptr_t DisableInterrupts() {
    uintptr_t mask;
    asm volatile(
        "mrs %0, PRIMASK\n\t"
        "cpsid i"
        : "=r"(mask)
        :
        : "memory"
    );
	return mask;
}