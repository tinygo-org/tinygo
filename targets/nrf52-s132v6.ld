
MEMORY
{
    FLASH_TEXT (rw) : ORIGIN = 0x00000000 + 0x00026000 , LENGTH = 256K - 0x00026000 /* .text */
    RAM (xrw)       : ORIGIN = 0x20000000 + 0x000039c0,  LENGTH = 64K  - 0x000039c0
}

_stack_size = 4K;

INCLUDE "targets/arm.ld"
