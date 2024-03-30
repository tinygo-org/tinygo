
MEMORY
{
  FLASH_TEXT (rx) : ORIGIN = 0x08000000, LENGTH = 1024K
  RAM (xrw)       : ORIGIN = 0x20000000, LENGTH = 96K
  RAM2 (xrw)      : ORIGIN = 0x10000000, LENGTH = 32K
}

_stack_size = 4K;

INCLUDE "targets/arm.ld"
