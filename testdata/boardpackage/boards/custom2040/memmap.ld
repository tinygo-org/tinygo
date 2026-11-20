MEMORY
{
    BOOT2_TEXT (rx) : ORIGIN = 0x10000000, LENGTH = 256
    FLASH_TEXT (rx) : ORIGIN = 0x10000000 + 256, LENGTH = 4M - 256
    RAM (rwx)       : ORIGIN = 0x20000000, LENGTH = 256k
}

_stack_size = 2K;

SECTIONS
{
    .boot2 : {
        __boot2_start__ = .;
        KEEP (*(.boot2));
        . = __boot2_start__ + 256 - 4;
        LONG(0)
    } > BOOT2_TEXT = 0x0
}

INCLUDE "targets/arm.ld"
