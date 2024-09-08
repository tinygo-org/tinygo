/* See Rust for a more complete reference: https://github.com/rp-rs/rp-hal/blob/main/rp235x-hal-examples/memory.x */
MEMORY
{
   /* 2MiB safe default.  */
    FLASH : ORIGIN = 0x10000000, LENGTH = 2048k
    /* RAM consists of 8 banks, SRAM0..SRAM7 with striped mapping. */
    SRAM : ORIGIN = 0x20000000, LENGTH = 512k
    /* Banks 8 and 9 use direct mapping which can be
       specailized for applications where predictable access time is beneficial.
       i.e: Separate stacks for core0 and core1. */
    SRAM4 : ORIGIN = 0x20080000, LENGTH = 4k
    SRAM5 : ORIGIN = 0x20081000, LENGTH = 4k
    /* Reserve exactly 256 bytes at start of flash for second stage bootloader */
    BOOT2_TEXT (rx) : ORIGIN = 0x10000000, LENGTH = 256
    FLASH_TEXT (rx) : ORIGIN = 0x10000000 + 256, LENGTH = 2048k - 256
    RAM (rwx)       : ORIGIN = 0x20000000, LENGTH = 512k
}

_stack_size = 2K;

SECTIONS
{
    /* Second stage bootloader is prepended to the image. It must be 256 bytes
       and checksummed.  The gap to the checksum is zero-padded.
    */
    .boot2 : {
        __boot2_start__ = .;
        KEEP (*(.boot2));

        /* Explicitly allocate space for CRC32 checksum at end of second stage
           bootloader
        */
        . = __boot2_start__ + 256 - 4;
        LONG(0)
    } > BOOT2_TEXT = 0x0

    /* The second stage will always enter the image at the start of .text.
       The debugger will use the ELF entry point, which is the _entry_point
       symbol if present, otherwise defaults to start of .text.
       This can be used to transfer control back to the bootrom on debugger
       launches only, to perform proper flash setup.
    */
}

INCLUDE "targets/arm.ld"
