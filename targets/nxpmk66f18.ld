
/* Unused, but here to silence a linker warning. */
ENTRY(Reset_Handler)

/* define memory layout */
MEMORY
{
    FLASH_TEXT (rx) : ORIGIN = 0x00000000, LENGTH = 1024K
    RAM (rwx)       : ORIGIN = 0x1FFF0000, LENGTH = 256K
}

_stack_size = 2K;

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
