ENTRY(_start)

SECTIONS
{
    /* Begin putting sections at 1 MiB, a conventional place for kernels to be
       loaded at by the bootloader. */
    . = 1M;

    /* First put the multiboot header, as it is required to be put very early
       early in the image or the bootloader won't recognize the file format.
       Next we'll put the .text section. */
    .text : ALIGN(4K)
    {
        *(.multiboot)
        *(.text .text.*)
    }

    /* Read-only data. */
    .rodata : ALIGN(4K)
    {
        *(.rodata .rodata.*)
    }

    /* Read-write data (initialized). */
    .data : ALIGN(4K)
    {
        _sdata = .;
        *(.data .data.*)
        _edata = .;
    }

    /* Read-write data (uninitialized) and stack. */
    .bss : ALIGN(4K)
    {
        _sbss = .;
        *(COMMON)
        *(.bss .bss.*)
        _ebss = .;
    }

    .stack : ALIGN(4K)
    {
        /* Define the stack here in the .bss section. */
        . = . + 16K;
        _stack_top = .;
    }

    /DISCARD/ :
    {
        *(.eh_frame)
    }
}

/* For the memory allocator. */
_heap_start = _ebss;
_heap_end = _ebss + 1M;
_globals_start = _sdata;
_globals_end = _ebss;
