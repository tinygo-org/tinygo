MEMORY
{
  FLASH   (rx) : ORIGIN = 0x08100000, LENGTH = 1024K /* M4 code/VTOR/ISRs */
  SRAMD2 (rwx) : ORIGIN = 0x10000000, LENGTH =  288K /* M4 stack/heap/globals */
  SRAMD3 (rwx) : ORIGIN = 0x38000000, LENGTH =   64K /* M7+M4 shared */
}

ENTRY(Reset_Handler);

_stack_size = 4K;

SECTIONS
{
  /* Program code and read-only data goes to FLASH. */
  .text : {

    _svectors = ABSOLUTE(.);
    /* TODO: copy to SRAM? */
    KEEP(*(.isr_vector))
    . = ALIGN(32);
    _evectors = .;
    *(.text*)
    *(.rodata*)
    . = ALIGN(8);

  } > FLASH

  .tinygo_stacksizes : {

    *(.tinygo_stacksizes)

  } > FLASH

  .stack (NOLOAD) : {

    . = ALIGN(8);
    . += _stack_size;
    _stack_top = .;

  } > SRAMD2

  .data : ALIGN(8) {

    _sdata = .;
    *(.data*);
    . = ALIGN(8);
    _edata = .;

  } > SRAMD2 AT > FLASH

  .bss : ALIGN(8) {

    _sbss = .;
    *(.bss*);
    *(COMMON);
    . = ALIGN(8);
    _ebss = .;

  } > SRAMD2

  /DISCARD/ : {

    *(.ARM.exidx*) /* causes spurious linker errors */

  }

  _sidata = LOADADDR(.data); /* Used for initialization */

  /* Remaining items are for the TinyGo memory allocator. */

  _heap_start = _ebss;
  _heap_end = ORIGIN(SRAMD2) + LENGTH(SRAMD2);

  _globals_start = _sdata;
  _globals_end = _ebss;

}
