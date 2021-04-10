
MEMORY
{
    FLASH_TEXT (rw) : ORIGIN = 0x00000000, LENGTH = 1M
    RAM (xrw)       : ORIGIN = 0x20000000, LENGTH = 256K
}

# Allocate 'Persistent RAM' from regular RAM (some chips have specific PRAM)
REGION_ALIAS("PRAM", RAM);

_stack_size = 4K;

INCLUDE "targets/arm.ld"
