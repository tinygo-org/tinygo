
MEMORY
{
    RAM (rwx)       : ORIGIN = 0x40000000, LENGTH = 0x20000 /* 128 KB */
}

REGION_ALIAS("FLASH_TEXT", RAM);

_stack_size = 2K;

INCLUDE "targets/riscv.ld"
