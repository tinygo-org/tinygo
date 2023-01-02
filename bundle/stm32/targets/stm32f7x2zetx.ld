
MEMORY
{
    FLASH_TEXT (rw) : ORIGIN = 0x08000000, LENGTH = 512K
    RAM (xrw)       : ORIGIN = 0x20000000, LENGTH = 256K
}

_stack_size = 4K;

INCLUDE "targets/arm.ld"
