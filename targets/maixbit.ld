
MEMORY
{
    RAM (xrw) : ORIGIN = 0x80000000, LENGTH = 6M
}

_stack_size = 2K;

SECTIONS
{
    .text :
    {
        . = ALIGN(16);
        KEEP(*(.init))
        . = ALIGN(16);
        *(.text.handleInterruptASM)
        *(.text)
        *(.text.*)
        *(.rodata)
        *(.rodata.*)
        . = ALIGN(16);
    } >RAM


    /* Start address (in flash) of .data, used by startup code. */
    _sidata = LOADADDR(.data);

    /* Globals with initial value */
    .data :
    {
        . = ALIGN(16);
        /* see https://gnu-mcu-eclipse.github.io/arch/riscv/programmer/#the-gp-global-pointer-register */
        PROVIDE( __global_pointer$ = . + (4K / 2) );
        _sdata = .;        /* used by startup code */
        *(.sdata)
        *(.data .data.*)
        . = ALIGN(16);
        _edata = .;        /* used by startup code */
    } >RAM

    /* Zero-initialized globals  */
    .bss :
    {
        . = ALIGN(16);
        _sbss = .;         /* used by startup code */
        *(.sbss)
        *(.bss .bss.*)
        *(COMMON)
        . = ALIGN(16);
        _ebss = .;         /* used by startup code */
    } >RAM

    /DISCARD/ :
    {
        *(.eh_frame)       /* causes 'no memory region specified' error in lld */
    }
}

PROVIDE(_stack_top = ORIGIN(RAM) + LENGTH(RAM));
_stack_bottom = _stack_top - _stack_size;

/* For the memory allocator. */
_heap_start = _ebss;
_heap_end = ORIGIN(RAM) + LENGTH(RAM) - _stack_size;
_globals_start = _sdata;
_globals_end = _ebss;
