
MEMORY
{
    /* Reserve exactly 256 bytes at start of flash for second stage bootloader */
    BOOT2_TEXT (rx) : ORIGIN = 0x10000000, LENGTH = 256
    FLASH_TEXT (rx) : ORIGIN = 0x10000000 + 256, LENGTH = FLASH_SIZE - 256
    RAM (rwx)       : ORIGIN = 0x20000000, LENGTH = 256k
}

_stack_size = 2K;

/* Default checksum for second stage bootloader to Zero if not specified */
boot2_checksum = DEFINED(boot2_checksum) ? boot2_checksum : 0x0;

SECTIONS
{
    /* Second stage bootloader is prepended to the image. It must be 256 bytes
       and checksummed.  The gap to the checksum is zero-padded.
    */
    .boot2 : {
        __boot2_start__ = .;
        KEEP (*(.boot2));

        /* Explicitly allocate space for CRC32 checksum at end of second stage
           bootloader and put checksum value (must be Little-Endian)
        */
        . = __boot2_start__ + 256 - 4;
        LONG(boot2_checksum)
    } > BOOT2_TEXT = 0x0

    /* The second stage will always enter the image at the start of .text.
       The debugger will use the ELF entry point, which is the _entry_point
       symbol if present, otherwise defaults to start of .text.
       This can be used to transfer control back to the bootrom on debugger
       launches only, to perform proper flash setup.
    */
}


INCLUDE "targets/arm.ld"
