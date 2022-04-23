
MEMORY
{
    /* Reserve exactly 256 bytes at start of flash for second stage bootloader */
    BOOT2_TEXT (rx) : ORIGIN = 0x10000000, LENGTH = 256
    FLASH_TEXT (rx) : ORIGIN = 0x10000000 + 256, LENGTH = 16384K - 256
    RAM (rwx)       : ORIGIN = 0x20000000, LENGTH = 256k
}

INCLUDE "targets/rp2040.ld"