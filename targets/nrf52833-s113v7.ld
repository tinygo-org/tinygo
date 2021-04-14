
MEMORY
{
    FLASH_TEXT (rw) : ORIGIN = 0x00000000 + 0x1C000, LENGTH = 0x80000 - 0x1C000
    RAM (xrw)       : ORIGIN = 0x20000000 + 0x1e20,  LENGTH = 0x20000 - 0x1e20
}

_stack_size = 4K + __softdevice_stack;

/* These values are needed for the Nordic SoftDevice. */
__app_ram_base = ORIGIN(RAM);
__softdevice_stack = DEFINED(__softdevice_stack) ? __softdevice_stack : 0;

INCLUDE "targets/arm.ld"
