MEMORY
{
  FLASH_TEXT (rx) : ORIGIN = 0x8000000, LENGTH = 0x10000
  RAM (xrw)       : ORIGIN = 0x20000000, LENGTH = 0x2000
}

_stack_size = 2K;

INCLUDE "targets/arm.ld"
