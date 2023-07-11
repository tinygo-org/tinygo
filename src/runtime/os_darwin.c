// Wrapper function because 'open' is a variadic function and variadic functions
// use a different (incompatible) calling convention on darwin/arm64.

#include <fcntl.h>
int syscall_libc_open(const char *pathname, int flags, mode_t mode) {
    return open(pathname, flags, mode);
}
