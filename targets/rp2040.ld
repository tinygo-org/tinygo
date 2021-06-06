
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
