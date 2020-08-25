MEMORY
{
	FLASH_TEXT (rx) : ORIGIN = 0x00000000, LENGTH = 62K
	RAM (rwx)       : ORIGIN = 0x1FFFF800, LENGTH = 8K
}

_stack_size = 512;

INCLUDE "targets/kinetis.ld"