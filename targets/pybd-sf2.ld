MEMORY
{
    /* Note: the bootloader lives in the first two sectors (32K) of flash). */
    FLASH_TEXT (rw) : ORIGIN = 0x08008000, LENGTH = 480K /* Sector 2-7 */
    RAM (xrw)       : ORIGIN = 0x20000000, LENGTH = 256K /* DTCM + SRAM1 + SRAM2 */
}

_stack_size = 2K;

INCLUDE "targets/arm.ld"
