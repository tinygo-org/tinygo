
MEMORY
{
    FLASH_TEXT (rw) : ORIGIN = 0x00000000, LENGTH = 512K /* .text */
    RAM (xrw)       : ORIGIN = 0x20000000, LENGTH = 64K
}

_stack_size = 2K;

INCLUDE "targets/arm.ld"
