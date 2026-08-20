MEMORY
{
  FLASH_TEXT (rx) : ORIGIN = 0x8000000, LENGTH = 0x40000
  RAM (xrw)       : ORIGIN = 0x20000000, LENGTH = 0x8000
}

_stack_size = 4K;

INCLUDE "targets/arm.ld"
