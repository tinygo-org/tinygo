
MEMORY
{
    FLASH_TEXT (rw) : ORIGIN = 0x00000000+0x26000, LENGTH = 796K /* SoftDevice S140. See https://learn.adafruit.com/introducing-the-adafruit-nrf52840-feather/hathach-memory-map */
    RAM (xrw)       : ORIGIN = 0x20004180, LENGTH = 37K
}

_stack_size = 2K;

INCLUDE "targets/arm.ld"
