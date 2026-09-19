//go:build none

// This file is included in the build, despite the //go:build line above.

#include <fcntl.h>

// sys/ioctl.h is not in lib/macos-minimal-sdk. Declare ioctl here.
extern int ioctl(int fd, unsigned long request, ...);

// Fixed-signature wrappers for the variadic libc imports. Variadic functions
// take stack arguments on darwin/arm64. See darwinVariadicImports in
// compiler/syscall.go.

int syscall_libc_open(uintptr_t pathname, uintptr_t flags, uintptr_t mode) {
    return open((const char *)pathname, (int)flags, (mode_t)mode);
}

int syscall_libc_ioctl(uintptr_t fd, uintptr_t request, uintptr_t arg) {
    return ioctl((int)fd, (unsigned long)request, (void *)arg);
}

// The third fcntl argument is an int for some commands and a pointer for
// other commands. The raw pointer-sized value is correct for both.
int syscall_libc_fcntl(uintptr_t fd, uintptr_t cmd, uintptr_t arg) {
    return fcntl((int)fd, (int)cmd, (void *)arg);
}

// x/sys calls openat through syscall6 with two trailing zero arguments. The
// two extra register arguments are harmless to a four-parameter callee.
int syscall_libc_openat(uintptr_t dirfd, uintptr_t pathname, uintptr_t flags, uintptr_t mode) {
    return openat((int)dirfd, (const char *)pathname, (int)flags, (mode_t)mode);
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
