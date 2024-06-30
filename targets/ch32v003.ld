MEMORY
{
    FLASH_TEXT (rw) : ORIGIN = 0x00000000, LENGTH = 16K
    RAM (xrw)       : ORIGIN = 0x20000000, LENGTH = 2K
}

_stack_size = 1K;

INCLUDE "targets/riscv.ld"
