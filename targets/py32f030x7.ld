
MEMORY
{
  FLASH_TEXT (rx) : ORIGIN = 0x08000000, LENGTH = 48K
  RAM (xrw)       : ORIGIN = 0x20000000, LENGTH = 6K
}

_stack_size = 512;

INCLUDE "targets/arm.ld"
