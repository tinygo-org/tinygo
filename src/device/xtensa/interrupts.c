#include <stdint.h>

void EnableInterrupts(uintptr_t mask) {
    // Wait for the PS write to complete, as XTOS_RESTORE_INTLEVEL does.
    // https://github.com/espressif/esp-idf/blob/master/components/xtensa/include/xtensa/xtruntime.h
    asm volatile(
        "wsr %0, PS\n\t"
        "rsync"
        :
        : "r"(mask)
        : "memory"
    );
}

uintptr_t DisableInterrupts() {
    uintptr_t mask;
    asm volatile(
        "rsil %0, 15"
        : "=r"(mask)
        :
        : "memory"
    );
	return mask;
}
