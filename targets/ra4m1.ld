
MEMORY
{
    FLASH_TEXT (rw) : ORIGIN = 0x00000000+0x4000, LENGTH = 0x00080000-0x4000  /* First 16KB used by bootloader */
    RAM (xrw)       : ORIGIN = 0x20000000, LENGTH = 0x00030000
}

_stack_size = 4K;

INCLUDE "targets/arm.ld"
