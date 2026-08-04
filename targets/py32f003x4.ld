
MEMORY
{
  FLASH_TEXT (rx) : ORIGIN = 0x08000000, LENGTH = 16K
  RAM (xrw)       : ORIGIN = 0x20000000, LENGTH = 2K
}

_stack_size = 512;

INCLUDE "targets/arm.ld"
