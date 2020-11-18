
MEMORY 
{
    FLASH_TEXT (rx) : ORIGIN = 0x00000000, LENGTH = 0x00200000
    RAM (rwx)       : ORIGIN = 0x20000000, LENGTH = 0x00040000
}

_stack_size = 4K;

INCLUDE "targets/arm.ld"