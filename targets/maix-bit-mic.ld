
MEMORY
{
    FLASH_TEXT (rw) : ORIGIN = 0x80000000, LENGTH = (6 * 1024 * 1024)
    RAM (xrw)       : ORIGIN = 0x80000000, LENGTH = (6 * 1024 * 1024)
}

_stack_size = 2K;

INCLUDE "targets/riscv.ld"
