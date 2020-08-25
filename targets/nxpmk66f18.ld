MEMORY
{
    FLASH_TEXT (rx) : ORIGIN = 0x00000000, LENGTH = 1024K
    RAM (rwx)       : ORIGIN = 0x1FFF0000, LENGTH = 256K
}

_stack_size = 2K;

INCLUDE "targets/kinetis.ld"
