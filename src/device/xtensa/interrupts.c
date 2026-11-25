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

uint32_t ReadINTENABLE() {
    uint32_t val;
    asm volatile(
        "rsr %0, INTENABLE"
        : "=r"(val)
    );
    return val;
}

void WriteINTENABLE(uint32_t val) {
    // Wait for the INTENABLE write to complete, as xt_ints_on and xt_ints_off do.
    // https://github.com/espressif/esp-idf/blob/master/components/xtensa/xtensa_intr_asm.S
    asm volatile(
        "wsr %0, INTENABLE\n\t"
        "rsync"
        :
        : "r"(val)
        : "memory"
    );
}

uint32_t SetINTENABLEBits(uint32_t mask) {
    // INTENABLE is 0 between xsr and wsr, so no interrupt can change it, as in xt_ints_on.
    // https://github.com/espressif/esp-idf/blob/master/components/xtensa/xtensa_intr_asm.S
    uint32_t old;
    uint32_t val;
    asm volatile(
        "movi %0, 0\n\t"
        "xsr %0, INTENABLE\n\t"
        "rsync\n\t"
        "or %1, %0, %2\n\t"
        "wsr %1, INTENABLE\n\t"
        "rsync"
        : "=&r"(old), "=&r"(val)
        : "r"(mask)
        : "memory"
    );
    return old;
}

uint32_t ReadINTERRUPT() {
    uint32_t val;
    asm volatile(
        "rsr %0, INTERRUPT"
        : "=r"(val)
    );
    return val;
}

void WriteINTCLEAR(uint32_t val) {
    asm volatile(
        "wsr %0, INTCLEAR"
        :
        : "r"(val)
        : "memory"
    );
}

void WriteVECBASE(uintptr_t val) {
    asm volatile(
        "wsr %0, VECBASE"
        :
        : "r"(val)
        : "memory"
    );
}

uintptr_t ReadPS() {
    uintptr_t val;
    asm volatile(
        "rsr %0, PS"
        : "=r"(val)
    );
    return val;
}
