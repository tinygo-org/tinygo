#include <stdio.h>

// Defined in the runtime package. Writes to the default console (usually, the
// first UART or an USB-CDC device).
int runtime_putchar(char, FILE*);

// Define stdin, stdout, and stderr as a single object.
// This object must not reside in ROM.
static FILE __stdio = FDEV_SETUP_STREAM(runtime_putchar, NULL, NULL, _FDEV_SETUP_WRITE);

// Define the underlying structs for stdin, stdout, and stderr.
FILE *const __iob[3] = { &__stdio, &__stdio, &__stdio };
