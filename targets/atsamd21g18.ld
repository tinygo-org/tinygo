
MEMORY
{
    FLASH_TEXT (rw) : ORIGIN = 0x00000000+0x2000, LENGTH = 0x00040000-0x2000  /* First 8KB used by bootloader */
    RAM (xrw)       : ORIGIN = 0x20000000, LENGTH = 0x00008000
}

_stack_size = 2K;

INCLUDE "targets/arm.ld"
