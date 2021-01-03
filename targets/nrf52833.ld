
MEMORY
{
    FLASH_TEXT (rw) : ORIGIN = 0x00000000, LENGTH = 0x80000
    RAM (xrw)       : ORIGIN = 0x20000000, LENGTH = 0x20000
}

_stack_size = 4K;

INCLUDE "targets/arm.ld"
