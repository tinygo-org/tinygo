
MEMORY 
{
   FLASH_TEXT (rx) : ORIGIN = 0x00010000, LENGTH = 2M - 64K /* start from 0x0, allowing bootloader */
   RAM (rwx) : ORIGIN = 0x20000000, LENGTH = 256K     /* full-size SRAM, 256K */
}

_stack_size = 4K;

INCLUDE "targets/arm.ld"