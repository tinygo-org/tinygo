MEMORY
{
  FLASH_TEXT (rx) : ORIGIN = 0x8000000, LENGTH = 0x80000
  RAM (xrw)       : ORIGIN = 0x20000000, LENGTH = 0x24000
}

_stack_size = 512;

INCLUDE "targets/arm.ld"
