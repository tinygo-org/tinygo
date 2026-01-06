
MEMORY
{
    FLASH_TEXT (rw) : ORIGIN = 0x08000000, LENGTH = 128K
    RAM (xrw)       : ORIGIN = 0x20000000, LENGTH = 144K
}

_stack_size = 4K;

INCLUDE "targets/arm.ld"