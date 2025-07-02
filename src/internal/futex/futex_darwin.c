//go:build none

// This file is manually included, to avoid CGo which would cause a circular
// import.

#include <stdint.h>

// This API isn't documented by Apple, but it is used by LLVM libc++ (so should
// be stable) and has been documented extensively here:
// https://outerproduct.net/futex-dictionary.html

int __ulock_wait(uint32_t operation, void *addr, uint64_t value, uint32_t timeout_us);
int __ulock_wait2(uint32_t operation, void *addr, uint64_t value, uint64_t timeout_ns, uint64_t value2);
int __ulock_wake(uint32_t operation, void *addr, uint64_t wake_value);

// Operation code.
#define UL_COMPARE_AND_WAIT 1

// Flags to the operation value.
#define ULF_WAKE_ALL 0x00000100
#define ULF_NO_ERRNO 0x01000000

void tinygo_futex_wait(uint32_t *addr, uint32_t cmp) {
    __ulock_wait(UL_COMPARE_AND_WAIT|ULF_NO_ERRNO, addr, (uint64_t)cmp, 0);
}

void tinygo_futex_wait_timeout(uint32_t *addr, uint32_t cmp, uint64_t timeout) {
    // Make sure that an accidental use of a zero timeout is not treated as an
    // infinite timeout. Return if it's zero since it wouldn't be waiting for
    // any significant time anyway.
    // Probably unnecessary, but guards against potential bugs.
    if (timeout == 0) {
        return;
    }

    // Note: __ulock_wait2 is available since MacOS 11.
    // I think that's fine, since the version before that (MacOS 10.15) is EOL
    // since 2022. Though if needed, we could certainly use __ulock_wait instead
    // and deal with the smaller timeout value.
    __ulock_wait2(UL_COMPARE_AND_WAIT|ULF_NO_ERRNO, addr, (uint64_t)cmp, timeout, 0);
}

void tinygo_futex_wake(uint32_t *addr) {
    __ulock_wake(UL_COMPARE_AND_WAIT|ULF_NO_ERRNO, addr, 0);
}

void tinygo_futex_wake_all(uint32_t *addr) {
    __ulock_wake(UL_COMPARE_AND_WAIT|ULF_NO_ERRNO|ULF_WAKE_ALL, addr, 0);
}
