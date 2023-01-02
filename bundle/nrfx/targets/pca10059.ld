
MEMORY
{
    FLASH_TEXT (rw) : ORIGIN = 0x00001000, LENGTH = 0xDF000
    RAM (xrw)       : ORIGIN = 0x20000008, LENGTH = 0x3FFF8
}

_stack_size = 2K;

INCLUDE "targets/arm.ld"
