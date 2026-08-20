MEMORY
{
  FLASH_TEXT (rx) : ORIGIN = 0x8000000, LENGTH = 0x6000
  RAM (xrw)       : ORIGIN = 0x20000000, LENGTH = 0xc00
}

_stack_size = 1K;

INCLUDE "targets/arm.ld"
