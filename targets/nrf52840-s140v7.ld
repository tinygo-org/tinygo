
MEMORY
{
    FLASH_TEXT (rw) : ORIGIN = 0x00000000 + 0x00027000, LENGTH = 1M - 0x00027000
    RAM (xrw)       : ORIGIN = 0x20000000 + 0x000039c0, LENGTH = 256K - 0x000039c0
}

_stack_size = 4K;

/* This value is needed by the Nordic SoftDevice. */
__app_ram_base = ORIGIN(RAM);

INCLUDE "targets/arm.ld"
