
MEMORY
{
    FLASH_TEXT (rw) : ORIGIN = 0x00000000+0x0000, LENGTH = 0x00100000-0x0000
    RAM (xrw)       : ORIGIN = 0x20000000, LENGTH = 0x00040000
}

_stack_size = 4K;

INCLUDE "targets/arm.ld"
