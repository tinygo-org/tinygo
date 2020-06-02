
MEMORY
{
    RAM (xrw) : ORIGIN = 0x80000000, LENGTH = (6 * 1024 * 1024)
}

REGION_ALIAS("FLASH_TEXT", RAM);

_stack_size = 2K;

INCLUDE "targets/riscv.ld"
