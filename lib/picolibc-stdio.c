// This file is included in the picolibc build.
// It makes stdio functions available to the C library.

#include <stdio.h>
#include <sys/cdefs.h>

// Defined in the runtime package. Writes to the default console (usually, the
// first UART or an USB-CDC device).
int runtime_putchar(char, FILE*);

// Define stdin, stdout, and stderr as a single object.
// This object must not reside in ROM.
static FILE __stdio = FDEV_SETUP_STREAM(runtime_putchar, NULL, NULL, _FDEV_SETUP_WRITE);

// Define the underlying structs for stdin, stdout, and stderr.
FILE *const stdin = &__stdio;
__strong_reference(stdin, stdout);
__strong_reference(stdin, stderr);
