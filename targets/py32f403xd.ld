MEMORY
{
  FLASH_TEXT (rx) : ORIGIN = 0x8000000, LENGTH = 0x60000
  RAM (xrw)       : ORIGIN = 0x20000000, LENGTH = 0x10000
}

_stack_size = 4K;

INCLUDE "targets/arm.ld"
