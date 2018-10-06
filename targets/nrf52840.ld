
MEMORY
{
    FLASH_TEXT (rw) : ORIGIN = 0x00000000, LENGTH = 1M
    RAM (xrw)       : ORIGIN = 0x20000000, LENGTH = 256K
}

_stack_size = 4K;

INCLUDE "targets/arm.ld"
