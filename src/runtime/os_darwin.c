//go:build none

// This file is included in the build, despite the //go:build line above.

#include <fcntl.h>

// Wrapper function because 'open' is a variadic function and variadic functions
// use a different (incompatible) calling convention on darwin/arm64.
// This function is referenced from the compiler, when it sees a
// syscall.libc_open_trampoline function.
int syscall_libc_open(const char *pathname, int flags, mode_t mode) {
    return open(pathname, flags, mode);
}

// The following functions are called by the runtime because Go can't call
// function pointers directly.

int tinygo_syscall(int (*fn)(uintptr_t a1, uintptr_t a2, uintptr_t a3), uintptr_t a1, uintptr_t a2, uintptr_t a3) {
    return fn(a1, a2, a3);
}

uintptr_t tinygo_syscallX(uintptr_t (*fn)(uintptr_t a1, uintptr_t a2, uintptr_t a3), uintptr_t a1, uintptr_t a2, uintptr_t a3) {
    return fn(a1, a2, a3);
}

int tinygo_syscall6(int (*fn)(uintptr_t a1, uintptr_t a2, uintptr_t a3, uintptr_t a4, uintptr_t a5, uintptr_t a6), uintptr_t a1, uintptr_t a2, uintptr_t a3, uintptr_t a4, uintptr_t a5, uintptr_t a6) {
    return fn(a1, a2, a3, a4, a5, a6);
}

uintptr_t tinygo_syscall6X(uintptr_t (*fn)(uintptr_t a1, uintptr_t a2, uintptr_t a3, uintptr_t a4, uintptr_t a5, uintptr_t a6), uintptr_t a1, uintptr_t a2, uintptr_t a3, uintptr_t a4, uintptr_t a5, uintptr_t a6) {
    return fn(a1, a2, a3, a4, a5, a6);
}
