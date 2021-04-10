
MEMORY
{
  FLASH_TEXT (rx) : ORIGIN = 0x08000000, LENGTH = 256K
  RAM (xrw)       : ORIGIN = 0x20000000, LENGTH = 64K
  PRAM (rw)       : ORIGIN = 0x10000000, LENGTH = 16k
}

_stack_size = 4K;

INCLUDE "targets/arm.ld"