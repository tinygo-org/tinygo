
MEMORY
{
  FLASH_TEXT (rx) : ORIGIN = 0x08000000, LENGTH = 32K
  RAM (xrw)       : ORIGIN = 0x20000000, LENGTH = 8K
}

# Allocate 'Persistent RAM' from regular RAM (some chips have specific PRAM)
REGION_ALIAS("PRAM", RAM);

_stack_size = 2K;

INCLUDE "targets/arm.ld"