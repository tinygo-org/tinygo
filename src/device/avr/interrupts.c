#include <stdint.h>

// SREG is I/O address 0x3f, see SREG in avr-libc include/avr/common.h.
// https://github.com/avrdudes/avr-libc/blob/main/include/avr/common.h

void EnableInterrupts(uint8_t sreg) {
    asm volatile(
        "out 0x3f, %0"
        :
        : "r"(sreg)
        : "memory"
    );
}

uint8_t DisableInterrupts() {
    uint8_t sreg;
    asm volatile(
        "in %0, 0x3f\n\t"
        "cli"
        : "=r"(sreg)
        :
        : "memory"
    );
    return sreg;
}
