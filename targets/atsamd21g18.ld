
MEMORY
{
    FLASH_TEXT (rw) : ORIGIN = 0x00000000, LENGTH = 0x00040000
    RAM (xrw)       : ORIGIN = 0x20000000, LENGTH = 0x00008000
}

_stack_size = 2K;

INCLUDE "targets/arm.ld"
