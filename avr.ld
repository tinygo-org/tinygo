
MEMORY
{
    FLASH_TEXT (rw) : ORIGIN = 0, LENGTH = 32256
    RAM (xrw)       : ORIGIN = 0, LENGTH = 2K
}

_stack_size = 512;

SECTIONS
{
    .text :
    {
        KEEP(*(.isr))
        KEEP(*(.reset))
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

    .data :
    {
        *(.data)
        *(.data*)
    } >RAM AT>FLASH_TEXT

    .bss :
    {
        _sbss = .;
        *(.bss)
        *(.bss*)
        *(COMMON)
        _ebss = .;
        _heap_start = .;
    } >RAM
}
