/* Linker script for AVRs with a unified flash and RAM address space. This
 * includes the ATtiny10 and the ATtiny1616.
 */

MEMORY
{
    FLASH_TEXT (x) : ORIGIN = 0,                    LENGTH = __flash_size
    FLASH_DATA (r) : ORIGIN = __mapped_flash_start, LENGTH = __flash_size
    RAM (xrw)      : ORIGIN = __ram_start,          LENGTH = __ram_size
}

ENTRY(main)

SECTIONS
{
    .text :
    {
        KEEP(*(.vectors))
        *(.text.__vector_RESET)
        KEEP(*(.text.__do_copy_data)) /* TODO: only use when __do_copy_data is requested */
        KEEP(*(.text.__do_clear_bss))
        KEEP(*(.text.main)) /* main must follow the reset handler */
        *(.text)
        *(.text.*)
    } > FLASH_TEXT

    /* Read-only data is stored in flash, but is read from an offset (0x4000 or
     * 0x8000 depending on the chip). This requires some weird math to get it in
     * the right place.
     */
    .rodata ORIGIN(FLASH_DATA) + ADDR(.text) + SIZEOF(.text):
    {
        *(.rodata)
        *(.rodata.*)
    } AT>FLASH_TEXT

    /* The address to which the data section should be copied by the startup
     * code.
     */
    __data_load_start = ORIGIN(FLASH_DATA) + LOADADDR(.data);

    .data :
    {
        __data_start = .;  /* used by startup code */
        *(.data)
        *(.data*)
        __data_end = .;    /* used by startup code */
    } >RAM AT>FLASH_TEXT

    .bss :
    {
        __bss_start = .;   /* used by startup code */
        *(.bss)
        *(.bss*)
        *(COMMON)
        __bss_end = .;     /* used by startup code */
    } >RAM
}
