
MEMORY
{
    RAM (xrw) : ORIGIN = 0xffffffff80000000, LENGTH = 6M
}

REGION_ALIAS("FLASH_TEXT", RAM);

_stack_size = 2K;

INCLUDE "targets/riscv.ld"
