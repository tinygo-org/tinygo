
MEMORY
{
  FLASH_TEXT (rx) : ORIGIN = 0x08000000, LENGTH = 512K
  RAM (xrw)       : ORIGIN = 0x20000000, LENGTH = 192K
  PRAM (rw)       : ORIGIN = 0x20030000, LENGTH = 64K
}

_stack_size = 4K;

INCLUDE "targets/arm.ld"