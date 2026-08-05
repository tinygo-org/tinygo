MEMORY
{
  FLASH_TEXT (rx) : ORIGIN = 0x8000000, LENGTH = 0x4000
  RAM (xrw)       : ORIGIN = 0x20000000, LENGTH = 0x800
}

_stack_size = 512;

INCLUDE "targets/arm.ld"
