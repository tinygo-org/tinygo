
/* SRAM0 is mirrored at 0x20070000 so that it joins SRAM1 at 0x20080000. */
MEMORY
{
    FLASH_TEXT (rx) : ORIGIN = 0x00080000, LENGTH = 512K
    RAM (rwx)       : ORIGIN = 0x20070000, LENGTH = 96K
}

_stack_size = 4K;

INCLUDE "targets/arm.ld"
