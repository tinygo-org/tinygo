
MEMORY
{
    FLASH_TEXT (rw) : ORIGIN = 0, LENGTH = __flash_size - _bootloader_size
    RAM (xrw)       : ORIGIN = 0, LENGTH = __ram_size
}

SECTIONS
{
    .text :
    {
        KEEP(*(.vectors))
        KEEP(*(.text.__vector_RESET))
        KEEP(*(.text.main)) /* main must follow the reset handler */
        *(.text.*)
        *(.rodata)
        *(.rodata.*)
    }

    .stack :
    {
        . += _stack_size;
        _stack_top = .;
    } >RAM

    _sidata = LOADADDR(.data);

    .data :
    {
        _sdata = .;        /* used by startup code */
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
