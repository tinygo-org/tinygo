
MEMORY
{
    FLASH_TEXT (rw) : ORIGIN = 0,                      LENGTH = __flash_size - _bootloader_size
    RAM (xrw)       : ORIGIN = 0x800000 + __ram_start, LENGTH = __ram_size
}

ENTRY(__vector_RESET)

SECTIONS
{
    .text :
    {
        KEEP(*(.vectors))
        KEEP(*(.text.__vector_RESET))
        KEEP(*(.text.main)) /* main must follow the reset handler */
        *(.text)
        *(.text.*)
        *(.progmem)
        *(.progmem.*)
        . = ALIGN(16); /* needed with ld.lld for some reasoon */
    }

    .stack (NOLOAD) :
    {
        . += _stack_size;
        _stack_top = .;
    } >RAM

    _sidata = LOADADDR(.data);

    .data :
    {
        _sdata = .;        /* used by startup code */
        *(.rodata)
        *(.rodata.*)
        *(.data)
        *(.data*)
        _edata = .;        /* used by startup code */
    } >RAM AT>FLASH_TEXT

    .bss :
    {
        _sbss = .;         /* used by startup code */
        *(.bss)
        *(.bss*)
        *(COMMON)
        _ebss = .;         /* used by startup code */
    } >RAM
}

/* For the memory allocator. */
_heap_start = _ebss;
_heap_end = ORIGIN(RAM) + LENGTH(RAM);
_globals_start = _sdata;
_globals_end = _ebss;
