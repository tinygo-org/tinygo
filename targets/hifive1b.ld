
MEMORY
{
    FLASH_TEXT (rw) : ORIGIN = 0x20010000, LENGTH = 0x6a120
    RAM (xrw)       : ORIGIN = 0x80000000, LENGTH = 0x4000
}

# Allocate 'Persistent RAM' from regular RAM (some chips have specific PRAM)
REGION_ALIAS("PRAM", RAM);

_stack_size = 2K;

INCLUDE "targets/riscv.ld"
