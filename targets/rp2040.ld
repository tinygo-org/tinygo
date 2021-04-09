
MEMORY
{
    RAM (rwx)       : ORIGIN = 0x20000000, LENGTH = 256k
}

_stack_size = 2K;

INCLUDE "targets/arm.ld"
