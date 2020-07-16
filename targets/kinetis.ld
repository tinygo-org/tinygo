/* define output sections */
SECTIONS
{
    /* Program code and read-only data goes to FLASH_TEXT. */
    .text :
    {
        /* vector table MUST start at 0x0 */
        . = 0;
        KEEP(*(.isr_vector))

        /* flash configuration MUST be at 0x400 */
        . = 0x400;
        KEEP(*(.flash_config))

        /* everything else */
        *(.text)
        *(.text.*)
        *(.rodata)
        *(.rodata.*)
        . = ALIGN(4);

    } >FLASH_TEXT = 0xFF
}

INCLUDE "targets/arm.ld"