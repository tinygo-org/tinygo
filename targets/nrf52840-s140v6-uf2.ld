
MEMORY
{
    FLASH_TEXT (rw) : ORIGIN = 0x00000000+0x26000, LENGTH = 0xED000-0x26000 /* SoftDevice S140. See https://learn.adafruit.com/introducing-the-adafruit-nrf52840-feather/hathach-memory-map. Application starts at 0x26000; user data starts at 0xED000 */
    RAM (xrw)       : ORIGIN = 0x20004180, LENGTH = 0x20040000-0x20004180
}

_stack_size = 2K;

/* This value is needed by the Nordic SoftDevice. */
__app_ram_base = ORIGIN(RAM);

INCLUDE "targets/arm.ld"
