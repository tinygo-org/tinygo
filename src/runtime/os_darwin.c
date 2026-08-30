//go:build none

// This file is included in the build, despite the //go:build line above.

#include <fcntl.h>
#include <stdint.h>

// Wrapper function because 'open' is a variadic function and variadic functions
// use a different (incompatible) calling convention on darwin/arm64.
// This function is referenced from the compiler, when it sees a
// syscall.libc_open_trampoline function.
int syscall_libc_open(const char *pathname, int flags, mode_t mode) {
    return open(pathname, flags, mode);
}

// Wrapper function for 'fcntl', whose third parameter is variadic as well.
// A call through a plain three-argument function pointer makes the callee read
// that argument from the stack, which holds an unrelated value.
//
// The argument is a uintptr_t so that the pointer commands reached through
// syscall.fcntlPtr use the same wrapper. Both spellings share
// libc_fcntl_trampoline, and on a little-endian target the int commands read
// the low half of the same stack slot.
int syscall_libc_fcntl(int fd, int cmd, uintptr_t arg) {
    return fcntl(fd, cmd, arg);
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
